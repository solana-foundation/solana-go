package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_AuthorizeWithSeed(t *testing.T) {
	args := VoteAuthorizeWithSeedArgs{
		AuthorizationType:               VoteAuthorize{Kind: VoteAuthorizeVoter},
		CurrentAuthorityDerivedKeyOwner: pubkeyOf(5),
		CurrentAuthorityDerivedKeySeed:  "seed-string",
		NewAuthority:                    pubkeyOf(6),
	}
	inst := NewAuthorizeWithSeedInstruction(args, pubkeyOf(1), pubkeyOf(2))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_AuthorizeWithSeed), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	a := decoded.Impl.(*AuthorizeWithSeed)
	require.Equal(t, VoteAuthorizeVoter, a.Args.AuthorizationType.Kind)
	require.Equal(t, "seed-string", a.Args.CurrentAuthorityDerivedKeySeed)
	require.Equal(t, pubkeyOf(5), a.Args.CurrentAuthorityDerivedKeyOwner)
	require.Equal(t, pubkeyOf(6), a.Args.NewAuthority)
}

func TestRoundTrip_AuthorizeCheckedWithSeed(t *testing.T) {
	args := VoteAuthorizeCheckedWithSeedArgs{
		AuthorizationType:               VoteAuthorize{Kind: VoteAuthorizeWithdrawer},
		CurrentAuthorityDerivedKeyOwner: pubkeyOf(5),
		CurrentAuthorityDerivedKeySeed:  "another-seed",
	}
	inst := NewAuthorizeCheckedWithSeedInstruction(args, pubkeyOf(1), pubkeyOf(2), pubkeyOf(3))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_AuthorizeCheckedWithSeed), data[:4])

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	a := decoded.Impl.(*AuthorizeCheckedWithSeed)
	require.Equal(t, VoteAuthorizeWithdrawer, a.Args.AuthorizationType.Kind)
	require.Equal(t, "another-seed", a.Args.CurrentAuthorityDerivedKeySeed)
}
