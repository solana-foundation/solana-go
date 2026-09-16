package token2022

import (
	"bytes"
	"testing"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text"
	"github.com/stretchr/testify/require"
)

// A UiAmountToAmount instruction with no payload decodes to the empty string;
// re-encoding and pretty-printing it must not dereference a nil pointer.
func TestUiAmountToAmount_EmptyPayloadRoundTrips(t *testing.T) {
	accs := []*ag_solanago.AccountMeta{{PublicKey: ag_solanago.PublicKey{1}}}
	ix, err := DecodeInstruction(accs, []byte{Instruction_UiAmountToAmount})
	require.NoError(t, err)
	require.NotPanics(t, func() {
		data, err := ix.Data()
		require.NoError(t, err)
		require.Equal(t, []byte{Instruction_UiAmountToAmount}, data)
		enc := text.NewTreeEncoder(new(bytes.Buffer), "")
		ix.EncodeToTree(enc)
	})
}
