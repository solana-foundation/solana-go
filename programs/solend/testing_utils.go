package solend

import (
	"bytes"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
)

func encodeT(data interface{}, buf *bytes.Buffer) error {
	if err := ag_binary.NewBinEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("unable to encode instruction: %w", err)
	}
	return nil
}

func decodeT(dst interface{}, data []byte) error {
	return ag_binary.NewBinDecoder(data).Decode(dst)
}
