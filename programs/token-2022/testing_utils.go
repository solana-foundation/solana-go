package token2022

import (
	"bytes"
	"fmt"

	ag_binary "github.com/gagliardetto/solana-go/binary"
)

func encodeT(data any, buf *bytes.Buffer) error {
	if err := ag_binary.NewBinEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("unable to encode instruction: %w", err)
	}
	return nil
}

func decodeT(dst any, data []byte) error {
	return ag_binary.NewBinDecoder(data).Decode(dst)
}
