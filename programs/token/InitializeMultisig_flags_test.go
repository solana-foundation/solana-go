package token

import (
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// The on-chain SPL Token InitializeMultisig / InitializeMultisig2 account spec
// marks the signer (member) accounts as read-only, non-signer:
//
//	InitializeMultisig
//	  0.   [WRITE] the multisignature account to initialize
//	  1.   []      rent sysvar
//	  2+N. []      the signer (member) accounts
//
//	InitializeMultisig2
//	  0.   [WRITE] the multisignature account to initialize
//	  1+N. []      the signer (member) accounts
//
// The member pubkeys are only recorded into the multisig account data; the
// instruction itself requires no signers, so the members must NOT carry the
// signer flag. Marking them as signers would force every prospective member to
// sign the setup transaction, which the program does not require.
func TestAccountFlags_InitializeMultisig(t *testing.T) {
	inst := NewInitializeMultisigInstructionBuilder().
		SetM(2).
		SetAccount(ag_solanago.PublicKey{1}).
		SetSysVarRentPubkeyAccount(ag_solanago.SysVarRentPubkey).
		AddSigners(ag_solanago.PublicKey{2}, ag_solanago.PublicKey{3})

	account := inst.GetAccount()
	require.True(t, account.IsWritable, "multisig account must be writable")
	require.False(t, account.IsSigner, "multisig account must not be a signer")

	rent := inst.GetSysVarRentPubkeyAccount()
	require.False(t, rent.IsWritable, "rent sysvar must not be writable")
	require.False(t, rent.IsSigner, "rent sysvar must not be a signer")

	require.Len(t, inst.Signers, 2)
	for i, s := range inst.Signers {
		require.False(t, s.IsSigner, "member account %d must not be a signer", i)
		require.False(t, s.IsWritable, "member account %d must not be writable", i)
	}
}

func TestAccountFlags_InitializeMultisig2(t *testing.T) {
	inst := NewInitializeMultisig2InstructionBuilder().
		SetM(2).
		SetAccount(ag_solanago.PublicKey{1}).
		AddSigners(ag_solanago.PublicKey{2}, ag_solanago.PublicKey{3})

	account := inst.GetAccount()
	require.True(t, account.IsWritable, "multisig account must be writable")
	require.False(t, account.IsSigner, "multisig account must not be a signer")

	require.Len(t, inst.Signers, 2)
	for i, s := range inst.Signers {
		require.False(t, s.IsSigner, "member account %d must not be a signer", i)
		require.False(t, s.IsWritable, "member account %d must not be writable", i)
	}
}
