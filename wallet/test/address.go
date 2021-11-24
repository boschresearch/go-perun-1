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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/io"
)

// TestAddress runs a test suite designed to test the general functionality of
// an address implementation.
func TestAddress(t *testing.T, s *Setup) {
	null := s.ZeroAddress
	addr, err := s.Backend.DecodeAddress(bytes.NewReader(s.AddressEncoded))
	assert.NoError(t, err, "Byte deserialization of address should work")

	// Test Address.String.
	nullString := null.String()
	addrString := addr.String()
	assert.Greater(t, len(nullString), 0)
	assert.Greater(t, len(addrString), 0)
	assert.NotEqual(t, addrString, nullString)

	// Test Address.Equals.
	assert.False(t, addr.Equals(null), "Expected inequality of zero, nonzero address")
	assert.True(t, null.Equals(null), "Expected equality of zero address to itself")

	// Test Address.Cmp.
	assert.Positive(t, addr.Cmp(null), "Expected addr > zero")
	assert.Zero(t, null.Cmp(null), "Expected zero = zero")
	assert.Negative(t, null.Cmp(addr), "Expected null < addr")

	// Test Address.Bytes.
	addrBytes := addr.Bytes()
	nullBytes := null.Bytes()
	assert.False(t, bytes.Equal(addrBytes, nullBytes), "Expected inequality of byte representations of nonzero and zero address")
	assert.True(t, bytes.Equal(addrBytes, addr.Bytes()), "Expected that byte representations do not change")

	// a.Equals(Decode(Encode(a)))
	t.Run("Serialize Equals Test", func(t *testing.T) {
		buff := new(bytes.Buffer)
		require.NoError(t, io.Encode(buff, addr))
		addr2, err := s.Backend.DecodeAddress(buff)
		require.NoError(t, err)

		assert.True(t, addr.Equals(addr2))
	})
}
