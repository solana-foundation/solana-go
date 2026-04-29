package vote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip_InitializeAccountV2(t *testing.T) {
	var blsPk [BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte
	var blsProof [BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE]byte
	for i := range blsPk {
		blsPk[i] = 0xAB
	}
	for i := range blsProof {
		blsProof[i] = 0xCD
	}

	init := VoteInitV2{
		NodePubkey:                          pubkeyOf(1),
		AuthorizedVoter:                     pubkeyOf(2),
		AuthorizedVoterBLSPubkey:            blsPk,
		AuthorizedVoterBLSProofOfPossession: blsProof,
		AuthorizedWithdrawer:                pubkeyOf(3),
		InflationRewardsCommissionBps:       500,
		BlockRevenueCommissionBps:           1000,
	}

	inst := NewInitializeAccountV2Instruction(init, pubkeyOf(10))
	data, err := encodeInst(inst)
	require.NoError(t, err)
	require.Equal(t, u32LE(Instruction_InitializeAccountV2), data[:4])

	// Data: 32+32+48+96+32+2+2 = 244 bytes + 4 tag
	require.Len(t, data, 4+244)

	decoded, err := DecodeInstruction(nil, data)
	require.NoError(t, err)
	ia := decoded.Impl.(*InitializeAccountV2)
	require.Equal(t, pubkeyOf(1), ia.VoteInit.NodePubkey)
	require.Equal(t, pubkeyOf(2), ia.VoteInit.AuthorizedVoter)
	require.Equal(t, blsPk, ia.VoteInit.AuthorizedVoterBLSPubkey)
	require.Equal(t, blsProof, ia.VoteInit.AuthorizedVoterBLSProofOfPossession)
	require.Equal(t, pubkeyOf(3), ia.VoteInit.AuthorizedWithdrawer)
	require.Equal(t, uint16(500), ia.VoteInit.InflationRewardsCommissionBps)
	require.Equal(t, uint16(1000), ia.VoteInit.BlockRevenueCommissionBps)
}
