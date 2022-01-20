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
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
)

func init() {
	wire.RegisterDecoder(wire.ChannelUpdate,
		func(r io.Reader) (wire.Msg, error) {
			var m MsgChannelUpdate
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelUpdateAcc,
		func(r io.Reader) (wire.Msg, error) {
			var m MsgChannelUpdateAcc
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelUpdateRej,
		func(r io.Reader) (wire.Msg, error) {
			var m MsgChannelUpdateRej
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.VirtualChannelFundingProposal,
		func(r io.Reader) (wire.Msg, error) {
			var m VirtualChannelFundingProposal
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.VirtualChannelSettlementProposal,
		func(r io.Reader) (wire.Msg, error) {
			var m virtualChannelSettlementProposal
			return &m, m.Decode(r)
		})
}

type (
	// ChannelMsg are all messages that can be routed to a particular channel
	// controller.
	ChannelMsg interface {
		wire.Msg
		ID() channel.ID
	}

	channelUpdateResMsg interface {
		ChannelMsg
		Ver() uint64
	}

	// MsgChannelUpdate is the wire message of a channel update proposal. It
	// additionally holds the signature on the proposed state.
	MsgChannelUpdate struct {
		ChannelUpdate
		// Sig is the signature on the proposed state by the peer sending the
		// ChannelUpdate.
		Sig wallet.Sig
	}

	// ChannelUpdateProposal represents an abstract update proposal message.
	ChannelUpdateProposal interface {
		wire.Msg
		perunio.Decoder
		Base() *MsgChannelUpdate
	}

	// MsgChannelUpdateAcc is the wire message sent as a positive reply to a
	// ChannelUpdate.  It references the channel ID and version and contains the
	// signature on the accepted new state by the sender.
	MsgChannelUpdateAcc struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Sig is the signature on the proposed new state by the sender.
		Sig wallet.Sig
	}

	// MsgChannelUpdateRej is the wire message sent as a negative reply to a
	// ChannelUpdate.  It references the channel ID and version and states a
	// reason for the rejection.
	MsgChannelUpdateRej struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Reason states why the sender rejectes the proposed new state.
		Reason string
	}
)

var (
	_ ChannelMsg          = (*MsgChannelUpdate)(nil)
	_ channelUpdateResMsg = (*MsgChannelUpdateAcc)(nil)
	_ channelUpdateResMsg = (*MsgChannelUpdateRej)(nil)
)

// Type returns this message's type: ChannelUpdate.
func (*MsgChannelUpdate) Type() wire.Type {
	return wire.ChannelUpdate
}

// Type returns this message's type: ChannelUpdateAcc.
func (*MsgChannelUpdateAcc) Type() wire.Type {
	return wire.ChannelUpdateAcc
}

// Type returns this message's type: ChannelUpdateRej.
func (*MsgChannelUpdateRej) Type() wire.Type {
	return wire.ChannelUpdateRej
}

// Base returns the core channel update message.
func (c *MsgChannelUpdate) Base() *MsgChannelUpdate {
	return c
}

func (c MsgChannelUpdate) Encode(w io.Writer) error {
	return perunio.Encode(w, c.State, c.ActorIdx, c.Sig)
}

func (c *MsgChannelUpdate) Decode(r io.Reader) (err error) {
	if c.State == nil {
		c.State = new(channel.State)
	}
	if err := perunio.Decode(r, c.State, &c.ActorIdx); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c MsgChannelUpdateAcc) Encode(w io.Writer) error {
	return perunio.Encode(w, c.ChannelID, c.Version, c.Sig)
}

func (c *MsgChannelUpdateAcc) Decode(r io.Reader) (err error) {
	if err := perunio.Decode(r, &c.ChannelID, &c.Version); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c MsgChannelUpdateRej) Encode(w io.Writer) error {
	return perunio.Encode(w, c.ChannelID, c.Version, c.Reason)
}

func (c *MsgChannelUpdateRej) Decode(r io.Reader) (err error) {
	return perunio.Decode(r, &c.ChannelID, &c.Version, &c.Reason)
}

// ID returns the id of the channel this update refers to.
func (c *MsgChannelUpdate) ID() channel.ID {
	return c.State.ID
}

// ID returns the id of the channel this update acceptance refers to.
func (c *MsgChannelUpdateAcc) ID() channel.ID {
	return c.ChannelID
}

// ID returns the id of the channel this update rejection refers to.
func (c *MsgChannelUpdateRej) ID() channel.ID {
	return c.ChannelID
}

// Ver returns the version of the state this update acceptance refers to.
func (c *MsgChannelUpdateAcc) Ver() uint64 {
	return c.Version
}

// Ver returns the version of the state this update rejection refers to.
func (c *MsgChannelUpdateRej) Ver() uint64 {
	return c.Version
}

/*
Virtual channel
*/

type (
	// VirtualChannelFundingProposal is a channel update that proposes the funding of a virtual channel.
	VirtualChannelFundingProposal struct {
		MsgChannelUpdate
		Initial  channel.SignedState
		IndexMap []channel.Index
	}

	// virtualChannelSettlementProposal is a channel update that proposes the settlement of a virtual channel.
	virtualChannelSettlementProposal struct {
		MsgChannelUpdate
		Final channel.SignedState
	}
)

// Type returns the message type.
func (*VirtualChannelFundingProposal) Type() wire.Type {
	return wire.VirtualChannelFundingProposal
}

func (m VirtualChannelFundingProposal) Encode(w io.Writer) (err error) {
	err = perunio.Encode(w,
		m.MsgChannelUpdate,
		m.Initial.Params,
		*m.Initial.State,
		indexMapWithLen(m.IndexMap),
	)
	if err != nil {
		return
	}

	return wallet.EncodeSparseSigs(w, m.Initial.Sigs)
}

func (m *VirtualChannelFundingProposal) Decode(r io.Reader) (err error) {
	m.Initial = channel.SignedState{
		Params: &channel.Params{},
		State:  &channel.State{},
	}
	err = perunio.Decode(r,
		&m.MsgChannelUpdate,
		m.Initial.Params,
		m.Initial.State,
		(*indexMapWithLen)(&m.IndexMap),
	)
	if err != nil {
		return
	}

	m.Initial.Sigs = make([]wallet.Sig, m.Initial.State.NumParts())
	return wallet.DecodeSparseSigs(r, &m.Initial.Sigs)
}

// Type returns the message type.
func (*virtualChannelSettlementProposal) Type() wire.Type {
	return wire.VirtualChannelSettlementProposal
}

func (m virtualChannelSettlementProposal) Encode(w io.Writer) (err error) {
	err = perunio.Encode(w,
		m.MsgChannelUpdate,
		m.Final.Params,
		*m.Final.State,
	)
	if err != nil {
		return
	}

	return wallet.EncodeSparseSigs(w, m.Final.Sigs)
}

func (m *virtualChannelSettlementProposal) Decode(r io.Reader) (err error) {
	m.Final = channel.SignedState{
		Params: &channel.Params{},
		State:  &channel.State{},
	}
	err = perunio.Decode(r,
		&m.MsgChannelUpdate,
		m.Final.Params,
		m.Final.State,
	)
	if err != nil {
		return
	}

	m.Final.Sigs = make([]wallet.Sig, m.Final.State.NumParts())
	return wallet.DecodeSparseSigs(r, &m.Final.Sigs)
}
