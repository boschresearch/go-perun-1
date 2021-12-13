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
	"fmt"

	"perun.network/go-perun/wallet"
)

// SigLen length of a signature in byte.
// ref https://godoc.org/github.com/ethereum/go-ethereum/crypto/secp256k1#Sign
// ref https://github.com/ethereum/go-ethereum/blob/54b271a86dd748f3b0bcebeaf678dc34e0d6177a/crypto/signature_cgo.go#L66
const SigLen = 65

// sigVSubtract value that is subtracted from the last byte of a signature if
// the last bytes exceeds it.
const sigVSubtract = 27

type Sig []byte

// MarshalBinary marshals the address into its binary representation.
// Error will always be nil, it is for implementing BinaryMarshaler.
func (s Sig) MarshalBinary() ([]byte, error) {
	return s[:], nil
}

// UnmarshalBinary unmarshals the address from its binary representation.
func (s *Sig) UnmarshalBinary(data []byte) error {
	if len(data) != SigLen {
		return fmt.Errorf("unexpected signature length %d, want %d", len(data), SigLen) //nolint: goerr113
	}
	copy(*s, data)
	return nil
}

// Clone returns a deep copy of the signature.
func (s Sig) Clone() wallet.Sig {
	clone := Sig(make([]byte, SigLen))
	copy(clone, s)
	return &clone
}
