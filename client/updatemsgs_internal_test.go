// Copyright 2019 - See NOTICE file for copyright holders.
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

package client

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	peruniotest "perun.network/go-perun/wire/perunio/test"
	pkgtest "polycry.pt/poly-go/test"
)

func newRandomMsgChannelUpdate(rng *rand.Rand) *MsgChannelUpdate {
	state := test.NewRandomState(rng)
	sig := newRandomSig(rng)
	return &MsgChannelUpdate{
		ChannelUpdate: ChannelUpdate{
			State:    state,
			ActorIdx: channel.Index(rng.Intn(state.NumParts())),
		},
		Sig: sig,
	}
}
func TestSerialization_VirtualChannelSettlementProposal(t *testing.T) {
	rng := pkgtest.Prng(t)
	for i := 0; i < 4; i++ {
		msgUp := newRandomMsgChannelUpdate(rng)
		params, state := test.NewRandomParamsAndState(rng)
		m := &virtualChannelSettlementProposal{
			MsgChannelUpdate: *msgUp,
			Final: channel.SignedState{
				Params: params,
				State:  state,
				Sigs:   newRandomSigs(rng, state.NumParts()),
			},
		}
		peruniotest.MsgSerializerTest(t, m)
	}
}

// newRandomSig generates a random account and then returns the signature on
// some random data.
func newRandomSig(rng *rand.Rand) wallet.Sig {
	acc := wallettest.NewRandomAccount(rng)
	data := make([]byte, 8)
	rng.Read(data)
	sig, err := acc.SignData(data)
	if err != nil {
		panic("signing error")
	}
	return sig
}

// newRandomSigs generates a list of random signatures.
func newRandomSigs(rng *rand.Rand, n int) (a []wallet.Sig) {
	a = make([]wallet.Sig, n)
	for i := range a {
		a[i] = newRandomSig(rng)
	}
	return
}

// newRandomstring returns a random string of length between minLen and
// minLen+maxLenDiff.
func newRandomString(rng *rand.Rand, minLen, maxLenDiff int) string {
	r := make([]byte, minLen+rng.Intn(maxLenDiff))
	rng.Read(r)
	return string(r)
}
