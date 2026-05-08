package rpc

import "github.com/solana-foundation/solana-go/v2"

// Redefined locally to avoid a cycle with the token2022 sub-package, which
// imports rpc. Canonical copies: token2022.MINT_SIZE / ACCOUNT_SIZE /
// AccountTypeMint.
const (
	tokenMintSize            = 82
	token2022AccountBaseSize = 165
	token2022AccountTypeMint = 1
)

// IsTokenMint reports whether acc holds an SPL Token or Token-2022 Mint,
// classifying by byte layout instead of a full borsh decode. Safe on nil
// acc and nil acc.Data.
//
// Token-2022 pads Mint records from 82 to 165 bytes and places a 1-byte
// AccountType discriminator (1 = Mint, 2 = Account) at offset 165 when
// extensions are present.
func IsTokenMint(acc *Account) bool {
	if acc == nil {
		return false
	}
	data := acc.Data.GetBinary()
	n := len(data)

	// Length check first: a 32-byte PublicKey compare costs more than a
	// length compare, so reject non-mint shapes before touching Owner.
	if n == tokenMintSize {
		return acc.Owner == solana.TokenProgramID ||
			acc.Owner == solana.Token2022ProgramID
	}
	if n > token2022AccountBaseSize &&
		data[token2022AccountBaseSize] == token2022AccountTypeMint {
		return acc.Owner == solana.Token2022ProgramID
	}
	return false
}
