package confidential

import "errors"

// Proof-generation errors
var (
	ErrNotEnoughFunds         = errors.New("zk: not enough funds in account")
	ErrIllegalAmountBitLength = errors.New("zk: amount has illegal bit length")
	ErrFeeCalculation         = errors.New("zk: fee calculation failed")
)
