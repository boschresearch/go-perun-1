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
	"bytes"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

// compile time check that we implement the perun Address interface.
var _ wallet.Address = (*Address)(nil)

// Address represents an ethereum address as a perun address.
type Address common.Address

// Bytes returns the address as a byte slice.
func (a *Address) Bytes() []byte {
	return (*common.Address)(a).Bytes()
}

// MarshalBinary marhals the address into a binary form.
// Error will always be nil, it is for implementing BinaryMarshaler.
func (a *Address) MarshalBinary() ([]byte, error) {
	return (*common.Address)(a).Bytes(), nil
}

// Decode decodes an address from a io.Reader. Part of the
// go-perun/pkg/io.Serializer interface.
func (a *Address) Decode(r io.Reader) error {
	var length int64
	err := perunio.Decode(r, &length)
	if err != nil {
		return err
	}
	if length != common.AddressLength {
		return fmt.Errorf("unexpected address length %d, want %d", length, common.AddressLength)
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	(*common.Address)(a).SetBytes(buf)
	return errors.Wrap(err, "error decoding address")
}

// String converts this address to a string.
func (a *Address) String() string {
	return (*common.Address)(a).String()
}

// Equals checks the equality of two addresses. The implementation must be
// equivalent to checking `Address.Cmp(Address) == 0`.
func (a *Address) Equals(addr wallet.Address) bool {
	return bytes.Equal(a.Bytes(), addr.(*Address).Bytes())
}

// Cmp checks ordering of two addresses.
//  0 if a==b,
// -1 if a < b,
// +1 if a > b.
// https://godoc.org/bytes#Compare
func (a *Address) Cmp(addr wallet.Address) int {
	return bytes.Compare(a.Bytes(), addr.(*Address).Bytes())
}

// AsEthAddr is a helper function to convert an address interface back into an
// ethereum address.
func AsEthAddr(a wallet.Address) common.Address {
	return common.Address(*a.(*Address))
}

// AsWalletAddr is a helper function to convert an ethereum address to an
// address interface.
func AsWalletAddr(addr common.Address) *Address {
	return (*Address)(&addr)
}
