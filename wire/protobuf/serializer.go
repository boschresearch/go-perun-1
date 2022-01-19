package protobuf

import (
	"math/big"
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
	case wire.SubChannelProposal:
		msg := env.Msg.(*client.SubChannelProposal)
		subChannelProposal, err := ToSubChannelProposal(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_SubChannelProposalMsg{
			SubChannelProposalMsg: subChannelProposal,
		}
	case wire.VirtualChannelProposal:
		msg := env.Msg.(*client.VirtualChannelProposal)
		virtualChannelProposal, err := ToVirtualChannelProposal(msg)
		if err != nil {
			return nil, err
		}
		grpcMsg = &Envelope_VirtualChannelProposalMsg{
			VirtualChannelProposalMsg: virtualChannelProposal,
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
		env.Msg, err = FromLedgerChannelProposal(protoEnv.GetLedgerChannelProposalMsg())
	case *Envelope_SubChannelProposalMsg:
		env.Msg, err = FromSubChannelProposal(protoEnv.GetSubChannelProposalMsg())
	case *Envelope_VirtualChannelProposalMsg:
		env.Msg, err = FromVirtualChannelProposal(protoEnv.GetVirtualChannelProposalMsg())
	}

	return env, err

}

func FromLedgerChannelProposal(p *LedgerChannelProposalMsg) (*client.LedgerChannelProposal, error) {
	baseChannelProposal, err := FromGrpcBaseChannelProposal(p.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	participant, err := FromGrpcWalletAddr(p.Participant)
	if err != nil {
		return nil, err
	}
	peers, err := FromGrpcWireAddrs(p.Peers)
	if err != nil {
		return nil, err
	}

	return &client.LedgerChannelProposal{
		BaseChannelProposal: baseChannelProposal,
		Participant:         participant,
		Peers:               peers,
	}, nil
}

func FromSubChannelProposal(p *SubChannelProposalMsg) (*client.SubChannelProposal, error) {
	baseChannelProposal, err := FromGrpcBaseChannelProposal(p.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	q := &client.SubChannelProposal{
		BaseChannelProposal: baseChannelProposal,
	}
	copy(q.Parent[:], p.Parent)
	return q, nil
}

func FromVirtualChannelProposal(p *VirtualChannelProposalMsg) (*client.VirtualChannelProposal, error) {
	baseChannelProposal, err := FromGrpcBaseChannelProposal(p.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	proposer, err := FromGrpcWalletAddr(p.Proposer)
	if err != nil {
		return nil, err
	}
	peers, err := FromGrpcWireAddrs(p.Peers)
	if err != nil {
		return nil, err
	}
	parents := make([]channel.ID, len(p.Parents))
	for i := range p.Parents {
		copy(parents[i][:], p.Parents[i])
	}
	indexMaps := make([][]channel.Index, len(p.IndexMaps.IndexMaps))
	for i := range p.Parents {
		indexMaps[i] = FromGrpcIndexMap(p.IndexMaps.IndexMaps[i].IndexMap)
	}

	return &client.VirtualChannelProposal{
		BaseChannelProposal: baseChannelProposal,
		Proposer:            proposer,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           indexMaps,
	}, nil
}

func FromGrpcWalletAddr(a []byte) (wallet.Address, error) {
	addr := wallet.NewAddress()
	err := addr.UnmarshalBinary(a)
	return addr, err
}

func FromGrpcWireAddrs(addrs [][]byte) (grpcAddrs []wire.Address, err error) {
	grpcAddrs = make([]wire.Address, len(addrs))
	for i := range addrs {
		grpcAddrs[i] = wire.NewAddress()
		err = grpcAddrs[i].UnmarshalBinary(addrs[i])
		if err != nil {
			return nil, err
		}
	}
	return grpcAddrs, nil
}

func FromGrpcBaseChannelProposal(p *BaseChannelProposal) (client.BaseChannelProposal, error) {
	var app channel.App
	var initData channel.Data
	var err error
	if len(p.App) > 0 {
		appDef := wallet.NewAddress()
		err := appDef.UnmarshalBinary(p.App)
		if err != nil {
			return client.BaseChannelProposal{}, errors.WithMessage(err, "unmarshalling app def")
		}
		app, err = channel.Resolve(appDef)
		if err != nil {
			return client.BaseChannelProposal{}, errors.WithMessage(err, "resolving app def")
		}

		initData = app.NewData()
		err = initData.UnmarshalBinary(p.InitData)
		if err != nil {
			return client.BaseChannelProposal{}, errors.WithMessage(err, "marshalling data")
		}
	} else {
		app = channel.NoApp()
		initData = channel.NoData()
	}

	initBals, err := FromGrpcAllocation(p.InitBals)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding init bals")
	}
	fundingAgreement := FromGrpcBalances(p.FundingAgreement)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding init bals")
	}
	bp := client.BaseChannelProposal{
		ChallengeDuration: p.ChallengeDuration,
		App:               app,
		InitData:          initData,
		InitBals:          &initBals,
		FundingAgreement:  fundingAgreement,
	}
	copy(bp.NonceShare[:], p.NonceShare)
	return bp, nil
}

func FromGrpcAllocation(a *Allocation) (channel.Allocation, error) {
	var err error
	assets := make([]channel.Asset, len(a.Assets))
	for i := range a.Assets {
		assets[i] = channel.NewAsset()
		err = assets[i].UnmarshalBinary(a.Assets[i])
		if err != nil {
			return channel.Allocation{}, errors.WithMessagef(err, "marshalling %d'th asset", i)
		}
	}

	locked := make([]channel.SubAlloc, len(a.Locked))
	for i := range a.Locked {
		locked[i], err = FromGrpcSubAlloc(a.Locked[i])
	}

	return channel.Allocation{
		Assets:   assets,
		Balances: FromGrpcBalances(a.Balances),
		Locked:   locked,
	}, nil
}

func FromGrpcBalances(bals *Balances) (grpcBals channel.Balances) {
	grpcBals = make([][]channel.Bal, len(bals.Balances))
	for i := range bals.Balances {
		grpcBals[i] = FromGrpcBalance(bals.Balances[i])
	}
	return grpcBals
}

func FromGrpcBalance(bal *Balance) []channel.Bal {
	grpcBal := make([]channel.Bal, len(bal.Balance))
	for j := range bal.Balance {
		grpcBal[j] = new(big.Int).SetBytes(bal.Balance[j])
	}
	return grpcBal
}

func FromGrpcSubAlloc(a *SubAlloc) (channel.SubAlloc, error) {
	subAlloc := channel.SubAlloc{
		Bals:     FromGrpcBalance(a.Bals),
		IndexMap: FromGrpcIndexMap(a.IndexMap.IndexMap),
	}
	if len(a.Id) != len(subAlloc.ID) {
		return channel.SubAlloc{}, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], a.Id)

	return subAlloc, nil
}

func FromGrpcIndexMap(im []uint32) []channel.Index {
	// In tests, include a test for type of index map is uint16
	indexMap := make([]channel.Index, len(im))
	for i := range im {
		indexMap[i] = channel.Index(uint16(im[i]))
	}
	return indexMap
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

func ToSubChannelProposal(p *client.SubChannelProposal) (*SubChannelProposalMsg, error) {
	baseChannelProposal, err := ToGrpcBaseChannelProposal(p.BaseChannelProposal)
	if err != nil {
		return nil, err
	}

	q := &SubChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,

		Parent: make([]byte, len(p.Parent)),
	}
	copy(q.Parent, p.Parent[:])

	return q, nil
}

func ToVirtualChannelProposal(p *client.VirtualChannelProposal) (*VirtualChannelProposalMsg, error) {
	baseChannelProposal, err := ToGrpcBaseChannelProposal(p.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	proposer, err := ToGrpcWalletAddr(p.Proposer)
	if err != nil {
		return nil, err
	}
	peers, err := ToGrpcWireAddrs(p.Peers)
	if err != nil {
		return nil, err
	}
	parents := make([][]byte, len(p.Parents))
	for i := range p.Parents {
		parents[i] = make([]byte, len(p.Parents[i]))
		copy(parents[i], p.Parents[i][:])
	}
	indexMaps := make([]*IndexMap, len(p.IndexMaps))
	for i := range p.IndexMaps {
		indexMaps[i] = &IndexMap{IndexMap: ToGrpcIndexMap(p.IndexMaps[i])}
	}

	return &VirtualChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Proposer:            proposer,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           &IndexMaps{IndexMaps: indexMaps},
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
	var app []byte
	var initData []byte
	var err error
	if !channel.IsNoApp(p.App) {
		app, err = p.App.Def().MarshalBinary()
		if err != nil {
			return nil, errors.WithMessage(err, "marshalling app")
		}
		initData, err = p.InitData.MarshalBinary()
		if err != nil {
			return nil, errors.WithMessage(err, "marshalling data")
		}
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
		if err != nil {
			return nil, err
		}
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

	bals, err := ToGrpcBalance(a.Bals)
	if err != nil {
		return nil, err
	}

	return &SubAlloc{
		Id:   a.ID[:],
		Bals: bals,
		IndexMap: &IndexMap{
			IndexMap: ToGrpcIndexMap(a.IndexMap),
		},
	}, nil
}

func ToGrpcIndexMap(im []channel.Index) []uint32 {
	indexMap := make([]uint32, len(im))
	for i := range im {
		indexMap[i] = uint32(im[i])
	}
	return indexMap
}
