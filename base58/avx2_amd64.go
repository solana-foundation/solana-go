//go:build amd64 && !purego

package base58

// Kernels in avx2_amd64.s. Encoders return the length n and clobber dst[n:].
// Decoders take 32..44 / 64..88 chars and return 0 (ok), 1 (ErrInvalidChar)
// or 2 (ErrValueTooLarge); leading-'1' validation is left to the caller.

//go:noescape
func encode32AVX2(src *[32]byte, dst *[48]byte) int

//go:noescape
func encode64AVX2(src *[64]byte, dst *[96]byte) int

//go:noescape
func decode32AVX2(s *byte, n int, dst *[32]byte) int

//go:noescape
func decode64AVX2(s *byte, n int, dst *[64]byte) int

// Tables widened to qwords (rows padded to 4-lane multiples) for VPMULUDQ.
var (
	encTable32q [binarySz32][8]uint64
	encTable64q [binarySz64][20]uint64
	decTable32q [intermediateSz32][8]uint64
	decTable64q [intermediateSz64][16]uint64
)

func init() {
	for i, row := range encTable32 {
		for k, v := range row {
			encTable32q[i][k] = uint64(v)
		}
	}
	for i, row := range encTable64 {
		for k, v := range row {
			encTable64q[i][k] = uint64(v)
		}
	}
	for i, row := range decTable32 {
		for k, v := range row {
			decTable32q[i][k] = uint64(v)
		}
	}
	for i, row := range decTable64 {
		for k, v := range row {
			decTable64q[i][k] = uint64(v)
		}
	}
}

// rep tiles b into a 32-byte vector constant.
//
//nolint:unused // referenced from avx2_amd64.s
func rep(b ...byte) (out [32]byte) {
	for i := range out {
		out[i] = b[i%len(b)]
	}
	return
}

//nolint:unused // referenced from avx2_amd64.s
var (
	avxBswap  = rep(3, 2, 1, 0, 7, 6, 5, 4, 11, 10, 9, 8, 15, 14, 13, 12)
	avxIota   = rep(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	avxSpread = rep(0, 1, 1, 1, 1, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1)
	avxP      = rep(0, 1, 2, 3, 5, 6, 7, 8, 10, 11, 12, 13, 0x80, 0x80, 0x80, 0x80)
	avxQ      = rep(4, 0x80, 0x80, 0x80, 9, 0x80, 0x80, 0x80, 14, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80)
	avxTail   = [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0x80}
	avxW1 = rep(58, 1)
	avxW2 = rep(0x24, 0x0D, 1, 0) // int16 {3364, 1}
	// valid base58 char iff lo[c&15]&hi[c>>4] != 0
	avxLoNib = rep(20, 31, 31, 31, 31, 31, 31, 31, 31, 29, 30, 10, 2, 10, 10, 8)
	avxHiNib = rep(0, 0, 0, 1, 2, 4, 8, 16, 0, 0, 0, 0, 0, 0, 0, 0)
	avxB8    = rep(8)
	avxB16   = rep(16)
	avxB21   = rep(21)
	avxB32   = rep(32)
	avxB43   = rep(43)
	avxB64   = rep(64)
	avxB73   = rep(73)
	avxB79   = rep(79)
	avxB96   = rep(96)
	avxB108  = rep(108)
	avxBm6   = rep(0xFA) // -6
	avxBm7   = rep(0xF9) // -7
	avxBm49  = rep(0xCF) // -'1'
	avxB15   = rep(0x0F)
	avxD58   = rep(58, 0, 0, 0)
	avxDivA  = rep(0x09, 0xCB, 0x3D, 0x8D, 0, 0, 0, 0) // 2369637129 = 2^37/58
	avxDivB  = rep(0x93, 0x20, 0xED, 0x4D, 0, 0, 0, 0) // 1307386003 = 2^42/58^2
)
