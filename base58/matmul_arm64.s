#include "textflag.h"

// func encodeMatMul32(src *[32]byte, intermediate *[9]uint64)
//
// Byte-swaps src into 8 x uint32, then computes the 8x8 matrix-vector
// multiply: intermediate[k+1] += bin[i] * encTable32[i][k]
//
// All intermediate values kept in registers to avoid memory traffic.
// Fully unrolled, zero entries skipped.
//
// ARM64 MADD semantics in Go asm: MADD Rm, Ra, Rn, Rd -> Rd = Ra + Rn * Rm
TEXT ·encodeMatMul32(SB), NOSPLIT|NOFRAME, $0-16
	MOVD	src+0(FP), R0
	MOVD	intermediate+8(FP), R2

	// Load and byte-swap 8 uint32 words from src.
	MOVWU	0(R0), R3
	REVW	R3, R3
	MOVWU	4(R0), R4
	REVW	R4, R4
	MOVWU	8(R0), R5
	REVW	R5, R5
	MOVWU	12(R0), R6
	REVW	R6, R6
	MOVWU	16(R0), R7
	REVW	R7, R7
	MOVWU	20(R0), R8
	REVW	R8, R8
	MOVWU	24(R0), R9
	REVW	R9, R9
	MOVWU	28(R0), R10
	REVW	R10, R10

	// Zero accumulators for intermediate[1..8].
	MOVD	$0, R11
	MOVD	$0, R12
	MOVD	$0, R13
	MOVD	$0, R14
	MOVD	$0, R15
	MOVD	$0, R16
	MOVD	$0, R17
	MOVD	$0, R19

	MOVD	$·encTable32(SB), R1

	// Row 0: intermediate[k+1] += bin[0] * encTable32[0][k], k=0..7
	MOVWU	0(R1), R20
	MADD	R3, R11, R20, R11
	MOVWU	4(R1), R20
	MADD	R3, R12, R20, R12
	MOVWU	8(R1), R20
	MADD	R3, R13, R20, R13
	MOVWU	12(R1), R20
	MADD	R3, R14, R20, R14
	MOVWU	16(R1), R20
	MADD	R3, R15, R20, R15
	MOVWU	20(R1), R20
	MADD	R3, R16, R20, R16
	MOVWU	24(R1), R20
	MADD	R3, R17, R20, R17
	MOVWU	28(R1), R20
	MADD	R3, R19, R20, R19

	// Row 1: k=1..7 (table[1][0] = 0)
	MOVWU	36(R1), R20
	MADD	R4, R12, R20, R12
	MOVWU	40(R1), R20
	MADD	R4, R13, R20, R13
	MOVWU	44(R1), R20
	MADD	R4, R14, R20, R14
	MOVWU	48(R1), R20
	MADD	R4, R15, R20, R15
	MOVWU	52(R1), R20
	MADD	R4, R16, R20, R16
	MOVWU	56(R1), R20
	MADD	R4, R17, R20, R17
	MOVWU	60(R1), R20
	MADD	R4, R19, R20, R19

	// Row 2: k=2..7
	MOVWU	72(R1), R20
	MADD	R5, R13, R20, R13
	MOVWU	76(R1), R20
	MADD	R5, R14, R20, R14
	MOVWU	80(R1), R20
	MADD	R5, R15, R20, R15
	MOVWU	84(R1), R20
	MADD	R5, R16, R20, R16
	MOVWU	88(R1), R20
	MADD	R5, R17, R20, R17
	MOVWU	92(R1), R20
	MADD	R5, R19, R20, R19

	// Row 3: k=3..7
	MOVWU	108(R1), R20
	MADD	R6, R14, R20, R14
	MOVWU	112(R1), R20
	MADD	R6, R15, R20, R15
	MOVWU	116(R1), R20
	MADD	R6, R16, R20, R16
	MOVWU	120(R1), R20
	MADD	R6, R17, R20, R17
	MOVWU	124(R1), R20
	MADD	R6, R19, R20, R19

	// Row 4: k=4..7
	MOVWU	144(R1), R20
	MADD	R7, R15, R20, R15
	MOVWU	148(R1), R20
	MADD	R7, R16, R20, R16
	MOVWU	152(R1), R20
	MADD	R7, R17, R20, R17
	MOVWU	156(R1), R20
	MADD	R7, R19, R20, R19

	// Row 5: k=5..7
	MOVWU	180(R1), R20
	MADD	R8, R16, R20, R16
	MOVWU	184(R1), R20
	MADD	R8, R17, R20, R17
	MOVWU	188(R1), R20
	MADD	R8, R19, R20, R19

	// Row 6: k=6..7
	MOVWU	216(R1), R20
	MADD	R9, R17, R20, R17
	MOVWU	220(R1), R20
	MADD	R9, R19, R20, R19

	// Row 7: k=7 only
	MOVWU	252(R1), R20
	MADD	R10, R19, R20, R19

	// Store intermediate[0..8].
	MOVD	$0, 0(R2)
	MOVD	R11, 8(R2)
	MOVD	R12, 16(R2)
	MOVD	R13, 24(R2)
	MOVD	R14, 32(R2)
	MOVD	R15, 40(R2)
	MOVD	R16, 48(R2)
	MOVD	R17, 56(R2)
	MOVD	R19, 64(R2)
	RET

// func decodeMatMul32(intermediate *[9]uint64, bin *[8]uint64)
//
// Computes: bin[k] = sum_i intermediate[i] * decTable32[i][k]
// Fully unrolled.
TEXT ·decodeMatMul32(SB), NOSPLIT|NOFRAME, $0-16
	MOVD	intermediate+0(FP), R0
	MOVD	bin+8(FP), R2

	// Load 9 intermediate values.
	MOVD	0(R0), R3
	MOVD	8(R0), R4
	MOVD	16(R0), R5
	MOVD	24(R0), R6
	MOVD	32(R0), R7
	MOVD	40(R0), R8
	MOVD	48(R0), R9
	MOVD	56(R0), R10
	MOVD	64(R0), R21

	// Zero accumulators for bin[0..7].
	MOVD	$0, R11
	MOVD	$0, R12
	MOVD	$0, R13
	MOVD	$0, R14
	MOVD	$0, R15
	MOVD	$0, R16
	MOVD	$0, R17
	MOVD	$0, R19

	MOVD	$·decTable32(SB), R1

	// Row 0: bin[k] += intermediate[0] * decTable32[0][k], k=0..6 (table[0][7]=0)
	MOVWU	0(R1), R20
	MADD	R3, R11, R20, R11
	MOVWU	4(R1), R20
	MADD	R3, R12, R20, R12
	MOVWU	8(R1), R20
	MADD	R3, R13, R20, R13
	MOVWU	12(R1), R20
	MADD	R3, R14, R20, R14
	MOVWU	16(R1), R20
	MADD	R3, R15, R20, R15
	MOVWU	20(R1), R20
	MADD	R3, R16, R20, R16
	MOVWU	24(R1), R20
	MADD	R3, R17, R20, R17

	// Row 1: k=1..6 (table[1][0]=0, table[1][7]=0)
	MOVWU	36(R1), R20
	MADD	R4, R12, R20, R12
	MOVWU	40(R1), R20
	MADD	R4, R13, R20, R13
	MOVWU	44(R1), R20
	MADD	R4, R14, R20, R14
	MOVWU	48(R1), R20
	MADD	R4, R15, R20, R15
	MOVWU	52(R1), R20
	MADD	R4, R16, R20, R16
	MOVWU	56(R1), R20
	MADD	R4, R17, R20, R17

	// Row 2: k=2..7
	MOVWU	72(R1), R20
	MADD	R5, R13, R20, R13
	MOVWU	76(R1), R20
	MADD	R5, R14, R20, R14
	MOVWU	80(R1), R20
	MADD	R5, R15, R20, R15
	MOVWU	84(R1), R20
	MADD	R5, R16, R20, R16
	MOVWU	88(R1), R20
	MADD	R5, R17, R20, R17
	MOVWU	92(R1), R20
	MADD	R5, R19, R20, R19

	// Row 3: k=3..7
	MOVWU	108(R1), R20
	MADD	R6, R14, R20, R14
	MOVWU	112(R1), R20
	MADD	R6, R15, R20, R15
	MOVWU	116(R1), R20
	MADD	R6, R16, R20, R16
	MOVWU	120(R1), R20
	MADD	R6, R17, R20, R17
	MOVWU	124(R1), R20
	MADD	R6, R19, R20, R19

	// Row 4: k=4..7
	MOVWU	144(R1), R20
	MADD	R7, R15, R20, R15
	MOVWU	148(R1), R20
	MADD	R7, R16, R20, R16
	MOVWU	152(R1), R20
	MADD	R7, R17, R20, R17
	MOVWU	156(R1), R20
	MADD	R7, R19, R20, R19

	// Row 5: k=5..7
	MOVWU	180(R1), R20
	MADD	R8, R16, R20, R16
	MOVWU	184(R1), R20
	MADD	R8, R17, R20, R17
	MOVWU	188(R1), R20
	MADD	R8, R19, R20, R19

	// Row 6: k=6..7
	MOVWU	216(R1), R20
	MADD	R9, R17, R20, R17
	MOVWU	220(R1), R20
	MADD	R9, R19, R20, R19

	// Row 7: k=7 only (table[7][7] = 656356768)
	MOVWU	252(R1), R20
	MADD	R10, R19, R20, R19

	// Row 8: table[8] = {0,0,0,0,0,0,0,1} -> bin[7] += intermediate[8]
	ADD	R21, R19, R19

	// Store bin[0..7].
	MOVD	R11, 0(R2)
	MOVD	R12, 8(R2)
	MOVD	R13, 16(R2)
	MOVD	R14, 24(R2)
	MOVD	R15, 32(R2)
	MOVD	R16, 40(R2)
	MOVD	R17, 48(R2)
	MOVD	R19, 56(R2)
	RET
