package ws

import (
	"strings"
	"testing"

	"github.com/gorilla/rpc/v2/json2"
	"github.com/stretchr/testify/require"
)

// A notification frame that omits params.result, sets it to null, or is a
// bare JSON null must surface as a decode error. It used to dereference a nil
// pointer inside the receive goroutine and crash the process.
func TestDecodeResponse_MissingResultDoesNotPanic(t *testing.T) {
	frames := map[string]string{
		"missing result": `{"jsonrpc":"2.0","method":"slotNotification","params":{"subscription":1}}`,
		"null result":    `{"jsonrpc":"2.0","method":"slotNotification","params":{"subscription":1,"result":null}}`,
		"missing params": `{"jsonrpc":"2.0","method":"slotNotification"}`,
		"null frame":     `null`,
	}

	for name, frame := range frames {
		t.Run(name+"/message", func(t *testing.T) {
			var res SlotResult
			require.NotPanics(t, func() {
				err := decodeResponseFromMessage([]byte(frame), &res)
				require.ErrorIs(t, err, json2.ErrNullResult)
			})
		})
		t.Run(name+"/reader", func(t *testing.T) {
			var res SlotResult
			require.NotPanics(t, func() {
				err := decodeResponseFromReader(strings.NewReader(frame), &res)
				require.ErrorIs(t, err, json2.ErrNullResult)
			})
		})
	}
}
