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
	"encoding/base64"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	gojson "github.com/goccy/go-json"

	"github.com/gagliardetto/solana-go/text"
)

type MessageAddressTableLookupSlice []MessageAddressTableLookup

// NumLookups returns the number of accounts from all the lookups.
func (lookups MessageAddressTableLookupSlice) NumLookups() int {
	count := 0
	for i := range lookups {
		count += len(lookups[i].ReadonlyIndexes)
		count += len(lookups[i].WritableIndexes)
	}
	return count
}

// NumWritableLookups returns the number of writable accounts
// across all the lookups (all the address tables).
func (lookups MessageAddressTableLookupSlice) NumWritableLookups() int {
	count := 0
	for i := range lookups {
		count += len(lookups[i].WritableIndexes)
	}
	return count
}

// GetTableIDs returns the list of all address table IDs.
func (lookups MessageAddressTableLookupSlice) GetTableIDs() PublicKeySlice {
	if lookups == nil {
		return nil
	}
	ids := make(PublicKeySlice, 0)
	for _, lookup := range lookups {
		ids.UniqueAppend(lookup.AccountKey)
	}
	return ids
}

type MessageAddressTableLookup struct {
	AccountKey      PublicKey       `json:"accountKey"` // The account key of the address table.
	WritableIndexes Uint8SliceAsNum `json:"writableIndexes"`
	ReadonlyIndexes Uint8SliceAsNum `json:"readonlyIndexes"`
}

// Uint8SliceAsNum is a slice of uint8s that can be marshaled as numbers instead of a byte slice.
type Uint8SliceAsNum []uint8

// MarshalJSON implements json.Marshaler.
func (slice Uint8SliceAsNum) MarshalJSON() ([]byte, error) {
	out := make([]uint16, len(slice))
	for i, idx := range slice {
		out[i] = uint16(idx)
	}
	return json.Marshal(out)
}

type MessageVersion int

const (
	MessageVersionLegacy MessageVersion = 0 // default
	MessageVersionV0     MessageVersion = 1 // v0
)

// messageVersionPrefix is the high bit mask used to indicate a versioned message.
// If the first byte has this bit set, the message is versioned; the remaining
// 7 bits encode the version number (0 for V0, 1 for V1, etc.).
// See: https://github.com/anza-xyz/solana-sdk/blob/master/message/src/versions/mod.rs
const messageVersionPrefix = 0x80

type Message struct {
	version MessageVersion
	// List of base-58 encoded public keys used by the transaction,
	// including by the instructions and for signatures.
	// The first `message.header.numRequiredSignatures` public keys must sign the transaction.
	AccountKeys PublicKeySlice `json:"accountKeys"` // static keys; static keys + dynamic keys if after resolution (i.e. call to `ResolveLookups()`)

	// Details the account types and signatures required by the transaction.
	Header MessageHeader `json:"header"`

	// A base-58 encoded hash of a recent block in the ledger used to
	// prevent transaction duplication and to give transactions lifetimes.
	RecentBlockhash Hash `json:"recentBlockhash"`

	// List of program instructions that will be executed in sequence
	// and committed in one atomic transaction if all succeed.
	Instructions []CompiledInstruction `json:"instructions"`

	// List of address table lookups used to load additional accounts for this transaction.
	AddressTableLookups MessageAddressTableLookupSlice `json:"addressTableLookups"`

	// The actual tables that contain the list of account pubkeys.
	// NOTE: you need to fetch these from the chain, and then call `SetAddressTables`
	// before you use this transaction -- otherwise, you will get a panic.
	addressTables map[PublicKey]PublicKeySlice

	resolved bool // if true, the lookups have been resolved, and the `AccountKeys` slice contains all the accounts (static + dynamic).
}

// SetAddressTables sets the actual address tables used by this message.
// Use `mx.GetAddressTableLookups().GetTableIDs()` to get the list of all address table IDs.
// NOTE: you can call this once.
func (mx *Message) SetAddressTables(tables map[PublicKey]PublicKeySlice) error {
	if mx.addressTables != nil {
		return fmt.Errorf("address tables already set")
	}
	mx.addressTables = tables
	return nil
}

// GetAddressTables returns the actual address tables used by this message.
// NOTE: you must have called `SetAddressTable` before being able to use this method.
func (mx *Message) GetAddressTables() map[PublicKey]PublicKeySlice {
	return mx.addressTables
}

var _ bin.EncoderDecoder = &Message{}

// SetVersion sets the message version.
// This method forces the message to be encoded in the specified version.
// NOTE: if you set lookups, the version will default to V0.
func (m *Message) SetVersion(version MessageVersion) (*Message, error) {
	// check if the version is valid
	switch version {
	case MessageVersionV0, MessageVersionLegacy:
	default:
		return nil, fmt.Errorf("invalid message version: %d", version)
	}
	m.version = version
	return m, nil
}

// GetVersion returns the message version.
func (m *Message) GetVersion() MessageVersion {
	return m.version
}

// SetAddressTableLookups (re)sets the lookups used by this message.
func (mx *Message) SetAddressTableLookups(lookups []MessageAddressTableLookup) *Message {
	mx.AddressTableLookups = lookups
	mx.version = MessageVersionV0
	return mx
}

// AddAddressTableLookup adds a new lookup to the message.
func (mx *Message) AddAddressTableLookup(lookup MessageAddressTableLookup) *Message {
	mx.AddressTableLookups = append(mx.AddressTableLookups, lookup)
	mx.version = MessageVersionV0
	return mx
}

// GetAddressTableLookups returns the lookups used by this message.
func (mx *Message) GetAddressTableLookups() MessageAddressTableLookupSlice {
	return mx.AddressTableLookups
}

func (mx *Message) NumLookups() int {
	if mx.AddressTableLookups == nil {
		return 0
	}
	return mx.AddressTableLookups.NumLookups()
}

func (mx *Message) NumWritableLookups() int {
	if mx.AddressTableLookups == nil {
		return 0
	}
	return mx.AddressTableLookups.NumWritableLookups()
}

func (mx Message) MarshalJSON() ([]byte, error) {
	if mx.version == MessageVersionLegacy {
		out := struct {
			AccountKeys     PublicKeySlice        `json:"accountKeys"`
			Header          MessageHeader         `json:"header"`
			RecentBlockhash Hash                  `json:"recentBlockhash"`
			Instructions    []CompiledInstruction `json:"instructions"`
		}{
			AccountKeys:     mx.AccountKeys,
			Header:          mx.Header,
			RecentBlockhash: mx.RecentBlockhash,
			Instructions:    mx.Instructions,
		}
		return json.Marshal(out)
	}
	// Versioned message:
	lookups := mx.AddressTableLookups
	if lookups == nil {
		lookups = MessageAddressTableLookupSlice{}
	}
	out := struct {
		AccountKeys         PublicKeySlice                 `json:"accountKeys"`
		Header              MessageHeader                  `json:"header"`
		RecentBlockhash     Hash                           `json:"recentBlockhash"`
		Instructions        []CompiledInstruction          `json:"instructions"`
		AddressTableLookups MessageAddressTableLookupSlice `json:"addressTableLookups"`
	}{
		AccountKeys:         mx.AccountKeys,
		Header:              mx.Header,
		RecentBlockhash:     mx.RecentBlockhash,
		Instructions:        mx.Instructions,
		AddressTableLookups: lookups,
	}
	return json.Marshal(out)
}

// UnmarshalJSON decodes the message from JSON and determines its version.
// The Solana RPC emits `addressTableLookups` only for versioned (V0+)
// messages; its presence in the JSON is what distinguishes V0 from legacy,
// since the private `version` field has no wire representation.
func (mx *Message) UnmarshalJSON(data []byte) error {
	// Decode `addressTableLookups` via a RawMessage pointer so presence of the
	// key can be detected in a single parse. A non-nil pointer means the key
	// was present in the JSON (even if its value is `null`), which selects V0.
	aux := struct {
		AccountKeys         PublicKeySlice        `json:"accountKeys"`
		Header              MessageHeader         `json:"header"`
		RecentBlockhash     Hash                  `json:"recentBlockhash"`
		Instructions        []CompiledInstruction `json:"instructions"`
		AddressTableLookups *gojson.RawMessage    `json:"addressTableLookups"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	mx.AccountKeys = aux.AccountKeys
	mx.Header = aux.Header
	mx.RecentBlockhash = aux.RecentBlockhash
	mx.Instructions = aux.Instructions

	if aux.AddressTableLookups == nil {
		mx.version = MessageVersionLegacy
		mx.AddressTableLookups = nil
		return nil
	}
	mx.version = MessageVersionV0
	return json.Unmarshal(*aux.AddressTableLookups, &mx.AddressTableLookups)
}

func (mx *Message) EncodeToTree(txTree treeout.Branches) {
	switch mx.version {
	case MessageVersionV0:
		txTree.Child("Version: v0")
	case MessageVersionLegacy:
		txTree.Child("Version: legacy")
	default:
		txTree.Child(text.Sf("Version (unknown): %v", mx.version))
	}
	txTree.Child(text.Sf("RecentBlockhash: %s", mx.RecentBlockhash))

	txTree.Child(fmt.Sprintf("AccountKeys[len=%v]", mx.numStaticAccounts()+mx.AddressTableLookups.NumLookups())).ParentFunc(func(accountKeysBranch treeout.Branches) {
		accountKeys, err := mx.AccountMetaList()
		if err != nil {
			accountKeysBranch.Child(text.RedBG(fmt.Sprintf("AccountMetaList: %s", err)))
		} else {
			for keyIndex, key := range accountKeys {
				isFromTable := mx.IsVersioned() && keyIndex >= mx.numStaticAccounts()
				if isFromTable {
					accountKeysBranch.Child(text.Sf("%s (from Address Table Lookup)", text.ColorizeBG(key.PublicKey.String())))
				} else {
					accountKeysBranch.Child(text.ColorizeBG(key.PublicKey.String()))
				}
			}
		}
	})

	if mx.IsVersioned() {
		txTree.Child(fmt.Sprintf("AddressTableLookups[len=%v]", len(mx.AddressTableLookups))).ParentFunc(func(lookupsBranch treeout.Branches) {
			for _, lookup := range mx.AddressTableLookups {
				lookupsBranch.Child(text.Sf("%s", text.ColorizeBG(lookup.AccountKey.String()))).ParentFunc(func(lookupBranch treeout.Branches) {
					lookupBranch.Child(text.Sf("WritableIndexes: %v", lookup.WritableIndexes))
					lookupBranch.Child(text.Sf("ReadonlyIndexes: %v", lookup.ReadonlyIndexes))
				})
			}
		})
	}

	txTree.Child("Header").ParentFunc(func(message treeout.Branches) {
		mx.Header.EncodeToTree(message)
	})
}

func (header *MessageHeader) EncodeToTree(mxBranch treeout.Branches) {
	mxBranch.Child(text.Sf("NumRequiredSignatures: %v", header.NumRequiredSignatures))
	mxBranch.Child(text.Sf("NumReadonlySignedAccounts: %v", header.NumReadonlySignedAccounts))
	mxBranch.Child(text.Sf("NumReadonlyUnsignedAccounts: %v", header.NumReadonlyUnsignedAccounts))
}

func (mx *Message) MarshalBinary() ([]byte, error) {
	switch mx.version {
	case MessageVersionV0:
		return mx.MarshalV0()
	case MessageVersionLegacy:
		return mx.MarshalLegacy()
	default:
		return nil, fmt.Errorf("invalid message version: %d", mx.version)
	}
}

func (mx *Message) MarshalLegacy() ([]byte, error) {
	// Estimate buffer size: 3 (header) + compactU16 + 32*keys + 32 (blockhash) + instructions.
	estimatedSize := 3 + 3 + 32*len(mx.AccountKeys) + 32 + 3
	for i := range mx.Instructions {
		estimatedSize += 1 + 3 + len(mx.Instructions[i].Accounts) + 3 + len(mx.Instructions[i].Data)
	}

	buf := make([]byte, 0, estimatedSize)
	buf = append(buf,
		mx.Header.NumRequiredSignatures,
		mx.Header.NumReadonlySignedAccounts,
		mx.Header.NumReadonlyUnsignedAccounts,
	)

	bin.EncodeCompactU16Length(&buf, len(mx.AccountKeys))
	for _, key := range mx.AccountKeys {
		buf = append(buf, key[:]...)
	}

	buf = append(buf, mx.RecentBlockhash[:]...)

	bin.EncodeCompactU16Length(&buf, len(mx.Instructions))
	for i := range mx.Instructions {
		buf = append(buf, byte(mx.Instructions[i].ProgramIDIndex))
		bin.EncodeCompactU16Length(&buf, len(mx.Instructions[i].Accounts))
		for _, accountIdx := range mx.Instructions[i].Accounts {
			buf = append(buf, byte(accountIdx))
		}

		bin.EncodeCompactU16Length(&buf, len(mx.Instructions[i].Data))
		buf = append(buf, mx.Instructions[i].Data...)
	}
	return buf, nil
}

func (mx *Message) MarshalV0() ([]byte, error) {
	// The actual Solana version number is the Go enum value minus 1
	// (MessageVersionV0=1 maps to Solana version 0).
	// The wire prefix is messageVersionPrefix (0x80) OR'd with the version number.
	solanaVersion := byte(mx.version - 1)
	if solanaVersion > 0x7F {
		return nil, fmt.Errorf("invalid message version: %d", mx.version)
	}

	staticAccountKeys := mx.getStaticKeys()

	// Estimate buffer size: 1 (version) + 3 (header) + compactU16 + 32*keys + 32 (blockhash) + instructions + lookups.
	estimatedSize := 1 + 3 + 3 + 32*len(staticAccountKeys) + 32 + 3
	for i := range mx.Instructions {
		estimatedSize += 1 + 3 + len(mx.Instructions[i].Accounts) + 3 + len(mx.Instructions[i].Data)
	}
	estimatedSize += 3 // lookups length
	for i := range mx.AddressTableLookups {
		estimatedSize += 32 + 3 + len(mx.AddressTableLookups[i].WritableIndexes) + 3 + len(mx.AddressTableLookups[i].ReadonlyIndexes)
	}

	// Write version prefix first to avoid a full-buffer prepend copy.
	buf := make([]byte, 0, estimatedSize)
	buf = append(buf, messageVersionPrefix|solanaVersion)

	buf = append(buf,
		mx.Header.NumRequiredSignatures,
		mx.Header.NumReadonlySignedAccounts,
		mx.Header.NumReadonlyUnsignedAccounts,
	)

	// Encode only the keys that are not in the address table lookups.
	bin.EncodeCompactU16Length(&buf, len(staticAccountKeys))
	for _, key := range staticAccountKeys {
		buf = append(buf, key[:]...)
	}

	buf = append(buf, mx.RecentBlockhash[:]...)

	bin.EncodeCompactU16Length(&buf, len(mx.Instructions))
	for i := range mx.Instructions {
		buf = append(buf, byte(mx.Instructions[i].ProgramIDIndex))
		bin.EncodeCompactU16Length(&buf, len(mx.Instructions[i].Accounts))
		for _, accountIdx := range mx.Instructions[i].Accounts {
			buf = append(buf, byte(accountIdx))
		}

		bin.EncodeCompactU16Length(&buf, len(mx.Instructions[i].Data))
		buf = append(buf, mx.Instructions[i].Data...)
	}

	bin.EncodeCompactU16Length(&buf, len(mx.AddressTableLookups))
	for i := range mx.AddressTableLookups {
		// write account pubkey
		buf = append(buf, mx.AddressTableLookups[i].AccountKey[:]...)
		// write writable indexes
		bin.EncodeCompactU16Length(&buf, len(mx.AddressTableLookups[i].WritableIndexes))
		buf = append(buf, mx.AddressTableLookups[i].WritableIndexes...)
		// write readonly indexes
		bin.EncodeCompactU16Length(&buf, len(mx.AddressTableLookups[i].ReadonlyIndexes))
		buf = append(buf, mx.AddressTableLookups[i].ReadonlyIndexes...)
	}

	return buf, nil
}

func (mx Message) MarshalWithEncoder(encoder *bin.Encoder) error {
	out, err := mx.MarshalBinary()
	if err != nil {
		return err
	}
	return encoder.WriteBytes(out, false)
}

func (mx Message) ToBase64() string {
	out, _ := mx.MarshalBinary()
	return base64.StdEncoding.EncodeToString(out)
}

func (mx *Message) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	// peek first byte to determine if this is a legacy or v0 message
	versionNum, err := decoder.Peek(1)
	if err != nil {
		return err
	}
	// If the high bit (0x80) is set, this is a versioned message;
	// otherwise it is a legacy message where this byte is numRequiredSignatures.
	if versionNum[0]&messageVersionPrefix == 0 {
		mx.version = MessageVersionLegacy
	} else {
		mx.version = MessageVersionV0
	}
	switch mx.version {
	case MessageVersionV0:
		return mx.UnmarshalV0(decoder)
	case MessageVersionLegacy:
		return mx.UnmarshalLegacy(decoder)
	default:
		return fmt.Errorf("invalid message version: %d", mx.version)
	}
}

func (mx *Message) UnmarshalBase64(b64 string) error {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}
	return mx.UnmarshalWithDecoder(bin.NewBinDecoder(b))
}

// GetAddressTableLookupAccounts associates the lookups with the accounts
// in the actual address tables, and returns the accounts.
// NOTE: you need to call `SetAddressTables` before calling this method,
// so that the lookups can be associated with the accounts in the address tables.
func (mx *Message) GetAddressTableLookupAccounts() (PublicKeySlice, error) {
	err := mx.checkPreconditions()
	if err != nil {
		return nil, err
	}
	numWritable := mx.AddressTableLookups.NumWritableLookups()
	numTotal := mx.AddressTableLookups.NumLookups()

	writable := make(PublicKeySlice, 0, numWritable)
	readonly := make(PublicKeySlice, 0, numTotal-numWritable)

	for i := range mx.AddressTableLookups {
		table, ok := mx.addressTables[mx.AddressTableLookups[i].AccountKey]
		if !ok {
			return nil, fmt.Errorf("address table lookup not found for account: %s", mx.AddressTableLookups[i].AccountKey)
		}
		for _, idx := range mx.AddressTableLookups[i].WritableIndexes {
			if int(idx) >= len(table) {
				return nil, fmt.Errorf("address table lookup index out of range: %d", idx)
			}
			writable = append(writable, table[idx])
		}
		for _, idx := range mx.AddressTableLookups[i].ReadonlyIndexes {
			if int(idx) >= len(table) {
				return nil, fmt.Errorf("address table lookup index out of range: %d", idx)
			}
			readonly = append(readonly, table[idx])
		}
	}

	return append(writable, readonly...), nil
}

// ResolveLookups resolves the address table lookups,
// and appends the resolved accounts to the `message.AccountKeys` field.
// NOTE: you need to call `SetAddressTables` before calling this method.
func (mx *Message) ResolveLookups() (err error) {
	if mx.resolved {
		return nil
	}
	// add accounts from the address table lookups
	atlAccounts, err := mx.GetAddressTableLookupAccounts()
	if err != nil {
		return err
	}
	mx.AccountKeys = append(mx.AccountKeys, atlAccounts...)
	mx.resolved = true

	return nil
}

var ErrAlreadyResolved = fmt.Errorf("lookups already resolved")

// ResolveLookupsWith resolves the address table lookups with the provided writable and readonly accounts,
// assuming that the order of the accounts is correct.
func (mx *Message) ResolveLookupsWith(writable, readonly PublicKeySlice) (err error) {
	if mx.resolved {
		return ErrAlreadyResolved
	}
	mx.AccountKeys = append(mx.AccountKeys, append(writable, readonly...)...)
	mx.resolved = true
	return nil
}

func (mx *Message) IsResolved() bool {
	return mx.resolved
}

// GetAllKeys returns ALL the message's account keys (including the keys from resolved address lookup tables).
func (mx *Message) GetAllKeys() (keys PublicKeySlice, err error) {
	if mx.resolved {
		// If the message has been resolved, then the account keys have already
		// been appended to the `AccountKeys` field of the message.
		return mx.AccountKeys, nil
	}
	// If not resolved, then we need to resolve the lookups first...
	atlAccounts, err := mx.GetAddressTableLookupAccounts()
	if err != nil {
		return keys, err
	}
	// Return a new slice to avoid mutating mx.AccountKeys' backing array.
	all := make(PublicKeySlice, len(mx.AccountKeys), len(mx.AccountKeys)+len(atlAccounts))
	copy(all, mx.AccountKeys)
	return append(all, atlAccounts...), nil
}

func (mx *Message) getStaticKeys() (keys PublicKeySlice) {
	if mx.resolved {
		// If the message has been resolved, then the account keys have already
		// been appended to the `AccountKeys` field of the message.
		return mx.AccountKeys[:mx.numStaticAccounts()]
	}
	return mx.AccountKeys
}

func (mx *Message) UnmarshalV0(decoder *bin.Decoder) (err error) {
	prefix, err := decoder.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read message version prefix: %w", err)
	}
	solanaVersion := prefix & 0x7F
	if solanaVersion != 0 {
		return fmt.Errorf("unsupported message version: %d", solanaVersion)
	}
	mx.version = MessageVersion(solanaVersion + 1) // map Solana version 0 -> MessageVersionV0 (1)

	// The middle of the message is the same as the legacy message:
	err = mx.UnmarshalLegacy(decoder)
	if err != nil {
		return err
	}

	addressTableLookupsLen, err := decoder.ReadCompactU16()
	if err != nil {
		return fmt.Errorf("failed to read address table lookups length: %w", err)
	}
	if addressTableLookupsLen > 0 {
		mx.AddressTableLookups = make([]MessageAddressTableLookup, addressTableLookupsLen)
		for i := range addressTableLookupsLen {
			// read account pubkey
			_, err = decoder.Read(mx.AddressTableLookups[i].AccountKey[:])
			if err != nil {
				return fmt.Errorf("failed to read account pubkey: %w", err)
			}

			// read writable indexes
			writableIndexesLen, err := decoder.ReadCompactU16()
			if err != nil {
				return fmt.Errorf("failed to read writable indexes length: %w", err)
			}
			if writableIndexesLen > decoder.Remaining() {
				return fmt.Errorf("writable indexes length is too large: %d", writableIndexesLen)
			}
			mx.AddressTableLookups[i].WritableIndexes = make([]byte, writableIndexesLen)
			_, err = decoder.Read(mx.AddressTableLookups[i].WritableIndexes)
			if err != nil {
				return fmt.Errorf("failed to read writable indexes: %w", err)
			}

			// read readonly indexes
			readonlyIndexesLen, err := decoder.ReadCompactU16()
			if err != nil {
				return fmt.Errorf("failed to read readonly indexes length: %w", err)
			}
			if readonlyIndexesLen > decoder.Remaining() {
				return fmt.Errorf("readonly indexes length is too large: %d", readonlyIndexesLen)
			}
			mx.AddressTableLookups[i].ReadonlyIndexes = make([]byte, readonlyIndexesLen)
			_, err = decoder.Read(mx.AddressTableLookups[i].ReadonlyIndexes)
			if err != nil {
				return fmt.Errorf("failed to read readonly indexes: %w", err)
			}
		}
	}
	return nil
}

func (mx *Message) UnmarshalLegacy(decoder *bin.Decoder) (err error) {
	{
		mx.Header.NumRequiredSignatures, err = decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode mx.Header.NumRequiredSignatures: %w", err)
		}
		mx.Header.NumReadonlySignedAccounts, err = decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode mx.Header.NumReadonlySignedAccounts: %w", err)
		}
		mx.Header.NumReadonlyUnsignedAccounts, err = decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode mx.Header.NumReadonlyUnsignedAccounts: %w", err)
		}
	}
	{
		numAccountKeys, err := decoder.ReadCompactU16()
		if err != nil {
			return fmt.Errorf("unable to decode numAccountKeys: %w", err)
		}
		if numAccountKeys > decoder.Remaining()/32 {
			return fmt.Errorf("numAccountKeys %d is too large for remaining bytes %d", numAccountKeys, decoder.Remaining())
		}
		mx.AccountKeys = make(PublicKeySlice, numAccountKeys)
		for i := range numAccountKeys {
			_, err := decoder.Read(mx.AccountKeys[i][:])
			if err != nil {
				return fmt.Errorf("unable to decode mx.AccountKeys[%d]: %w", i, err)
			}
		}
	}
	{
		_, err := decoder.Read(mx.RecentBlockhash[:])
		if err != nil {
			return fmt.Errorf("unable to decode mx.RecentBlockhash: %w", err)
		}
	}
	{
		numInstructions, err := decoder.ReadCompactU16()
		if err != nil {
			return fmt.Errorf("unable to decode numInstructions: %w", err)
		}
		if numInstructions > decoder.Remaining() {
			return fmt.Errorf("numInstructions %d is greater than remaining bytes %d", numInstructions, decoder.Remaining())
		}
		mx.Instructions = make([]CompiledInstruction, numInstructions)
		for instructionIndex := range numInstructions {
			programIDIndex, err := decoder.ReadUint8()
			if err != nil {
				return fmt.Errorf("unable to decode mx.Instructions[%d].ProgramIDIndex: %w", instructionIndex, err)
			}
			mx.Instructions[instructionIndex].ProgramIDIndex = uint16(programIDIndex)

			{
				numAccounts, err := decoder.ReadCompactU16()
				if err != nil {
					return fmt.Errorf("unable to decode numAccounts for ix[%d]: %w", instructionIndex, err)
				}
				if numAccounts > decoder.Remaining() {
					return fmt.Errorf("ix[%v]: numAccounts %d is greater than remaining bytes %d", instructionIndex, numAccounts, decoder.Remaining())
				}
				mx.Instructions[instructionIndex].Accounts = make([]uint16, numAccounts)
				for i := range numAccounts {
					accountIndex, err := decoder.ReadUint8()
					if err != nil {
						return fmt.Errorf("unable to decode accountIndex for ix[%d].Accounts[%d]: %w", instructionIndex, i, err)
					}
					mx.Instructions[instructionIndex].Accounts[i] = uint16(accountIndex)
				}
			}
			{
				dataLen, err := decoder.ReadCompactU16()
				if err != nil {
					return fmt.Errorf("unable to decode dataLen for ix[%d]: %w", instructionIndex, err)
				}
				if dataLen > decoder.Remaining() {
					return fmt.Errorf("ix[%v]: dataLen %d is greater than remaining bytes %d", instructionIndex, dataLen, decoder.Remaining())
				}
				dataBytes, err := decoder.ReadBytes(dataLen)
				if err != nil {
					return fmt.Errorf("unable to decode dataBytes for ix[%d]: %w", instructionIndex, err)
				}
				mx.Instructions[instructionIndex].Data = (Base58)(dataBytes)
			}
		}
	}

	return nil
}

func (m *Message) checkPreconditions() error {
	// if this is versioned,
	// and there are > 0 lookups,
	// but the address table is empty,
	// then we can't build the account meta list:
	if m.IsVersioned() && m.AddressTableLookups.NumLookups() > 0 && len(m.addressTables) == 0 {
		return fmt.Errorf("cannot build account meta list without address tables")
	}

	return nil
}

func (m *Message) AccountMetaList() (AccountMetaSlice, error) {
	err := m.checkPreconditions()
	if err != nil {
		return nil, err
	}
	accountKeys, err := m.GetAllKeys()
	if err != nil {
		return nil, err
	}
	out := make(AccountMetaSlice, len(accountKeys))

	for i, a := range accountKeys {
		out[i] = &AccountMeta{
			PublicKey:  a,
			IsSigner:   m.accountIndexIsSigner(i),
			IsWritable: m.uncheckedAccountIndexIsWritable(i),
		}
	}

	return out, nil
}

func (m *Message) IsVersioned() bool {
	return m.version != MessageVersionLegacy
}

// Signers returns the pubkeys of all accounts that are signers.
func (m *Message) Signers() PublicKeySlice {
	numSigners := int(m.Header.NumRequiredSignatures)
	if numSigners > len(m.AccountKeys) {
		numSigners = len(m.AccountKeys)
	}
	// Signers are always the first NumRequiredSignatures keys — no need to iterate all.
	out := make(PublicKeySlice, numSigners)
	copy(out, m.AccountKeys[:numSigners])
	return out
}

// Writable returns the pubkeys of all accounts that are writable.
func (m *Message) Writable() (out PublicKeySlice, err error) {
	err = m.checkPreconditions()
	if err != nil {
		return nil, err
	}
	accountKeys, err := m.GetAllKeys()
	if err != nil {
		return nil, err
	}

	for i, a := range accountKeys {
		if m.uncheckedAccountIndexIsWritable(i) {
			out = append(out, a)
		}
	}

	return out, nil
}

// ResolveProgramIDIndex resolves the program ID index to a program ID.
// DEPRECATED: use `Program(index)` instead.
func (m *Message) ResolveProgramIDIndex(programIDIndex uint16) (PublicKey, error) {
	return m.Program(programIDIndex)
}

// Program returns the program key at the given index.
func (m *Message) Program(programIDIndex uint16) (PublicKey, error) {
	// programIDIndex always in AccountKeys
	if int(programIDIndex) < len(m.AccountKeys) {
		return m.AccountKeys[programIDIndex], nil
	}
	return PublicKey{}, fmt.Errorf("programID index not found %d", programIDIndex)
}

// Account returns the account at the given index.
func (m *Message) Account(index uint16) (PublicKey, error) {
	if int(index) < len(m.AccountKeys) {
		return m.AccountKeys[index], nil
	}
	allKeys, err := m.GetAllKeys()
	if err != nil {
		return PublicKey{}, err
	}
	if int(index) < len(allKeys) {
		return allKeys[index], nil
	}
	return PublicKey{}, fmt.Errorf("account index not found %d", index)
}

// GetAccountIndex returns the index of the given account (first occurrence of the account).
func (m *Message) GetAccountIndex(account PublicKey) (uint16, error) {
	err := m.checkPreconditions()
	if err != nil {
		return 0, err
	}
	accountKeys, err := m.GetAllKeys()
	if err != nil {
		return 0, err
	}

	for idx, a := range accountKeys {
		if a == account {
			return uint16(idx), nil
		}
	}

	return 0, fmt.Errorf("account not found: %s", account)
}

func (m *Message) HasAccount(account PublicKey) (bool, error) {
	err := m.checkPreconditions()
	if err != nil {
		return false, err
	}
	accountKeys, err := m.GetAllKeys()
	if err != nil {
		return false, err
	}

	for _, a := range accountKeys {
		if a == account {
			return true, nil
		}
	}

	return false, nil
}

func (m *Message) IsSigner(account PublicKey) bool {
	// signers always in AccountKeys
	for idx, acc := range m.AccountKeys {
		if acc == account {
			return m.accountIndexIsSigner(idx)
		}
	}
	return false
}

func (m *Message) accountIndexIsSigner(index int) bool {
	return index < int(m.Header.NumRequiredSignatures)
}

// numStaticAccounts returns the number of accounts that are always present in the
// account keys list (i.e. all the accounts that are NOT in the lookup table).
func (m *Message) numStaticAccounts() int {
	if !m.resolved {
		return len(m.AccountKeys)
	}
	return len(m.AccountKeys) - m.NumLookups()
}

// IsWritableStatic checks if the account is a writable account in the static accounts list, ignoring the accounts in the address table lookups.
func (m *Message) IsWritableStatic(account PublicKey) bool {
	// check only the static accounts (i.e. not the ones in the address table lookups); no check preconditions needed.
	accountKeys := m.getStaticKeys()
	index := -1
	for idx, acc := range accountKeys {
		if acc == account {
			index = idx
			break
		}
	}
	if index < 0 {
		return false
	}
	h := m.Header
	if index >= int(h.NumRequiredSignatures) {
		// Use int arithmetic to avoid underflow (Rust uses saturating_sub here).
		numWritableUnsigned := max(m.numStaticAccounts()-int(h.NumRequiredSignatures)-int(h.NumReadonlyUnsignedAccounts), 0)
		return index-int(h.NumRequiredSignatures) < numWritableUnsigned
	}
	return index < max(int(h.NumRequiredSignatures)-int(h.NumReadonlySignedAccounts), 0)
}

func (m *Message) IsWritable(account PublicKey) (bool, error) {
	err := m.checkPreconditions()
	if err != nil {
		return false, err
	}
	accountKeys, err := m.GetAllKeys()
	if err != nil {
		return false, err
	}

	index := -1
	for idx, acc := range accountKeys {
		if acc == account {
			index = idx
			break
		}
	}
	if index < 0 {
		return false, nil
	}
	return m.uncheckedAccountIndexIsWritable(index), nil
}

// uncheckedAccountIndexIsWritable checks if the account at the given index is writable.
// It does not check preconditions; the caller must ensure index is valid
// within the full (static + lookup) account list.
func (m *Message) uncheckedAccountIndexIsWritable(index int) bool {
	h := m.Header
	numStaticAccts := m.numStaticAccounts()

	if index >= numStaticAccts {
		// Account is in lookups: writable lookups come first.
		return index-numStaticAccts < m.AddressTableLookups.NumWritableLookups()
	} else if index >= int(h.NumRequiredSignatures) {
		// Use int arithmetic to avoid underflow (Rust uses saturating_sub here).
		numWritableUnsigned := max(numStaticAccts-int(h.NumRequiredSignatures)-int(h.NumReadonlyUnsignedAccounts), 0)
		return index-int(h.NumRequiredSignatures) < numWritableUnsigned
	}
	// Use int arithmetic to avoid uint8 underflow (Rust uses saturating_sub here).
	return index < max(int(h.NumRequiredSignatures)-int(h.NumReadonlySignedAccounts), 0)
}

func (m *Message) signerKeys() PublicKeySlice {
	return m.AccountKeys[0:m.Header.NumRequiredSignatures]
}

// ProgramIDs returns the deduplicated list of program IDs used by the message's instructions.
func (m *Message) ProgramIDs() PublicKeySlice {
	seen := make(map[PublicKey]struct{})
	out := make(PublicKeySlice, 0, len(m.Instructions))
	for i := range m.Instructions {
		idx := m.Instructions[i].ProgramIDIndex
		if int(idx) < len(m.AccountKeys) {
			key := m.AccountKeys[idx]
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				out = append(out, key)
			}
		}
	}
	return out
}

// IsInstructionAccount returns true if the given account index is referenced as
// an account (not a program) in any instruction.
func (m *Message) IsInstructionAccount(index uint16) bool {
	for i := range m.Instructions {
		for _, acctIdx := range m.Instructions[i].Accounts {
			if acctIdx == index {
				return true
			}
		}
	}
	return false
}

// ProgramPosition returns the 0-based position of the account at the given index
// among all program IDs invoked by the message. Returns (pos, true) if found,
// or (0, false) if the account is not a program ID.
func (m *Message) ProgramPosition(index uint16) (int, bool) {
	pos := 0
	seen := make(map[uint16]struct{})
	for i := range m.Instructions {
		progIdx := m.Instructions[i].ProgramIDIndex
		if _, ok := seen[progIdx]; !ok {
			seen[progIdx] = struct{}{}
			if progIdx == index {
				return pos, true
			}
			pos++
		}
	}
	return 0, false
}

// NumReadonlyAccounts returns the total number of readonly accounts (signed + unsigned)
// in the static account list.
func (m *Message) NumReadonlyAccounts() int {
	return int(m.Header.NumReadonlySignedAccounts) + int(m.Header.NumReadonlyUnsignedAccounts)
}

// GetIxSigners returns the set of account keys that are both signers and referenced
// as accounts (not program) in the instruction at the given index.
func (m *Message) GetIxSigners(ixIndex int) PublicKeySlice {
	if ixIndex < 0 || ixIndex >= len(m.Instructions) {
		return nil
	}
	ix := m.Instructions[ixIndex]
	seen := make(map[PublicKey]struct{}, len(ix.Accounts))
	var out PublicKeySlice
	for _, acctIdx := range ix.Accounts {
		if m.accountIndexIsSigner(int(acctIdx)) && int(acctIdx) < len(m.AccountKeys) {
			key := m.AccountKeys[acctIdx]
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				out = append(out, key)
			}
		}
	}
	return out
}

type MessageHeader struct {
	// The total number of signatures required to make the transaction valid.
	// The signatures must match the first `numRequiredSignatures` of `message.account_keys`.
	NumRequiredSignatures uint8 `json:"numRequiredSignatures"`

	// The last numReadonlySignedAccounts of the signed keys are read-only accounts.
	// Programs may process multiple transactions that load read-only accounts within
	// a single PoH entry, but are not permitted to credit or debit lamports or modify
	// account data.
	// Transactions targeting the same read-write account are evaluated sequentially.
	NumReadonlySignedAccounts uint8 `json:"numReadonlySignedAccounts"`

	// The last `numReadonlyUnsignedAccounts` of the unsigned keys are read-only accounts.
	NumReadonlyUnsignedAccounts uint8 `json:"numReadonlyUnsignedAccounts"`
}
