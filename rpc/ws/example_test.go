package ws_test

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func ExampleClient_AccountSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}
	program := solana.MustPublicKeyFromBase58("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin") // serum

	sub, err := client.AccountSubscribe(
		program,
		"",
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_AccountSubscribeWithOpts() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}
	program := solana.MustPublicKeyFromBase58("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin") // serum

	// You can specify the data encoding of the returned accounts:
	sub, err := client.AccountSubscribeWithOpts(
		program,
		"",
		solana.EncodingBase64,
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_BlockSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.TestNet_WS)
	if err != nil {
		panic(err)
	}

	// NOTE: This subscription is unstable and only available if the validator
	// was started with the --rpc-pubsub-enable-block-subscription flag.
	sub, err := client.BlockSubscribe(
		ws.NewBlockSubscribeFilterAll(),
		&ws.BlockSubscribeOpts{
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_LogsSubscribeMentions() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}
	program := solana.MustPublicKeyFromBase58("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin") // serum

	// Subscribe to log events that mention the provided pubkey:
	sub, err := client.LogsSubscribeMentions(
		program,
		rpc.CommitmentProcessed,
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_LogsSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}

	// Subscribe to all log events:
	sub, err := client.LogsSubscribe(
		ws.LogsSubscribeFilterAll,
		rpc.CommitmentProcessed,
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_ProgramSubscribeWithOpts() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}
	program := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA") // token

	sub, err := client.ProgramSubscribeWithOpts(
		program,
		rpc.CommitmentProcessed,
		solana.EncodingBase64Zstd,
		nil,
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)

		decodedBinary := got.Value.Account.Data.GetBinary()
		if decodedBinary != nil {
			// spew.Dump(decodedBinary)
		}

		// or if you requested solana.EncodingJSONParsed and it is supported:
		rawJSON := got.Value.Account.Data.GetRawJSON()
		if rawJSON != nil {
			// spew.Dump(rawJSON)
		}
	}
}

func ExampleClient_RootSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.TestNet_WS)
	if err != nil {
		panic(err)
	}

	sub, err := client.RootSubscribe()
	if err != nil {
		panic(err)
	}

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_SignatureSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.TestNet_WS)
	if err != nil {
		panic(err)
	}

	txSig := solana.MustSignatureFromBase58("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	sub, err := client.SignatureSubscribe(
		txSig,
		"",
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_SlotSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.TestNet_WS)
	if err != nil {
		panic(err)
	}

	sub, err := client.SlotSubscribe()
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_SlotsUpdatesSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.TestNet_WS)
	if err != nil {
		panic(err)
	}

	// NOTE: This subscription is unstable; the format of this subscription
	// may change in the future and it may not always be supported.
	sub, err := client.SlotsUpdatesSubscribe()
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}

func ExampleClient_VoteSubscribe() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}

	// NOTE: this subscription must be enabled by the node you're connecting to.
	// This subscription is disabled by default.
	sub, err := client.VoteSubscribe()
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		spew.Dump(got)
	}
}
