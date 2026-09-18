package addresslookuptable

import (
	"encoding/binary"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

// The address count in ExtendLookupTable data comes from the wire. Before it
// was bounded by the bytes actually present, a 12-byte payload could request
// a 2^60-address slice (panic) or a 4 GiB one (OOM) from any process decoding
// block instructions; failed transactions are included in blocks, so anyone
// can put such a payload in front of an indexer.
func TestDecodeInstruction_ExtendLookupTableRejectsImpossibleCount(t *testing.T) {
	accs := make([]*solana.AccountMeta, 4)
	for i := range accs {
		accs[i] = &solana.AccountMeta{PublicKey: solana.PublicKey{byte(i + 1)}}
	}
	for _, count := range []uint64{1 << 60, 1 << 27, 1} {
		data := binary.LittleEndian.AppendUint32(nil, 2) // ExtendLookupTable
		data = binary.LittleEndian.AppendUint64(data, count)
		require.NotPanics(t, func() {
			_, err := DecodeInstruction(accs, data)
			require.Error(t, err, "count %d with no address bytes", count)
		})
	}
}
