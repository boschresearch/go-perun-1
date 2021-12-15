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

package wire_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
	wiretest "perun.network/go-perun/wire/test"
	"polycry.pt/poly-go/test"
)

var nilDecoder = func(io.Reader) (perunio.Msg, error) { return nil, nil }

func TestType_Valid_String(t *testing.T) {
	test.OnlyOnce(t)

	const testTypeVal uint8 = 25
	const testTypeStr string = "testTypeA"
	testType := testTypeVal
	assert.False(t, perunio.Valid(testType), "unregistered type should not be valid")
	// assert.Equal(t, strconv.Itoa(testTypeVal), testType.String(),
	// 	"unregistered type's String() should return its integer value")

	perunio.RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr)
	assert.True(t, perunio.Valid(testType), "registered type should be valid")
	// assert.Equal(t, testTypeStr, testType.String(),
	// 	"registered type's String() should be 'testType'")
}

func TestRegisterExternalDecoder(t *testing.T) {
	test.OnlyOnce(t)

	const testTypeVal, testTypeStr = 251, "testTypeB"

	perunio.RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr)
	assert.Panics(t,
		func() { perunio.RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr) },
		"second registration of same type should fail",
	)
	assert.Panics(t,
		func() { perunio.RegisterExternalDecoder(perunio.Ping, nilDecoder, "PingFail") },
		"registration of internal type should fail",
	)
}

func TestEnvelope_EncodeDecode(t *testing.T) {
	ping := wiretest.NewRandomEnvelope(test.Prng(t), wire.NewPingMsg())
	wiretest.GenericSerializerTest(t, ping)
}
