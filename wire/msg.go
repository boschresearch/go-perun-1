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

package wire

import (
	"io"
)

var serializer Serializer

// SetSerializer is.
func SetSerializer(s Serializer) {
	if serializer != nil {
		panic("serializer already set")
	}
	serializer = s
}

type (
	// Serializer is.
	Serializer interface {
		EncodeEnvelope(w io.Writer, env *Envelope) error
		DecodeEnvelope(r io.Reader, env *Envelope) error
		EncodeMsg(msg Msg, w io.Writer) error
		DecodeMsg(r io.Reader) (Msg, error)
	}

	// Msg is.
	Msg interface {
		Type() uint8
	}
	// An Envelope encapsulates a message with routing information, that is, the
	// sender and intended recipient.
	Envelope struct {
		Sender    Address // Sender of the message.
		Recipient Address // Recipient of the message.
		// Msg contained in this Envelope. Not embedded so Envelope doesn't implement Msg.
		Msg Msg
	}
)

// Encode encodes an Envelope into an io.Writer.
func (env *Envelope) Encode(w io.Writer) error {
	return serializer.EncodeEnvelope(w, env)
	// if err := perunio.Encode(w, env.Sender, env.Recipient); err != nil {
	// 	return err
	// }
	// return Encode(env.Msg, w)
}

// Decode decodes an Envelope from an io.Reader.
func (env *Envelope) Decode(r io.Reader) (err error) {
	return serializer.DecodeEnvelope(r, env)
	// env.Sender = NewAddress()
	// if err = perunio.Decode(r, env.Sender); err != nil {
	// 	return err
	// }
	// env.Recipient = NewAddress()
	// if err = perunio.Decode(r, env.Recipient); err != nil {
	// 	return err
	// }
	// env.Msg, err = Decode(r)
	// return err
}

// Encode encodes a message into an io.Writer. It also encodes the
// message type whereas the Msg.Encode implementation is assumed not to write
// the type.
func Encode(msg Msg, w io.Writer) error {
	return serializer.EncodeMsg(msg, w)
}

// 	// Encode the message type and payload
// 	return perunio.Encode(w, byte(msg.Type()), msg)
// }

// Decode decodes a message from an io.Reader.
func Decode(r io.Reader) (Msg, error) {
	return serializer.DecodeMsg(r)
}

// 	var t Type
// 	if err := perunio.Decode(r, (*byte)(&t)); err != nil {
// 		return nil, errors.WithMessage(err, "failed to decode message Type")
// 	}

// 	if !t.Valid() {
// 		return nil, errors.Errorf("wire: no decoder known for message Type): %v", t)
// 	}
// 	return decoders[t](r)
// }

// Enumeration of message categories known to the Perun framework.
const (
	Ping uint8 = iota
	Pong
	Shutdown
	AuthResponse
	LedgerChannelProposal
	LedgerChannelProposalAcc
	SubChannelProposal
	SubChannelProposalAcc
	VirtualChannelProposal
	VirtualChannelProposalAcc
	ChannelProposalRej
	ChannelUpdate
	VirtualChannelFundingProposal
	VirtualChannelSettlementProposal
	ChannelUpdateAcc
	ChannelUpdateRej
	ChannelSync
	LastType // upper bound on the message types of the Perun wire protocol
)
