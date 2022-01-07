package protobuf

import (
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"perun.network/go-perun/wire"
)

func EncodeEnvelope(env wire.Envelope) ([]byte, error) {
	sender, err := env.Sender.MarshalBinary()
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling sender address")
	}
	recipient, err := env.Recipient.MarshalBinary()
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling recipient address")
	}

	var msg isEnvelope_Msg
	switch env.Msg.Type() {
	case wire.Ping:
		pingMsg := env.Msg.(*wire.PingMsg)
		msg = &Envelope_PingMsg{
			PingMsg: &PingMsg{
				Created: pingMsg.Created.UnixNano(),
			},
		}
	}

	protoEnv := Envelope{
		Sender:    sender,
		Recipient: recipient,
		Msg:       msg,
	}

	data, err := proto.Marshal(&protoEnv)
	return data, errors.Wrap(err, "marshalling envelope")
}

func DecodeEnvelope(data []byte) (wire.Envelope, error) {
	var protoEnv Envelope
	if err := proto.Unmarshal(data, &protoEnv); err != nil {
		return wire.Envelope{}, errors.Wrap(err, "unmarshalling envelope")
	}

	env := wire.Envelope{}
	env.Sender = wire.NewAddress()
	if err := env.Sender.UnmarshalBinary(protoEnv.Sender); err != nil {
		return wire.Envelope{}, errors.Wrap(err, "unmarshalling sender address")
	}
	env.Recipient = wire.NewAddress()
	if err := env.Recipient.UnmarshalBinary(protoEnv.Recipient); err != nil {
		return wire.Envelope{}, errors.Wrap(err, "unmarshalling recipient address")
	}

	switch protoEnv.Msg.(type) {
	case *Envelope_PingMsg:
		env.Msg = &wire.PingMsg{
			wire.PingPongMsg{
				Created: time.Unix(0, protoEnv.GetPingMsg().Created),
			},
		}
	}

	return env, nil

}
