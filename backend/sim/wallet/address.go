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

package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

// AddressBinaryLength is the length of the binary representation of Address.
const AddressBinaryLength = 64

// Address represents a simulated address.
type Address ecdsa.PublicKey

// compile time check that we implement the perun Address interface.
var _ wallet.Address = (*Address)(nil)

// NewRandomAddress creates a new address using the randomness
// provided by rng.
func NewRandomAddress(rng io.Reader) *Address {
	privateKey, err := ecdsa.GenerateKey(curve, rng)

	if err != nil {
		log.Panicf("Creation of account failed with error", err)
	}

	return &Address{
		Curve: privateKey.Curve,
		X:     privateKey.X,
		Y:     privateKey.Y,
	}
}

// Bytes converts this address to bytes.
func (a *Address) Bytes() []byte {
	data := a.ByteArray()
	return data[:]
}

// ByteArray converts an address into a 64-byte array. The returned array
// consists of two 32-byte chunks representing the public key's X and Y values.
func (a *Address) ByteArray() (data [64]byte) {
	xb := a.X.Bytes()
	yb := a.Y.Bytes()

	// Left-pad with 0 bytes.
	copy(data[32-len(xb):32], xb)
	copy(data[64-len(yb):64], yb)

	return data
}

// String converts this address to a human-readable string.
func (a *Address) String() string {
	// Encode the address directly instead of using Address.Bytes() because
	// * some addresses may have a very short encoding, e.g., the null address,
	// * the Address.Bytes() output may contain encoding information, e.g., the
	//   length.
	bs := make([]byte, 4)
	copy(bs, a.X.Bytes())

	return "0x" + hex.EncodeToString(bs)
}

// Equals checks the equality of two addresses. The implementation must be
// equivalent to checking `Address.Cmp(Address) == 0`.
func (a *Address) Equals(addr wallet.Address) bool {
	b := addr.(*Address)
	return (a.X.Cmp(b.X) == 0) && (a.Y.Cmp(b.Y) == 0)
}

// Cmp checks the ordering of two addresses according to following definition:
//   -1 if (a.X <  addr.X) || ((a.X == addr.X) && (a.Y < addr.Y))
//    0 if (a.X == addr.X) && (a.Y == addr.Y)
//   +1 if (a.X >  addr.X) || ((a.X == addr.X) && (a.Y > addr.Y))
// So the X coordinate is weighted higher.
func (a *Address) Cmp(addr wallet.Address) int {
	b := addr.(*Address)
	const EQ = 0
	xCmp, yCmp := a.X.Cmp(b.X), a.Y.Cmp(b.Y)
	if xCmp != EQ {
		return xCmp
	}
	return yCmp
}

// MarshalBinary marshals the address into a binary form.
// Error will always be nil, it is for implementing BinaryMarshaler.
func (a *Address) MarshalBinary() ([]byte, error) {
	data := a.ByteArray()
	return data[:], nil
}

// UnmarshalBinary unmarshalled the address from its binary representation.
func (a *Address) UnmarshalBinary(data []byte) error {
	if len(data) != AddressBinaryLength {
		return fmt.Errorf("unexpected address length %d, want %d", len(data), AddressBinaryLength)
	}

	a.X = new(big.Int).SetBytes(data[:32])
	a.Y = new(big.Int).SetBytes(data[32:])
	a.Curve = curve

	return nil
}
