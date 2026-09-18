package vote

import (
	"encoding/binary"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// The u64 element counts in vote instruction and account data come from the
// wire. Before they were bounded by the bytes actually present, a 12-byte
// UpdateVoteState payload could request a 2^60-element slice (panic) or a
// multi-GiB one (OOM) from any process decoding block instructions.
func TestDecodeInstruction_UpdateVoteStateRejectsImpossibleLockoutCount(t *testing.T) {
	accs := []*solana.AccountMeta{{PublicKey: solana.PublicKey{1}}, {PublicKey: solana.PublicKey{2}}}
	for _, count := range []uint64{1 << 60, 1 << 27, 2} {
		data := binary.LittleEndian.AppendUint32(nil, 8) // UpdateVoteState
		data = binary.LittleEndian.AppendUint64(data, count)
		require.NotPanics(t, func() {
			_, err := DecodeInstruction(accs, data)
			require.Error(t, err, "count %d with no lockout bytes", count)
		})
	}
}

func TestVoteStateDecodersRejectImpossibleCounts(t *testing.T) {
	huge := binary.LittleEndian.AppendUint64(nil, 1<<60)
	require.NotPanics(t, func() {
		var av AuthorizedVoters
		require.Error(t, av.UnmarshalWithDecoder(bin.NewBinDecoder(huge)))
	})
	require.NotPanics(t, func() {
		_, err := decodeEpochCredits(bin.NewBinDecoder(huge))
		require.Error(t, err)
	})
}
