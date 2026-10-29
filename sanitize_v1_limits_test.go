package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// raisedV1Limits is the SIMD-0596 (Draft) address limit.
var raisedV1Limits = V1Limits{
	MaxSignatures:   MaxSignaturesV1,
	MaxAddresses:    96,
	MaxInstructions: MaxInstructionsV1,
}

// v1MessageWithAddresses returns a valid v1 message with n account addresses:
// one writable signer (the fee payer), a program at index 1, and n-2 readonly
// unsigned accounts.
func v1MessageWithAddresses(t *testing.T, n int) Message {
	t.Helper()
	msg := createTestV1Message(t)
	msg.AccountKeys = uniqueKeys(n)
	msg.Header = MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1}
	msg.Instructions[0].ProgramIDIndex = 1
	msg.Instructions[0].Accounts = []uint16{0, 2}
	return msg
}

func TestDefaultV1Limits(t *testing.T) {
	assert.Equal(t, V1Limits{MaxSignatures: 12, MaxAddresses: 64, MaxInstructions: 64}, DefaultV1Limits())
}

// Sanitize and SanitizeWithV1Limits(DefaultV1Limits()) agree on both sides of
// every default limit.
func TestV1SanitizeWithV1Limits_DefaultMatchesSanitize(t *testing.T) {
	cases := map[string]func(t *testing.T) Message{
		"64 addresses": func(t *testing.T) Message { return v1MessageWithAddresses(t, 64) },
		"65 addresses": func(t *testing.T) Message { return v1MessageWithAddresses(t, 65) },
		"12 signatures": func(t *testing.T) Message {
			msg := createTestV1Message(t)
			msg.Header.NumRequiredSignatures = 12
			msg.AccountKeys = uniqueKeys(14)
			msg.Instructions[0] = CompiledInstruction{ProgramIDIndex: 13, Accounts: []uint16{0}}
			return msg
		},
		"13 signatures": func(t *testing.T) Message {
			msg := createTestV1Message(t)
			msg.Header.NumRequiredSignatures = 13
			msg.AccountKeys = uniqueKeys(15)
			msg.Instructions[0] = CompiledInstruction{ProgramIDIndex: 14, Accounts: []uint16{0}}
			return msg
		},
		"64 instructions": func(t *testing.T) Message { return v1MessageWithInstructions(t, 64) },
		"65 instructions": func(t *testing.T) Message { return v1MessageWithInstructions(t, 65) },
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			msg := build(t)
			want := msg.Sanitize()
			got := msg.SanitizeWithV1Limits(DefaultV1Limits())
			if want == nil {
				require.NoError(t, got)
				return
			}
			require.Error(t, got)
			assert.Equal(t, want.Error(), got.Error())
			assert.True(t, IsSanitizeError(got))
		})
	}
}

func v1MessageWithInstructions(t *testing.T, n int) Message {
	t.Helper()
	msg := createTestV1Message(t)
	msg.Instructions = make([]CompiledInstruction, n)
	for i := range msg.Instructions {
		msg.Instructions[i] = CompiledInstruction{ProgramIDIndex: 1, Accounts: []uint16{0, 2}}
	}
	return msg
}

func TestV1SanitizeWithV1Limits_RaisedAddressLimit(t *testing.T) {
	msg := v1MessageWithAddresses(t, 96)
	require.NoError(t, msg.SanitizeWithV1Limits(raisedV1Limits))

	err := msg.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many addresses: 96 > max 64")

	msg = v1MessageWithAddresses(t, 97)
	err = msg.SanitizeWithV1Limits(raisedV1Limits)
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
	assert.Contains(t, err.Error(), "too many addresses: 97 > max 96")
}

func TestV1SanitizeWithV1Limits_SignatureAndInstructionLimits(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Header.NumRequiredSignatures = 2
	msg.AccountKeys = uniqueKeys(4)
	msg.Instructions[0] = CompiledInstruction{ProgramIDIndex: 3, Accounts: []uint16{0}}
	require.NoError(t, msg.SanitizeWithV1Limits(DefaultV1Limits()))

	limits := DefaultV1Limits()
	limits.MaxSignatures = 1
	err := msg.SanitizeWithV1Limits(limits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many signatures: required_signatures(2) > max 1")

	msg = v1MessageWithInstructions(t, 65)
	limits = DefaultV1Limits()
	limits.MaxInstructions = 65
	require.NoError(t, msg.SanitizeWithV1Limits(limits))
	limits.MaxInstructions = 1
	err = msg.SanitizeWithV1Limits(limits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many instructions: 65 > max 1")
}

// The zero value fails closed instead of meaning "no limit".
func TestV1SanitizeWithV1Limits_ZeroLimitsRejectEveryV1Message(t *testing.T) {
	msg := createTestV1Message(t)
	require.NoError(t, msg.Sanitize())
	err := msg.SanitizeWithV1Limits(V1Limits{})
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
}

// Raising the policy limits never relaxes the other v1 checks.
func TestV1SanitizeWithV1Limits_OtherChecksUnchanged(t *testing.T) {
	permissive := V1Limits{MaxSignatures: 255, MaxAddresses: 255, MaxInstructions: 255}

	msg := createTestV1Message(t)
	msg.AccountKeys[1] = msg.AccountKeys[0]
	err := msg.SanitizeWithV1Limits(permissive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")

	msg = createTestV1Message(t)
	msg.TransactionConfig = msg.TransactionConfig.WithHeapSize(MinHeapSizeV1 + 1)
	err = msg.SanitizeWithV1Limits(permissive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "heap size")

	msg = createTestV1Message(t)
	msg.AddAddressTableLookup(MessageAddressTableLookup{AccountKey: newUniqueKey(), WritableIndexes: []uint8{0}})
	err = msg.SanitizeWithV1Limits(permissive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "address table lookups")

	msg = createTestV1Message(t)
	msg.Header.NumReadonlySignedAccounts = 1
	err = msg.SanitizeWithV1Limits(permissive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no writable signer")

	msg = createTestV1Message(t)
	msg.Instructions[0].Accounts = []uint16{0, 99}
	err = msg.SanitizeWithV1Limits(permissive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account index 99 out of bounds")
}

func TestSanitizeWithV1Limits_LegacyAndV0IgnoreLimits(t *testing.T) {
	legacy := Message{
		Header:      MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
		AccountKeys: PublicKeySlice(uniqueKeys(70)),
		Instructions: []CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint16{0}},
		},
	}
	legacy.version = MessageVersionLegacy
	require.NoError(t, legacy.Sanitize())
	require.NoError(t, legacy.SanitizeWithV1Limits(V1Limits{}))

	v0 := Message{
		Header:      MessageHeader{NumRequiredSignatures: 1},
		AccountKeys: PublicKeySlice{newUniqueKey()},
	}
	v0.version = MessageVersionV0
	require.NoError(t, v0.Sanitize())
	require.NoError(t, v0.SanitizeWithV1Limits(V1Limits{}))

	// Invalid legacy messages stay invalid whatever the limits.
	legacy.Header.NumReadonlySignedAccounts = 1
	require.Error(t, legacy.SanitizeWithV1Limits(V1Limits{MaxSignatures: 255, MaxAddresses: 255, MaxInstructions: 255}))
}

func TestTransactionSanitizeWithV1Limits(t *testing.T) {
	tx := &Transaction{
		Message:    v1MessageWithAddresses(t, 96),
		Signatures: []Signature{{}},
	}
	err := tx.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many addresses")
	require.NoError(t, tx.SanitizeWithV1Limits(raisedV1Limits))

	// The signature-count checks still run before the message checks.
	tx.Signatures = nil
	err = tx.SanitizeWithV1Limits(raisedV1Limits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not enough signatures")

	tx.Signatures = []Signature{{}, {}}
	err = tx.SanitizeWithV1Limits(raisedV1Limits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many signatures: required 1, got 2")
}

// The wire format is not limited to 64 addresses, so a 96-address v1 message
// round-trips; only the policy limit decides whether it is valid.
func TestV1_96AddressMessageRoundTrip(t *testing.T) {
	msg := v1MessageWithAddresses(t, 96)
	raw := mustMarshalV1(t, &msg)
	back := mustUnmarshalV1(t, raw)

	assert.Equal(t, MessageVersionV1, back.GetVersion())
	assert.Equal(t, msg.AccountKeys, back.AccountKeys)
	assert.Equal(t, msg.Header, back.Header)
	assert.Equal(t, msg.Instructions, back.Instructions)
	assert.Equal(t, raw, mustMarshalV1(t, &back))

	require.Error(t, back.Sanitize())
	require.NoError(t, back.SanitizeWithV1Limits(raisedV1Limits))
}

func TestNewTransaction_V1Limits(t *testing.T) {
	payer := newUniqueKey()
	prog := newUniqueKey()
	bh := uniqueHash()
	metas := AccountMetaSlice{Meta(payer).WRITE().SIGNER()}
	for range 94 {
		metas = append(metas, Meta(newUniqueKey()))
	}
	ix := NewInstruction(prog, metas, nil)

	_, err := NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer), TransactionMessageVersion(MessageVersionV1))
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))

	tx, err := NewTransaction([]Instruction{ix}, bh,
		TransactionPayer(payer),
		TransactionMessageVersion(MessageVersionV1),
		TransactionV1Limits(raisedV1Limits),
	)
	require.NoError(t, err)
	assert.Equal(t, MessageVersionV1, tx.Message.GetVersion())
	assert.Len(t, tx.Message.AccountKeys, 96)
	require.NoError(t, tx.Message.SanitizeWithV1Limits(raisedV1Limits))

	// The option does not select v1 by itself.
	tx, err = NewTransaction([]Instruction{ix}, bh, TransactionPayer(payer), TransactionV1Limits(V1Limits{}))
	require.NoError(t, err)
	assert.Equal(t, MessageVersionLegacy, tx.Message.GetVersion())
}

// unboundedV1Limits are policy limits above every v1 wire count, so only the
// wire format bounds what SanitizeWithV1Limits accepts.
var unboundedV1Limits = V1Limits{MaxSignatures: 1 << 20, MaxAddresses: 1 << 20, MaxInstructions: 1 << 20}

// Whatever the policy limits, SanitizeWithV1Limits never accepts a v1 message
// that MarshalV1 cannot encode: v1 stores the address and instruction counts
// in one byte each.
func TestV1SanitizeWithV1Limits_WireCountsBoundEveryLimit(t *testing.T) {
	cases := []struct {
		name    string
		build   func(t *testing.T) Message
		wantErr string
	}{
		{"255 addresses", func(t *testing.T) Message { return v1MessageWithAddresses(t, 255) }, ""},
		{"256 addresses", func(t *testing.T) Message { return v1MessageWithAddresses(t, 256) }, "too many addresses: 256 > wire limit 255"},
		{"255 instructions", func(t *testing.T) Message { return v1MessageWithInstructions(t, 255) }, ""},
		{"256 instructions", func(t *testing.T) Message { return v1MessageWithInstructions(t, 256) }, "too many instructions: 256 > wire limit 255"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.build(t)
			sanitizeErr := msg.SanitizeWithV1Limits(unboundedV1Limits)
			_, marshalErr := msg.MarshalV1()
			if tc.wantErr == "" {
				require.NoError(t, sanitizeErr)
				require.NoError(t, marshalErr, "an accepted message must be encodable")
				return
			}
			require.Error(t, sanitizeErr)
			assert.True(t, IsSanitizeError(sanitizeErr))
			assert.Contains(t, sanitizeErr.Error(), tc.wantErr)
			require.Error(t, marshalErr, "the wire format cannot carry this message either")
		})
	}
}

// The default limits still report the policy error first, so the wire-count
// check never changes what Sanitize returns.
func TestV1Sanitize_DefaultErrorUnchangedPastWireCount(t *testing.T) {
	msg := v1MessageWithAddresses(t, 256)
	err := msg.Sanitize()
	require.Error(t, err)
	assert.Equal(t, "too many addresses: 256 > max 64", err.Error())

	msg = v1MessageWithInstructions(t, 256)
	err = msg.Sanitize()
	require.Error(t, err)
	assert.Equal(t, "too many instructions: 256 > max 64", err.Error())
}

// A zero field is an ordinary maximum count. V1Limits{} rejects every v1
// message, and so does a zero MaxSignatures or MaxAddresses alone, because
// every v1 message has a fee payer; a zero MaxInstructions still accepts a
// message with no instructions.
func TestV1SanitizeWithV1Limits_ZeroFieldSemantics(t *testing.T) {
	msg := createTestV1Message(t)
	msg.Instructions = nil
	require.NoError(t, msg.Sanitize())

	limits := DefaultV1Limits()
	limits.MaxInstructions = 0
	require.NoError(t, msg.SanitizeWithV1Limits(limits))
	oneInstruction := v1MessageWithInstructions(t, 1)
	require.Error(t, oneInstruction.SanitizeWithV1Limits(limits))

	limits = DefaultV1Limits()
	limits.MaxSignatures = 0
	require.Error(t, msg.SanitizeWithV1Limits(limits))

	limits = DefaultV1Limits()
	limits.MaxAddresses = 0
	require.Error(t, msg.SanitizeWithV1Limits(limits))

	require.Error(t, msg.SanitizeWithV1Limits(V1Limits{}))
}

// newV1TransferToAccounts returns an instruction that references the payer
// plus extra unique accounts, so a v1 message built from it lists
// extra+2 addresses (payer, the extra accounts, and the program).
func newV1TransferToAccounts(payer PublicKey, extra int) Instruction {
	metas := AccountMetaSlice{Meta(payer).WRITE().SIGNER()}
	for range extra {
		metas = append(metas, Meta(newUniqueKey()))
	}
	return NewInstruction(newUniqueKey(), metas, nil)
}

// NewTransaction applies the same wire-count bound as SanitizeWithV1Limits,
// so unbounded limits cannot build an unencodable v1 transaction.
func TestNewTransaction_V1LimitsCannotExceedWireCounts(t *testing.T) {
	payer := newUniqueKey()
	opts := []TransactionOption{
		TransactionPayer(payer),
		TransactionMessageVersion(MessageVersionV1),
		TransactionV1Limits(unboundedV1Limits),
	}

	tx, err := NewTransaction([]Instruction{newV1TransferToAccounts(payer, 253)}, uniqueHash(), opts...)
	require.NoError(t, err)
	require.Len(t, tx.Message.AccountKeys, 255)
	_, err = tx.Message.MarshalV1()
	require.NoError(t, err)

	_, err = NewTransaction([]Instruction{newV1TransferToAccounts(payer, 254)}, uniqueHash(), opts...)
	require.Error(t, err)
	assert.True(t, IsSanitizeError(err))
	assert.Contains(t, err.Error(), "too many addresses: 256 > wire limit 255")
}

// NewTransaction does not record the limits it was given: Sanitize keeps
// using DefaultV1Limits, so a transaction built with raised limits has to be
// checked again with the same limits.
func TestNewTransaction_V1LimitsAreNotRecorded(t *testing.T) {
	payer := newUniqueKey()
	tx, err := NewTransaction([]Instruction{newV1TransferToAccounts(payer, 94)}, uniqueHash(),
		TransactionPayer(payer),
		TransactionMessageVersion(MessageVersionV1),
		TransactionV1Limits(raisedV1Limits),
	)
	require.NoError(t, err)
	tx.Signatures = make([]Signature, tx.Message.Header.NumRequiredSignatures)

	err = tx.Sanitize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many addresses: 96 > max 64")
	require.NoError(t, tx.SanitizeWithV1Limits(raisedV1Limits))
}
