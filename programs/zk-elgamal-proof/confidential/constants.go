package confidential

import token2022 "github.com/gagliardetto/solana-go/programs/token-2022"

const (
	// The number of bits in account or mint balance.
	BalanceBitLength = 64

	// The number of bits in the low part of a transfer amount.
	AmountLoBitLength = 16
	// The number of bits in the high part of a transfer amount.
	AmountHiBitLength = 32
	// Highest possible value of an encrypted transfer amount.
	MaxAmount = 1<<(AmountLoBitLength+AmountHiBitLength) - 1
	// PadBitLength is the headroom in proving balance and amount fit into 128 bits.
	PadBitLength = 128 - BalanceBitLength - AmountLoBitLength - AmountHiBitLength

	// The number of bits in the low part of a transfer fee.
	FeeAmountLoBitLength = 16
	// The number of bits in the high part of a transfer fee.
	FeeAmountHiBitLength = 32
	// MaxFeeBasisPoints is the maximum possible fee in basis points: 100%.
	MaxFeeBasisPoints = uint64(token2022.MaxFeeBasisPoints)
	// deltaBitLength bounds the fee rounding error certified by the percentage-with-cap proof.
	deltaBitLength = 16
	// netAmountBitLength covers the transfer amount less the fee.
	netAmountBitLength = 64
)
