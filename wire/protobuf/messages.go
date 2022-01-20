package protobuf

import (
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"perun.network/go-perun/client"
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

	var grpcMsg isEnvelope_Msg
	switch env.Msg.Type() {
	case wire.Ping:
		msg := env.Msg.(*wire.PingMsg)
		grpcMsg = &Envelope_PingMsg{
			PingMsg: &PingMsg{
				Created: msg.Created.UnixNano(),
			},
		}
	case wire.Pong:
		msg := env.Msg.(*wire.PongMsg)
		grpcMsg = &Envelope_PongMsg{
			PongMsg: &PongMsg{
				Created: msg.Created.UnixNano(),
			},
		}
	case wire.Shutdown:
		msg := env.Msg.(*wire.ShutdownMsg)
		grpcMsg = &Envelope_ShutdownMsg{
			ShutdownMsg: &ShutdownMsg{
				Reason: msg.Reason,
			},
		}
	case wire.AuthResponse:
		grpcMsg = &Envelope_AuthResponseMsg{
			AuthResponseMsg: &AuthResponseMsg{},
		}
	case wire.LedgerChannelProposal:
		msg := env.Msg.(*client.LedgerChannelProposal)
		ledgerChannelProposal, err := fromLedgerChannelProposal(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_LedgerChannelProposalMsg{
			LedgerChannelProposalMsg: ledgerChannelProposal,
		}
	case wire.LedgerChannelProposalAcc:
		msg := env.Msg.(*client.LedgerChannelProposalAcc)
		ledgerChannelProposalAcc, err := fromLedgerChannelProposalAcc(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_LedgerChannelProposalAccMsg{
			LedgerChannelProposalAccMsg: ledgerChannelProposalAcc,
		}

	case wire.SubChannelProposal:
		msg := env.Msg.(*client.SubChannelProposal)
		subChannelProposal, err := fromSubChannelProposal(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_SubChannelProposalMsg{
			SubChannelProposalMsg: subChannelProposal,
		}
	case wire.SubChannelProposalAcc:
		msg := env.Msg.(*client.SubChannelProposalAcc)
		grpcMsg = &Envelope_SubChannelProposalAccMsg{
			SubChannelProposalAccMsg: fromSubChannelProposalAcc(msg),
		}
	case wire.VirtualChannelProposal:
		msg := env.Msg.(*client.VirtualChannelProposal)
		virtualChannelProposal, err := fromVirtualChannelProposal(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_VirtualChannelProposalMsg{
			VirtualChannelProposalMsg: virtualChannelProposal,
		}
	case wire.VirtualChannelProposalAcc:
		msg := env.Msg.(*client.VirtualChannelProposalAcc)
		virtualChannelProposalAcc, err := fromVirtualChannelProposalAcc(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_VirtualChannelProposalAccMsg{
			VirtualChannelProposalAccMsg: virtualChannelProposalAcc,
		}
	case wire.ChannelProposalRej:
		msg := env.Msg.(*client.ChannelProposalRej)
		grpcMsg = &Envelope_ChannelProposalRejMsg{
			ChannelProposalRejMsg: fromChannelProposalRej(msg),
		}
	case wire.ChannelUpdateRej:
		msg := env.Msg.(*client.MsgChannelUpdateRej)
		grpcMsg = &Envelope_ChannelUpdateRejMsg{
			ChannelUpdateRejMsg: fromChannelUpdateRej(msg),
		}
	}

	protoEnv := Envelope{
		Sender:    sender,
		Recipient: recipient,
		Msg:       grpcMsg,
	}

	data, err := proto.Marshal(&protoEnv)
	return data, errors.Wrap(err, "marshalling envelope")
}

func DecodeEnvelope(data []byte) (wire.Envelope, error) {
	var err error
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
	case *Envelope_PongMsg:
		env.Msg = &wire.PongMsg{
			wire.PingPongMsg{
				Created: time.Unix(0, protoEnv.GetPongMsg().Created),
			},
		}
	case *Envelope_ShutdownMsg:
		env.Msg = &wire.ShutdownMsg{
			Reason: protoEnv.GetShutdownMsg().Reason,
		}
	case *Envelope_AuthResponseMsg:
		env.Msg = &wire.AuthResponseMsg{}

	case *Envelope_LedgerChannelProposalMsg:
		env.Msg, err = toLedgerChannelProposal(protoEnv.GetLedgerChannelProposalMsg())

	case *Envelope_LedgerChannelProposalAccMsg:
		env.Msg, err = toLedgerChannelProposalAcc(protoEnv.GetLedgerChannelProposalAccMsg())

	case *Envelope_SubChannelProposalMsg:
		env.Msg, err = toSubChannelProposal(protoEnv.GetSubChannelProposalMsg())

	case *Envelope_SubChannelProposalAccMsg:
		env.Msg = toSubChannelProposalAcc(protoEnv.GetSubChannelProposalAccMsg())

	case *Envelope_VirtualChannelProposalMsg:
		env.Msg, err = toVirtualChannelProposal(protoEnv.GetVirtualChannelProposalMsg())

	case *Envelope_VirtualChannelProposalAccMsg:
		env.Msg, err = toVirtualChannelProposalAcc(protoEnv.GetVirtualChannelProposalAccMsg())

	case *Envelope_ChannelProposalRejMsg:
		env.Msg = toChannelProposalRej(protoEnv.GetChannelProposalRejMsg())
	case *Envelope_ChannelUpdateRejMsg:
		env.Msg = toChannelUpdateRej(protoEnv.GetChannelUpdateRejMsg())
	}

	return env, err

}
