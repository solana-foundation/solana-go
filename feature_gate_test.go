package solana

import (
	"bytes"
	"encoding/hex"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Real mainnet-beta account data of FeatureReduceSlotTimeTo350ms
// (iBRL5RuWhw4yqaAZu96RUULHckHTZAoe2b77qaV38JZ), activated at slot 440208000.
const mainnet350msFeatureHex = "01800a3d1a00000000"

func TestDecodeFeature_ActivatedMainnetVector(t *testing.T) {
	data, err := hex.DecodeString(mainnet350msFeatureHex)
	require.NoError(t, err)

	f, err := FeatureFromAccount(FeatureProgramID, data)
	require.NoError(t, err)
	require.True(t, f.IsActive())
	assert.Equal(t, uint64(440208000), *f.ActivatedAt)

	// Encode round-trip is byte-identical.
	var buf bytes.Buffer
	require.NoError(t, f.MarshalWithEncoder(bin.NewBinEncoder(&buf)))
	assert.Equal(t, data, buf.Bytes())
}

func TestDecodeFeature_Pending(t *testing.T) {
	f, err := DecodeFeature([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0})
	require.NoError(t, err)
	assert.False(t, f.IsActive())
	assert.Nil(t, f.ActivatedAt)

	var buf bytes.Buffer
	require.NoError(t, f.MarshalWithEncoder(bin.NewBinEncoder(&buf)))
	assert.Equal(t, []byte{0}, buf.Bytes())
}

func TestDecodeFeature_Invalid(t *testing.T) {
	_, err := DecodeFeature(nil)
	require.Error(t, err)

	_, err = DecodeFeature([]byte{1, 2, 3}) // tag=1 but truncated value
	require.Error(t, err)

	_, err = DecodeFeature([]byte{2, 0, 0, 0, 0, 0, 0, 0, 0}) // invalid option tag
	require.Error(t, err)

	// Trailing bytes are tolerated (bincode parity).
	f, err := DecodeFeature([]byte{1, 5, 0, 0, 0, 0, 0, 0, 0, 0xFF})
	require.NoError(t, err)
	assert.Equal(t, uint64(5), *f.ActivatedAt)

	// Wrong owner.
	_, err = FeatureFromAccount(TokenProgramID, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0})
	require.Error(t, err)

	// Owner check also enforces the minimum size (upstream from_account),
	// unlike lenient DecodeFeature which accepts a bare pending tag.
	_, err = FeatureFromAccount(FeatureProgramID, []byte{0})
	require.Error(t, err)
	_, err = DecodeFeature([]byte{0})
	require.NoError(t, err)
}

func TestFeatureSet(t *testing.T) {
	var nilSet FeatureSet
	assert.False(t, nilSet.IsActive(FeatureReduceSlotTimeTo350ms))
	_, ok := nilSet.ActivatedSlot(FeatureReduceSlotTimeTo350ms)
	assert.False(t, ok)

	fs := FeatureSet{FeatureReduceSlotTimeTo350ms: 440208000}
	assert.True(t, fs.IsActive(FeatureReduceSlotTimeTo350ms))
	slot, ok := fs.ActivatedSlot(FeatureReduceSlotTimeTo350ms)
	require.True(t, ok)
	assert.Equal(t, uint64(440208000), slot)
	assert.False(t, fs.IsActive(FeatureReduceSlotTimeTo300ms))
}
