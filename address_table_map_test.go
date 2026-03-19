package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testTableKey1 = MustPublicKeyFromBase58("8Vaso6eE1pWktDHwy2qQBB1fhjmBgwzhoXQKe1sxtFjn")
	testTableKey2 = MustPublicKeyFromBase58("FqtwFavD9v99FvoaZrY14bGatCQa9ChsMVphEUNAWHeG")
	testAddr1     = MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	testAddr2     = MustPublicKeyFromBase58("JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4")
	testAddr3     = MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
)

func TestAddressTableMapFromMap_RoundTrip(t *testing.T) {
	tables := map[PublicKey]PublicKeySlice{
		testTableKey1: {testAddr1, testAddr2},
		testTableKey2: {testAddr3},
	}

	om := addressTableMapFromMap(tables)
	require.Equal(t, 2, om.Len())

	result := addressTableMapToMap(om)
	assert.Equal(t, tables, result)
}

func TestAddressTableMapFromMap_Empty(t *testing.T) {
	om := addressTableMapFromMap(map[PublicKey]PublicKeySlice{})
	assert.Equal(t, 0, om.Len())
	assert.Equal(t, map[PublicKey]PublicKeySlice{}, addressTableMapToMap(om))
}

func TestAddressTableMapFromSlice_PreservesOrder(t *testing.T) {
	entries := []AddressTableEntry{
		{TableKey: testTableKey1, Addresses: PublicKeySlice{testAddr1, testAddr2}},
		{TableKey: testTableKey2, Addresses: PublicKeySlice{testAddr3}},
	}

	om := addressTableMapFromSlice(entries)
	require.Equal(t, 2, om.Len())

	// Verify insertion order is preserved.
	result := addressTableMapToSlice(om)
	require.Len(t, result, 2)
	assert.Equal(t, testTableKey1, result[0].TableKey)
	assert.Equal(t, PublicKeySlice{testAddr1, testAddr2}, result[0].Addresses)
	assert.Equal(t, testTableKey2, result[1].TableKey)
	assert.Equal(t, PublicKeySlice{testAddr3}, result[1].Addresses)
}

func TestAddressTableMapFromSlice_Empty(t *testing.T) {
	om := addressTableMapFromSlice([]AddressTableEntry{})
	assert.Equal(t, 0, om.Len())
	assert.Equal(t, []AddressTableEntry{}, addressTableMapToSlice(om))
}

func TestAddressTableMapToSlice_RoundTrip(t *testing.T) {
	entries := []AddressTableEntry{
		{TableKey: testTableKey2, Addresses: PublicKeySlice{testAddr3}},
		{TableKey: testTableKey1, Addresses: PublicKeySlice{testAddr1, testAddr2}},
	}

	result := addressTableMapToSlice(addressTableMapFromSlice(entries))
	assert.Equal(t, entries, result)
}

func TestAddressTableMapToMap_NilSafe(t *testing.T) {
	// nil ordered map should return empty plain map without panicking.
	assert.NotPanics(t, func() {
		result := addressTableMapToMap(nil)
		assert.Equal(t, map[PublicKey]PublicKeySlice{}, result)
	})
}

func TestAddressTableMapToSlice_NilSafe(t *testing.T) {
	assert.NotPanics(t, func() {
		result := addressTableMapToSlice(nil)
		assert.Equal(t, []AddressTableEntry{}, result)
	})
}

// TestAddressTableMapFromSlice_OrderDeterminesTablePriority verifies that when
// the same address appears in two tables, the first entry in the slice wins.
func TestAddressTableMapFromSlice_OrderDeterminesTablePriority(t *testing.T) {
	sharedAddr := testAddr1

	entries := []AddressTableEntry{
		{TableKey: testTableKey1, Addresses: PublicKeySlice{sharedAddr, testAddr2}},
		{TableKey: testTableKey2, Addresses: PublicKeySlice{sharedAddr, testAddr3}},
	}

	om := addressTableMapFromSlice(entries)

	// Build the same addressLookupKeysMap that NewTransaction builds to confirm
	// table1 wins for the shared address.
	addressLookupKeysMap := make(map[PublicKey]addressTablePubkeyWithIndex)
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		for i, addr := range pair.Value {
			if _, exists := addressLookupKeysMap[addr]; !exists {
				addressLookupKeysMap[addr] = addressTablePubkeyWithIndex{
					addressTable: pair.Key,
					index:        uint8(i),
				}
			}
		}
	}

	entry := addressLookupKeysMap[sharedAddr]
	assert.Equal(t, testTableKey1, entry.addressTable, "first slice entry should win for shared address")
	assert.Equal(t, uint8(0), entry.index)
}
