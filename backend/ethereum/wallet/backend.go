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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

// compile-time check that the ethereum backend implements the perun backend.
var _ wallet.Backend = (*Backend)(nil)

// NewAddress returns a variable of type Address, which can be used
// for unmarshalling an address from its binary representation.
func (b *Backend) NewAddress() wallet.Address {
	addr := Address{}
	return &addr
}

// NewAddress returns a variable of type Address, which can be used
// for unmarshalling an address from its binary representation.
func (*Backend) NewSig() wallet.Sig {
	sig := Sig(make([]byte, SigLen))
	return &sig
}

// VerifySignature verifies a signature.
func (*Backend) VerifySignature(msg []byte, sig wallet.Sig, a wallet.Address) (bool, error) {
	return VerifySignature(msg, sig, a)
}

// VerifySignature verifies if a signature was made by this account.
func VerifySignature(msg []byte, sig1 wallet.Sig, a wallet.Address) (bool, error) {
	sig, ok := sig1.(*Sig)
	if !ok {
		return false, errors.New("signature was not of expected type")
	}
	hash := PrefixedHash(msg)
	sigCopy := make([]byte, SigLen)
	copy(sigCopy, *sig)
	if sigCopy[SigLen-1] >= sigVSubtract {
		sigCopy[SigLen-1] -= sigVSubtract
	}
	pk, err := crypto.SigToPub(hash, sigCopy)
	if err != nil {
		return false, errors.WithStack(err)
	}
	addr := crypto.PubkeyToAddress(*pk)
	return a.Equal((*Address)(&addr)), nil
}

// PrefixedHash adds an ethereum specific prefix to the hash of given data, rehashes the results
// and returns it.
func PrefixedHash(data []byte) []byte {
	hash := crypto.Keccak256(data)
	prefix := []byte("\x19Ethereum Signed Message:\n32")
	return crypto.Keccak256(prefix, hash)
}
