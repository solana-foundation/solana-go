package computebudget

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetLoadedAccountsDataSizeLimitInstruction(t *testing.T) {
	t.Run("should reject zero bytes", func(t *testing.T) {
		_, err := NewSetLoadedAccountsDataSizeLimitInstruction(0).ValidateAndBuild()
		require.Error(t, err)
	})

	t.Run("should build loaded accounts data size limit ix", func(t *testing.T) {
		ix, err := NewSetLoadedAccountsDataSizeLimitInstruction(1000).ValidateAndBuild()
		require.NoError(t, err)

		require.Equal(t, ProgramID, ix.ProgramID())
		require.Equal(t, 0, len(ix.Accounts()))

		data, err := ix.Data()
		require.NoError(t, err)
		require.Equal(t, []byte{0x4, 0xe8, 0x3, 0x0, 0x0}, data)
	})
}
