package token2022

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

// This file ports the extension helper math from the Rust
// spl-token-2022-interface crate: transfer-fee calculation, interest-bearing
// and scaled-UI-amount conversions. The implementations mirror the Rust
// semantics exactly, including f64 rounding behavior, ceiling division in
// 128-bit arithmetic, and saturating float-to-u64 casts.

// MaxFeeBasisPoints is the maximum possible fee in basis points: 100%.
const MaxFeeBasisPoints uint16 = 10_000

const (
	oneInBasisPoints = uint64(MaxFeeBasisPoints)
	// secondsPerYear mirrors the Rust constant: 60 * 60 * 24 * 365.24.
	secondsPerYear = 31_556_736.0
)

// maxU64AsF64 is u64::MAX as f64 in Rust: 2^64 (the nearest representable value).
const maxU64AsF64 = float64(1 << 64)

// --- shared helpers ---

// ceilDiv128 returns ceil((hi,lo) / d) as a 128-bit quotient, mirroring the
// Rust TransferFee::ceil_div over u128. d must be nonzero.
func ceilDiv128(hi, lo, d uint64) (qhi, qlo uint64) {
	lo, carry := bits.Add64(lo, d-1, 0)
	hi += carry // cannot overflow: hi < 2^64-1 for all callers
	qhi = hi / d
	rem := hi % d
	qlo, _ = bits.Div64(rem, lo, d)
	return qhi, qlo
}

// checkedSubI64 mirrors Rust i64::checked_sub.
func checkedSubI64(a, b int64) (int64, bool) {
	diff := a - b
	if (b > 0 && diff > a) || (b < 0 && diff < a) {
		return 0, false
	}
	return diff, true
}

// saturatingF64ToU64 mirrors the Rust `as u64` cast: values of 2^64 or above
// saturate to u64::MAX, negative values and NaN become 0. The explicit NaN
// branch matters: a bare Go conversion of NaN is implementation-defined.
func saturatingF64ToU64(v float64) uint64 {
	if math.IsNaN(v) {
		return 0
	}
	if v >= maxU64AsF64 {
		return math.MaxUint64
	}
	if v <= 0 {
		return 0
	}
	return uint64(v)
}

// formatFixedF64 formats v with the given number of fraction digits, matching
// Rust's format!("{:.N}", v) (including "inf"/"-inf"/"NaN" spellings).
func formatFixedF64(v float64, decimals uint8) string {
	switch {
	case math.IsInf(v, 1):
		return "inf"
	case math.IsInf(v, -1):
		return "-inf"
	case math.IsNaN(v):
		return "NaN"
	}
	return strconv.FormatFloat(v, 'f', int(decimals), 64)
}

// trimUiAmountString mirrors the Rust trim_ui_amount_string helper: for a
// nonzero decimals value, excess trailing zeroes and an unneeded decimal
// point are trimmed.
func trimUiAmountString(uiAmount string, decimals uint8) string {
	if decimals > 0 {
		uiAmount = strings.TrimRight(uiAmount, "0")
		uiAmount = strings.TrimRight(uiAmount, ".")
	}
	return uiAmount
}

// parseUiAmountF64 parses a UI amount string the way Rust's f64::from_str
// does. Go's strconv.ParseFloat additionally accepts digit-separating
// underscores ("1_000") and hexadecimal floats ("0x1p4"), which Rust rejects;
// those are filtered out for parity. (A "p" exponent is only reachable
// through the hex prefix, so rejecting "x" covers it.)
func parseUiAmountF64(uiAmount string) (float64, error) {
	if strings.ContainsAny(uiAmount, "_xX") {
		return 0, fmt.Errorf("%w: %q", ErrInvalidUiAmount, uiAmount)
	}
	v, err := strconv.ParseFloat(uiAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidUiAmount, uiAmount)
	}
	return v, nil
}

// --- TransferFee math ---

// CalculateFee calculates the transfer fee for the given pre-fee amount,
// rounding up and capping at MaximumFee.
func (f TransferFee) CalculateFee(preFeeAmount uint64) (uint64, error) {
	bps := uint64(f.TransferFeeBasisPoints)
	if bps == 0 || preFeeAmount == 0 {
		return 0, nil
	}
	hi, lo := bits.Mul64(preFeeAmount, bps)
	qhi, qlo := ceilDiv128(hi, lo, oneInBasisPoints)
	if qhi != 0 {
		return 0, ErrCalculationOverflow
	}
	return min(qlo, f.MaximumFee), nil
}

// CalculatePostFeeAmount calculates the gross transfer amount after deducting fees.
func (f TransferFee) CalculatePostFeeAmount(preFeeAmount uint64) (uint64, error) {
	fee, err := f.CalculateFee(preFeeAmount)
	if err != nil {
		return 0, err
	}
	if fee > preFeeAmount {
		return 0, ErrCalculationOverflow
	}
	return preFeeAmount - fee, nil
}

// CalculatePreFeeAmount calculates the transfer amount that will result in
// the specified net transfer amount.
//
// The original transfer amount may not always be unique due to rounding; in
// that case the smaller amount is chosen. On large net transfer amounts the
// original amount may overflow, in which case an error is returned.
func (f TransferFee) CalculatePreFeeAmount(postFeeAmount uint64) (uint64, error) {
	maximumFee := f.MaximumFee
	bps := uint64(f.TransferFeeBasisPoints)
	switch {
	case bps == 0:
		// no fee, same amount
		return postFeeAmount, nil
	case postFeeAmount == 0:
		// 0 out, 0 in
		return 0, nil
	case bps == oneInBasisPoints:
		// 100%, cap at max fee
		sum, carry := bits.Add64(maximumFee, postFeeAmount, 0)
		if carry != 0 {
			return 0, ErrCalculationOverflow
		}
		return sum, nil
	default:
		if bps > oneInBasisPoints {
			// denominator would underflow (checked_sub in Rust)
			return 0, ErrCalculationOverflow
		}
		hi, lo := bits.Mul64(postFeeAmount, oneInBasisPoints)
		qhi, qlo := ceilDiv128(hi, lo, oneInBasisPoints-bps)
		// (raw_pre_fee_amount - post_fee_amount) >= maximum_fee, in 128 bits.
		dlo, borrow := bits.Sub64(qlo, postFeeAmount, 0)
		dhi, borrow := bits.Sub64(qhi, 0, borrow)
		if borrow != 0 {
			return 0, ErrCalculationOverflow
		}
		if dhi > 0 || dlo >= maximumFee {
			sum, carry := bits.Add64(postFeeAmount, maximumFee, 0)
			if carry != 0 {
				return 0, ErrCalculationOverflow
			}
			return sum, nil
		}
		if qhi != 0 {
			// pre-fee amount does not fit in u64
			return 0, ErrCalculationOverflow
		}
		return qlo, nil
	}
}

// CalculateInverseFee calculates the fee that would produce the given output.
//
// Note: this is not an exact inverse of CalculateFee; only
// CalculateFee(x) >= CalculateInverseFee(x - CalculateFee(x)) holds.
func (f TransferFee) CalculateInverseFee(postFeeAmount uint64) (uint64, error) {
	preFeeAmount, err := f.CalculatePreFeeAmount(postFeeAmount)
	if err != nil {
		return 0, err
	}
	return f.CalculateFee(preFeeAmount)
}

// GetEpochFee returns the fee schedule active in the given epoch.
func (s TransferFeeConfigState) GetEpochFee(epoch uint64) TransferFee {
	if epoch >= s.NewerTransferFee.Epoch {
		return s.NewerTransferFee
	}
	return s.OlderTransferFee
}

// CalculateEpochFee calculates the fee for the given epoch and input amount.
func (s TransferFeeConfigState) CalculateEpochFee(epoch uint64, preFeeAmount uint64) (uint64, error) {
	return s.GetEpochFee(epoch).CalculateFee(preFeeAmount)
}

// CalculateInverseEpochFee calculates the fee for the given epoch and output amount.
func (s TransferFeeConfigState) CalculateInverseEpochFee(epoch uint64, postFeeAmount uint64) (uint64, error) {
	return s.GetEpochFee(epoch).CalculateInverseFee(postFeeAmount)
}

// --- InterestBearingConfig math ---

func (s InterestBearingConfigState) preUpdateExp() (float64, bool) {
	timespan, ok := checkedSubI64(s.LastUpdateTimestamp, s.InitializationTimestamp)
	if !ok {
		return 0, false
	}
	exponent := float64(s.PreUpdateAverageRate) * float64(timespan) / secondsPerYear / float64(oneInBasisPoints)
	return math.Exp(exponent), true
}

func (s InterestBearingConfigState) postUpdateExp(unixTimestamp int64) (float64, bool) {
	timespan, ok := checkedSubI64(unixTimestamp, s.LastUpdateTimestamp)
	if !ok {
		return 0, false
	}
	exponent := float64(s.CurrentRate) * float64(timespan) / secondsPerYear / float64(oneInBasisPoints)
	return math.Exp(exponent), true
}

// totalScale mirrors the Rust total_scale. Note: Go's math.Pow10 is
// correctly rounded, while Rust's 10_f64.powi can double-round for exponents
// where 10^n is not exactly representable (23-27, 29-31); results may differ
// in the last ulp for such uncommon decimals values.
func (s InterestBearingConfigState) totalScale(decimals uint8, unixTimestamp int64) (float64, bool) {
	pre, ok := s.preUpdateExp()
	if !ok {
		return 0, false
	}
	post, ok := s.postUpdateExp(unixTimestamp)
	if !ok {
		return 0, false
	}
	return pre * post / math.Pow10(int(decimals)), true
}

// AmountToUiAmount converts a raw amount to its UI representation, accruing
// continuously-compounded interest up to the given unix timestamp. Excess
// zeroes and an unneeded decimal point are trimmed.
func (s InterestBearingConfigState) AmountToUiAmount(amount uint64, decimals uint8, unixTimestamp int64) (string, error) {
	scale, ok := s.totalScale(decimals, unixTimestamp)
	if !ok {
		return "", ErrCalculationOverflow
	}
	scaledAmountWithInterest := float64(amount) * scale
	uiAmount := formatFixedF64(scaledAmountWithInterest, decimals)
	return trimUiAmountString(uiAmount, decimals), nil
}

// UiAmountToAmount converts a UI representation of a token amount back to its
// raw amount, mirroring the Rust try_ui_amount_into_amount (the result is
// rounded and saturates at u64::MAX).
func (s InterestBearingConfigState) UiAmountToAmount(uiAmount string, decimals uint8, unixTimestamp int64) (uint64, error) {
	scaledAmount, err := parseUiAmountF64(uiAmount)
	if err != nil {
		return 0, err
	}
	scale, ok := s.totalScale(decimals, unixTimestamp)
	if !ok {
		return 0, ErrCalculationOverflow
	}
	amount := scaledAmount / scale
	if amount > maxU64AsF64 || amount < 0 || math.IsNaN(amount) {
		return 0, fmt.Errorf("%w: %q", ErrInvalidUiAmount, uiAmount)
	}
	// Rounding must happen last; rounding earlier gives wrong "inf" answers.
	return saturatingF64ToU64(math.Round(amount)), nil
}

// TimeWeightedAverageRate returns the time-weighted average of the current
// and average rates, in basis points.
func (s InterestBearingConfigState) TimeWeightedAverageRate(currentTimestamp int64) (int16, error) {
	// Mirrors the Rust i128 arithmetic exactly via big.Int: overflow is
	// impossible mid-computation, only the final i16 conversion can fail.
	r1 := big.NewInt(int64(s.PreUpdateAverageRate))
	t1 := new(big.Int).Sub(big.NewInt(s.LastUpdateTimestamp), big.NewInt(s.InitializationTimestamp))
	r2 := big.NewInt(int64(s.CurrentRate))
	t2 := new(big.Int).Sub(big.NewInt(currentTimestamp), big.NewInt(s.LastUpdateTimestamp))
	totalTimespan := new(big.Int).Add(t1, t2)

	var averageRate *big.Int
	if totalTimespan.Sign() == 0 {
		// Happens in testing situations; just use the new rate since the
		// earlier one was never practically used.
		averageRate = r2
	} else {
		sum := new(big.Int).Add(new(big.Int).Mul(r1, t1), new(big.Int).Mul(r2, t2))
		averageRate = new(big.Int).Quo(sum, totalTimespan)
	}
	if !averageRate.IsInt64() {
		return 0, ErrCalculationOverflow
	}
	v := averageRate.Int64()
	if v < math.MinInt16 || v > math.MaxInt16 {
		return 0, ErrCalculationOverflow
	}
	return int16(v), nil
}

// --- ScaledUiAmountConfig math ---

// CurrentMultiplier returns the multiplier active at the given unix timestamp.
func (s ScaledUiAmountState) CurrentMultiplier(unixTimestamp int64) float64 {
	if unixTimestamp >= s.NewMultiplierEffectiveTimestamp {
		return s.NewMultiplier
	}
	return s.Multiplier
}

// totalMultiplier mirrors the Rust total_multiplier; see the math.Pow10
// rounding note on InterestBearingConfigState.totalScale.
func (s ScaledUiAmountState) totalMultiplier(decimals uint8, unixTimestamp int64) float64 {
	return s.CurrentMultiplier(unixTimestamp) / math.Pow10(int(decimals))
}

// AmountToUiAmount converts a raw amount to its UI representation using the
// given decimals field. The value is converted to a float and truncated
// towards zero; excess zeroes and an unneeded decimal point are trimmed.
func (s ScaledUiAmountState) AmountToUiAmount(amount uint64, decimals uint8, unixTimestamp int64) (string, error) {
	scaledAmount := float64(amount) * s.CurrentMultiplier(unixTimestamp)
	truncatedAmount := math.Trunc(scaledAmount) / math.Pow10(int(decimals))
	uiAmount := formatFixedF64(truncatedAmount, decimals)
	return trimUiAmountString(uiAmount, decimals), nil
}

// UiAmountToAmount converts a UI representation of a token amount back to its
// raw amount. The string is parsed to a float, scaled, and truncated towards
// zero (saturating at u64::MAX), mirroring the Rust try_ui_amount_into_amount.
func (s ScaledUiAmountState) UiAmountToAmount(uiAmount string, decimals uint8, unixTimestamp int64) (uint64, error) {
	scaledAmount, err := parseUiAmountF64(uiAmount)
	if err != nil {
		return 0, err
	}
	amount := scaledAmount / s.totalMultiplier(decimals, unixTimestamp)
	if amount > maxU64AsF64 || amount < 0 || math.IsNaN(amount) {
		return 0, fmt.Errorf("%w: %q", ErrInvalidUiAmount, uiAmount)
	}
	// Truncation must happen last; truncating earlier gives wrong "inf" answers.
	return saturatingF64ToU64(math.Trunc(amount)), nil
}
