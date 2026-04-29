package vote

import (
	"bytes"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/require"
)

// --- short_vec length encoding ---

func TestCompactU16Len_OneByte(t *testing.T) {
	cases := []int{0, 1, 0x7f}
	for _, n := range cases {
		buf := new(bytes.Buffer)
		require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), n))
		require.Equal(t, []byte{byte(n)}, buf.Bytes(), "n=%d", n)

		dec := bin.NewBinDecoder(buf.Bytes())
		got, err := dec.ReadCompactU16Length()
		require.NoError(t, err)
		require.Equal(t, n, got)
	}
}

func TestCompactU16Len_TwoBytes(t *testing.T) {
	// 0x80 encodes as [0x80, 0x01]
	buf := new(bytes.Buffer)
	require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), 0x80))
	require.Equal(t, []byte{0x80, 0x01}, buf.Bytes())

	// 0x3fff encodes as [0xff, 0x7f]
	buf.Reset()
	require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), 0x3fff))
	require.Equal(t, []byte{0xff, 0x7f}, buf.Bytes())

	// Round-trip both
	for _, n := range []int{0x80, 0x3fff} {
		buf := new(bytes.Buffer)
		require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), n))
		got, err := bin.NewBinDecoder(buf.Bytes()).ReadCompactU16Length()
		require.NoError(t, err)
		require.Equal(t, n, got)
	}
}

func TestCompactU16Len_ThreeBytes(t *testing.T) {
	// 0x4000 encodes as [0x80, 0x80, 0x01]
	buf := new(bytes.Buffer)
	require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), 0x4000))
	require.Equal(t, []byte{0x80, 0x80, 0x01}, buf.Bytes())

	// 0xffff encodes as [0xff, 0xff, 0x03]
	buf.Reset()
	require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), 0xffff))
	require.Equal(t, []byte{0xff, 0xff, 0x03}, buf.Bytes())

	// Round-trip
	for _, n := range []int{0x4000, 0xffff} {
		buf := new(bytes.Buffer)
		require.NoError(t, writeCompactU16Len(bin.NewBinEncoder(buf), n))
		got, err := bin.NewBinDecoder(buf.Bytes()).ReadCompactU16Length()
		require.NoError(t, err)
		require.Equal(t, n, got)
	}
}

// --- varint u64 encoding ---

func TestVarintU64_Zero(t *testing.T) {
	buf := new(bytes.Buffer)
	require.NoError(t, writeVarintU64(bin.NewBinEncoder(buf), 0))
	require.Equal(t, []byte{0x00}, buf.Bytes())

	got, err := readVarintU64(bin.NewBinDecoder(buf.Bytes()))
	require.NoError(t, err)
	require.Equal(t, uint64(0), got)
}

func TestVarintU64_Small(t *testing.T) {
	// 0x7f is the largest single-byte varint
	buf := new(bytes.Buffer)
	require.NoError(t, writeVarintU64(bin.NewBinEncoder(buf), 0x7f))
	require.Equal(t, []byte{0x7f}, buf.Bytes())
}

func TestVarintU64_TwoBytes(t *testing.T) {
	// 0x80 -> [0x80, 0x01]
	buf := new(bytes.Buffer)
	require.NoError(t, writeVarintU64(bin.NewBinEncoder(buf), 0x80))
	require.Equal(t, []byte{0x80, 0x01}, buf.Bytes())

	got, err := readVarintU64(bin.NewBinDecoder(buf.Bytes()))
	require.NoError(t, err)
	require.Equal(t, uint64(0x80), got)
}

func TestVarintU64_MaxU64(t *testing.T) {
	buf := new(bytes.Buffer)
	require.NoError(t, writeVarintU64(bin.NewBinEncoder(buf), ^uint64(0)))
	// Round-trip
	got, err := readVarintU64(bin.NewBinDecoder(buf.Bytes()))
	require.NoError(t, err)
	require.Equal(t, ^uint64(0), got)
}

func TestVarintU64_RoundTripRange(t *testing.T) {
	values := []uint64{
		0, 1, 127, 128, 255, 16383, 16384, 1 << 20, 1 << 30, 1 << 62, ^uint64(0) - 1,
	}
	for _, v := range values {
		buf := new(bytes.Buffer)
		require.NoError(t, writeVarintU64(bin.NewBinEncoder(buf), v))
		got, err := readVarintU64(bin.NewBinDecoder(buf.Bytes()))
		require.NoError(t, err)
		require.Equal(t, v, got, "v=%d", v)
	}
}

// --- CompactVoteStateUpdate wire format ---

// Verifies the delta encoding: given lockouts at absolute slots [100, 101, 105]
// with root=99, the encoded offsets should be [1, 1, 4].
func TestCompactVoteStateUpdate_DeltaOffsets(t *testing.T) {
	root := uint64(99)
	update := &VoteStateUpdate{
		Lockouts: []Lockout{
			{Slot: 100, ConfirmationCount: 3},
			{Slot: 101, ConfirmationCount: 2},
			{Slot: 105, ConfirmationCount: 1},
		},
		Root: &root,
		Hash: hashOf(0x00),
	}
	buf := new(bytes.Buffer)
	require.NoError(t, marshalCompactVoteStateUpdate(bin.NewBinEncoder(buf), update))

	// Expected layout:
	//   root:          u64 LE  = 99, 00, 00, 00, 00, 00, 00, 00 (8 bytes)
	//   lockout_count: u8      = 3 (fits in one byte)
	//   offset(1):     varint  = 0x01
	//   conf_count:    u8      = 3
	//   offset(1):     varint  = 0x01
	//   conf_count:    u8      = 2
	//   offset(4):     varint  = 0x04
	//   conf_count:    u8      = 1
	//   hash:          32 bytes of 0x00
	//   timestamp:     u8      = 0 (None)
	expected := []byte{
		99, 0, 0, 0, 0, 0, 0, 0, // root
		3,    // short_vec length
		1, 3, // offset 1, confirmation 3
		1, 2, // offset 1, confirmation 2
		4, 1, // offset 4, confirmation 1
	}
	expected = append(expected, make([]byte, 32)...) // hash
	expected = append(expected, 0)                   // timestamp None

	require.Equal(t, expected, buf.Bytes())

	// Round-trip reconstructs absolute slots
	decoded := new(VoteStateUpdate)
	require.NoError(t, unmarshalCompactVoteStateUpdate(bin.NewBinDecoder(buf.Bytes()), decoded))
	require.Len(t, decoded.Lockouts, 3)
	require.Equal(t, uint64(100), decoded.Lockouts[0].Slot)
	require.Equal(t, uint64(101), decoded.Lockouts[1].Slot)
	require.Equal(t, uint64(105), decoded.Lockouts[2].Slot)
	require.NotNil(t, decoded.Root)
	require.Equal(t, uint64(99), *decoded.Root)
}

func TestCompactVoteStateUpdate_NoRoot(t *testing.T) {
	update := &VoteStateUpdate{
		Lockouts: []Lockout{},
		Root:     nil,
		Hash:     hashOf(0x00),
	}
	buf := new(bytes.Buffer)
	require.NoError(t, marshalCompactVoteStateUpdate(bin.NewBinEncoder(buf), update))

	// Root should be u64::MAX when nil
	expected := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0}
	expected = append(expected, make([]byte, 32)...)
	expected = append(expected, 0)
	require.Equal(t, expected, buf.Bytes())

	decoded := new(VoteStateUpdate)
	require.NoError(t, unmarshalCompactVoteStateUpdate(bin.NewBinDecoder(buf.Bytes()), decoded))
	require.Nil(t, decoded.Root)
	require.Empty(t, decoded.Lockouts)
}

// TestTowerSync_CompactWireFormat verifies that TowerSync's wire format
// differs from the non-compact VoteStateUpdate wire format (smaller) and
// that it includes the trailing block_id.
func TestTowerSync_CompactWireFormat(t *testing.T) {
	root := uint64(99)
	sync := TowerSyncUpdate{
		Lockouts: []Lockout{
			{Slot: 100, ConfirmationCount: 3},
			{Slot: 101, ConfirmationCount: 2},
		},
		Root:    &root,
		Hash:    hashOf(0xAA),
		BlockID: hashOf(0xBB),
	}
	buf := new(bytes.Buffer)
	require.NoError(t, sync.MarshalWithEncoder(bin.NewBinEncoder(buf)))

	// Layout:
	//   u64 root = 99              (8 bytes)
	//   u8  lockout_count = 2      (1 byte)
	//   varint 1, u8 3             (2 bytes)
	//   varint 1, u8 2             (2 bytes)
	//   hash (32 bytes of 0xAA)
	//   u8  timestamp None         (1 byte)
	//   block_id (32 bytes of 0xBB)
	// Total: 8 + 1 + 2 + 2 + 32 + 1 + 32 = 78 bytes
	require.Len(t, buf.Bytes(), 78)

	// Last 32 bytes should be block_id
	blockID := hashOf(0xBB)
	require.Equal(t, blockID[:], buf.Bytes()[len(buf.Bytes())-32:])

	// Round-trip
	decoded := new(TowerSyncUpdate)
	require.NoError(t, decoded.UnmarshalWithDecoder(bin.NewBinDecoder(buf.Bytes())))
	require.Len(t, decoded.Lockouts, 2)
	require.Equal(t, uint64(100), decoded.Lockouts[0].Slot)
	require.Equal(t, uint64(101), decoded.Lockouts[1].Slot)
	require.Equal(t, hashOf(0xAA), decoded.Hash)
	require.Equal(t, hashOf(0xBB), decoded.BlockID)
}

// --- nil marshal safety ---

func TestMarshal_NilFields_ReturnError(t *testing.T) {
	// Each of these has a nil required field — MarshalWithEncoder should
	// return an error instead of panicking.
	cases := []struct {
		name string
		inst interface{ Build() *Instruction }
	}{
		{"InitializeAccount", &InitializeAccount{}},
		{"Authorize", &Authorize{}},
		{"UpdateVoteState", &UpdateVoteState{}},
		{"UpdateVoteStateSwitch", &UpdateVoteStateSwitch{}},
		{"CompactUpdateVoteState", &CompactUpdateVoteState{}},
		{"CompactUpdateVoteStateSwitch", &CompactUpdateVoteStateSwitch{}},
		{"TowerSync", &TowerSync{}},
		{"TowerSyncSwitch", &TowerSyncSwitch{}},
		{"AuthorizeWithSeed", &AuthorizeWithSeed{}},
		{"AuthorizeCheckedWithSeed", &AuthorizeCheckedWithSeed{}},
		{"InitializeAccountV2", &InitializeAccountV2{}},
		{"UpdateCommission", &UpdateCommission{}},
		{"UpdateCommissionCollector", &UpdateCommissionCollector{}},
		{"UpdateCommissionBps", &UpdateCommissionBps{}},
		{"DepositDelegatorRewards", &DepositDelegatorRewards{}},
		{"AuthorizeChecked", &AuthorizeChecked{}},
		{"VoteSwitch", &VoteSwitch{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.inst.Build().Data()
			require.Error(t, err, "expected error for nil marshal of %s", tc.name)
		})
	}
}
