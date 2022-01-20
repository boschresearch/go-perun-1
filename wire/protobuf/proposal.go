package protobuf

import (
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

func toLedgerChannelProposal(in *LedgerChannelProposalMsg) (*client.LedgerChannelProposal, error) {
	baseChannelProposal, err := toBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	participant, err := toWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}
	peers, err := toWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}

	return &client.LedgerChannelProposal{
		BaseChannelProposal: baseChannelProposal,
		Participant:         participant,
		Peers:               peers,
	}, nil
}

func toLedgerChannelProposalAcc(in *LedgerChannelProposalAccMsg) (*client.LedgerChannelProposalAcc, error) {
	participant, err := toWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}

	return &client.LedgerChannelProposalAcc{
		BaseChannelProposalAcc: toBaseChannelProposalAcc(in.BaseChannelProposalAcc),
		Participant:            participant,
	}, nil
}

func toSubChannelProposal(in *SubChannelProposalMsg) (*client.SubChannelProposal, error) {
	baseChannelProposal, err := toBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	out := &client.SubChannelProposal{
		BaseChannelProposal: baseChannelProposal,
	}
	copy(out.Parent[:], in.Parent)
	return out, nil
}

func toSubChannelProposalAcc(in *SubChannelProposalAccMsg) *client.SubChannelProposalAcc {
	return &client.SubChannelProposalAcc{
		BaseChannelProposalAcc: toBaseChannelProposalAcc(in.BaseChannelProposalAcc),
	}
}

func toVirtualChannelProposal(in *VirtualChannelProposalMsg) (*client.VirtualChannelProposal, error) {
	baseChannelProposal, err := toBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	proposer, err := toWalletAddr(in.Proposer)
	if err != nil {
		return nil, err
	}
	peers, err := toWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}
	parents := make([]channel.ID, len(in.Parents))
	for i := range in.Parents {
		copy(parents[i][:], in.Parents[i])
	}
	indexMaps := make([][]channel.Index, len(in.IndexMaps.IndexMaps))
	for i := range in.Parents {
		indexMaps[i] = toIndexMap(in.IndexMaps.IndexMaps[i].IndexMap)
	}

	return &client.VirtualChannelProposal{
		BaseChannelProposal: baseChannelProposal,
		Proposer:            proposer,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           indexMaps,
	}, nil
}

func toVirtualChannelProposalAcc(in *VirtualChannelProposalAccMsg) (*client.VirtualChannelProposalAcc, error) {
	responder, err := toWalletAddr(in.Responder)
	if err != nil {
		return nil, err
	}

	return &client.VirtualChannelProposalAcc{
		BaseChannelProposalAcc: toBaseChannelProposalAcc(in.BaseChannelProposalAcc),
		Responder:              responder,
	}, nil
}

func toChannelProposalRej(in *ChannelProposalRejMsg) (out *client.ChannelProposalRej) {
	out = &client.ChannelProposalRej{}
	copy(out.ProposalID[:], in.ProposalID)
	out.Reason = in.Reason
	return
}

func toWalletAddr(in []byte) (wallet.Address, error) {
	out := wallet.NewAddress()
	return out, out.UnmarshalBinary(in)
}

func toWalletAddrs(in [][]byte) ([]wallet.Address, error) {
	out := make([]wallet.Address, len(in))
	for i := range in {
		out[i] = wallet.NewAddress()
		err := out[i].UnmarshalBinary(in[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func toWireAddrs(in [][]byte) ([]wire.Address, error) {
	out := make([]wire.Address, len(in))
	for i := range in {
		out[i] = wire.NewAddress()
		err := out[i].UnmarshalBinary(in[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func toBaseChannelProposalAcc(in *BaseChannelProposalAcc) (out client.BaseChannelProposalAcc) {
	copy(out.ProposalID[:], in.ProposalID)
	copy(out.NonceShare[:], in.NonceShare)
	return
}

func toBaseChannelProposal(in *BaseChannelProposal) (client.BaseChannelProposal, error) {
	initBals, err := toAllocation(in.InitBals)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding init bals")
	}
	fundingAgreement := toBalances(in.FundingAgreement)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding init bals")
	}
	app, initData, err := toAppAndData(in.App, in.InitData)
	out := client.BaseChannelProposal{
		ChallengeDuration: in.ChallengeDuration,
		App:               app,
		InitData:          initData,
		InitBals:          &initBals,
		FundingAgreement:  fundingAgreement,
	}
	copy(out.NonceShare[:], in.NonceShare)
	return out, nil
}

func toAppAndData(appIn, dataIn []byte) (appOut channel.App, dataOut channel.Data, err error) {
	if len(appIn) == 0 {
		appOut = channel.NoApp()
		dataOut = channel.NoData()
		return
	}
	appDef := wallet.NewAddress()
	err = appDef.UnmarshalBinary(appIn)
	if err != nil {
		return
	}
	appOut, err = channel.Resolve(appDef)
	if err != nil {
		return
	}
	dataOut = appOut.NewData()
	err = dataOut.UnmarshalBinary(dataIn)
	return
}

func toAllocation(in *Allocation) (channel.Allocation, error) {
	var err error
	assets := make([]channel.Asset, len(in.Assets))
	for i := range in.Assets {
		assets[i] = channel.NewAsset()
		err = assets[i].UnmarshalBinary(in.Assets[i])
		if err != nil {
			return channel.Allocation{}, errors.WithMessagef(err, "marshalling %d'th asset", i)
		}
	}

	locked := make([]channel.SubAlloc, len(in.Locked))
	for i := range in.Locked {
		locked[i], err = toSubAlloc(in.Locked[i])
	}

	return channel.Allocation{
		Assets:   assets,
		Balances: toBalances(in.Balances),
		Locked:   locked,
	}, nil
}

func toBalances(in *Balances) (out channel.Balances) {
	out = make([][]channel.Bal, len(in.Balances))
	for i := range in.Balances {
		out[i] = toBalance(in.Balances[i])
	}
	return out
}

func toBalance(in *Balance) []channel.Bal {
	out := make([]channel.Bal, len(in.Balance))
	for j := range in.Balance {
		out[j] = new(big.Int).SetBytes(in.Balance[j])
	}
	return out
}

func toSubAlloc(in *SubAlloc) (channel.SubAlloc, error) {
	subAlloc := channel.SubAlloc{
		Bals:     toBalance(in.Bals),
		IndexMap: toIndexMap(in.IndexMap.IndexMap),
	}
	if len(in.Id) != len(subAlloc.ID) {
		return channel.SubAlloc{}, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], in.Id)

	return subAlloc, nil
}

func toIndexMap(in []uint32) []channel.Index {
	// In tests, include a test for type of index map is uint16
	indexMap := make([]channel.Index, len(in))
	for i := range in {
		indexMap[i] = channel.Index(uint16(in[i]))
	}
	return indexMap
}

func fromLedgerChannelProposal(in *client.LedgerChannelProposal) (*LedgerChannelProposalMsg, error) {
	baseChannelProposal, err := fromBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	participant, err := fromWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}
	peers, err := fromWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}

	return &LedgerChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Participant:         participant,
		Peers:               peers,
	}, nil
}

func fromLedgerChannelProposalAcc(in *client.LedgerChannelProposalAcc) (*LedgerChannelProposalAccMsg, error) {
	baseChannelProposalAcc := fromBaseChannelProposalAcc(in.BaseChannelProposalAcc)
	participant, err := fromWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}

	return &LedgerChannelProposalAccMsg{
		BaseChannelProposalAcc: baseChannelProposalAcc,
		Participant:            participant,
	}, nil
}

func fromSubChannelProposal(in *client.SubChannelProposal) (*SubChannelProposalMsg, error) {
	baseChannelProposal, err := fromBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}

	out := &SubChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Parent:              make([]byte, len(in.Parent)),
	}
	copy(out.Parent, in.Parent[:])
	return out, nil
}

func fromSubChannelProposalAcc(in *client.SubChannelProposalAcc) *SubChannelProposalAccMsg {
	return &SubChannelProposalAccMsg{
		BaseChannelProposalAcc: fromBaseChannelProposalAcc(in.BaseChannelProposalAcc),
	}
}

func fromVirtualChannelProposal(in *client.VirtualChannelProposal) (*VirtualChannelProposalMsg, error) {
	baseChannelProposal, err := fromBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	proposer, err := fromWalletAddr(in.Proposer)
	if err != nil {
		return nil, err
	}
	peers, err := fromWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}
	parents := make([][]byte, len(in.Parents))
	for i := range in.Parents {
		parents[i] = make([]byte, len(in.Parents[i]))
		copy(parents[i], in.Parents[i][:])
	}
	indexMaps := make([]*IndexMap, len(in.IndexMaps))
	for i := range in.IndexMaps {
		indexMaps[i] = &IndexMap{IndexMap: fromIndexMap(in.IndexMaps[i])}
	}

	return &VirtualChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Proposer:            proposer,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           &IndexMaps{IndexMaps: indexMaps},
	}, nil
}

func fromVirtualChannelProposalAcc(in *client.VirtualChannelProposalAcc) (*VirtualChannelProposalAccMsg, error) {
	baseChannelProposalAcc := fromBaseChannelProposalAcc(in.BaseChannelProposalAcc)
	responder, err := fromWalletAddr(in.Responder)
	if err != nil {
		return nil, err
	}

	return &VirtualChannelProposalAccMsg{
		BaseChannelProposalAcc: baseChannelProposalAcc,
		Responder:              responder,
	}, nil
}

func fromChannelProposalRej(in *client.ChannelProposalRej) (out *ChannelProposalRejMsg) {
	out = &ChannelProposalRejMsg{}
	out.ProposalID = make([]byte, len(in.ProposalID))
	copy(out.ProposalID, in.ProposalID[:])
	out.Reason = in.Reason
	return
}

func fromWalletAddr(in wallet.Address) ([]byte, error) {
	return in.MarshalBinary()
}

func fromWireAddrs(in []wire.Address) (out [][]byte, err error) {
	out = make([][]byte, len(in))
	for i := range in {
		out[i], err = in[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func fromBaseChannelProposalAcc(in client.BaseChannelProposalAcc) (out *BaseChannelProposalAcc) {
	out = &BaseChannelProposalAcc{}
	out.ProposalID = make([]byte, len(in.ProposalID))
	out.NonceShare = make([]byte, len(in.NonceShare))
	copy(out.ProposalID, in.ProposalID[:])
	copy(out.NonceShare, in.NonceShare[:])
	return
}

func fromBaseChannelProposal(in client.BaseChannelProposal) (*BaseChannelProposal, error) {
	initBals, err := fromAllocation(*in.InitBals)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding init bals")
	}
	fundingAgreement, err := fromBalances(in.FundingAgreement)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding init bals")
	}
	app, initData, err := fromAppAndData(in.App, in.InitData)
	if err != nil {
		return nil, err
	}
	return &BaseChannelProposal{
		ChallengeDuration: in.ChallengeDuration,
		NonceShare:        in.NonceShare[:],
		App:               app,
		InitData:          initData,
		InitBals:          initBals,
		FundingAgreement:  fundingAgreement,
	}, nil
}

func fromAppAndData(appIn channel.App, dataIn channel.Data) (appOut, dataOut []byte, err error) {
	if channel.IsNoApp(appIn) {
		return
	}
	appOut, err = appIn.Def().MarshalBinary()
	if err != nil {
		err = errors.WithMessage(err, "marshalling app")
		return
	}
	dataOut, err = dataIn.MarshalBinary()
	err = errors.WithMessage(err, "marshalling app")
	return
}

func fromAllocation(in channel.Allocation) (*Allocation, error) {
	assets := make([][]byte, len(in.Assets))
	var err error
	for i := range in.Assets {
		assets[i], err = in.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "marshalling %d'th asset", i)
		}
	}
	bals, err := fromBalances(in.Balances)
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling balances")
	}
	locked := make([]*SubAlloc, len(in.Locked))
	for i := range in.Locked {
		locked[i], err = fromSubAlloc(in.Locked[i])
	}

	return &Allocation{
		Assets:   assets,
		Balances: bals,
		Locked:   locked,
	}, nil
}

func fromBalances(in channel.Balances) (out *Balances, err error) {
	out = &Balances{
		Balances: make([]*Balance, len(in)),
	}
	for i := range in {
		out.Balances[i], err = fromBalance(in[i])
		if err != nil {
			return nil, err
		}
	}
	return
}

func fromBalance(in []channel.Bal) (*Balance, error) {
	out := &Balance{
		Balance: make([][]byte, len(in)),
	}
	for j := range in {
		if in[j] == nil {
			return nil, errors.New("logic error: tried to encode nil big.Int")
		}
		if in[j].Sign() == -1 {
			return nil, errors.New("encoding of negative big.Int not implemented")
		}
		out.Balance[j] = in[j].Bytes()
	}
	return out, nil
}

func fromSubAlloc(in channel.SubAlloc) (*SubAlloc, error) {
	bals, err := fromBalance(in.Bals)
	if err != nil {
		return nil, err
	}
	return &SubAlloc{
		Id:   in.ID[:],
		Bals: bals,
		IndexMap: &IndexMap{
			IndexMap: fromIndexMap(in.IndexMap),
		},
	}, nil
}

func fromIndexMap(in []channel.Index) []uint32 {
	out := make([]uint32, len(in))
	for i := range in {
		out[i] = uint32(in[i])
	}
	return out
}
