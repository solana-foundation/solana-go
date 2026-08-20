package base58

// testAlphabet is kept local to the regression suite so the oracle does not
// share lookup tables or implementation code with github.com/fluxrpc/base58.
const testAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// encodeVariable is a deliberately simple long-division Base58 encoder used
// as an independent oracle for the fixed-width optimized paths.
func encodeVariable(src []byte) string {
	zeros := 0
	for zeros < len(src) && src[zeros] == 0 {
		zeros++
	}

	// ceil(n * log(256) / log(58)), with integer arithmetic and slack.
	size := (len(src)-zeros)*138/100 + 1
	digits := make([]byte, size)
	high := size - 1

	for i := zeros; i < len(src); i++ {
		j := size - 1
		for carry := uint32(src[i]); j > high || carry != 0; j-- {
			carry += 256 * uint32(digits[j])
			digits[j] = byte(carry % 58)
			carry /= 58
			if j == 0 {
				break
			}
		}
		high = j
	}

	first := 0
	for first < size && digits[first] == 0 {
		first++
	}

	out := make([]byte, zeros+size-first)
	for i := range zeros {
		out[i] = testAlphabet[0]
	}
	for i := zeros; first < size; i++ {
		out[i] = testAlphabet[digits[first]]
		first++
	}
	return string(out)
}
