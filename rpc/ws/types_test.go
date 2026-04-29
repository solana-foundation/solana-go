package ws

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRequest_IDWithinJSONSafeInteger(t *testing.T) {
	for range 2_000 {
		req := newRequest(nil, "slotSubscribe", nil, false)
		require.LessOrEqual(t, req.ID, maxJSONSafeInteger)
	}
}

func TestNewRequest_ShortIDWithinInt31(t *testing.T) {
	for range 2_000 {
		req := newRequest(nil, "slotSubscribe", nil, true)
		require.LessOrEqual(t, req.ID, uint64(math.MaxInt32))
	}
}

func TestGetUint64_AcceptsNumberAndString(t *testing.T) {
	numberPayload := []byte(`{"id":3338220398172203928}`)
	id, err := getUint64(numberPayload, "id")
	require.NoError(t, err)
	require.Equal(t, uint64(3338220398172203928), id)

	stringPayload := []byte(`{"id":"3338220398172203928"}`)
	id, err = getUint64(stringPayload, "id")
	require.NoError(t, err)
	require.Equal(t, uint64(3338220398172203928), id)
}
