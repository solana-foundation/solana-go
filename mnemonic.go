// Copyright 2026 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package solana

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	voied25519 "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/tyler-smith/go-bip39"
)

// SolanaDerivationPath is the default BIP-44 derivation path used by Phantom
// and most Solana wallets when generating a key from a BIP-39 mnemonic:
// m/44'/501'/0'/0'.
const SolanaDerivationPath = "m/44'/501'/0'/0'"

// PrivateKeyFromMnemonic derives a PrivateKey from a BIP-39 mnemonic using
// the default Solana derivation path (m/44'/501'/0'/0'). The passphrase may
// be empty; when set, it must match the passphrase used when the mnemonic
// was generated.
func PrivateKeyFromMnemonic(mnemonic, passphrase string) (PrivateKey, error) {
	return PrivateKeyFromMnemonicAtPath(mnemonic, passphrase, SolanaDerivationPath)
}

// PrivateKeyFromMnemonicAtPath derives a PrivateKey from a BIP-39 mnemonic at
// the given SLIP-0010 derivation path. All path segments must be hardened
// (suffixed with ' or h); SLIP-0010 does not define non-hardened derivation
// for ed25519.
func PrivateKeyFromMnemonicAtPath(mnemonic, passphrase, path string) (PrivateKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic")
	}
	seed := bip39.NewSeed(mnemonic, passphrase)
	return PrivateKeyFromSeedAtPath(seed, path)
}

// PrivateKeyFromSeedAtPath derives a PrivateKey from a 16..64 byte seed
// (typically a 64 byte BIP-39 seed) using the given SLIP-0010 derivation
// path. All path segments must be hardened.
func PrivateKeyFromSeedAtPath(seed []byte, path string) (PrivateKey, error) {
	indices, err := parseDerivationPath(path)
	if err != nil {
		return nil, err
	}
	key, chainCode, err := slip10MasterKey(seed)
	if err != nil {
		return nil, err
	}
	for _, index := range indices {
		key, chainCode = slip10ChildKey(key, chainCode, index)
	}
	return PrivateKey(voied25519.NewKeyFromSeed(key)), nil
}

// NewWalletFromMnemonic creates a Wallet whose private key is derived from a
// BIP-39 mnemonic using the default Solana derivation path m/44'/501'/0'/0'.
func NewWalletFromMnemonic(mnemonic, passphrase string) (*Wallet, error) {
	pk, err := PrivateKeyFromMnemonic(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}
	return &Wallet{PrivateKey: pk}, nil
}

const slip10HardenedOffset = uint32(0x80000000)

// parseDerivationPath parses a SLIP-0010 derivation path of the form
// m/idx'/idx'/... Each index must be hardened (suffixed with ' or h). An
// empty path or "m" returns no indices, meaning the master key is used.
func parseDerivationPath(path string) ([]uint32, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "m" || path == "/" {
		return nil, nil
	}
	path = strings.TrimPrefix(path, "m")
	path = strings.TrimPrefix(path, "/")
	segments := strings.Split(path, "/")
	indices := make([]uint32, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			return nil, fmt.Errorf("invalid derivation path %q: empty segment", path)
		}
		hardened := false
		switch seg[len(seg)-1] {
		case '\'', 'h', 'H':
			hardened = true
			seg = seg[:len(seg)-1]
		}
		n, err := strconv.ParseUint(seg, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid derivation path %q: %w", path, err)
		}
		if n >= uint64(slip10HardenedOffset) {
			return nil, fmt.Errorf("invalid derivation path %q: index %d out of range", path, n)
		}
		if !hardened {
			return nil, fmt.Errorf("invalid derivation path %q: SLIP-0010 ed25519 requires all segments to be hardened", path)
		}
		indices = append(indices, uint32(n)+slip10HardenedOffset)
	}
	return indices, nil
}

// slip10MasterKey derives the SLIP-0010 master key for ed25519 from a seed,
// per https://github.com/satoshilabs/slips/blob/master/slip-0010.md.
func slip10MasterKey(seed []byte) (key, chainCode []byte, err error) {
	if l := len(seed); l < 16 || l > 64 {
		return nil, nil, fmt.Errorf("invalid seed length %d (want 16..64 bytes)", l)
	}
	mac := hmac.New(sha512.New, []byte("ed25519 seed"))
	mac.Write(seed)
	sum := mac.Sum(nil)
	return sum[:32], sum[32:], nil
}

// slip10ChildKey derives a hardened SLIP-0010 child key for ed25519. The
// caller must ensure index >= slip10HardenedOffset; non-hardened derivation
// is not defined for ed25519.
func slip10ChildKey(parentKey, parentChainCode []byte, index uint32) (key, chainCode []byte) {
	mac := hmac.New(sha512.New, parentChainCode)
	mac.Write([]byte{0x00})
	mac.Write(parentKey)
	_ = binary.Write(mac, binary.BigEndian, index)
	sum := mac.Sum(nil)
	return sum[:32], sum[32:]
}
