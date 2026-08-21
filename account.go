// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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
	"bytes"
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
)

// Wallet is a wrapper around a PrivateKey
type Wallet struct {
	PrivateKey PrivateKey
}

func NewWallet() *Wallet {
	privateKey, err := NewRandomPrivateKey()
	if err != nil {
		panic(fmt.Sprintf("failed to generate private key: %s", err))
	}
	return &Wallet{
		PrivateKey: privateKey,
	}
}

func WalletFromPrivateKeyBase58(privateKey string) (*Wallet, error) {
	k, err := PrivateKeyFromBase58(privateKey)
	if err != nil {
		return nil, fmt.Errorf("account from private key: private key from b58: %w", err)
	}
	return &Wallet{
		PrivateKey: k,
	}, nil
}

func (a *Wallet) PublicKey() PublicKey {
	return a.PrivateKey.PublicKey()
}

type AccountMeta struct {
	PublicKey  PublicKey
	IsWritable bool
	IsSigner   bool
}

// Meta intializes a new AccountMeta with the provided pubKey.
func Meta(
	pubKey PublicKey,
) *AccountMeta {
	return &AccountMeta{
		PublicKey: pubKey,
	}
}

// WRITE sets IsWritable to true.
func (meta *AccountMeta) WRITE() *AccountMeta {
	meta.IsWritable = true
	return meta
}

// SIGNER sets IsSigner to true.
func (meta *AccountMeta) SIGNER() *AccountMeta {
	meta.IsSigner = true
	return meta
}

func NewAccountMeta(
	pubKey PublicKey,
	WRITE bool,
	SIGNER bool,
) *AccountMeta {
	return &AccountMeta{
		PublicKey:  pubKey,
		IsWritable: WRITE,
		IsSigner:   SIGNER,
	}
}

func (a AccountMeta) less(act *AccountMeta) bool {
	if a.IsSigner != act.IsSigner {
		return a.IsSigner
	}
	if a.IsWritable != act.IsWritable {
		return a.IsWritable
	}

	return bytes.Compare(a.PublicKey[:], act.PublicKey[:]) < 0
}

type AccountMetaSlice []*AccountMeta

func (slice *AccountMetaSlice) Append(account *AccountMeta) {
	*slice = append(*slice, account)
}

func (slice *AccountMetaSlice) SetAccounts(accounts []*AccountMeta) error {
	*slice = accounts
	return nil
}

func (slice AccountMetaSlice) GetAccounts() []*AccountMeta {
	out := make([]*AccountMeta, 0, len(slice))
	for i := range slice {
		if slice[i] != nil {
			out = append(out, slice[i])
		}
	}
	return out
}

// Get returns the AccountMeta at the desired index.
// If the index is not present, it returns nil.
func (slice AccountMetaSlice) Get(index int) *AccountMeta {
	if len(slice) > index {
		return slice[index]
	}
	return nil
}

// GetSigners returns the accounts that are signers.
func (slice AccountMetaSlice) GetSigners() []*AccountMeta {
	signers := make([]*AccountMeta, 0, len(slice))
	for _, ac := range slice {
		if ac.IsSigner {
			signers = append(signers, ac)
		}
	}
	return signers
}

// GetKeys returns the pubkeys of all AccountMeta.
func (slice AccountMetaSlice) GetKeys() PublicKeySlice {
	keys := make(PublicKeySlice, 0, len(slice))
	for _, ac := range slice {
		keys = append(keys, ac.PublicKey)
	}
	return keys
}

func (slice AccountMetaSlice) Len() int {
	return len(slice)
}

func (slice AccountMetaSlice) SplitFrom(index int) (AccountMetaSlice, AccountMetaSlice) {
	if index < 0 {
		panic("negative index")
	}
	if index == 0 {
		return AccountMetaSlice{}, slice
	}
	if index > len(slice)-1 {
		return slice, AccountMetaSlice{}
	}

	firstLen, secondLen := calcSplitAtLengths(len(slice), index)

	first := make(AccountMetaSlice, firstLen)
	copy(first, slice[:index])

	second := make(AccountMetaSlice, secondLen)
	copy(second, slice[index:])

	return first, second
}

func calcSplitAtLengths(total int, index int) (int, int) {
	if index == 0 {
		return 0, total
	}
	if index > total-1 {
		return total, 0
	}
	return index, total - index
}

// Account is the on-chain state of an account: lamports, raw data, owner,
// executable flag, and rent epoch.
//
// This is a port of the upstream anza-xyz/solana-sdk `account` crate's
// `Account` struct. It is the plain binary-state counterpart of rpc.Account
// (which is shaped for JSON-RPC responses): use this type when working with
// bincode-encoded account state, test fixtures, or off-chain tooling that
// mirrors validator-side types.
//
// The bincode codec below matches the Rust serde encoding of
// solana_account::Account exactly (fields in declaration order,
// little-endian integers, Vec<u8> as u64 length + bytes).
//
// The json tags are a plain debug/convenience form only; the resulting JSON
// is not compatible with the JSON-RPC shape of rpc.Account (in particular,
// Data encodes as base64 rather than the RPC data envelope).
//
// NOTE: MarshalWithEncoder/UnmarshalWithDecoder are custom-codec hooks that
// every gagliardetto/binary encoder dispatches to, so a borsh or compact-u16
// container with an Account field still emits the bincode layout for it.
// Likewise, embedding Account anonymously in another struct promotes these
// methods to the outer type, which would then (de)serialize only the Account
// fields; give such wrapper structs their own codec methods.
type Account struct {
	// Lamports in the account.
	Lamports uint64 `json:"lamports"`

	// Data held in this account.
	Data []byte `json:"data"`

	// Owner is the program that owns this account. If executable, the
	// program that loads this account.
	Owner PublicKey `json:"owner"`

	// Executable indicates whether this account's data contains a loaded
	// program (and is now read-only).
	Executable bool `json:"executable"`

	// RentEpoch is the epoch at which this account will next owe rent.
	RentEpoch uint64 `json:"rentEpoch"`
}

func (a Account) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(a.Lamports, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(uint64(len(a.Data)), binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteBytes(a.Data, false); err != nil {
		return err
	}
	if err := encoder.WriteBytes(a.Owner[:], false); err != nil {
		return err
	}
	if err := encoder.WriteBool(a.Executable); err != nil {
		return err
	}
	return encoder.WriteUint64(a.RentEpoch, binary.LittleEndian)
}

func (a *Account) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	if a.Lamports, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	dataLen, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	// Check against Remaining() before converting to int: on 32-bit
	// platforms int(dataLen) would truncate, which ReadNBytes could not
	// detect. It also yields a clear error for malformed input.
	if dataLen > uint64(decoder.Remaining()) {
		return fmt.Errorf("account data length %d exceeds remaining buffer %d", dataLen, decoder.Remaining())
	}
	if dataLen == 0 {
		// Canonical nil for empty data keeps decode(encode(x)) exact.
		a.Data = nil
	} else {
		data, err := decoder.ReadNBytes(int(dataLen))
		if err != nil {
			return err
		}
		// ReadNBytes returns a view into the decoder's buffer; clone so the
		// account neither observes later mutations of the caller's input nor
		// can corrupt it through append.
		a.Data = bytes.Clone(data)
	}
	owner, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	a.Owner = PublicKeyFromBytes(owner)
	// Read the bool byte strictly: Rust bincode rejects values other than
	// 0 and 1, while decoder.ReadBool accepts any non-zero byte as true.
	execByte, err := decoder.ReadByte()
	if err != nil {
		return err
	}
	if execByte > 1 {
		return fmt.Errorf("invalid bool byte %d for executable flag", execByte)
	}
	a.Executable = execByte == 1
	a.RentEpoch, err = decoder.ReadUint64(binary.LittleEndian)
	return err
}

// MarshalBinary bincode-encodes the account, matching the Rust serde
// encoding of solana_account::Account.
func (a Account) MarshalBinary() ([]byte, error) {
	// The output size is known exactly, so a pre-sized buffer makes the
	// encode a single allocation instead of geometric growth.
	buf := bytes.NewBuffer(make([]byte, 0, 8+8+len(a.Data)+32+1+8))
	if err := a.MarshalWithEncoder(bin.NewBinEncoder(buf)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary decodes a bincode-encoded account. Trailing bytes after
// the account are tolerated: Rust's bincode::deserialize free function is
// declared with allow_trailing_bytes(), and Agave decodes account state
// through it (e.g. Account::deserialize_data), so leniency IS the parity
// behavior.
func (a *Account) UnmarshalBinary(data []byte) error {
	return a.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeAccount decodes a bincode-encoded solana_account::Account. Trailing
// bytes are tolerated; see UnmarshalBinary.
func DecodeAccount(data []byte) (*Account, error) {
	var a Account
	if err := a.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to decode account: %w", err)
	}
	return &a, nil
}
