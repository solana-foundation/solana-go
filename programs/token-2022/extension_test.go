package token2022

import (
	"bytes"
	"encoding/binary"
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_require "github.com/stretchr/testify/require"
)

// Helper: create a [32]byte filled with a single value.
func pubkeyOf(v byte) ag_solanago.PublicKey {
	var pk ag_solanago.PublicKey
	for i := range pk {
		pk[i] = v
	}
	return pk
}

// Helper: build expected bytes from parts.
func concat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func repeatByte(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func u64LE(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func u16LE(v uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return b
}

// optionalPubkeyBytes encodes a COption<Pubkey> for instruction data:
// 1 byte discriminator + 32 bytes.
func optionalPubkeyBytes(pk *ag_solanago.PublicKey) []byte {
	if pk == nil {
		return concat([]byte{0}, repeatByte(0, 32))
	}
	return concat([]byte{1}, pk[:])
}

// ===================================================================
// Tests ported from solana-program/token-2022 interface/src/instruction.rs
// ===================================================================

// Test vectors from the official Rust tests for token-2022-specific instructions.
// The existing instruction_test.go covers base token instructions (0-24) via fuzz.
// These tests verify exact byte-level compatibility with the Rust implementation.

func TestOfficialVector_InitializeMintCloseAuthority(t *testing.T) {
	pk10 := pubkeyOf(10)
	inst := NewInitializeMintCloseAuthorityInstruction(pk10, pubkeyOf(0))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := concat([]byte{25}, optionalPubkeyBytes(&pk10))
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_CreateNativeMint(t *testing.T) {
	inst := NewCreateNativeMintInstructionBuilder().
		SetPayerAccount(pubkeyOf(1)).
		SetNativeMintAccount(pubkeyOf(2)).
		SetSystemProgramAccount(ag_solanago.SystemProgramID)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{31}
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_InitializePermanentDelegate(t *testing.T) {
	pk11 := pubkeyOf(11)
	inst := NewInitializePermanentDelegateInstruction(pk11, pubkeyOf(0))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := concat([]byte{35}, pk11[:])
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_GetAccountDataSize_Empty(t *testing.T) {
	inst := NewGetAccountDataSizeInstruction(nil, pubkeyOf(0))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{21}
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_GetAccountDataSize_WithExtensions(t *testing.T) {
	inst := NewGetAccountDataSizeInstruction(
		[]ExtensionType{ExtensionTransferFeeConfig, ExtensionTransferFeeAmount},
		pubkeyOf(0),
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	// [21, 1, 0, 2, 0] -- tag 21, then u16 LE for each extension type
	expected := concat([]byte{21}, u16LE(1), u16LE(2))
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_UnwrapLamports_None(t *testing.T) {
	inst := NewUnwrapLamportsInstruction(nil, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3), nil)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{45, 0}
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_UnwrapLamports_Some(t *testing.T) {
	amount := uint64(1)
	inst := NewUnwrapLamportsInstruction(&amount, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3), nil)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := concat([]byte{45, 1}, u64LE(1))
	ag_require.Equal(t, expected, data)
}

// ===================================================================
// TransferFee extension instruction tests
// Ported from interface/src/extension/transfer_fee/instruction.rs
// ===================================================================

func TestOfficialVector_TransferFee_InitializeConfig(t *testing.T) {
	pk11 := pubkeyOf(11)
	inst := NewInitializeTransferFeeConfigInstruction(
		&pk11,      // authority
		nil,        // withdraw_withheld_authority = None
		111,        // basis_points
		^uint64(0), // max_fee = u64::MAX
		pubkeyOf(0),
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	// [26, 0, <authority_some>, <withdraw_none>, bps_u16, max_fee_u64]
	expected := concat(
		[]byte{26, 0}, // instruction tag 26, sub-instruction 0
		optionalPubkeyBytes(&pk11),
		optionalPubkeyBytes(nil),
		u16LE(111),
		u64LE(^uint64(0)),
	)
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_TransferFee_TransferCheckedWithFee(t *testing.T) {
	inst := NewTransferCheckedWithFeeInstruction(
		24, 24, 23,
		pubkeyOf(1), pubkeyOf(2), pubkeyOf(3), pubkeyOf(4), nil,
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := concat(
		[]byte{26, 1}, // instruction tag 26, sub-instruction 1
		u64LE(24),     // amount
		[]byte{24},    // decimals
		u64LE(23),     // fee
	)
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_TransferFee_WithdrawFromMint(t *testing.T) {
	inst := NewWithdrawWithheldTokensFromMintInstruction(
		pubkeyOf(1), pubkeyOf(2), pubkeyOf(3), nil,
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{26, 2}
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_TransferFee_HarvestToMint(t *testing.T) {
	inst := NewHarvestWithheldTokensToMintInstruction(pubkeyOf(1))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{26, 4}
	ag_require.Equal(t, expected, data)
}

func TestOfficialVector_TransferFee_SetTransferFee(t *testing.T) {
	inst := NewSetTransferFeeInstruction(
		^uint16(0), ^uint64(0),
		pubkeyOf(1), pubkeyOf(2), nil,
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := concat(
		[]byte{26, 5},
		u16LE(^uint16(0)),
		u64LE(^uint64(0)),
	)
	ag_require.Equal(t, expected, data)
}

// ===================================================================
// Extension instruction round-trip tests
// ===================================================================

func TestRoundTrip_DefaultAccountState(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		inst := NewInitializeDefaultAccountStateInstruction(AccountStateFrozen, pubkeyOf(1))
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)
		ag_require.Equal(t, byte(Instruction_DefaultAccountStateExtension), data[0])
		ag_require.Equal(t, byte(DefaultAccountState_Initialize), data[1])
		ag_require.Equal(t, byte(AccountStateFrozen), data[2])

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*DefaultAccountStateExtension)
		ag_require.Equal(t, DefaultAccountState_Initialize, ext.SubInstruction)
		ag_require.Equal(t, AccountStateFrozen, *ext.State)
	})

	t.Run("Update", func(t *testing.T) {
		inst := NewUpdateDefaultAccountStateInstruction(
			AccountStateInitialized, pubkeyOf(1), pubkeyOf(2), nil,
		)
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*DefaultAccountStateExtension)
		ag_require.Equal(t, DefaultAccountState_Update, ext.SubInstruction)
		ag_require.Equal(t, AccountStateInitialized, *ext.State)
	})
}

func TestRoundTrip_MemoTransfer(t *testing.T) {
	inst := NewEnableMemoTransferInstruction(pubkeyOf(1), pubkeyOf(2), nil)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, byte(Instruction_MemoTransferExtension), data[0])
	ag_require.Equal(t, byte(MemoTransfer_Enable), data[1])

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*MemoTransferExtension)
	ag_require.Equal(t, MemoTransfer_Enable, ext.SubInstruction)
}

func TestRoundTrip_CpiGuard(t *testing.T) {
	inst := NewDisableCpiGuardInstruction(pubkeyOf(1), pubkeyOf(2), nil)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, byte(Instruction_CpiGuardExtension), data[0])
	ag_require.Equal(t, byte(CpiGuard_Disable), data[1])

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*CpiGuardExtension)
	ag_require.Equal(t, CpiGuard_Disable, ext.SubInstruction)
}

func TestRoundTrip_InterestBearing(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		pk := pubkeyOf(5)
		inst := NewInitializeInterestBearingMintInstruction(&pk, 100, pubkeyOf(1))
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*InterestBearingMintExtension)
		ag_require.Equal(t, InterestBearingMint_Initialize, ext.SubInstruction)
		ag_require.Equal(t, &pk, ext.RateAuthority)
		ag_require.Equal(t, int16(100), ext.Rate)
	})

	t.Run("UpdateRate", func(t *testing.T) {
		inst := NewUpdateInterestRateInstruction(250, pubkeyOf(1), pubkeyOf(2), nil)
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*InterestBearingMintExtension)
		ag_require.Equal(t, InterestBearingMint_UpdateRate, ext.SubInstruction)
		ag_require.Equal(t, int16(250), ext.Rate)
	})
}

func TestRoundTrip_TransferHook(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		auth := pubkeyOf(5)
		hookProg := pubkeyOf(6)
		inst := NewInitializeTransferHookInstruction(&auth, &hookProg, pubkeyOf(1))
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*TransferHookExtension)
		ag_require.Equal(t, TransferHook_Initialize, ext.SubInstruction)
		ag_require.Equal(t, &auth, ext.Authority)
		ag_require.Equal(t, &hookProg, ext.HookProgramID)
	})

	t.Run("Update", func(t *testing.T) {
		hookProg := pubkeyOf(7)
		inst := NewUpdateTransferHookInstruction(&hookProg, pubkeyOf(1), pubkeyOf(2), nil)
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*TransferHookExtension)
		ag_require.Equal(t, TransferHook_Update, ext.SubInstruction)
		ag_require.Equal(t, &hookProg, ext.HookProgramID)
	})
}

func TestRoundTrip_MetadataPointer(t *testing.T) {
	auth := pubkeyOf(3)
	metaAddr := pubkeyOf(4)
	inst := NewInitializeMetadataPointerInstruction(&auth, &metaAddr, pubkeyOf(1))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*MetadataPointerExtension)
	ag_require.Equal(t, MetadataPointer_Initialize, ext.SubInstruction)
	ag_require.Equal(t, &auth, ext.Authority)
	ag_require.Equal(t, &metaAddr, ext.MetadataAddress)
}

func TestRoundTrip_GroupPointer(t *testing.T) {
	auth := pubkeyOf(3)
	groupAddr := pubkeyOf(4)
	inst := NewInitializeGroupPointerInstruction(&auth, &groupAddr, pubkeyOf(1))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*GroupPointerExtension)
	ag_require.Equal(t, GroupPointer_Initialize, ext.SubInstruction)
	ag_require.Equal(t, &auth, ext.Authority)
	ag_require.Equal(t, &groupAddr, ext.GroupAddress)
}

func TestRoundTrip_GroupMemberPointer(t *testing.T) {
	auth := pubkeyOf(3)
	memberAddr := pubkeyOf(4)
	inst := NewInitializeGroupMemberPointerInstruction(&auth, &memberAddr, pubkeyOf(1))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*GroupMemberPointerExtension)
	ag_require.Equal(t, GroupMemberPointer_Initialize, ext.SubInstruction)
	ag_require.Equal(t, &auth, ext.Authority)
	ag_require.Equal(t, &memberAddr, ext.MemberAddress)
}

func TestRoundTrip_Pausable(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		inst := NewInitializePausableInstruction(pubkeyOf(1))
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)
		ag_require.Equal(t, byte(Instruction_PausableExtension), data[0])
		ag_require.Equal(t, byte(Pausable_Initialize), data[1])

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*PausableExtension)
		ag_require.Equal(t, Pausable_Initialize, ext.SubInstruction)
	})

	t.Run("Pause", func(t *testing.T) {
		inst := NewPauseInstruction(pubkeyOf(1), pubkeyOf(2), nil)
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*PausableExtension)
		ag_require.Equal(t, Pausable_Pause, ext.SubInstruction)
	})

	t.Run("Resume", func(t *testing.T) {
		inst := NewResumeInstruction(pubkeyOf(1), pubkeyOf(2), nil)
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*PausableExtension)
		ag_require.Equal(t, Pausable_Resume, ext.SubInstruction)
	})
}

func TestRoundTrip_PermissionedBurn(t *testing.T) {
	inst := NewInitializePermissionedBurnInstruction(pubkeyOf(1))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, byte(Instruction_PermissionedBurnExtension), data[0])
	ag_require.Equal(t, byte(PermissionedBurn_Initialize), data[1])

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*PermissionedBurnExtension)
	ag_require.Equal(t, PermissionedBurn_Initialize, ext.SubInstruction)
}

func TestRoundTrip_InitializeNonTransferableMint(t *testing.T) {
	inst := NewInitializeNonTransferableMintInstruction(pubkeyOf(1))
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{32}
	ag_require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "InitializeNonTransferableMint", InstructionIDToName(decoded.TypeID.Uint8()))
}

func TestRoundTrip_Reallocate(t *testing.T) {
	inst := NewReallocateInstruction(
		[]ExtensionType{ExtensionTransferFeeConfig, ExtensionMemoTransfer},
		pubkeyOf(1), pubkeyOf(2), pubkeyOf(3), nil,
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)
	ag_require.Equal(t, byte(Instruction_Reallocate), data[0])

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	realloc := decoded.Impl.(*Reallocate)
	ag_require.Equal(t, []ExtensionType{ExtensionTransferFeeConfig, ExtensionMemoTransfer}, realloc.ExtensionTypes)
}

func TestRoundTrip_WithdrawExcessLamports(t *testing.T) {
	inst := NewWithdrawExcessLamportsInstruction(
		pubkeyOf(1), pubkeyOf(2), pubkeyOf(3), nil,
	)
	built := inst.Build()
	data, err := built.Data()
	ag_require.NoError(t, err)

	expected := []byte{38}
	ag_require.Equal(t, expected, data)

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "WithdrawExcessLamports", InstructionIDToName(decoded.TypeID.Uint8()))
}

func TestRoundTrip_ScaledUiAmount(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		auth := pubkeyOf(5)
		inst := NewInitializeScaledUiAmountInstruction(&auth, 1.5, pubkeyOf(1))
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*ScaledUiAmountExtension)
		ag_require.Equal(t, ScaledUiAmount_Initialize, ext.SubInstruction)
		ag_require.Equal(t, &auth, ext.Authority)
		ag_require.Equal(t, 1.5, ext.Multiplier)
	})

	t.Run("UpdateMultiplier", func(t *testing.T) {
		inst := NewUpdateScaledUiAmountMultiplierInstruction(
			2.0, 1000000, pubkeyOf(1), pubkeyOf(2), nil,
		)
		built := inst.Build()
		data, err := built.Data()
		ag_require.NoError(t, err)

		decoded, err := DecodeInstruction(nil, data)
		ag_require.NoError(t, err)
		ext := decoded.Impl.(*ScaledUiAmountExtension)
		ag_require.Equal(t, ScaledUiAmount_UpdateMultiplier, ext.SubInstruction)
		ag_require.Equal(t, 2.0, ext.Multiplier)
		ag_require.Equal(t, int64(1000000), ext.EffectiveTimestamp)
	})
}

// ===================================================================
// Account/Mint state deserialization tests
// Test vectors from interface/src/state.rs and extension/mod.rs
// ===================================================================

// TEST_MINT_SLICE from the official repo:
// mint_authority=Some([1;32]), supply=42, decimals=7, is_initialized=true, freeze_authority=Some([2;32])
var testMintSlice = []byte{
	1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 42, 0, 0, 0, 0, 0, 0, 0, 7, 1, 1, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
}

func TestOfficialVector_MintDeserialization(t *testing.T) {
	ag_require.Equal(t, 82, len(testMintSlice))

	mint := new(Mint)
	err := mint.UnmarshalWithDecoder(ag_binary.NewBinDecoder(testMintSlice))
	ag_require.NoError(t, err)

	expectedAuthority := pubkeyOf(1)
	ag_require.NotNil(t, mint.MintAuthority)
	ag_require.Equal(t, expectedAuthority, *mint.MintAuthority)
	ag_require.Equal(t, uint64(42), mint.Supply)
	ag_require.Equal(t, uint8(7), mint.Decimals)
	ag_require.True(t, mint.IsInitialized)
	expectedFreeze := pubkeyOf(2)
	ag_require.NotNil(t, mint.FreezeAuthority)
	ag_require.Equal(t, expectedFreeze, *mint.FreezeAuthority)

	// Round-trip
	buf := new(bytes.Buffer)
	err = mint.MarshalWithEncoder(ag_binary.NewBinEncoder(buf))
	ag_require.NoError(t, err)
	ag_require.Equal(t, testMintSlice, buf.Bytes())
}

// TEST_ACCOUNT_SLICE from the official repo:
// mint=[1;32], owner=[2;32], amount=3, delegate=Some([4;32]), state=Frozen(2),
// is_native=Some(5), delegated_amount=6, close_authority=Some([7;32])
var testAccountSlice = []byte{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 3, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 2, 1, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0,
	0, 6, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
}

func TestOfficialVector_AccountDeserialization(t *testing.T) {
	ag_require.Equal(t, 165, len(testAccountSlice))

	acct := new(Account)
	err := acct.UnmarshalWithDecoder(ag_binary.NewBinDecoder(testAccountSlice))
	ag_require.NoError(t, err)

	ag_require.Equal(t, pubkeyOf(1), acct.Mint)
	ag_require.Equal(t, pubkeyOf(2), acct.Owner)
	ag_require.Equal(t, uint64(3), acct.Amount)
	expectedDelegate := pubkeyOf(4)
	ag_require.NotNil(t, acct.Delegate)
	ag_require.Equal(t, expectedDelegate, *acct.Delegate)
	ag_require.Equal(t, AccountStateFrozen, acct.State)
	ag_require.NotNil(t, acct.IsNative)
	ag_require.Equal(t, uint64(5), *acct.IsNative)
	ag_require.Equal(t, uint64(6), acct.DelegatedAmount)
	expectedClose := pubkeyOf(7)
	ag_require.NotNil(t, acct.CloseAuthority)
	ag_require.Equal(t, expectedClose, *acct.CloseAuthority)

	// Round-trip
	buf := new(bytes.Buffer)
	err = acct.MarshalWithEncoder(ag_binary.NewBinEncoder(buf))
	ag_require.NoError(t, err)
	ag_require.Equal(t, testAccountSlice, buf.Bytes())
}

// ===================================================================
// Extension TLV parsing tests
// From interface/src/extension/mod.rs
// ===================================================================

func TestParseMintWithExtensions(t *testing.T) {
	// Build MINT_WITH_EXTENSION: base mint + padding to 165 + account type byte + TLV
	// 82 bytes mint + 83 bytes zero padding + 1 byte AccountType::Mint + 4 byte TLV header + 32 byte data
	mintData := make([]byte, 0, 215)
	mintData = append(mintData, testMintSlice...)
	mintData = append(mintData, make([]byte, 83)...) // zero padding to offset 165
	mintData = append(mintData, AccountTypeMint)     // account type byte at offset 165
	// TLV entry: type=3 (MintCloseAuthority), length=32
	mintData = append(mintData, 3, 0, 32, 0) // u16 LE type, u16 LE length
	mintData = append(mintData, repeatByte(1, 32)...)

	ag_require.Equal(t, 202, len(mintData))

	mint, extensions, err := ParseMintWithExtensions(mintData)
	ag_require.NoError(t, err)
	ag_require.NotNil(t, mint)
	ag_require.Equal(t, uint64(42), mint.Supply)
	ag_require.Len(t, extensions, 1)
	ag_require.Equal(t, ExtensionMintCloseAuthority, extensions[0].Type)
	ag_require.Equal(t, uint16(32), extensions[0].Length)
	ag_require.Equal(t, repeatByte(1, 32), extensions[0].Data)
}

func TestParseAccountWithExtensions(t *testing.T) {
	// Build ACCOUNT_WITH_EXTENSION: base account + AccountType::Account + TLV
	acctData := make([]byte, 0, 171)
	acctData = append(acctData, testAccountSlice...)
	acctData = append(acctData, AccountTypeAccount) // account type byte at offset 165
	// TLV entry: type=15 (TransferHookAccount), length=1
	acctData = append(acctData, 15, 0, 1, 0) // u16 LE type, u16 LE length
	acctData = append(acctData, 1)           // transferring=true

	ag_require.Equal(t, 171, len(acctData))

	acct, extensions, err := ParseAccountWithExtensions(acctData)
	ag_require.NoError(t, err)
	ag_require.NotNil(t, acct)
	ag_require.Equal(t, uint64(3), acct.Amount)
	ag_require.Equal(t, AccountStateFrozen, acct.State)
	ag_require.Len(t, extensions, 1)
	ag_require.Equal(t, ExtensionTransferHookAccount, extensions[0].Type)
	ag_require.Equal(t, uint16(1), extensions[0].Length)
	ag_require.Equal(t, []byte{1}, extensions[0].Data)
}

// ===================================================================
// TokenMetadata state serialization round-trip
// ===================================================================

func TestTokenMetadata_RoundTrip(t *testing.T) {
	meta := TokenMetadataState{
		UpdateAuthority: NewOptionalPubkey(&ag_solanago.PublicKey{1, 2, 3}),
		Mint:            pubkeyOf(10),
		Name:            "Test Token",
		Symbol:          "TST",
		Uri:             "https://example.com/metadata.json",
		AdditionalMetadata: []MetadataField{
			{Key: "description", Value: "A test token"},
			{Key: "image", Value: "https://example.com/image.png"},
		},
	}

	buf := new(bytes.Buffer)
	err := meta.MarshalWithEncoder(ag_binary.NewBinEncoder(buf))
	ag_require.NoError(t, err)

	got := new(TokenMetadataState)
	err = got.UnmarshalWithDecoder(ag_binary.NewBinDecoder(buf.Bytes()))
	ag_require.NoError(t, err)

	ag_require.Equal(t, meta.UpdateAuthority, got.UpdateAuthority)
	ag_require.Equal(t, meta.Mint, got.Mint)
	ag_require.Equal(t, meta.Name, got.Name)
	ag_require.Equal(t, meta.Symbol, got.Symbol)
	ag_require.Equal(t, meta.Uri, got.Uri)
	ag_require.Equal(t, meta.AdditionalMetadata, got.AdditionalMetadata)
}

// ===================================================================
// Instruction ID constants verification
// Ensure our iota-based constants match the official program values.
// ===================================================================

func TestInstructionIDValues(t *testing.T) {
	ag_require.Equal(t, uint8(0), Instruction_InitializeMint)
	ag_require.Equal(t, uint8(1), Instruction_InitializeAccount)
	ag_require.Equal(t, uint8(2), Instruction_InitializeMultisig)
	ag_require.Equal(t, uint8(3), Instruction_Transfer)
	ag_require.Equal(t, uint8(4), Instruction_Approve)
	ag_require.Equal(t, uint8(5), Instruction_Revoke)
	ag_require.Equal(t, uint8(6), Instruction_SetAuthority)
	ag_require.Equal(t, uint8(7), Instruction_MintTo)
	ag_require.Equal(t, uint8(8), Instruction_Burn)
	ag_require.Equal(t, uint8(9), Instruction_CloseAccount)
	ag_require.Equal(t, uint8(10), Instruction_FreezeAccount)
	ag_require.Equal(t, uint8(11), Instruction_ThawAccount)
	ag_require.Equal(t, uint8(12), Instruction_TransferChecked)
	ag_require.Equal(t, uint8(13), Instruction_ApproveChecked)
	ag_require.Equal(t, uint8(14), Instruction_MintToChecked)
	ag_require.Equal(t, uint8(15), Instruction_BurnChecked)
	ag_require.Equal(t, uint8(16), Instruction_InitializeAccount2)
	ag_require.Equal(t, uint8(17), Instruction_SyncNative)
	ag_require.Equal(t, uint8(18), Instruction_InitializeAccount3)
	ag_require.Equal(t, uint8(19), Instruction_InitializeMultisig2)
	ag_require.Equal(t, uint8(20), Instruction_InitializeMint2)
	ag_require.Equal(t, uint8(21), Instruction_GetAccountDataSize)
	ag_require.Equal(t, uint8(22), Instruction_InitializeImmutableOwner)
	ag_require.Equal(t, uint8(23), Instruction_AmountToUiAmount)
	ag_require.Equal(t, uint8(24), Instruction_UiAmountToAmount)
	ag_require.Equal(t, uint8(25), Instruction_InitializeMintCloseAuthority)
	ag_require.Equal(t, uint8(26), Instruction_TransferFeeExtension)
	ag_require.Equal(t, uint8(27), Instruction_ConfidentialTransferExtension)
	ag_require.Equal(t, uint8(28), Instruction_DefaultAccountStateExtension)
	ag_require.Equal(t, uint8(29), Instruction_Reallocate)
	ag_require.Equal(t, uint8(30), Instruction_MemoTransferExtension)
	ag_require.Equal(t, uint8(31), Instruction_CreateNativeMint)
	ag_require.Equal(t, uint8(32), Instruction_InitializeNonTransferableMint)
	ag_require.Equal(t, uint8(33), Instruction_InterestBearingMintExtension)
	ag_require.Equal(t, uint8(34), Instruction_CpiGuardExtension)
	ag_require.Equal(t, uint8(35), Instruction_InitializePermanentDelegate)
	ag_require.Equal(t, uint8(36), Instruction_TransferHookExtension)
	ag_require.Equal(t, uint8(37), Instruction_ConfidentialTransferFeeExtension)
	ag_require.Equal(t, uint8(38), Instruction_WithdrawExcessLamports)
	ag_require.Equal(t, uint8(39), Instruction_MetadataPointerExtension)
	ag_require.Equal(t, uint8(40), Instruction_GroupPointerExtension)
	ag_require.Equal(t, uint8(41), Instruction_GroupMemberPointerExtension)
	ag_require.Equal(t, uint8(42), Instruction_ConfidentialMintBurnExtension)
	ag_require.Equal(t, uint8(43), Instruction_ScaledUiAmountExtension)
	ag_require.Equal(t, uint8(44), Instruction_PausableExtension)
	ag_require.Equal(t, uint8(45), Instruction_UnwrapLamports)
	ag_require.Equal(t, uint8(46), Instruction_PermissionedBurnExtension)
}
