// Copyright 2021 github.com/gagliardetto
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

// Durable transaction nonces.
//
// This is a port of the upstream anza-xyz/solana-sdk `nonce` and
// `nonce-account` crates: the typed durable-nonce account state model
// (DurableNonce / NonceData / NonceState / NonceVersions) plus the off-chain
// helpers used to decode and verify a nonce account fetched from the cluster.
//
// It complements the transaction-level durable-nonce detection in this package
// (Transaction.UsesDurableNonce / GetNonceAccount) and the system program's
// nonce instruction builders (programs/system).

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
)

// durableNonceHashPrefix is the domain-separation prefix mixed into a blockhash
// when deriving a durable nonce. Durable nonces and real blockhashes live in
// separate hash domains. Mirrors DURABLE_NONCE_HASH_PREFIX in
// solana-sdk nonce/src/state.rs.
const durableNonceHashPrefix = "DURABLE_NONCE"

// NonceStateSize is the serialized size, in bytes, of a nonce account's data.
// It is constant regardless of whether the account is initialized: the runtime
// always allocates this many bytes. Mirrors solana_nonce::state::State::size().
const NonceStateSize = 80

// NonceVersion identifies the version of a nonce account's state. Only Current
// nonces have durable-nonce and blockhash domains separated and can therefore
// authorize durable-nonce transactions.
type NonceVersion uint32

const (
	NonceVersionLegacy  NonceVersion = 0
	NonceVersionCurrent NonceVersion = 1
)

// NonceStateKind identifies whether a nonce account has been initialized.
type NonceStateKind uint32

const (
	NonceStateUninitialized NonceStateKind = 0
	NonceStateInitialized   NonceStateKind = 1
)

// DurableNonce is the 32-byte value stored in an initialized nonce account and
// used as the recent_blockhash field of a durable-nonce transaction.
//
// Although it is used in the recent_blockhash slot, it is NOT a real blockhash:
// it is derived from one via DurableNonceFromBlockhash with a domain prefix.
type DurableNonce Hash

// DurableNonceFromBlockhash derives a durable nonce from a blockhash, computing
// sha256("DURABLE_NONCE" || blockhash). Mirrors DurableNonce::from_blockhash.
func DurableNonceFromBlockhash(blockhash Hash) DurableNonce {
	h := sha256.New()
	h.Write([]byte(durableNonceHashPrefix))
	h.Write(blockhash[:])
	var out DurableNonce
	copy(out[:], h.Sum(nil))
	return out
}

// AsHash returns the durable nonce as a Hash, i.e. the value to place in a
// transaction's recent_blockhash field. Mirrors DurableNonce::as_hash.
func (d DurableNonce) AsHash() Hash { return Hash(d) }

// Equals reports whether two durable nonces are equal.
func (d DurableNonce) Equals(other DurableNonce) bool { return d == other }

// IsZero reports whether the durable nonce is the zero value.
func (d DurableNonce) IsZero() bool { return d == DurableNonce{} }

func (d DurableNonce) String() string { return Hash(d).String() }

// NonceData is the initialized data of a durable transaction nonce account.
type NonceData struct {
	// Authority is the address allowed to sign transactions that consume and
	// advance this nonce.
	Authority PublicKey
	// DurableNonce is the current nonce value.
	DurableNonce DurableNonce
	// FeeCalculator records the fee in effect when the nonce was last advanced.
	FeeCalculator FeeCalculator
}

// NewNonceData builds nonce data from an authority, durable nonce and the
// lamports-per-signature fee. Mirrors Data::new.
func NewNonceData(authority PublicKey, durableNonce DurableNonce, lamportsPerSignature uint64) NonceData {
	return NonceData{
		Authority:     authority,
		DurableNonce:  durableNonce,
		FeeCalculator: FeeCalculator{LamportsPerSignature: lamportsPerSignature},
	}
}

// Blockhash returns the hash to use as a transaction's recent_blockhash when
// spending this nonce. Mirrors Data::blockhash.
func (d NonceData) Blockhash() Hash { return d.DurableNonce.AsHash() }

// LamportsPerSignature returns the per-signature fee recorded for the next
// transaction that uses this nonce. Mirrors Data::get_lamports_per_signature.
func (d NonceData) LamportsPerSignature() uint64 { return d.FeeCalculator.LamportsPerSignature }

func (d NonceData) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	if err = encoder.WriteBytes(d.Authority[:], false); err != nil {
		return err
	}
	if err = encoder.WriteBytes(d.DurableNonce[:], false); err != nil {
		return err
	}
	return d.FeeCalculator.MarshalWithEncoder(encoder)
}

func (d *NonceData) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	auth, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	d.Authority = PublicKeyFromBytes(auth)
	nonce, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(d.DurableNonce[:], nonce)
	return d.FeeCalculator.UnmarshalWithDecoder(decoder)
}

// NonceState is the state of a durable transaction nonce account: a
// discriminated union that is either Uninitialized or Initialized with NonceData.
type NonceState struct {
	Kind NonceStateKind
	// Data is meaningful only when Kind == NonceStateInitialized.
	Data NonceData
}

// NewInitializedNonceState builds an initialized nonce state.
// Mirrors State::new_initialized.
func NewInitializedNonceState(authority PublicKey, durableNonce DurableNonce, lamportsPerSignature uint64) NonceState {
	return NonceState{
		Kind: NonceStateInitialized,
		Data: NewNonceData(authority, durableNonce, lamportsPerSignature),
	}
}

// IsInitialized reports whether the nonce account has been initialized.
func (s NonceState) IsInitialized() bool { return s.Kind == NonceStateInitialized }

func (s NonceState) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	if err = encoder.WriteUint32(uint32(s.Kind), binary.LittleEndian); err != nil {
		return err
	}
	if s.Kind == NonceStateInitialized {
		return s.Data.MarshalWithEncoder(encoder)
	}
	return nil
}

func (s *NonceState) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	kind, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	switch NonceStateKind(kind) {
	case NonceStateUninitialized:
		s.Kind = NonceStateUninitialized
		s.Data = NonceData{}
		return nil
	case NonceStateInitialized:
		s.Kind = NonceStateInitialized
		return s.Data.UnmarshalWithDecoder(decoder)
	default:
		return fmt.Errorf("invalid nonce state discriminant: %d", kind)
	}
}

// NonceVersions wraps a NonceState together with its version. It is the type
// stored in a nonce account's data. Mirrors solana_nonce::versions::Versions.
type NonceVersions struct {
	Version NonceVersion
	State   NonceState
}

// NewNonceVersions wraps a state in a Current version. Mirrors Versions::new.
func NewNonceVersions(state NonceState) NonceVersions {
	return NonceVersions{Version: NonceVersionCurrent, State: state}
}

func (v NonceVersions) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	if err = encoder.WriteUint32(uint32(v.Version), binary.LittleEndian); err != nil {
		return err
	}
	return v.State.MarshalWithEncoder(encoder)
}

func (v *NonceVersions) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	ver, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	if NonceVersion(ver) > NonceVersionCurrent {
		return fmt.Errorf("invalid nonce version: %d", ver)
	}
	v.Version = NonceVersion(ver)
	return v.State.UnmarshalWithDecoder(decoder)
}

// MarshalBinary returns the bincode-serialized nonce versions (8 bytes when
// uninitialized, NonceStateSize bytes when initialized).
func (v NonceVersions) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := v.MarshalWithEncoder(bin.NewBinEncoder(buf)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary decodes nonce versions from bincode-serialized bytes,
// satisfying encoding.BinaryUnmarshaler as the counterpart to MarshalBinary.
// Like DecodeNonceVersions it tolerates trailing bytes (e.g. account padding).
func (v *NonceVersions) UnmarshalBinary(data []byte) error {
	return v.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// MarshalAccountData returns the on-chain account data for this nonce, zero-
// padded to NonceStateSize bytes the way the runtime allocates it.
func (v NonceVersions) MarshalAccountData() ([]byte, error) {
	b, err := v.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if len(b) > NonceStateSize {
		return nil, fmt.Errorf("nonce data too large: %d bytes", len(b))
	}
	out := make([]byte, NonceStateSize)
	copy(out, b)
	return out, nil
}

// DecodeNonceVersions decodes a nonce account's data into NonceVersions. Any
// trailing bytes (e.g. the zero padding present because the account is
// allocated at NonceStateSize) are ignored. The account-verification helpers
// use decodeNonceVersionsStrict instead, to match upstream bincode semantics.
func DecodeNonceVersions(data []byte) (*NonceVersions, error) {
	var v NonceVersions
	if err := v.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
		return nil, err
	}
	return &v, nil
}

// decodeNonceVersionsStrict decodes nonce versions and, mirroring upstream
// bincode::deserialize, rejects any input with bytes left over after the state.
// This is what nonce-account's verify_nonce_account / lamports_per_signature_of
// rely on: a real nonce account serializes to exactly its data length, so
// trailing bytes mean the account is not a valid nonce.
func decodeNonceVersionsStrict(data []byte) (*NonceVersions, error) {
	dec := bin.NewBinDecoder(data)
	var v NonceVersions
	if err := v.UnmarshalWithDecoder(dec); err != nil {
		return nil, err
	}
	if dec.Remaining() != 0 {
		return nil, fmt.Errorf("nonce account: %d trailing byte(s) after state", dec.Remaining())
	}
	return &v, nil
}

// VerifyRecentBlockhash checks that recentBlockhash matches this nonce and
// returns the nonce data if so. Legacy versions never verify, because durable
// nonces in the legacy blockhash domain are invalid. Mirrors
// Versions::verify_recent_blockhash.
func (v NonceVersions) VerifyRecentBlockhash(recentBlockhash Hash) (*NonceData, bool) {
	if v.Version != NonceVersionCurrent {
		return nil, false
	}
	if v.State.Kind != NonceStateInitialized {
		return nil, false
	}
	if v.State.Data.Blockhash() != recentBlockhash {
		return nil, false
	}
	data := v.State.Data
	return &data, true
}

// LamportsPerSignature returns the per-signature fee of an initialized nonce, or
// (0, false) when uninitialized.
func (v NonceVersions) LamportsPerSignature() (uint64, bool) {
	if v.State.Kind != NonceStateInitialized {
		return 0, false
	}
	return v.State.Data.LamportsPerSignature(), true
}

// Upgrade migrates a Legacy nonce into the Current durable-nonce domain. It
// returns (_, false) when there is nothing to upgrade: the nonce is already
// Current, or it is an uninitialized Legacy nonce (which is upgraded on
// initialization instead). Mirrors Versions::upgrade.
func (v NonceVersions) Upgrade() (NonceVersions, bool) {
	if v.Version != NonceVersionLegacy || v.State.Kind != NonceStateInitialized {
		return NonceVersions{}, false
	}
	data := v.State.Data
	data.DurableNonce = DurableNonceFromBlockhash(data.DurableNonce.AsHash())
	return NonceVersions{
		Version: NonceVersionCurrent,
		State:   NonceState{Kind: NonceStateInitialized, Data: data},
	}, true
}

// ErrNonceUninitialized is returned by Authorize when the nonce account has not
// been initialized.
var ErrNonceUninitialized = errors.New("nonce account is uninitialized")

// MissingRequiredSignatureError is returned by Authorize when the current nonce
// authority is not among the provided signers.
type MissingRequiredSignatureError struct {
	// Authority is the nonce account's current authority whose signature is
	// required.
	Authority PublicKey
}

func (e *MissingRequiredSignatureError) Error() string {
	return fmt.Sprintf("missing required signature for nonce authority %s", e.Authority)
}

// Authorize sets a new authority on the nonce account. The current authority
// must be present in signers. The version variant is preserved because the
// durable_nonce field is not changed here. Mirrors Versions::authorize.
func (v NonceVersions) Authorize(signers []PublicKey, newAuthority PublicKey) (NonceVersions, error) {
	if v.State.Kind != NonceStateInitialized {
		return NonceVersions{}, ErrNonceUninitialized
	}
	current := v.State.Data
	if !PublicKeySlice(signers).Contains(current.Authority) {
		return NonceVersions{}, &MissingRequiredSignatureError{Authority: current.Authority}
	}
	return NonceVersions{
		Version: v.Version,
		State: NonceState{
			Kind: NonceStateInitialized,
			Data: NewNonceData(newAuthority, current.DurableNonce, current.LamportsPerSignature()),
		},
	}, nil
}

// SystemAccountKind classifies a system-program-owned account.
type SystemAccountKind int

const (
	SystemAccountKindSystem SystemAccountKind = iota
	SystemAccountKindNonce
)

// VerifyNonceAccount decodes a system-owned nonce account and verifies that
// recentBlockhash matches its current nonce, returning the nonce data when
// valid. The account owner must be the System program. Mirrors
// nonce_account::verify_nonce_account.
func VerifyNonceAccount(owner PublicKey, data []byte, recentBlockhash Hash) (*NonceData, bool) {
	if !owner.Equals(SystemProgramID) {
		return nil, false
	}
	v, err := decodeNonceVersionsStrict(data)
	if err != nil {
		return nil, false
	}
	return v.VerifyRecentBlockhash(recentBlockhash)
}

// LamportsPerSignatureOf decodes a nonce account's data and returns the
// per-signature fee it records, or (0, false) when uninitialized or undecodable.
//
// Mirrors nonce_account::lamports_per_signature_of, which likewise does NOT
// verify the account owner; this function only takes the data bytes and cannot.
// Pass data you have already confirmed belongs to a System-owned nonce account
// (e.g. via GetSystemAccountKind or VerifyNonceAccount), or treat the result as
// advisory.
func LamportsPerSignatureOf(data []byte) (uint64, bool) {
	v, err := decodeNonceVersionsStrict(data)
	if err != nil {
		return 0, false
	}
	return v.LamportsPerSignature()
}

// GetSystemAccountKind classifies a system-program-owned account as a plain
// system account or an initialized nonce account, returning (_, false) for any
// other account (wrong owner, or data that is neither empty nor a valid
// initialized nonce). Mirrors nonce_account::get_system_account_kind.
func GetSystemAccountKind(owner PublicKey, data []byte) (SystemAccountKind, bool) {
	if !owner.Equals(SystemProgramID) {
		return 0, false
	}
	if len(data) == 0 {
		return SystemAccountKindSystem, true
	}
	if len(data) == NonceStateSize {
		v, err := decodeNonceVersionsStrict(data)
		if err != nil {
			return 0, false
		}
		if v.State.Kind == NonceStateInitialized {
			return SystemAccountKindNonce, true
		}
	}
	return 0, false
}
