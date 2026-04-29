package token2022

import (
	"io"

	gojson "github.com/goccy/go-json"
)

var json = struct {
	Marshal       func(v any) ([]byte, error)
	MarshalIndent func(v any, prefix, indent string) ([]byte, error)
	Unmarshal     func(data []byte, v any) error
	NewDecoder    func(r io.Reader) *gojson.Decoder
	NewEncoder    func(w io.Writer) *gojson.Encoder
}{
	Marshal:       gojson.Marshal,
	MarshalIndent: gojson.MarshalIndent,
	Unmarshal:     gojson.Unmarshal,
	NewDecoder:    gojson.NewDecoder,
	NewEncoder:    gojson.NewEncoder,
}
