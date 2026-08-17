package solana

import (
	"errors"
	"fmt"
)

// sanitizeError represents a message or transaction validation error.
type sanitizeError struct {
	msg string
}

func (e *sanitizeError) Error() string {
	return e.msg
}

func newSanitizeError(format string, args ...any) error {
	return &sanitizeError{msg: fmt.Sprintf(format, args...)}
}

// IsSanitizeError reports whether err is a sanitization validation error.
func IsSanitizeError(err error) bool {
	var se *sanitizeError
	return errors.As(err, &se)
}

// maxAccountKeys is the maximum number of accounts a message can reference.
// Account indices are encoded as u8, so the limit is 256.
const maxAccountKeys = 256

// Sanitize validates the structural integrity of a Message.
// Ported from solana-sdk/message legacy.rs, v0/mod.rs sanitize() and v1 validate().
func (m *Message) Sanitize() error {
	switch m.version {
	case MessageVersionV1:
		return m.sanitizeV1()
	case MessageVersionV0:
		return m.sanitizeV0()
	default:
		return m.sanitizeLegacy()
	}
}

// sanitizeHeader checks the header against the number of static keys.
func (m *Message) sanitizeHeader(numStaticKeys int, keysLabel string) error {
	// Signing area and read-only non-signing area should not overlap.
	if int(m.Header.NumRequiredSignatures)+int(m.Header.NumReadonlyUnsignedAccounts) > numStaticKeys {
		return newSanitizeError("header references more accounts than available: required_signatures(%d) + readonly_unsigned(%d) > %s(%d)",
			m.Header.NumRequiredSignatures, m.Header.NumReadonlyUnsignedAccounts, keysLabel, numStaticKeys)
	}
	// There should be at least 1 RW fee-payer account.
	if m.Header.NumReadonlySignedAccounts >= m.Header.NumRequiredSignatures {
		return newSanitizeError("no writable signer: readonly_signed(%d) >= required_signatures(%d)",
			m.Header.NumReadonlySignedAccounts, m.Header.NumRequiredSignatures)
	}
	return nil
}

// sanitizeInstructions checks program and account indices against the given bounds.
func (m *Message) sanitizeInstructions(maxProgramIdx, maxAccountIdx int, programErr string) error {
	for i, ci := range m.Instructions {
		if int(ci.ProgramIDIndex) > maxProgramIdx {
			return newSanitizeError("instruction %d: program_id_index %d "+programErr, i, ci.ProgramIDIndex, maxProgramIdx)
		}
		// A program cannot be the payer.
		if ci.ProgramIDIndex == 0 {
			return newSanitizeError("instruction %d: program_id_index cannot be 0 (fee payer)", i)
		}
		for _, ai := range ci.Accounts {
			if int(ai) > maxAccountIdx {
				return newSanitizeError("instruction %d: account index %d out of bounds (max %d)", i, ai, maxAccountIdx)
			}
		}
	}
	return nil
}

func (m *Message) sanitizeLegacy() error {
	numKeys := len(m.AccountKeys)
	if err := m.sanitizeHeader(numKeys, "account_keys"); err != nil {
		return err
	}
	return m.sanitizeInstructions(numKeys-1, numKeys-1, "out of bounds (max %d)")
}

func (m *Message) sanitizeV0() error {
	numStaticKeys := len(m.AccountKeys)
	if err := m.sanitizeHeader(numStaticKeys, "static_keys"); err != nil {
		return err
	}

	// Count dynamic keys from address table lookups.
	numDynamicKeys := 0
	for _, lookup := range m.AddressTableLookups {
		numLookupIndexes := len(lookup.WritableIndexes) + len(lookup.ReadonlyIndexes)
		// Each lookup table must be used to load at least one account.
		if numLookupIndexes == 0 {
			return newSanitizeError("address table lookup for %s loads no accounts", lookup.AccountKey)
		}
		numDynamicKeys += numLookupIndexes
	}

	if numStaticKeys == 0 {
		return newSanitizeError("message has no account keys")
	}

	// The combined number of static and dynamic account keys must be <= 256
	// since account indices are encoded as u8.
	totalKeys := numStaticKeys + numDynamicKeys
	if totalKeys > maxAccountKeys {
		return newSanitizeError("total account keys %d exceeds maximum %d", totalKeys, maxAccountKeys)
	}

	// Program IDs must be in static keys only (not from lookup tables).
	return m.sanitizeInstructions(numStaticKeys-1, totalKeys-1, "exceeds static keys (max %d)")
}

// sanitizeV1 validates a V1 (SIMD-0385) message.
// Ported from solana-sdk message/src/versions/v1/message.rs validate() (+ #699 heap bounds).
func (m *Message) sanitizeV1() error {
	if len(m.AddressTableLookups) > 0 {
		return newSanitizeError("v1 messages do not support address table lookups (got %d)", len(m.AddressTableLookups))
	}
	if int(m.Header.NumRequiredSignatures) > MaxSignaturesV1 {
		return newSanitizeError("too many signatures: required_signatures(%d) > max %d", m.Header.NumRequiredSignatures, MaxSignaturesV1)
	}
	if len(m.Instructions) > MaxInstructionsV1 {
		return newSanitizeError("too many instructions: %d > max %d", len(m.Instructions), MaxInstructionsV1)
	}
	numKeys := len(m.AccountKeys)
	if numKeys > MaxAddressesV1 {
		return newSanitizeError("too many addresses: %d > max %d", numKeys, MaxAddressesV1)
	}
	if err := m.sanitizeHeader(numKeys, "account_keys"); err != nil {
		return err
	}
	if m.HasDuplicates() {
		return newSanitizeError("duplicate addresses found in message")
	}
	// Invalid mask bits only exist on the wire and are rejected by UnmarshalV1.
	if hs := m.TransactionConfig.HeapSize; hs != nil {
		if *hs%1024 != 0 {
			return newSanitizeError("heap size %d is not a multiple of 1024", *hs)
		}
		if *hs < MinHeapSizeV1 || *hs > MaxHeapSizeV1 {
			return newSanitizeError("heap size %d out of bounds [%d, %d]", *hs, MinHeapSizeV1, MaxHeapSizeV1)
		}
	}
	if numKeys == 0 {
		return newSanitizeError("message has no account keys")
	}
	for i, ci := range m.Instructions {
		if len(ci.Accounts) > 255 {
			return newSanitizeError("instruction %d: too many accounts (%d), max 255", i, len(ci.Accounts))
		}
		if len(ci.Data) > 65535 {
			return newSanitizeError("instruction %d: data too large (%d bytes), max 65535", i, len(ci.Data))
		}
	}
	return m.sanitizeInstructions(numKeys-1, numKeys-1, "out of bounds (max %d)")
}

// HasDuplicates checks if the message has duplicate account keys.
// Uses O(n^2) comparison but requires no heap allocation, which is faster
// for the typically small number of accounts in a message.
// Ported from solana-sdk/message/legacy.rs has_duplicates().
func (m *Message) HasDuplicates() bool {
	keys := m.AccountKeys
	for i := 1; i < len(keys); i++ {
		for j := i; j < len(keys); j++ {
			if keys[i-1].Equals(keys[j]) {
				return true
			}
		}
	}
	return false
}

// Sanitize validates the structural integrity of a Transaction.
// It checks that the signature count matches the message header and
// that the message itself is valid.
// Ported from solana-sdk/transaction: lib.rs and versioned/mod.rs sanitize().
func (tx *Transaction) Sanitize() error {
	numSigs := len(tx.Signatures)
	numRequired := int(tx.Message.Header.NumRequiredSignatures)
	numStaticKeys := len(tx.Message.AccountKeys)

	// Signature count must exactly match num_required_signatures.
	if numRequired > numSigs {
		return newSanitizeError("not enough signatures: required %d, got %d", numRequired, numSigs)
	}
	if numRequired < numSigs {
		return newSanitizeError("too many signatures: required %d, got %d", numRequired, numSigs)
	}

	// Signatures must not exceed static account keys count
	// (signatures are verified before lookup keys are loaded).
	if numSigs > numStaticKeys {
		return newSanitizeError("more signatures (%d) than static account keys (%d)", numSigs, numStaticKeys)
	}

	return tx.Message.Sanitize()
}

// VerifyWithResults verifies each signature independently and returns
// a per-signature boolean result.
// Ported from solana-sdk/transaction/lib.rs verify_with_results().
func (tx *Transaction) VerifyWithResults() ([]bool, error) {
	msg, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	results := make([]bool, len(tx.Signatures))
	for i, sig := range tx.Signatures {
		if i < len(tx.Message.AccountKeys) {
			results[i] = sig.Verify(tx.Message.AccountKeys[i], msg)
		}
	}
	return results, nil
}

// isAdvanceNonceInstructionData checks if the instruction data starts with
// the AdvanceNonceAccount discriminant (u32 LE value 4).
func isAdvanceNonceInstructionData(data []byte) bool {
	return len(data) >= 4 && data[0] == 4 && data[1] == 0 && data[2] == 0 && data[3] == 0
}

// nonceAdvanceInstruction returns the first instruction if it is a
// System Program AdvanceNonceAccount instruction, or nil otherwise.
func (tx *Transaction) nonceAdvanceInstruction() *CompiledInstruction {
	if len(tx.Message.Instructions) == 0 {
		return nil
	}
	ix := &tx.Message.Instructions[0]

	// Check that the program is the System Program.
	if int(ix.ProgramIDIndex) >= len(tx.Message.AccountKeys) {
		return nil
	}
	if !tx.Message.AccountKeys[ix.ProgramIDIndex].Equals(SystemProgramID) {
		return nil
	}
	if !isAdvanceNonceInstructionData(ix.Data) {
		return nil
	}
	return ix
}

// UsesDurableNonce checks whether this transaction uses a durable nonce
// by inspecting the first instruction. Returns true if the first instruction
// is a System Program AdvanceNonceAccount instruction.
// Ported from solana-sdk/transaction: uses_durable_nonce().
func (tx *Transaction) UsesDurableNonce() bool {
	return tx.nonceAdvanceInstruction() != nil
}

// GetNonceAccount returns the public key of the nonce account if this
// transaction uses a durable nonce. The nonce account is the first account
// of the first instruction (the AdvanceNonceAccount instruction).
// Returns the zero PublicKey and false if this is not a nonce transaction.
func (tx *Transaction) GetNonceAccount() (PublicKey, bool) {
	ix := tx.nonceAdvanceInstruction()
	if ix == nil {
		return PublicKey{}, false
	}
	if len(ix.Accounts) == 0 {
		return PublicKey{}, false
	}
	nonceAccountIdx := ix.Accounts[0]
	if int(nonceAccountIdx) >= len(tx.Message.AccountKeys) {
		return PublicKey{}, false
	}
	return tx.Message.AccountKeys[nonceAccountIdx], true
}
