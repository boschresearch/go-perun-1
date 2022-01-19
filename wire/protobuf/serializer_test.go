package protobuf_test

import (
	"math/rand"

	_ "perun.network/go-perun/backend/sim/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

// func Test_Msg_Encode_Decode(t *testing.T) {

// 	tests := []struct {
// 		name string
// 		msg  wire.Msg
// 	}{
// 		{name: "ping", msg: wire.NewPingMsg()},
// 		{name: "pong", msg: wire.NewPongMsg()},
// 		{name: "shutdown", msg: &wire.ShutdownMsg{"m2384ordkln fb30954390582"}},
// 		{name: "authResponse", msg: &wire.AuthResponseMsg{}},
// 		{name: "ledgerChannelProposal", msg:
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			rng := pkgtest.Prng(t)

// 			envelope := newEnvelope(rng)
// 			envelope.Msg = tc.msg

// 			data, err := protobuf.EncodeEnvelope(envelope)
// 			require.NoError(t, err)
// 			require.NotNil(t, data)

// 			gotEnvelope, err := protobuf.DecodeEnvelope(data)
// 			require.NoError(t, err)
// 			assert.EqualValues(t, envelope, gotEnvelope)
// 		})
// 	}

// }

//	*Envelope_LedgerChannelProposalMsg
//	*Envelope_LedgerChannelProposalAccMsg
//	*Envelope_SubChannelProposalMsg
//	*Envelope_SubChannelProposalAccMsg
//	*Envelope_VirtualChannelProposalMsg
//	*Envelope_VirtualChannelProposalAccMsg
//	*Envelope_ChannelProposalRejMsg
//	*Envelope_ChannelUpdateMsg
//	*Envelope_VirtualChannelFundingProposalMsg
//	*Envelope_VirtualChannelSettlementProposalMsg
//	*Envelope_ChannelUpdateAccMsg
//	*Envelope_ChannelUpdateRejMsg
//	*Envelope_ChannelSyncMsg
func newEnvelope(rng *rand.Rand) wire.Envelope {
	return wire.Envelope{
		Sender:    wallettest.NewRandomAddress(rng),
		Recipient: wallettest.NewRandomAddress(rng),
	}
}
