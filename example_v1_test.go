package solana_test

import (
	"crypto/ed25519"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// ExampleNewTransaction_v1 builds, signs, serializes and decodes a V1
// (SIMD-0385) transaction with an inline compute budget configuration.
func ExampleNewTransaction_v1() {
	// Deterministic keys for the example; use solana.NewWallet() in real code.
	payerKey := solana.PrivateKey(ed25519.NewKeyFromSeed(make([]byte, 32)))
	payer := payerKey.PublicKey()
	recipient := solana.MustPublicKeyFromBase58("2mHtsPqiHkQKKh6t2Q1jGwYQ8vG7ULfF7c9k4t9BvGkw")
	programID := solana.SystemProgramID

	// Any Instruction works; this is a raw System Program transfer of 1 lamport.
	data := []byte{2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}
	ix := solana.NewInstruction(programID, solana.AccountMetaSlice{
		solana.Meta(payer).WRITE().SIGNER(),
		solana.Meta(recipient).WRITE(),
	}, data)

	// A zero blockhash for the example; fetch it from the RPC in real code.
	var blockhash solana.Hash

	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix},
		blockhash,
		solana.TransactionPayer(payer),
		// Selects v1; unset CU limit means 0, PriorityFee is total lamports.
		solana.TransactionV1Config(solana.TransactionConfig{}.
			WithComputeUnitLimit(20_000).
			WithPriorityFee(1_000)),
	)
	if err != nil {
		panic(err)
	}
	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(payer) {
			return &payerKey
		}
		return nil
	}); err != nil {
		panic(err)
	}

	// Wire: 0x81 prefix first, signatures last (no length prefix).
	wire, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	decoded, err := solana.TransactionFromBytes(wire)
	if err != nil {
		panic(err)
	}
	fmt.Println("version:", decoded.Message.GetVersion() == solana.MessageVersionV1)
	fmt.Println("prefix: 0x" + fmt.Sprintf("%02x", wire[0]))
	fmt.Println("compute unit limit:", *decoded.Message.TransactionConfig.ComputeUnitLimit)
	fmt.Println("priority fee:", *decoded.Message.TransactionConfig.PriorityFee)
	fmt.Println("heap size requested:", decoded.Message.TransactionConfig.HeapSize != nil)
	fmt.Println("size ok:", len(wire) <= solana.MaxTransactionSizeV1)
	fmt.Println("signatures valid:", decoded.VerifySignatures() == nil)
	// Output:
	// version: true
	// prefix: 0x81
	// compute unit limit: 20000
	// priority fee: 1000
	// heap size requested: false
	// size ok: true
	// signatures valid: true
}
