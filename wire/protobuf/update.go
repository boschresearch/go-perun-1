package protobuf

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
)

func toVirtualChannelFundingProposalMsg(in *VirtualChannelFundingProposalMsg) (
	*client.VirtualChannelFundingProposal,
	error,
) {
	channelUpdate, err := toChannelUpdate(in.ChannelUpdateMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding update message")
	}

	initial, err := toSignedState(in.Initial)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding signed state")
	}

	return &client.VirtualChannelFundingProposal{
		MsgChannelUpdate: *channelUpdate,
		Initial:          *initial,
		IndexMap:         toIndexMap(in.IndexMap.IndexMap),
	}, nil
}

func toChannelUpdate(in *ChannelUpdateMsg) (*client.MsgChannelUpdate, error) {
	state, err := toState(in.GetChannelUpdate().GetState())
	if err != nil {
		return nil, errors.WithMessage(err, "encoding state")
	}
	out := &client.MsgChannelUpdate{
		ChannelUpdate: client.ChannelUpdate{
			State:    state,
			ActorIdx: channel.Index(in.ChannelUpdate.ActorIdx), // write a test for safe conversion.
		},
		Sig: wallet.Sig(make([]byte, len(in.Sig))),
	}
	copy(out.Sig, in.Sig)
	return out, nil
}

func toChannelUpdateAcc(in *ChannelUpdateAccMsg) (out *client.MsgChannelUpdateAcc) {
	out = &client.MsgChannelUpdateAcc{
		Version: in.Version,
		Sig:     wallet.Sig(make([]byte, len(in.Sig))),
	}
	copy(out.ChannelID[:], in.ChannelID)
	copy(out.Sig, in.Sig)
	return
}

func toChannelUpdateRej(in *ChannelUpdateRejMsg) (out *client.MsgChannelUpdateRej) {
	out = &client.MsgChannelUpdateRej{
		Version: in.Version,
		Reason:  in.Reason,
	}
	copy(out.ChannelID[:], in.ChannelID)
	return
}

func toSignedState(in *SignedState) (*channel.SignedState, error) {
	params, err := toParams(in.Params)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding update message")
	}
	state, err := toState(in.State)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding update message")
	}

	sigs := make([][]byte, len(in.Sigs))
	for i := range in.Sigs {
		sigs[i] = make([]byte, len(in.Sigs[i]))
		copy(sigs[i], in.Sigs[i])
	}

	return &channel.SignedState{
		Params: params,
		State:  state,
		Sigs:   sigs,
	}, nil
}

func toParams(in *Params) (*channel.Params, error) {
	app, err := toApp(in.App)
	if err != nil {
		return nil, err
	}
	parts, err := toWalletAddrs(in.Parts)
	if err != nil {
		return nil, err
	}

	return &channel.Params{
		ChallengeDuration: in.ChallengeDuration,
		Parts:             parts,
		App:               app,
		Nonce:             new(big.Int).SetBytes(in.Nonce),
		LedgerChannel:     in.LedgerChannel,
		VirtualChannel:    in.VirtualChannel,
	}, nil
}

func toApp(appIn []byte) (appOut channel.App, err error) {
	if len(appIn) == 0 {
		appOut = channel.NoApp()
		return
	}
	appDef := wallet.NewAddress()
	err = appDef.UnmarshalBinary(appIn)
	if err != nil {
		return
	}
	appOut, err = channel.Resolve(appDef)
	return
}

func toState(in *State) (*channel.State, error) {
	allocation, err := toAllocation(in.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding allocation")
	}
	app, data, err := toAppAndData(in.App, in.Data)
	if err != nil {
		return nil, err
	}

	out := &channel.State{
		Version:    in.Version,
		App:        app,
		Allocation: allocation,
		Data:       data,
		IsFinal:    in.IsFinal,
	}
	copy(out.ID[:], in.Id)
	fmt.Println(out.ID)
	return out, nil
}

func fromVirtualChannelFundingProposalMsg(in *client.VirtualChannelFundingProposal) (
	*VirtualChannelFundingProposalMsg,
	error,
) {
	channelUpdate, err := fromChannelUpdate(&in.MsgChannelUpdate)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding update message")
	}

	initial, err := fromSignedState(&in.Initial)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding signed state")
	}

	return &VirtualChannelFundingProposalMsg{
		ChannelUpdateMsg: channelUpdate,
		Initial:          initial,
		IndexMap:         &IndexMap{IndexMap: fromIndexMap(in.IndexMap)},
	}, nil
}

func fromChannelUpdate(in *client.MsgChannelUpdate) (*ChannelUpdateMsg, error) {
	state, err := fromState(in.ChannelUpdate.State)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding state")
	}
	out := &ChannelUpdateMsg{
		ChannelUpdate: &ChannelUpdate{
			State:    state,
			ActorIdx: uint32(in.ChannelUpdate.ActorIdx),
		},
		Sig: wallet.Sig(make([]byte, len(in.Sig))),
	}
	copy(out.Sig, in.Sig)
	return out, nil
}

func fromChannelUpdateRej(in *client.MsgChannelUpdateRej) (out *ChannelUpdateRejMsg) {
	out = &ChannelUpdateRejMsg{
		ChannelID: make([]byte, len(in.ChannelID)),
		Version:   in.Version,
		Reason:    in.Reason,
	}
	copy(out.ChannelID, in.ChannelID[:])
	return
}

func fromChannelUpdateAcc(in *client.MsgChannelUpdateAcc) (out *ChannelUpdateAccMsg) {
	out = &ChannelUpdateAccMsg{
		ChannelID: make([]byte, len(in.ChannelID)),
		Version:   in.Version,
		Sig:       make([]byte, len(in.Sig)),
	}
	copy(out.ChannelID, in.ChannelID[:])
	copy(out.Sig, in.Sig)
	return
}

func fromSignedState(in *channel.SignedState) (*SignedState, error) {
	params, err := fromParams(in.Params)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding update message")
	}
	state, err := fromState(in.State)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding update message")
	}

	sigs := make([][]byte, len(in.Sigs))
	for i := range in.Sigs {
		sigs[i] = make([]byte, len(in.Sigs[i]))
		copy(sigs[i], in.Sigs[i])
	}

	return &SignedState{
		Params: params,
		State:  state,
		Sigs:   sigs,
	}, nil
}

func fromParams(in *channel.Params) (*Params, error) {
	app, err := fromApp(in.App)
	if err != nil {
		return nil, err
	}
	parts, err := fromWalletAddrs(in.Parts)
	if err != nil {
		return nil, err
	}
	if in.Nonce.Sign() == -1 {
		return nil, errors.New("encoding of negative big.Int not implemented")
	}

	return &Params{
		ChallengeDuration: in.ChallengeDuration,
		Parts:             parts,
		App:               app,
		Nonce:             in.Nonce.Bytes(),
		LedgerChannel:     in.LedgerChannel,
		VirtualChannel:    in.VirtualChannel,
	}, nil
}

func fromApp(appIn channel.App) (appOut []byte, err error) {
	if channel.IsNoApp(appIn) {
		return
	}
	appOut, err = appIn.Def().MarshalBinary()
	err = errors.WithMessage(err, "marshalling app")
	return
}

func fromWalletAddrs(in []wallet.Address) (out [][]byte, err error) {
	out = make([][]byte, len(in))
	for i := range in {
		out[i], err = in[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func fromState(in *channel.State) (*State, error) {
	allocation, err := fromAllocation(in.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding allocation")
	}
	app, data, err := fromAppAndData(in.App, in.Data)
	if err != nil {
		return nil, err
	}

	out := &State{
		Id:         make([]byte, len(in.ID)),
		Version:    in.Version,
		App:        app,
		Allocation: allocation,
		Data:       data,
		IsFinal:    in.IsFinal,
	}
	copy(out.Id, in.ID[:])
	fmt.Println(out.Id)
	return out, nil
}
