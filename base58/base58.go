// Package base58 provides Base58 encoding and decoding optimized for the
// fixed-width values commonly used by Solana.
//
// Deprecated: use github.com/fluxrpc/base58 directly in new code. This
// package remains as a compatibility shim for existing solana-go users.
package base58

import fluxbase58 "github.com/fluxrpc/base58"

const (
	// EncodedMaxLen32 is the maximum Base58-encoded length of a 32-byte value.
	EncodedMaxLen32 = fluxbase58.EncodedMaxLen32
	// EncodedMaxLen64 is the maximum Base58-encoded length of a 64-byte value.
	EncodedMaxLen64 = fluxbase58.EncodedMaxLen64
)

var (
	// ErrInvalidChar indicates that an encoded value contains a character
	// outside the Base58 alphabet.
	ErrInvalidChar = fluxbase58.ErrInvalidChar
	// ErrInvalidLength indicates that an encoded value cannot fit the target
	// fixed-width output.
	ErrInvalidLength = fluxbase58.ErrInvalidLength
	// ErrValueTooLarge indicates that a decoded value exceeds the target size.
	ErrValueTooLarge = fluxbase58.ErrValueTooLarge
	// ErrLeadingZeros indicates a mismatch between leading '1' characters and
	// leading zero bytes.
	ErrLeadingZeros = fluxbase58.ErrLeadingZeros
)

// Encode encodes a byte slice as Base58.
func Encode(src []byte) string {
	return fluxbase58.Encode(src)
}

// Decode decodes a Base58 string into a byte slice.
func Decode(encoded string) ([]byte, error) {
	return fluxbase58.Decode(encoded)
}

// Encode32 encodes a 32-byte value using the fixed-width fast path.
func Encode32(src *[32]byte) string {
	return fluxbase58.Encode32(src)
}

// Decode32 decodes a Base58 string into a 32-byte value.
func Decode32(encoded string, dst *[32]byte) error {
	return fluxbase58.Decode32(encoded, dst)
}

// Encode64 encodes a 64-byte value using the fixed-width fast path.
func Encode64(src *[64]byte) string {
	return fluxbase58.Encode64(src)
}

// Decode64 decodes a Base58 string into a 64-byte value.
func Decode64(encoded string, dst *[64]byte) error {
	return fluxbase58.Decode64(encoded, dst)
}

// AppendEncode32 appends the Base58 encoding of src to dst.
func AppendEncode32(dst []byte, src *[32]byte) []byte {
	return fluxbase58.AppendEncode32(dst, src)
}

// AppendEncode64 appends the Base58 encoding of src to dst.
func AppendEncode64(dst []byte, src *[64]byte) []byte {
	return fluxbase58.AppendEncode64(dst, src)
}
