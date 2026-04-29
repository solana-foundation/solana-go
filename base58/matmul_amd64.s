#include "textflag.h"

// func encodeMatMul32(src *[32]byte, intermediate *[9]uint64)
//
// Byte-swaps src into 8 x uint32, then computes:
//   intermediate[k+1] += bin[i] * encTable32[i][k]   for i,k in 0..7
//
// All 8 accumulators live in R8..R15. bin[i] is loaded fresh per row into BX.
// Table entries are 32-bit, loaded into AX with MOVL (zero-extending to 64 bits).
// Fully unrolled, zero table entries skipped.
//
// Register map:
//   SI  = src pointer
//   DI  = intermediate pointer
//   DX  = encTable32 base
//   BX  = current bin[i] (zero-extended u32)
//   AX  = scratch: table entry, then product
//   R8..R15 = intermediate[1..8] accumulators
TEXT ·encodeMatMul32(SB), NOSPLIT|NOFRAME, $0-16
	MOVQ	src+0(FP), SI
	MOVQ	intermediate+8(FP), DI
	LEAQ	·encTable32(SB), DX

	// Zero accumulators.
	XORQ	R8, R8
	XORQ	R9, R9
	XORQ	R10, R10
	XORQ	R11, R11
	XORQ	R12, R12
	XORQ	R13, R13
	XORQ	R14, R14
	XORQ	R15, R15

	// Row 0: bin[0] = bswap(src[0..4])
	MOVL	0(SI), BX
	BSWAPL	BX
	MOVL	0(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R8
	MOVL	4(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R9
	MOVL	8(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R10
	MOVL	12(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	16(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	20(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	24(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	28(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 1: bin[1], k=1..7 (table[1][0] = 0)
	MOVL	4(SI), BX
	BSWAPL	BX
	MOVL	36(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R9
	MOVL	40(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R10
	MOVL	44(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	48(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	52(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	56(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	60(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 2: bin[2], k=2..7
	MOVL	8(SI), BX
	BSWAPL	BX
	MOVL	72(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R10
	MOVL	76(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	80(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	84(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	88(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	92(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 3: bin[3], k=3..7
	MOVL	12(SI), BX
	BSWAPL	BX
	MOVL	108(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	112(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	116(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	120(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	124(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 4: bin[4], k=4..7
	MOVL	16(SI), BX
	BSWAPL	BX
	MOVL	144(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	148(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	152(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	156(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 5: bin[5], k=5..7
	MOVL	20(SI), BX
	BSWAPL	BX
	MOVL	180(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	184(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	188(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 6: bin[6], k=6..7
	MOVL	24(SI), BX
	BSWAPL	BX
	MOVL	216(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	220(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 7: bin[7], k=7 only
	MOVL	28(SI), BX
	BSWAPL	BX
	MOVL	252(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Store intermediate[0..8]. intermediate[0] = 0.
	MOVQ	$0, 0(DI)
	MOVQ	R8, 8(DI)
	MOVQ	R9, 16(DI)
	MOVQ	R10, 24(DI)
	MOVQ	R11, 32(DI)
	MOVQ	R12, 40(DI)
	MOVQ	R13, 48(DI)
	MOVQ	R14, 56(DI)
	MOVQ	R15, 64(DI)
	RET

// func decodeMatMul32(intermediate *[9]uint64, bin *[8]uint64)
//
// Computes: bin[k] = sum_i intermediate[i] * decTable32[i][k]
//
// intermediate[i] values are guaranteed < 58^5 (< 2^30), so the low 32 bits
// of each intermediate contain the full value. We load with MOVL for zero-
// extension.
//
// Register map:
//   SI  = intermediate pointer
//   DI  = bin pointer
//   DX  = decTable32 base
//   BX  = current intermediate[i]
//   AX  = scratch: table entry, then product
//   R8..R15 = bin[0..7] accumulators
TEXT ·decodeMatMul32(SB), NOSPLIT|NOFRAME, $0-16
	MOVQ	intermediate+0(FP), SI
	MOVQ	bin+8(FP), DI
	LEAQ	·decTable32(SB), DX

	// Zero accumulators.
	XORQ	R8, R8
	XORQ	R9, R9
	XORQ	R10, R10
	XORQ	R11, R11
	XORQ	R12, R12
	XORQ	R13, R13
	XORQ	R14, R14
	XORQ	R15, R15

	// Row 0: intermediate[0], k=0..6 (table[0][7] = 0)
	MOVL	0(SI), BX
	MOVL	0(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R8
	MOVL	4(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R9
	MOVL	8(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R10
	MOVL	12(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	16(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	20(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	24(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14

	// Row 1: intermediate[1], k=1..6 (table[1][0] = 0, table[1][7] = 0)
	MOVL	8(SI), BX
	MOVL	36(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R9
	MOVL	40(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R10
	MOVL	44(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	48(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	52(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	56(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14

	// Row 2: intermediate[2], k=2..7
	MOVL	16(SI), BX
	MOVL	72(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R10
	MOVL	76(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	80(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	84(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	88(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	92(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 3: intermediate[3], k=3..7
	MOVL	24(SI), BX
	MOVL	108(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R11
	MOVL	112(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	116(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	120(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	124(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 4: intermediate[4], k=4..7
	MOVL	32(SI), BX
	MOVL	144(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R12
	MOVL	148(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	152(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	156(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 5: intermediate[5], k=5..7
	MOVL	40(SI), BX
	MOVL	180(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R13
	MOVL	184(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	188(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 6: intermediate[6], k=6..7
	MOVL	48(SI), BX
	MOVL	216(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R14
	MOVL	220(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 7: intermediate[7], k=7 only
	MOVL	56(SI), BX
	MOVL	252(DX), AX
	IMULQ	BX, AX
	ADDQ	AX, R15

	// Row 8: table[8] = {0,0,0,0,0,0,0,1} -> bin[7] += intermediate[8]
	MOVQ	64(SI), BX
	ADDQ	BX, R15

	// Store bin[0..7].
	MOVQ	R8, 0(DI)
	MOVQ	R9, 8(DI)
	MOVQ	R10, 16(DI)
	MOVQ	R11, 24(DI)
	MOVQ	R12, 32(DI)
	MOVQ	R13, 40(DI)
	MOVQ	R14, 48(DI)
	MOVQ	R15, 56(DI)
	RET
