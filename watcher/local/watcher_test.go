// Copyright 2021 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ethChannel "perun.network/go-perun/backend/ethereum/channel"
	_ "perun.network/go-perun/backend/ethereum/channel/test" // For initializing channeltest
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/watcher"
	"perun.network/go-perun/watcher/internal/mocks"
	"perun.network/go-perun/watcher/local"
)

func Test_StartWatching(t *testing.T) {
	rng := test.Prng(t)
	rs := &mocks.RegisterSubscriber{}
	rs.On("Subscribe", mock.Anything, mock.Anything).Return(&ethChannel.RegisteredSub{}, nil)

	t.Run("happy/ledger_channel", func(t *testing.T) {
		w := newWatcher(t, rs)
		params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		signedState := makeSignedStateWDummySigs(params, state)

		statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), signedState)

		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		_, _, err = w.StartWatchingLedgerChannel(context.TODO(), signedState)
		require.Error(t, err, "StartWatching twice for the same channel must error")
	})

	t.Run("happy/sub_channel", func(t *testing.T) {
		w := newWatcher(t, rs)

		// Start watching for ledger channel.
		parentParams, parentState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentState)
		statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), parentSignedState)
		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		// Start watching for a sub-channel with unknown parent ch id.
		childParams, childState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		childSignedState := makeSignedStateWDummySigs(childParams, childState)
		randomID := channeltest.NewRandomChannelID(rng)
		_, _, err = w.StartWatchingSubChannel(context.TODO(), randomID, childSignedState)
		require.Error(t, err, "Start watching for a sub-channel with unknown parent ch id must error")

		// Start watching for a sub-channel with known parent ch id.
		_, _, err = w.StartWatchingSubChannel(context.TODO(), parentState.ID, childSignedState)
		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		// Repeat Start watching for the sub-channel.
		_, _, err = w.StartWatchingSubChannel(context.TODO(), parentState.ID, childSignedState)
		require.Error(t, err, "StartWatching twice for the same channel must error")
	})
}

func Test_Watcher_Working(t *testing.T) {
	rng := test.Prng(t)

	t.Run("ledger_channel_without_sub_channel", func(t *testing.T) {
		// Send a registered event on the adjudicator subscription, with the latest state.
		// Watcher should relay the event and not refute.
		t.Run("happy/latest_state_registered", func(t *testing.T) {
			// Setup
			params, txs := randomTxsForSingleCh(rng, 3)
			adjSub, trigger := setupAdjudicatorSub(makeRegisteredEvents(txs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSub, nil)
			w := newWatcher(t, rs)

			// Publish both the states to the watcher.
			statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
			require.NoError(t, statesPub.Publish(txs[1]))
			require.NoError(t, statesPub.Publish(txs[2]))

			// Trigger events and assert.
			triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
			rs.AssertExpectations(t)
		})

		// Send a registered event on the adjudicator subscription,
		// with a state newer than the latest state (published to the watcher),
		// Watch should relay the event and not refute.
		t.Run("happy/newer_than_latest_state_registered", func(t *testing.T) {
			// Setup
			params, txs := randomTxsForSingleCh(rng, 3)
			adjSub, trigger := setupAdjudicatorSub(makeRegisteredEvents(txs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSub, nil)
			w := newWatcher(t, rs)

			// Publish only one of the two newly created off-chain states to the watcher.
			statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
			require.NoError(t, statesPub.Publish(txs[1]))

			// Trigger adjudicator events with a state newer than the latest state (published to the watcher).
			triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
			rs.AssertExpectations(t)
		})

		// First, send a registered event on the adjudicator subscription,
		// with a state older than the latest state (published to the watcher),
		// Watch should relay the event and refute by registering the latest state.
		//
		// Next, send a registered event on adjudicator subscription,
		// with the state that was registered.
		// This time, watcher should relay the event and not dispute.
		t.Run("error/older_state_registered", func(t *testing.T) {
			// Setup
			params, txs := randomTxsForSingleCh(rng, 3)
			adjSub, trigger := setupAdjudicatorSub(makeRegisteredEvents(txs[1], txs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSub, nil)
			setupRegistered(t, rs, &channelTree{txs[2], []channel.Transaction{}})
			w := newWatcher(t, rs)

			// Publish both the states to the watcher.
			statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
			require.NoError(t, statesPub.Publish(txs[1]))
			require.NoError(t, statesPub.Publish(txs[2]))

			// Trigger adjudicator events with an older state and assert if Register was called once.
			triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
			time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
			rs.AssertNumberOfCalls(t, "Register", 1)

			// Trigger adjudicator events with the registered state and assert.
			triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
			rs.AssertExpectations(t)
		})
	})

	t.Run("ledger_channel_with_sub_channel", func(t *testing.T) {
		// For both, the parent and the sub-channel,
		// send a registered event on the adjudicator subscription, with the latest state.
		// Watcher should relay the events and not refute.
		t.Run("happy/latest_state_registered", func(t *testing.T) {
			// Setup
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

			adjSubParent, triggerParent := setupAdjudicatorSub(makeRegisteredEvents(parentTxs[2])...)
			adjSubChild, triggerChild := setupAdjudicatorSub(makeRegisteredEvents(childTxs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubParent, nil).Once()
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubChild, nil).Once()

			w := newWatcher(t, rs)

			// Parent: Publish both the states to the watcher.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesPubParent, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesPubParent.Publish(parentTxs[1]))
			require.NoError(t, statesPubParent.Publish(parentTxs[2]))

			// Child: Publish both the states to the watcher.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			statesPubChild, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, statesPubChild.Publish(childTxs[1]))
			require.NoError(t, statesPubChild.Publish(childTxs[2]))

			// Parent and child: Trigger adjudicator events with the latest states and assert.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			rs.AssertExpectations(t)
		})

		// For both, the parent and the sub-channel,
		// send a registered event on the adjudicator subscription,
		// with a state newer than the latest state (published to the watcher),
		// Watch should relay the event and not refute.
		t.Run("happy/newer_than_latest_state_registered", func(t *testing.T) {
			// Setup
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

			adjSubParent, triggerParent := setupAdjudicatorSub(makeRegisteredEvents(parentTxs[2])...)
			adjSubChild, triggerChild := setupAdjudicatorSub(makeRegisteredEvents(childTxs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubParent, nil).Once()
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubChild, nil).Once()

			w := newWatcher(t, rs)

			// Parent: Publish only one of the two newly created off-chain states to the watcher.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesPubParent, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesPubParent.Publish(parentTxs[1]))

			// Child: Publish only one of the two newly created off-chain states to the watcher.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			statesPubChild, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, statesPubChild.Publish(childTxs[1]))

			// Parent, Child: Trigger adjudicator events with a state newer than
			// the latest state (published to the watcher).
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			rs.AssertExpectations(t)
		})

		// For both, the parent and the sub-channel,
		//
		// First, send a registered event on the adjudicator subscription,
		// with a state older than the latest state (published to the watcher),
		// Watch should relay the event and refute by registering the latest state.
		//
		// Next, send a registered event on adjudicator subscription,
		// with the state that was registered.
		// This time, watcher should relay the event and not dispute.
		t.Run("happy/older_state_registered", func(t *testing.T) {
			// Setup
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

			adjSubParent, triggerParent := setupAdjudicatorSub(makeRegisteredEvents(parentTxs[1], parentTxs[2])...)
			adjSubChild, triggerChild := setupAdjudicatorSub(makeRegisteredEvents(childTxs[1], childTxs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubParent, nil).Once()
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubChild, nil).Once()
			setupRegistered(t, rs, &channelTree{parentTxs[2], []channel.Transaction{childTxs[2]}})

			w := newWatcher(t, rs)

			// Parent: Publish both the states to the watcher.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			parentStatesPub, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, parentStatesPub.Publish(parentTxs[1]))
			require.NoError(t, parentStatesPub.Publish(parentTxs[2]))

			// Child: Publish both the states to the watcher.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			childStatesPub, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, childStatesPub.Publish(childTxs[1]))
			require.NoError(t, childStatesPub.Publish(childTxs[2]))

			// Parent and Child: Trigger adjudicator events with an older state and assert if Register was called once.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
			rs.AssertNumberOfCalls(t, "Register", 1)

			// Parent and Child: Trigger adjudicator events with the registered state and assert.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			rs.AssertExpectations(t)
		})

		// First, for both, the parent and the sub-channel,
		// Send a registered event on the adjudicator subscription,
		// with a state older than the latest state (published to the watcher),
		// Watch should relay the event and refute by registering the latest state.
		//
		// Next, for the sub-channel, publish another off-chain state to the watcher.
		//
		// Next, for both, the parent and the sub-channel,
		// Send a registered event on adjudicator subscription,
		// with the state that was registered.
		// This time, because the registered state is older than the latest state for the sub-channel,
		// Watch should relay the event and refute by registering the latest state.
		//
		// Next, for both, the parent and the sub-channel,
		// send a registered event on adjudicator subscription,
		// with the state that was registered.
		// This time, watcher should relay the event and not dispute.
		t.Run("happy/older_state_registered_then_newer_state_received", func(t *testing.T) {
			// Setup
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 4)
			// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

			adjSubParent, triggerParent := setupAdjudicatorSub(makeRegisteredEvents(parentTxs[1], parentTxs[2], parentTxs[2])...)
			adjSubChild, triggerChild := setupAdjudicatorSub(makeRegisteredEvents(childTxs[1], childTxs[2], childTxs[3])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubParent, nil).Once()
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubChild, nil).Once()
			setupRegistered(t, rs,
				&channelTree{parentTxs[2], []channel.Transaction{childTxs[2]}},
				&channelTree{parentTxs[2], []channel.Transaction{childTxs[3]}})

			w := newWatcher(t, rs)

			// Parent: Publish both the states to the watcher.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			parentStatesPub, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, parentStatesPub.Publish(parentTxs[1]))
			require.NoError(t, parentStatesPub.Publish(parentTxs[2]))

			// Child: Publish both the states to the watcher.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			childStatesPub, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, childStatesPub.Publish(childTxs[1]))
			require.NoError(t, childStatesPub.Publish(childTxs[2]))

			// Parent and Child: Trigger adjudicator events with an older state and assert if Register was called once.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
			rs.AssertNumberOfCalls(t, "Register", 1)

			// Child: After register was called, publish a new state to the watcher.
			require.NoError(t, childStatesPub.Publish(childTxs[3]))
			// Parent and Child: Trigger adjudicator events with the registered state and assert if Register was called once.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
			rs.AssertNumberOfCalls(t, "Register", 2)

			// Parent and Child: Trigger adjudicator events with the newly registered state and assert.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
			rs.AssertExpectations(t)
		})
	})
}

func newWatcher(t *testing.T, rs channel.RegisterSubscriber) *local.Watcher {
	t.Helper()

	w, err := local.NewWatcher(local.Config{RegisterSubscriber: rs})
	require.NoError(t, err)
	require.NotNil(t, w)
	return w
}

func makeSignedStateWDummySigs(params *channel.Params, state *channel.State) channel.SignedState {
	return channel.SignedState{Params: params, State: state}
}

// randomTxsForSingleCh returns "n" transactions for a random channel.
func randomTxsForSingleCh(rng *rand.Rand, n int) (*channel.Params, []channel.Transaction) {
	params, initialState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0),
		channeltest.WithNumParts(2), channeltest.WithNumAssets(1), channeltest.WithIsFinal(false), channeltest.WithNumLocked(0))

	txs := make([]channel.Transaction, n)
	for i := range txs {
		txs[i] = channel.Transaction{State: initialState.Clone()}
		txs[i].State.Version = uint64(i)
	}
	return params, txs
}

// adjEventSource is used for triggering events on the mock adjudicator subscription.
//
// Adjudicator subscription blocks until it is closed.
type adjEventSource struct {
	handle    chan chan time.Time
	adjEvents chan channel.AdjudicatorEvent
}

func (t *adjEventSource) trigger() channel.AdjudicatorEvent {
	select {
	case handles := <-t.handle:
		close(handles)
		return <-t.adjEvents
	default:
		panic(fmt.Sprintf("Number of triggers exceeded maximum limit of %d", cap(t.handle)))
	}
}

// setupAdjudicatorSub returns an adjudicator subscription and a trigger.
//
// On each trigger, an event is sent on the mock adjudicator subscription and
// the corresponding transaction is returned.
//
// If trigger is triggered more times than the number of transactions, it panics.
//
// After all triggers are used, the subscription blocks.
func setupAdjudicatorSub(adjEvents ...channel.AdjudicatorEvent) (*mocks.AdjudicatorSubscription, adjEventSource) {
	adjSub := &mocks.AdjudicatorSubscription{}
	triggers := adjEventSource{
		handle:    make(chan chan time.Time, len(adjEvents)),
		adjEvents: make(chan channel.AdjudicatorEvent, len(adjEvents)),
	}

	for i := range adjEvents {
		handle := make(chan time.Time)
		triggers.handle <- handle
		triggers.adjEvents <- adjEvents[i]

		adjSub.On("Next").Return(adjEvents[i]).WaitUntil(handle).Once()
	}
	// Set up a transaction, that cannot be triggered.
	handle := make(chan time.Time)
	adjSub.On("Next").Return(channel.RegisteredEvent{}).WaitUntil(handle).Once()

	return adjSub, triggers
}

type channelTree struct {
	rootTx channel.Transaction
	subTxs []channel.Transaction
}

func setupRegistered(t *testing.T, rs *mocks.RegisterSubscriber, channelTrees ...*channelTree) {
	limit := len(channelTrees)
	mtx := sync.Mutex{}
	iChannelTree := 0

	rs.On("Register", mock.Anything,
		mock.MatchedBy(func(req channel.AdjudicatorReq) bool {
			mtx.Lock()
			if iChannelTree >= limit {
				return false
			}
			return assertEqualAdjudicatorReq(t, req, channelTrees[iChannelTree].rootTx.State)
		}),
		mock.MatchedBy(func(subStates []channel.SignedState) bool {
			defer func() {
				iChannelTree += 1
				mtx.Unlock()
			}()
			if iChannelTree >= limit {
				return false
			}
			return assertEqualSignedStates(t, subStates, channelTrees[iChannelTree].subTxs)
		})).Return(nil).Times(limit)
}

func assertEqualAdjudicatorReq(t *testing.T, got channel.AdjudicatorReq, want *channel.State) bool {
	if nil != got.Tx.State.Equal(want) {
		t.Logf("Got %+v, expected %+v", got.Tx.State, want)
		return false
	}
	return true
}

func assertEqualSignedStates(t *testing.T, got []channel.SignedState, want []channel.Transaction) bool {
	if len(got) != len(want) {
		t.Logf("Got %d sub states, expected %d sub states", len(got), len(want))
		return false
	}
	for iSubState := range got {
		t.Logf("Got %+v, expected %+v", got[iSubState].State, want[iSubState].State)
		if nil != got[iSubState].State.Equal(want[iSubState].State) {
			return false
		}
	}
	return true
}

func makeRegisteredEvents(txs ...channel.Transaction) []channel.AdjudicatorEvent {
	events := make([]channel.AdjudicatorEvent, len(txs))
	for i, tx := range txs {
		events[i] = &channel.RegisteredEvent{
			State: tx.State,
			Sigs:  tx.Sigs,
			AdjudicatorEventBase: channel.AdjudicatorEventBase{
				IDV:      tx.State.ID,
				TimeoutV: &channel.ElapsedTimeout{},
				VersionV: tx.State.Version,
			},
		}
	}
	return events
}

func startWatchingForLedgerChannel(
	t *testing.T,
	w *local.Watcher,
	signedState channel.SignedState,
) (watcher.StatesPub, watcher.AdjudicatorSub) {
	statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), signedState)

	require.NoError(t, err)
	require.NotNil(t, statesPub)
	require.NotNil(t, eventsSub)

	return statesPub, eventsSub
}

func startWatchingForSubChannel(
	t *testing.T,
	w *local.Watcher,
	signedState channel.SignedState,
	parentID channel.ID,
) (watcher.StatesPub, watcher.AdjudicatorSub) {
	statesPub, eventsSub, err := w.StartWatchingSubChannel(context.TODO(), parentID, signedState)

	require.NoError(t, err)
	require.NotNil(t, eventsSub)

	return statesPub, eventsSub
}

func triggerAdjEventAndExpectNotification(
	t *testing.T,
	trigger adjEventSource,
	eventsForClient watcher.AdjudicatorSub,
) {
	wantEvent := trigger.trigger()
	t.Logf("waiting for adjudicator event for ch %x, version: %v", wantEvent.ID(), wantEvent.Version())
	gotEvent := <-eventsForClient.EventStream()
	require.EqualValues(t, gotEvent, wantEvent)
	t.Logf("received adjudicator event for ch %x, version: %v", wantEvent.ID(), wantEvent.Version())
}
