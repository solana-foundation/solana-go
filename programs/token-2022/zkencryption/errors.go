package zkencryption

import "errors"

var (
	ErrSeedTooShort          = errors.New("zkencryption: seed is too short")
	ErrSeedTooLong           = errors.New("zkencryption: seed is too long")
	ErrDefaultSignature      = errors.New("zkencryption: refusing to derive key from default (all-zero) signature")
	ErrInvalidScalarEncoding = errors.New("zkencryption: scalar wide-reduction failed")
)
