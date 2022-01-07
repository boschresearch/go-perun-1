package protobuf_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "perun.network/go-perun/backend/sim/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/protobuf"
	pkgtest "polycry.pt/poly-go/test"
)

func Test_PingMsg(t *testing.T) {
	rng := pkgtest.Prng(t)

	envelope := newEnvelope(rng)
	envelope.Msg = wire.NewPingMsg()

	data, err := protobuf.EncodeEnvelope(envelope)
	require.NoError(t, err)
	require.NotNil(t, data)

	gotEnvelope, err := protobuf.DecodeEnvelope(data)
	require.NoError(t, err)
	assert.EqualValues(t, envelope, gotEnvelope)
}

func newEnvelope(rng *rand.Rand) wire.Envelope {
	return wire.Envelope{
		Sender:    wallettest.NewRandomAddress(rng),
		Recipient: wallettest.NewRandomAddress(rng),
	}
}
