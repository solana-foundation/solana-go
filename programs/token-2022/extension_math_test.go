package token2022

import (
	"math"
	"testing"

	ag_require "github.com/stretchr/testify/require"
)

// The tests in this file are ported 1:1 from the Rust
// spl-token-2022-interface crate:
//   - src/extension/transfer_fee/mod.rs
//   - src/extension/interest_bearing_mint/mod.rs
//   - src/extension/scaled_ui_amount/mod.rs
// The Rust proptest property tests (round_trip_fee_calculation,
// inverse_fee_relationship, time_weighted_average_calc, amount_to_ui_amount)
// are not ported; the deterministic tests are ported verbatim.

const (
	testNewerEpoch = uint64(100)
	testOlderEpoch = uint64(1)
	// INT_SECONDS_PER_YEAR in the Rust tests: 6 * 6 * 24 * 36524.
	intSecondsPerYear = int64(6 * 6 * 24 * 36524)
	testDecimals      = uint8(2)
)

// testTransferFeeConfig mirrors the Rust test_transfer_fee_config fixture.
func testTransferFeeConfig() TransferFeeConfigState {
	authority := pubkeyOf(10)
	withdraw := pubkeyOf(11)
	return TransferFeeConfigState{
		TransferFeeConfigAuthority: NewOptionalPubkey(&authority),
		WithdrawWithheldAuthority:  NewOptionalPubkey(&withdraw),
		WithheldAmount:             math.MaxUint64,
		OlderTransferFee: TransferFee{
			Epoch:                  testOlderEpoch,
			MaximumFee:             10,
			TransferFeeBasisPoints: 100,
		},
		NewerTransferFee: TransferFee{
			Epoch:                  testNewerEpoch,
			MaximumFee:             5_000,
			TransferFeeBasisPoints: 1,
		},
	}
}

func TestTransferFee_EpochFee(t *testing.T) {
	config := testTransferFeeConfig()
	// during epoch 100 and after, use newer transfer fee
	ag_require.Equal(t, testNewerEpoch, config.GetEpochFee(testNewerEpoch).Epoch)
	ag_require.Equal(t, testNewerEpoch, config.GetEpochFee(testNewerEpoch+1).Epoch)
	ag_require.Equal(t, testNewerEpoch, config.GetEpochFee(math.MaxUint64).Epoch)
	// before that, use older transfer fee
	ag_require.Equal(t, testOlderEpoch, config.GetEpochFee(testNewerEpoch-1).Epoch)
	ag_require.Equal(t, testOlderEpoch, config.GetEpochFee(testOlderEpoch).Epoch)
	ag_require.Equal(t, testOlderEpoch, config.GetEpochFee(testOlderEpoch+1).Epoch)
}

// requireFee returns a helper that unwraps (value, error) pairs from the fee
// calculation methods, failing the test on error.
func requireFee(t *testing.T) func(uint64, error) uint64 {
	return func(v uint64, err error) uint64 {
		t.Helper()
		ag_require.NoError(t, err)
		return v
	}
}

func TestTransferFee_CalculateFeeMax(t *testing.T) {
	fee := requireFee(t)
	one := oneInBasisPoints
	transferFee := TransferFee{Epoch: 0, MaximumFee: 5_000, TransferFeeBasisPoints: 1}
	maximumFee := transferFee.MaximumFee
	// hit maximum fee
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateFee(math.MaxUint64)))
	// at exactly the max
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateFee(maximumFee*one)))
	// one token above, normally rounds up, but we're at the max
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateFee(maximumFee*one+1)))
	// one token below, rounds up to the max
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateFee(maximumFee*one-1)))
}

func TestTransferFee_CalculateFeeMin(t *testing.T) {
	fee := requireFee(t)
	one := oneInBasisPoints
	transferFee := TransferFee{Epoch: 0, MaximumFee: 5_000, TransferFeeBasisPoints: 1}
	minimumFee := uint64(1)
	// hit minimum fee even with 1 token
	ag_require.Equal(t, minimumFee, fee(transferFee.CalculateFee(1)))
	// still minimum at 2 tokens
	ag_require.Equal(t, minimumFee, fee(transferFee.CalculateFee(2)))
	// still minimum at 10_000 tokens
	ag_require.Equal(t, minimumFee, fee(transferFee.CalculateFee(one)))
	// 2 token fee at 10_001
	ag_require.Equal(t, minimumFee+1, fee(transferFee.CalculateFee(one+1)))
	// zero is always zero
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(0)))
}

func TestTransferFee_CalculateFeeZero(t *testing.T) {
	fee := requireFee(t)
	one := oneInBasisPoints
	transferFee := TransferFee{Epoch: 0, MaximumFee: math.MaxUint64, TransferFeeBasisPoints: 0}
	// always zero fee
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(0)))
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(math.MaxUint64)))
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(1)))
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(one)))

	transferFee = TransferFee{Epoch: 0, MaximumFee: 0, TransferFeeBasisPoints: MaxFeeBasisPoints}
	// always zero fee
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(0)))
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(math.MaxUint64)))
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(1)))
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculateFee(one)))
}

func TestTransferFee_CalculateFeeExactOutMax(t *testing.T) {
	fee := requireFee(t)
	one := oneInBasisPoints
	transferFee := TransferFee{Epoch: 0, MaximumFee: 5_000, TransferFeeBasisPoints: 1}
	maximumFee := transferFee.MaximumFee
	// hit maximum fee
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateInverseFee(math.MaxUint64-maximumFee)))
	// at exactly the max
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateInverseFee(maximumFee*one-maximumFee)))
	// one token above, normally rounds up, but we're at the max
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateInverseFee(maximumFee*one-maximumFee+1)))
	// one token below, rounds up to the max
	ag_require.Equal(t, maximumFee, fee(transferFee.CalculateInverseFee(maximumFee*one-maximumFee-1)))
}

func TestTransferFee_CalculatePreFeeAmountEdgeCases(t *testing.T) {
	fee := requireFee(t)
	maximumFee := uint64(5_000)
	transferFee := TransferFee{Epoch: 0, MaximumFee: maximumFee, TransferFeeBasisPoints: MaxFeeBasisPoints}

	// 0 zero out, 0 in
	ag_require.Equal(t, uint64(0), fee(transferFee.CalculatePreFeeAmount(0)))

	// cap at max fee
	ag_require.Equal(t, 1+maximumFee, fee(transferFee.CalculatePreFeeAmount(1)))

	// no fee same amount
	transferFee = TransferFee{Epoch: 0, MaximumFee: maximumFee, TransferFeeBasisPoints: 0}
	ag_require.Equal(t, uint64(1), fee(transferFee.CalculatePreFeeAmount(1)))
}

func TestTransferFee_CalculateFeeExactOutMin(t *testing.T) {
	fee := requireFee(t)
	one := oneInBasisPoints
	transferFee := TransferFee{Epoch: 0, MaximumFee: 5_000, TransferFeeBasisPoints: 1}
	minimumFee := uint64(1)
	// hit minimum fee even with 1 token
	ag_require.Equal(t, minimumFee, fee(transferFee.CalculateInverseFee(1)))
	// still minimum at 2 tokens
	ag_require.Equal(t, minimumFee, fee(transferFee.CalculateInverseFee(2)))
	// still minimum at 9_999 tokens
	ag_require.Equal(t, minimumFee, fee(transferFee.CalculateInverseFee(one-1)))
	// 2 token fee at 10_000
	ag_require.Equal(t, minimumFee+1, fee(transferFee.CalculateInverseFee(one)))
}

// --- InterestBearingConfig ---

func TestInterestBearing_SecondsPerYear(t *testing.T) {
	ag_require.Equal(t, 31_556_736.0, secondsPerYear)
	ag_require.Equal(t, int64(31_556_736), intSecondsPerYear)
}

func TestInterestBearing_SpecificAmountToUiAmount(t *testing.T) {
	const one = uint64(1_000_000_000_000_000_000)
	// constant 5%
	config := InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             500,
	}
	// 1 year at 5% gives a total of exp(0.05) = 1.0512710963760241
	ui, err := config.AmountToUiAmount(one, 18, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "1.051271096376024117", ui)
	// with 1 decimal place
	ui, err = config.AmountToUiAmount(one, 19, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "0.1051271096376024117", ui)
	// with 10 decimal places
	ui, err = config.AmountToUiAmount(one, 28, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "0.0000000001051271096376024175", ui) // different digits at the end!

	// huge amount with 10 decimal places
	ui, err = config.AmountToUiAmount(10_000_000_000, 10, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "1.0512710964", ui)

	// negative
	config = InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    -500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             -500,
	}
	// 1 year at -5% gives a total of exp(-0.05) = 0.951229424500714
	ui, err = config.AmountToUiAmount(one, 18, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "0.951229424500713905", ui)

	// net out
	config = InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    -500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             500,
	}
	// 1 year at -5% and 1 year at 5% gives a total of 1
	ui, err = config.AmountToUiAmount(1, 0, intSecondsPerYear*2)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "1", ui)

	// huge values
	config = InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             500,
	}
	ui, err = config.AmountToUiAmount(math.MaxUint64, 0, intSecondsPerYear*2)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "20386805083448098816", ui)
	ui, err = config.AmountToUiAmount(math.MaxUint64, 0, intSecondsPerYear*10_000)
	ag_require.NoError(t, err)
	// there's an underflow risk, but it works!
	ag_require.Equal(t, "258917064265813826192025834755112557504850551118283225815045099303279643822914042296793377611277551888244755303462190670431480816358154467489350925148558569427069926786360814068189956495940285398273555561779717914539956777398245259214848", ui)
}

func TestInterestBearing_SpecificUiAmountToAmount(t *testing.T) {
	// constant 5%
	config := InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             500,
	}
	// 1 year at 5% gives a total of exp(0.05) = 1.0512710963760241
	amount, err := config.UiAmountToAmount("1.0512710963760241", 0, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)
	// with 1 decimal place
	amount, err = config.UiAmountToAmount("0.10512710963760241", 1, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)
	// with 10 decimal places
	amount, err = config.UiAmountToAmount("0.00000000010512710963760242", 10, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)

	// huge amount with 10 decimal places
	amount, err = config.UiAmountToAmount("1.0512710963760241", 10, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(10_000_000_000), amount)

	// negative
	config = InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    -500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             -500,
	}
	// 1 year at -5% gives a total of exp(-0.05) = 0.951229424500714
	amount, err = config.UiAmountToAmount("0.951229424500714", 0, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)

	// net out
	config = InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    -500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             500,
	}
	// 1 year at -5% and 1 year at 5% gives a total of 1
	amount, err = config.UiAmountToAmount("1", 0, intSecondsPerYear*2)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)

	// huge values
	config = InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    500,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             500,
	}
	amount, err = config.UiAmountToAmount("20386805083448100000", 0, intSecondsPerYear*2)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(math.MaxUint64), amount)
}

func TestInterestBearing_SpecificAmountToUiAmountNoInterest(t *testing.T) {
	config := InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    0,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             0,
	}
	for _, tc := range []struct {
		amount   uint64
		expected string
	}{
		{23, "0.23"}, {110, "1.1"}, {4200, "42"}, {0, "0"},
	} {
		ui, err := config.AmountToUiAmount(tc.amount, testDecimals, intSecondsPerYear)
		ag_require.NoError(t, err)
		ag_require.Equal(t, tc.expected, ui)
	}
}

func TestInterestBearing_SpecificUiAmountToAmountNoInterest(t *testing.T) {
	config := InterestBearingConfigState{
		InitializationTimestamp: 0,
		PreUpdateAverageRate:    0,
		LastUpdateTimestamp:     intSecondsPerYear,
		CurrentRate:             0,
	}
	for _, tc := range []struct {
		uiAmount string
		expected uint64
	}{
		{"0.23", 23}, {"0.20", 20}, {"0.2000", 20}, {".2", 20},
		{"1.1", 110}, {"1.10", 110}, {"42", 4200}, {"42.", 4200}, {"0", 0},
	} {
		amount, err := config.UiAmountToAmount(tc.uiAmount, testDecimals, intSecondsPerYear)
		ag_require.NoError(t, err)
		ag_require.Equal(t, tc.expected, amount)
	}

	// this is invalid with normal mints, but rounding for this mint makes it ok
	amount, err := config.UiAmountToAmount("0.111", testDecimals, intSecondsPerYear)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(11), amount)

	// fail if invalid ui_amount passed in
	for _, uiAmount := range []string{"", ".", "0.t"} {
		_, err := config.UiAmountToAmount(uiAmount, testDecimals, intSecondsPerYear)
		ag_require.ErrorIs(t, err, ErrInvalidUiAmount)
	}
}

func TestParseUiAmount_RejectsGoLiteralSyntax(t *testing.T) {
	// Rust's f64::from_str rejects digit-separating underscores and hex
	// floats; Go's strconv.ParseFloat accepts them, so the parser filters
	// them out explicitly for parity.
	config := ScaledUiAmountState{
		Multiplier:                      1.0,
		NewMultiplierEffectiveTimestamp: 1,
	}
	for _, uiAmount := range []string{"1_000", "1_0.5", "0x1p4", "0X1P4"} {
		_, err := config.UiAmountToAmount(uiAmount, 0, 0)
		ag_require.ErrorIs(t, err, ErrInvalidUiAmount, uiAmount)
	}
}

// --- ScaledUiAmountConfig ---

func TestScaledUiAmount_MultiplierChoice(t *testing.T) {
	multiplier := 5.0
	newMultiplier := 10.0
	effectiveTimestamp := int64(1)
	config := ScaledUiAmountState{
		Multiplier:                      multiplier,
		NewMultiplier:                   newMultiplier,
		NewMultiplierEffectiveTimestamp: effectiveTimestamp,
	}
	ag_require.Equal(t, newMultiplier, config.CurrentMultiplier(effectiveTimestamp))
	ag_require.Equal(t, multiplier, config.CurrentMultiplier(effectiveTimestamp-1))
	ag_require.Equal(t, multiplier, config.CurrentMultiplier(0))
	ag_require.Equal(t, multiplier, config.CurrentMultiplier(math.MinInt64))
	ag_require.Equal(t, newMultiplier, config.CurrentMultiplier(math.MaxInt64))
}

func TestScaledUiAmount_SpecificAmountToUiAmount(t *testing.T) {
	// 5x
	config := ScaledUiAmountState{
		Multiplier:                      5.0,
		NewMultiplierEffectiveTimestamp: 1,
	}
	ui, err := config.AmountToUiAmount(1, 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "5", ui)
	// with 1 decimal place
	ui, err = config.AmountToUiAmount(1, 1, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "0.5", ui)
	// with 10 decimal places
	ui, err = config.AmountToUiAmount(1, 10, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "0.0000000005", ui)

	// huge amount with 10 decimal places
	ui, err = config.AmountToUiAmount(10_000_000_000, 10, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "5", ui)

	// huge values
	config = ScaledUiAmountState{
		Multiplier:                      math.MaxFloat64,
		NewMultiplierEffectiveTimestamp: 1,
	}
	ui, err = config.AmountToUiAmount(math.MaxUint64, 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "inf", ui)

	// truncation
	config = ScaledUiAmountState{
		Multiplier:                      0.99,
		NewMultiplierEffectiveTimestamp: 1,
	}
	// This is really 0.99999... but it gets truncated
	ui, err = config.AmountToUiAmount(101, 2, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, "0.99", ui)
}

func TestScaledUiAmount_SpecificUiAmountToAmount(t *testing.T) {
	// constant 5x
	config := ScaledUiAmountState{
		Multiplier:                      5.0,
		NewMultiplierEffectiveTimestamp: 1,
	}
	amount, err := config.UiAmountToAmount("5.0", 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)
	// with 1 decimal place
	amount, err = config.UiAmountToAmount("0.500000000", 1, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)
	// with 10 decimal places
	amount, err = config.UiAmountToAmount("0.00000000050000000000000000", 10, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)

	// huge amount with 10 decimal places
	amount, err = config.UiAmountToAmount("5.0000000000000000", 10, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(10_000_000_000), amount)

	// huge values
	amount, err = config.UiAmountToAmount("92233720368547758075", 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(math.MaxUint64), amount)

	config = ScaledUiAmountState{
		Multiplier:                      math.MaxFloat64,
		NewMultiplierEffectiveTimestamp: 1,
	}
	// scientific notation "e"
	amount, err = config.UiAmountToAmount("1.7976931348623157e308", 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(1), amount)

	config = ScaledUiAmountState{
		Multiplier:                      9.745314011399998e288,
		NewMultiplierEffectiveTimestamp: 1,
	}
	amount, err = config.UiAmountToAmount("1.7976931348623157e308", 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(math.MaxUint64), amount)
	// scientific notation "E"
	amount, err = config.UiAmountToAmount("1.7976931348623157E308", 0, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(math.MaxUint64), amount)

	// this is unfortunate, but underflows can happen due to floats
	config = ScaledUiAmountState{
		Multiplier:                      1.0,
		NewMultiplierEffectiveTimestamp: 1,
	}
	amount, err = config.UiAmountToAmount("18446744073709551616", 0, 0) // u64::MAX + 1
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(math.MaxUint64), amount)

	// overflow u64 fail
	config = ScaledUiAmountState{
		Multiplier:                      0.1,
		NewMultiplierEffectiveTimestamp: 1,
	}
	_, err = config.UiAmountToAmount("18446744073709551615", 0, 0)
	ag_require.ErrorIs(t, err, ErrInvalidUiAmount)

	for _, failUiAmount := range []string{"-0.0000000000000000000001", "inf", "-inf", "NaN"} {
		_, err = config.UiAmountToAmount(failUiAmount, 0, 0)
		ag_require.ErrorIs(t, err, ErrInvalidUiAmount)
	}

	// truncation
	config = ScaledUiAmountState{
		Multiplier:                      0.99,
		NewMultiplierEffectiveTimestamp: 1,
	}
	// There are a few possibilities for what "0.99" means, it could be 101
	// or 100 underlying tokens, but the result gives the fewest possible
	// tokens that give that UI amount.
	amount, err = config.UiAmountToAmount("0.99", 2, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(100), amount)
}

func TestScaledUiAmount_SpecificAmountToUiAmountNoScale(t *testing.T) {
	config := ScaledUiAmountState{
		Multiplier:                      1.0,
		NewMultiplierEffectiveTimestamp: 1,
	}
	for _, tc := range []struct {
		amount   uint64
		expected string
	}{
		{23, "0.23"}, {110, "1.1"}, {4200, "42"}, {0, "0"},
	} {
		ui, err := config.AmountToUiAmount(tc.amount, testDecimals, 0)
		ag_require.NoError(t, err)
		ag_require.Equal(t, tc.expected, ui)
	}
}

func TestScaledUiAmount_SpecificUiAmountToAmountNoScale(t *testing.T) {
	config := ScaledUiAmountState{
		Multiplier:                      1.0,
		NewMultiplierEffectiveTimestamp: 1,
	}
	for _, tc := range []struct {
		uiAmount string
		expected uint64
	}{
		{"0.23", 23}, {"0.20", 20}, {"0.2000", 20}, {".2", 20},
		{"1.1", 110}, {"1.10", 110}, {"42", 4200}, {"42.", 4200}, {"0", 0},
	} {
		amount, err := config.UiAmountToAmount(tc.uiAmount, testDecimals, 0)
		ag_require.NoError(t, err)
		ag_require.Equal(t, tc.expected, amount)
	}

	// this is invalid with normal mints, but truncation for this mint makes it ok
	amount, err := config.UiAmountToAmount("0.111", testDecimals, 0)
	ag_require.NoError(t, err)
	ag_require.Equal(t, uint64(11), amount)

	// fail if invalid ui_amount passed in
	for _, uiAmount := range []string{"", ".", "0.t"} {
		_, err := config.UiAmountToAmount(uiAmount, testDecimals, 0)
		ag_require.ErrorIs(t, err, ErrInvalidUiAmount)
	}
}
