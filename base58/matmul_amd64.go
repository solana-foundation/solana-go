//go:build amd64

package base58

// Scalar fallback used when AVX2 is unavailable (see avx2_amd64.s).

//go:noescape
func encodeMatMul32(src *[32]byte, intermediate *[intermediateSz32]uint64)

//go:noescape
func decodeMatMul32(intermediate *[intermediateSz32]uint64, bin *[binarySz32]uint64)
