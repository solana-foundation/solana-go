package token2022

import (
	"encoding/binary"
	"math"
	"testing"

	ag_require "github.com/stretchr/testify/require"
)

// Extension instructions that take an optional address encode it as
// MaybeNull<Address>: 32 bytes, all zeros meaning none, with no option flag.
// The expected bytes follow the instruction data structs in
// solana-program/token-2022 interface/src/extension/*/instruction.rs.

func dataOf(t *testing.T, inst interface{ Build() *Instruction }) []byte {
	t.Helper()
	data, err := inst.Build().Data()
	ag_require.NoError(t, err)
	return data
}

func TestVector_MaybeNullPubkeyInstructions(t *testing.T) {
	mint := pubkeyOf(1)
	owner := pubkeyOf(2)
	auth := pubkeyOf(3)
	addr := pubkeyOf(4)

	basisPoints := int16(-25)
	rate := make([]byte, 2)
	binary.LittleEndian.PutUint16(rate, uint16(basisPoints))
	multiplier := make([]byte, 8)
	binary.LittleEndian.PutUint64(multiplier, math.Float64bits(1.5))

	cases := []struct {
		name     string
		inst     interface{ Build() *Instruction }
		expected []byte
	}{
		{"MetadataPointer.Initialize", NewInitializeMetadataPointerInstruction(&auth, &addr, mint),
			concat([]byte{39, 0}, auth[:], addr[:])},
		{"MetadataPointer.Initialize without authority", NewInitializeMetadataPointerInstruction(nil, &addr, mint),
			concat([]byte{39, 0}, maybeNullPubkeyBytes(nil), addr[:])},
		{"MetadataPointer.Update", NewUpdateMetadataPointerInstruction(&addr, mint, owner, nil),
			concat([]byte{39, 1}, addr[:])},
		{"TransferHook.Initialize", NewInitializeTransferHookInstruction(&auth, &addr, mint),
			concat([]byte{36, 0}, auth[:], addr[:])},
		{"TransferHook.Update to none", NewUpdateTransferHookInstruction(nil, mint, owner, nil),
			concat([]byte{36, 1}, maybeNullPubkeyBytes(nil))},
		{"GroupPointer.Initialize", NewInitializeGroupPointerInstruction(&auth, &addr, mint),
			concat([]byte{40, 0}, auth[:], addr[:])},
		{"GroupPointer.Update", NewUpdateGroupPointerInstruction(&addr, mint, owner, nil),
			concat([]byte{40, 1}, addr[:])},
		{"GroupMemberPointer.Initialize", NewInitializeGroupMemberPointerInstruction(&auth, &addr, mint),
			concat([]byte{41, 0}, auth[:], addr[:])},
		{"GroupMemberPointer.Update", NewUpdateGroupMemberPointerInstruction(&addr, mint, owner, nil),
			concat([]byte{41, 1}, addr[:])},
		{"InterestBearingMint.Initialize", NewInitializeInterestBearingMintInstruction(&auth, basisPoints, mint),
			concat([]byte{33, 0}, auth[:], rate)},
		{"ScaledUiAmount.Initialize", NewInitializeScaledUiAmountInstruction(&auth, 1.5, mint),
			concat([]byte{43, 0}, auth[:], multiplier)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ag_require.Equal(t, c.expected, dataOf(t, c.inst))
		})
	}
}

func TestRoundTrip_MaybeNullPubkeyNone(t *testing.T) {
	addr := pubkeyOf(4)
	decoded, err := DecodeInstruction(nil, dataOf(t, NewInitializeMetadataPointerInstruction(nil, &addr, pubkeyOf(1))))
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*MetadataPointerExtension)
	ag_require.Nil(t, ext.Authority, "all zeros decodes as none")
	ag_require.Equal(t, &addr, ext.MetadataAddress)
}

func TestRoundTrip_TransferFeeConfigWithoutAuthorities(t *testing.T) {
	data := dataOf(t, NewInitializeTransferFeeConfigInstruction(nil, nil, 111, 5, pubkeyOf(1)))
	ag_require.Equal(t, concat([]byte{26, 0}, []byte{0}, []byte{0}, u16LE(111), u64LE(5)), data)

	decoded, err := DecodeInstruction(nil, data)
	ag_require.NoError(t, err)
	ext := decoded.Impl.(*TransferFeeExtension)
	ag_require.Nil(t, ext.TransferFeeConfigAuthority)
	ag_require.Nil(t, ext.WithdrawWithheldAuthority)
	ag_require.Equal(t, uint16(111), ext.TransferFeeBasisPoints)
	ag_require.Equal(t, uint64(5), ext.MaximumFee)
}
