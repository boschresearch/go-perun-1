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

package proxy

import (
	"io"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
)

type serializer struct{}

func init() {
	wire.SetSerializer(serializer{})
}

func (s serializer) EncodeMsg(msg wire.Msg, w io.Writer) error {
	// Encode the message type and payload
	return perunio.Encode(w, msg.Type(), msg)
}

func (s serializer) DecodeMsg(r io.Reader) (wire.Msg, error) {
	var t uint8
	if err := perunio.Decode(r, &t); err != nil {
		return nil, errors.WithMessage(err, "failed to decode message Type")
	}

	if !perunio.Valid(t) {
		return nil, errors.Errorf("wire: no decoder known for message Type): %v", t)
	}
	return perunio.Decoders[t](r)
}

// Encode encodes an Envelope into an io.Writer.
func (s serializer) EncodeEnvelope(w io.Writer, env *wire.Envelope) error {
	if err := perunio.Encode(w, env.Sender, env.Recipient); err != nil {
		return err
	}
	return s.EncodeMsg(env.Msg, w)
}

// Decode decodes an Envelope from an io.Reader.
func (s serializer) DecodeEnvelope(r io.Reader, env *wire.Envelope) (err error) {
	env.Sender = wire.NewAddress()
	if err = perunio.Decode(r, env.Sender); err != nil {
		return err
	}
	env.Recipient = wire.NewAddress()
	if err = perunio.Decode(r, env.Recipient); err != nil {
		return err
	}
	env.Msg, err = s.DecodeMsg(r)
	return err
}
