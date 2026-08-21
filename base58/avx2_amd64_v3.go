//go:build amd64.v3 && !purego

package base58

// GOAMD64=v3 or higher guarantees AVX2; the scalar paths compile away.
const useAVX2 = true
