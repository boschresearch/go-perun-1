package protobuf

import (
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
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
		ledgerChannelProposal, err := ToLedgerChannelProposal(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_LedgerChannelProposalMsg{
			LedgerChannelProposalMsg: ledgerChannelProposal,
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

	}

	return env, nil

}

func ToLedgerChannelProposal(p *client.LedgerChannelProposal) (*LedgerChannelProposalMsg, error) {
	baseChannelProposal, err := ToGrpcBaseChannelProposal(p.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	participant, err := ToGrpcWalletAddr(p.Participant)
	if err != nil {
		return nil, err
	}
	peers, err := ToGrpcWireAddrs(p.Peers)
	if err != nil {
		return nil, err
	}

	return &LedgerChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Participant:         participant,
		Peers:               peers,
	}, nil
}

func ToGrpcWalletAddr(a wallet.Address) ([]byte, error) {
	return a.MarshalBinary()
}

func ToGrpcWireAddrs(addrs []wire.Address) (grpcAddrs [][]byte, err error) {
	grpcAddrs = make([][]byte, len(addrs))
	for i := range addrs {
		grpcAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
	}
	return grpcAddrs, nil
}

func ToGrpcBaseChannelProposal(p client.BaseChannelProposal) (*BaseChannelProposal, error) {
	app, err := p.App.Def().MarshalBinary()
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling app")
	}
	initData, err := p.InitData.MarshalBinary()
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling data")
	}
	initBals, err := ToGrpcAllocation(*p.InitBals)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding init bals")
	}
	fundingAgreement, err := ToGrpcBalances(p.FundingAgreement)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding init bals")
	}
	return &BaseChannelProposal{
		ChallengeDuration: p.ChallengeDuration,
		NonceShare:        p.NonceShare[:],
		App:               app,
		InitData:          initData,
		InitBals:          initBals,
		FundingAgreement:  fundingAgreement,
	}, nil
}

func ToGrpcAllocation(a channel.Allocation) (*Allocation, error) {
	var err error
	assets := make([][]byte, len(a.Assets))
	for i := range a.Assets {
		assets[i], err = a.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "marshalling %d'th asset", i)
		}
	}

	bals, err := ToGrpcBalances(a.Balances)
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling balances")
	}

	locked := make([]*SubAlloc, len(a.Locked))
	for i := range a.Locked {
		locked[i], err = ToGrpcSubAlloc(a.Locked[i])
	}

	return &Allocation{
		Assets:   assets,
		Balances: bals,
		Locked:   locked,
	}, nil
}

func ToGrpcBalances(bals channel.Balances) (grpcBals *Balances, err error) {
	grpcBals = &Balances{
		Balances: make([]*Balance, len(bals)),
	}
	for i := range bals {
		grpcBals.Balances[i], err = ToGrpcBalance(bals[i])
		return nil, err
	}
	return grpcBals, nil
}

func ToGrpcBalance(bal []channel.Bal) (*Balance, error) {
	grpcBal := &Balance{
		Balance: make([][]byte, len(bal)),
	}
	for j := range bal {
		if bal[j] == nil {
			return nil, errors.New("logic error: tried to encode nil big.Int")
		}
		if bal[j].Sign() == -1 {
			return nil, errors.New("encoding of negative big.Int not implemented")
		}
		grpcBal.Balance[j] = bal[j].Bytes()
	}
	return grpcBal, nil
}

func ToGrpcSubAlloc(a channel.SubAlloc) (*SubAlloc, error) {
	indexMap := make([]uint32, len(a.IndexMap))
	for i := range a.IndexMap {
		indexMap[i] = uint32(a.IndexMap[i])
	}

	bals, err := ToGrpcBalance(a.Bals)
	if err != nil {
		return nil, err
	}

	return &SubAlloc{
		Id:   a.ID[:],
		Bals: bals,
		IndexMap: &IndexMap{
			IndexMap: indexMap,
		},
	}, nil
}
