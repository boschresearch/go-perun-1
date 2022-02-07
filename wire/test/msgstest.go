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

package test

import (
	"testing"

	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"

	pkgtest "polycry.pt/poly-go/test"
)

// ControlMsgsTest runs serialization tests on control messages.
func ControlMsgsTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()
	serializerTest(t, wire.NewPingMsg())
	serializerTest(t, wire.NewPongMsg())
	serializerTest(t, &wire.ShutdownMsg{Reason: "m2384ordkln fb30954390582"})
}

// AuthMsgsTest runs serialization tests on auth message.
func AuthMsgsTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()

	rng := pkgtest.Prng(t)
	serializerTest(t, wire.NewAuthResponseMsg(wallettest.NewRandomAccount(rng)))
}
