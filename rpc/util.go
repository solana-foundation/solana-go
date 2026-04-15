package rpc

import "github.com/gagliardetto/solana-go"

func IsTokenMint(acc *Account) bool {
	data := acc.Data.GetBinary()
	n := len(data)

	switch acc.Owner {
	case solana.TokenProgramID:
		return n == 82
	case solana.Token2022ProgramID:
		if n == 82 {
			return true //Normal Mint
		}
		if n <= 165 {
			return false //Normal Token Account
		}
		return data[165] == 1 // Mint Extensions
	}

	return false
}
