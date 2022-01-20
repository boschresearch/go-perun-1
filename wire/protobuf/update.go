package protobuf

import (
	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
)

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
	return out, nil
}
