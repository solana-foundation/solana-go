package rpc_test

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/token"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func ExampleClient_GetAccountInfo() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	{
		pubKey := solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt") // serum token
		// basic usage
		resp, err := client.GetAccountInfo(
			context.TODO(),
			pubKey,
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(resp)

		var mint token.Mint
		// Account{}.Data.GetBinary() returns the *decoded* binary data
		// regardless the original encoding (it can handle them all).
		err = bin.NewBinDecoder(resp.GetBinary()).Decode(&mint)
		if err != nil {
			panic(err)
		}
		spew.Dump(mint)
		// NOTE: The supply is mint.Supply, with the mint.Decimals:
		// mint.Supply = 9998022451607088
		// mint.Decimals = 6
		// ... which means that the supply is 9998022451.607088
	}
	{
		// Or you can use `GetAccountDataInto` which does all of the above in one call:
		pubKey := solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt") // serum token
		var mint token.Mint
		// Get the account, and decode its data into the provided mint object:
		err := client.GetAccountDataInto(
			context.TODO(),
			pubKey,
			&mint,
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(mint)
	}
	{
		pubKey := solana.MustPublicKeyFromBase58("4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R") // raydium token
		// advanced usage
		resp, err := client.GetAccountInfoWithOpts(
			context.TODO(),
			pubKey,
			// You can specify more options here:
			&rpc.GetAccountInfoOpts{
				Encoding:   solana.EncodingBase64Zstd,
				Commitment: rpc.CommitmentFinalized,
				// You can get just a part of the account data by specify a DataSlice:
				// DataSlice: &rpc.DataSlice{
				//  Offset: pointer.ToUint64(0),
				//  Length: pointer.ToUint64(1024),
				// },
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(resp)

		var mint token.Mint
		err = bin.NewBinDecoder(resp.GetBinary()).Decode(&mint)
		if err != nil {
			panic(err)
		}
		spew.Dump(mint)
	}
}

func ExampleClient_GetBalance() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	out, err := client.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
	spew.Dump(out.Value) // total lamports on the account; 1 sol = 1000000000 lamports

	var lamportsOnAccount = new(big.Float).SetUint64(uint64(out.Value))
	// Convert lamports to sol:
	var solBalance = new(big.Float).Quo(lamportsOnAccount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	// WARNING: this is not a precise conversion.
	fmt.Println("◎", solBalance.Text('f', 10))
}

func ExampleClient_GetBlock() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	{
		out, err := client.GetBlock(context.TODO(), uint64(example.Context.Slot))
		if err != nil {
			panic(err)
		}
		// spew.Dump(out) // NOTE: This generates a lot of output.
		spew.Dump(len(out.Transactions))
	}

	{
		includeRewards := false
		out, err := client.GetBlockWithOpts(
			context.TODO(),
			uint64(example.Context.Slot),
			// You can specify more options here:
			&rpc.GetBlockOpts{
				Encoding:   solana.EncodingBase64,
				Commitment: rpc.CommitmentFinalized,
				// Get only signatures:
				TransactionDetails: rpc.TransactionDetailsSignatures,
				// Exclude rewards:
				Rewards: &includeRewards,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
}

func ExampleClient_GetBlockCommitment() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	out, err := client.GetBlockCommitment(
		context.TODO(),
		uint64(example.Context.Slot),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetBlockHeight() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetBlockHeight(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetBlockProduction() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	{
		out, err := client.GetBlockProduction(context.TODO())
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
	{
		out, err := client.GetBlockProductionWithOpts(
			context.TODO(),
			&rpc.GetBlockProductionOpts{
				Commitment: rpc.CommitmentFinalized,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
}

func ExampleClient_GetBlockTime() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetLatestBlockhash(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}

	out, err := client.GetBlockTime(
		context.TODO(),
		uint64(example.Context.Slot),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
	spew.Dump(out.Time().Format(time.RFC1123))
}

func ExampleClient_GetBlocks() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetLatestBlockhash(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}

	endSlot := uint64(example.Context.Slot)
	out, err := client.GetBlocks(
		context.TODO(),
		uint64(example.Context.Slot-3),
		&endSlot,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetBlocksWithLimit() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetLatestBlockhash(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}

	limit := uint64(4)
	out, err := client.GetBlocksWithLimit(
		context.TODO(),
		uint64(example.Context.Slot-10),
		limit,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetClusterNodes() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetClusterNodes(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetEpochInfo() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetEpochInfo(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetEpochSchedule() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetEpochSchedule(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetFeeForMessage() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetFeeForMessage(
		context.Background(),
		"AQABAgIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBAQAA",
		rpc.CommitmentProcessed,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(example)
}

func ExampleClient_GetFirstAvailableBlock() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetFirstAvailableBlock(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetGenesisHash() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetGenesisHash(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetHealth() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetHealth(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
	spew.Dump(out == rpc.HealthOk)
}

func ExampleClient_GetHighestSnapshotSlot() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetHighestSnapshotSlot(
		context.Background(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(example)
}

func ExampleClient_GetLatestBlockhash() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	example, err := client.GetLatestBlockhash(
		context.Background(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(example)
}

func ExampleClient_GetIdentity() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetIdentity(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetInflationGovernor() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetInflationGovernor(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetInflationRate() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetInflationRate(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetInflationReward() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("6dmNQ5jwLeLk5REvio1JcMshcbvkYMwy26sJ8pbkvStu")

	out, err := client.GetInflationReward(
		context.TODO(),
		[]solana.PublicKey{
			pubKey,
		},
		&rpc.GetInflationRewardOpts{
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetLargestAccounts() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetLargestAccounts(
		context.TODO(),
		rpc.CommitmentFinalized,
		rpc.LargestAccountsFilterCirculating,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetLeaderSchedule() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetLeaderSchedule(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out) // NOTE: this creates a lot of output
}

func ExampleClient_GetMaxRetransmitSlot() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetMaxRetransmitSlot(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetMaxShredInsertSlot() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetMaxShredInsertSlot(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetMinimumBalanceForRentExemption() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	dataSize := uint64(1024 * 9)
	out, err := client.GetMinimumBalanceForRentExemption(
		context.TODO(),
		dataSize,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetMultipleAccounts() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	{
		out, err := client.GetMultipleAccounts(
			context.TODO(),
			solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt"),  // serum token
			solana.MustPublicKeyFromBase58("4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R"), // raydium token
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
	{
		out, err := client.GetMultipleAccountsWithOpts(
			context.TODO(),
			[]solana.PublicKey{
				solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt"),  // serum token
				solana.MustPublicKeyFromBase58("4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R"), // raydium token
			},
			&rpc.GetMultipleAccountsOpts{
				Encoding:   solana.EncodingBase64Zstd,
				Commitment: rpc.CommitmentFinalized,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
	}
}

func ExampleClient_GetProgramAccounts() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetProgramAccounts(
		context.TODO(),
		solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(len(out))
}

func ExampleClient_GetRecentPerformanceSamples() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	limit := uint(3)
	out, err := client.GetRecentPerformanceSamples(
		context.TODO(),
		&limit,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetRecentPrioritizationFees() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetRecentPrioritizationFees(
		context.TODO(),
		[]solana.PublicKey{
			solana.MustPublicKeyFromBase58("q5BgreVhTyBH1QCeriVb7kQYEPneanFXPLjvyjdf8M3"),
		},
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetSignatureStatuses() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetSignatureStatuses(
		context.TODO(),
		true,
		// All the transactions you want the get the status for:
		solana.MustSignatureFromBase58("2CwH8SqVZWFa1EvsH7vJXGFors1NdCuWJ7Z85F8YqjCLQ2RuSHQyeGKkfo1Tj9HitSTeLoMWnxpjxF2WsCH8nGWh"),
		solana.MustSignatureFromBase58("5YJHZPeHZuZjhunBc1CCB1NDRNf2tTJNpdb3azGsR7PfyEncCDhr95wG8EWrvjNXBc4wCKixkheSbCxoC2NCG3X7"),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetSignaturesForAddress() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetSignaturesForAddress(
		context.TODO(),
		solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetSlot() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetSlot(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetSlotLeader() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetSlotLeader(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetSlotLeaders() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	recent, err := client.GetLatestBlockhash(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}

	out, err := client.GetSlotLeaders(
		context.TODO(),
		uint64(recent.Context.Slot),
		10,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetStakeMinimumDelegation() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetStakeMinimumDelegation(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetSupply() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetSupply(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetTokenAccountBalance() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("EzK5qLWhftu8Z2znVa5fozVtobbjhd8Gdu9hQHpC8bec")
	out, err := client.GetTokenAccountBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetTokenAccountsByDelegate() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("AfkALUPjQp8R1rUwE6KhT38NuTYWCncwwHwcJu7UtAfV")
	out, err := client.GetTokenAccountsByDelegate(
		context.TODO(),
		pubKey,
		&rpc.GetTokenAccountsConfig{
			Mint: solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112").ToPointer(),
		},
		nil,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetTokenAccountsByOwner() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("7HZaCWazgTuuFuajxaaxGYbGnyVKwxvsJKue1W4Nvyro")
	out, err := client.GetTokenAccountsByOwner(
		context.TODO(),
		pubKey,
		&rpc.GetTokenAccountsConfig{
			Mint: solana.WrappedSol.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{
			Encoding: solana.EncodingBase64Zstd,
		},
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)

	{
		tokenAccounts := make([]token.Account, 0)
		for _, rawAccount := range out.Value {
			var tokAcc token.Account

			data := rawAccount.Account.Data.GetBinary()
			dec := bin.NewBinDecoder(data)
			err := dec.Decode(&tokAcc)
			if err != nil {
				panic(err)
			}
			tokenAccounts = append(tokenAccounts, tokAcc)
		}
		spew.Dump(tokenAccounts)
	}
}

func ExampleClient_GetTokenLargestAccounts() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt") // serum token
	out, err := client.GetTokenLargestAccounts(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetTokenSupply() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58("SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt") // serum token
	out, err := client.GetTokenSupply(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetTransaction() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	txSig := solana.MustSignatureFromBase58("4bjVLV1g9SAfv7BSAdNnuSPRbSscADHFe4HegL6YVcuEBMY83edLEvtfjE4jfr6rwdLwKBQbaFiGgoLGtVicDzHq")
	{
		out, err := client.GetTransaction(
			context.TODO(),
			txSig,
			&rpc.GetTransactionOpts{
				Encoding: solana.EncodingBase64,
			},
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)
		spew.Dump(out.Transaction.GetBinary())

		decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
		if err != nil {
			panic(err)
		}
		spew.Dump(decodedTx)
	}
	{
		out, err := client.GetTransaction(
			context.TODO(),
			txSig,
			nil,
		)
		if err != nil {
			panic(err)
		}
		spew.Dump(out)

		decodedTx, err := out.Transaction.GetTransaction()
		if err != nil {
			panic(err)
		}
		spew.Dump(decodedTx)
	}
}

func ExampleClient_GetTransactionCount() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetTransactionCount(
		context.TODO(),
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetVersion() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetVersion(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_GetVoteAccounts() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.GetVoteAccounts(
		context.TODO(),
		&rpc.GetVoteAccountsOpts{
			VotePubkey: solana.MustPublicKeyFromBase58("vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu").ToPointer(),
		},
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_IsBlockhashValid() {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)

	blockHash := solana.MustHashFromBase58("J7rBdM6AecPDEZp8aPq5iPSNKVkU5Q76F3oAV4eW5wsW")
	out, err := client.IsBlockhashValid(
		context.TODO(),
		blockHash,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
	spew.Dump(out.Value) // true or false

	fmt.Println("is blockhash valid:", out.Value)
}

func ExampleClient_MinimumLedgerSlot() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	out, err := client.MinimumLedgerSlot(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}

func ExampleClient_RequestAirdrop() {
	endpoint := rpc.TestNet_RPC
	client := rpc.New(endpoint)

	amount := solana.LAMPORTS_PER_SOL // 1 sol
	pubKey := solana.MustPublicKeyFromBase58("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	out, err := client.RequestAirdrop(
		context.TODO(),
		pubKey,
		amount,
		"",
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
}
