package token2022_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"

	token2022 "github.com/gagliardetto/solana-go/programs/token-2022"
)

// xStockMintAccountData is the on-chain account data of the token-2022 mint
// Xs3oZwbHvqis4NYcf4YKWmEia2eC84wSiVrcYcTqpH8 (SpaceX xStock), as returned by
// getAccountInfo with base64 encoding. The mint carries eight extensions:
// MetadataPointer, PermanentDelegate, DefaultAccountState, ScaledUiAmount,
// Pausable, ConfidentialTransferMint, TransferHook, and TokenMetadata.
const xStockMintAccountData = "AQAAAGVqQkIv6okUBqQZ0dHeCPQqhHlBtaGulevOYZrDFyk0XOODfu4yAAAIAQEAAAD/3+wbzSzTg5PITaoIyRzA041nf/jQq3tdAz8A9zLMMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARIAQABD+fHuLje4B+pFy3TmAJcivsAJKWAm5ORVi0NeJKXpxgfoA5u0rUz3nDg+Q8DFHdGFN0BRK3nQ+UISkjRR0AODDAAgAEP58e4uN7gH6kXLdOYAlyK+wAkpYCbk5FWLQ14kpenGBgABAAEZADgABm9ZIlHMR3R4JaWa0UIupDVz9SjaXe4q94ErMU+ZReMAAAAAAADwPwAAAAAAAAAAAAAAAAAA8D8aACEA/9/sG80s04OTyE2qCMkcwNONZ3/40Kt7XQM/APcyzDAABABBAEP58e4uN7gH6kXLdOYAlyK+wAkpYCbk5FWLQ14kpenGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADgBAAEP58e4uN7gH6kXLdOYAlyK+wAkpYCbk5FWLQ14kpenGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAATAKYAQ/nx7i43uAfqRct05gCXIr7ACSlgJuTkVYtDXiSl6cYH6AObtK1M95w4PkPAxR3RhTdAUSt50PlCEpI0UdADgw0AAABTcGFjZVggeFN0b2NrBQAAAFNQQ1h4RAAAAGh0dHBzOi8veHN0b2Nrcy1tZXRhZGF0YS5iYWNrZWQuZmkvdG9rZW5zL1NvbGFuYS9TUENYeC9tZXRhZGF0YS5qc29uAAAAAA=="

// Decode a real mainnet token-2022 mint into a typed struct, walk every one
// of its eight extensions (in their on-chain TLV order), and re-encode the
// account byte-exactly. Extension fields are nil when an extension is absent;
// optional pubkeys inside extensions return nil from Get() when unset.
func ExampleDecodeMintWithExtensions() {
	data, err := base64.StdEncoding.DecodeString(xStockMintAccountData)
	if err != nil {
		log.Fatal(err)
	}

	mint, err := token2022.DecodeMintWithExtensions(data)
	if err != nil {
		log.Fatal(err)
	}

	// Base mint state (the same 82-byte layout as the original SPL token).
	fmt.Println("== base mint ==")
	fmt.Println("supply:", mint.Mint.Supply)
	fmt.Println("decimals:", mint.Mint.Decimals)
	fmt.Println("mint authority:", mint.Mint.MintAuthority)
	fmt.Println("freeze authority:", mint.Mint.FreezeAuthority)

	fmt.Println("== MetadataPointer ==")
	fmt.Println("authority:", mint.MetadataPointer.Authority.Get())
	fmt.Println("metadata address:", mint.MetadataPointer.MetadataAddress.Get())

	fmt.Println("== PermanentDelegate ==")
	fmt.Println("delegate:", mint.PermanentDelegate.Delegate.Get())

	fmt.Println("== DefaultAccountState ==")
	fmt.Println("state for new token accounts:", mint.DefaultAccountState.State)

	fmt.Println("== ScaledUiAmount ==")
	fmt.Println("authority:", mint.ScaledUiAmount.Authority.Get())
	fmt.Println("multiplier:", mint.ScaledUiAmount.Multiplier)
	fmt.Println("new multiplier:", mint.ScaledUiAmount.NewMultiplier,
		"effective at unix time", mint.ScaledUiAmount.NewMultiplierEffectiveTimestamp)
	// The extension also converts raw amounts to UI amounts (unix timestamp
	// 0 already selects the new 1.0 multiplier here).
	uiAmount, err := mint.ScaledUiAmount.AmountToUiAmount(mint.Mint.Supply, mint.Mint.Decimals, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ui supply:", uiAmount)

	fmt.Println("== Pausable ==")
	fmt.Println("authority:", mint.Pausable.Authority.Get())
	fmt.Println("paused:", mint.Pausable.Paused)

	fmt.Println("== ConfidentialTransferMint ==")
	fmt.Println("authority:", mint.ConfidentialTransferMint.Authority.Get())
	fmt.Println("auto approve new accounts:", mint.ConfidentialTransferMint.AutoApproveNewAccounts)
	fmt.Println("auditor configured:", mint.ConfidentialTransferMint.AuditorElGamalPubkey != [32]byte{})

	fmt.Println("== TransferHook ==")
	fmt.Println("authority:", mint.TransferHook.Authority.Get())
	fmt.Println("hook program:", mint.TransferHook.ProgramID.Get())

	fmt.Println("== TokenMetadata ==")
	fmt.Println("update authority:", mint.TokenMetadata.UpdateAuthority.Get())
	fmt.Println("mint:", mint.TokenMetadata.Mint)
	fmt.Println("name:", mint.TokenMetadata.Name)
	fmt.Println("symbol:", mint.TokenMetadata.Symbol)
	fmt.Println("uri:", mint.TokenMetadata.Uri)
	fmt.Println("additional metadata fields:", len(mint.TokenMetadata.AdditionalMetadata))

	// MarshalBinary reproduces the on-chain bytes exactly, preserving the
	// original TLV extension order.
	encoded, err := mint.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("round-trip byte-exact:", bytes.Equal(encoded, data))

	// Output:
	// == base mint ==
	// supply: 55999906177884
	// decimals: 8
	// mint authority: 7pt9tkctJPK7PPNQJ77GKg8ZffSF6QxoMiCFYHxrtaCj
	// freeze authority: JDq14BWvqCRFNu1krb12bcRpbGtJZ1FLEakMw6FdxJNs
	// == MetadataPointer ==
	// authority: 5aMNNLQJwAEeoemTEMkv5NVjqKwvvefRYCQ5Z67HFvEq
	// metadata address: Xs3oZwbHvqis4NYcf4YKWmEia2eC84wSiVrcYcTqpH8
	// == PermanentDelegate ==
	// delegate: 5aMNNLQJwAEeoemTEMkv5NVjqKwvvefRYCQ5Z67HFvEq
	// == DefaultAccountState ==
	// state for new token accounts: Initialized
	// == ScaledUiAmount ==
	// authority: S7vYFFWH6BjJyEsdrPQpqpYTqLTrPRK6KW3VwsJuRaS
	// multiplier: 1
	// new multiplier: 1 effective at unix time 0
	// ui supply: 559999.06177884
	// == Pausable ==
	// authority: JDq14BWvqCRFNu1krb12bcRpbGtJZ1FLEakMw6FdxJNs
	// paused: false
	// == ConfidentialTransferMint ==
	// authority: 5aMNNLQJwAEeoemTEMkv5NVjqKwvvefRYCQ5Z67HFvEq
	// auto approve new accounts: false
	// auditor configured: false
	// == TransferHook ==
	// authority: 5aMNNLQJwAEeoemTEMkv5NVjqKwvvefRYCQ5Z67HFvEq
	// hook program: <nil>
	// == TokenMetadata ==
	// update authority: 5aMNNLQJwAEeoemTEMkv5NVjqKwvvefRYCQ5Z67HFvEq
	// mint: Xs3oZwbHvqis4NYcf4YKWmEia2eC84wSiVrcYcTqpH8
	// name: SpaceX xStock
	// symbol: SPCXx
	// uri: https://xstocks-metadata.backed.fi/tokens/Solana/SPCXx/metadata.json
	// additional metadata fields: 0
	// round-trip byte-exact: true
}
