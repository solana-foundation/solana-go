package ws

import (
	stdjson "encoding/json"
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

// TestResponseDecodesLargeSubscriptionID covers issue #286: validator
// notifications carry the subscription id as a JSON uint64. Before the fix
// params.Subscription was typed `int`, which fails encoding/json decode on
// 32-bit builds and on any value above math.MaxInt64. The field must use
// uint64 to round-trip safely.
func TestResponseDecodesLargeSubscriptionID(t *testing.T) {
	// Use the largest JSON-safe-but-still-uint64 value the server might emit.
	// math.MaxUint64 itself stresses the regression most directly.
	payload := []byte(`{
		"jsonrpc": "2.0",
		"params": {
			"result": null,
			"subscription": 18446744073709551615
		}
	}`)
	var r response
	require.NoError(t, stdjson.Unmarshal(payload, &r))
	require.NotNil(t, r.Params)
	require.Equal(t, uint64(math.MaxUint64), r.Params.Subscription)
}
