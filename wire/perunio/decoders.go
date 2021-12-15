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

package perunio

import (
	"fmt"
	"io"
)

// Decoders ...
var Decoders = make(map[uint8]func(io.Reader) (Msg, error))

// Msg .
type Msg interface {
	// Type returns the message's type.
	Type() uint8
	// encoding of payload. Type byte should not be encoded.
	Encoder
}

// RegisterDecoder sets the decoder of messages of Type `t`.
func RegisterDecoder(t uint8, decoder func(io.Reader) (Msg, error)) {
	if Decoders[t] != nil {
		panic(fmt.Sprintf("wire: decoder for Type %v already set", t))
	}

	Decoders[t] = decoder
}

// RegisterExternalDecoder sets the decoder of messages of external type `t`.
// This is like RegisterDecoder but for message types not part of the Perun wire
// protocol and thus not known natively. This can be used by users of the
// framework to create additional message types and send them over the same
// peer connection. It also comes in handy to register types for testing.
func RegisterExternalDecoder(t uint8, decoder func(io.Reader) (Msg, error), name string) {
	if t < LastType {
		panic("external decoders can only be registered for alien types")
	}
	RegisterDecoder(t, decoder)
	// above registration panics if already set, so we don't need to check the
	// next assignment.
	typeNames[t] = name
}

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

var typeNames = map[uint8]string{
	Ping:                             "Ping",
	Pong:                             "Pong",
	Shutdown:                         "Shutdown",
	AuthResponse:                     "AuthResponse",
	LedgerChannelProposal:            "LedgerChannelProposal",
	LedgerChannelProposalAcc:         "LedgerChannelProposalAcc",
	SubChannelProposal:               "SubChannelProposal",
	SubChannelProposalAcc:            "SubChannelProposalAcc",
	VirtualChannelProposal:           "VirtualChannelProposal",
	VirtualChannelProposalAcc:        "VirtualChannelProposalAcc",
	ChannelProposalRej:               "ChannelProposalRej",
	ChannelUpdate:                    "ChannelUpdate",
	VirtualChannelFundingProposal:    "VirtualChannelFundingProposal",
	VirtualChannelSettlementProposal: "VirtualChannelSettlementProposal",
	ChannelUpdateAcc:                 "ChannelUpdateAcc",
	ChannelUpdateRej:                 "ChannelUpdateRej",
	ChannelSync:                      "ChannelSync",
}

// String returns the name of a message type if it is valid and name known
// or otherwise its numerical representation.
// func (t Type) String() string {
// 	name, ok := typeNames[t]
// 	if !ok {
// 		return strconv.Itoa(int(t))
// 	}
// 	return name
// }

// Valid checks whether a decoder is known for the type.
func Valid(t uint8) bool {
	_, ok := Decoders[t]
	return ok
}
