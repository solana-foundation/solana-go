# Solana SDK library for Go

[![GoDoc](https://pkg.go.dev/badge/github.com/gagliardetto/solana-go?status.svg)](https://pkg.go.dev/github.com/gagliardetto/solana-go@v1.16.0?tab=doc)
[![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/gagliardetto/solana-go?include_prereleases&label=release-tag)](https://github.com/gagliardetto/solana-go/releases)
[![Build Status](https://github.com/gagliardetto/solana-go/workflows/tests/badge.svg?branch=main)](https://github.com/gagliardetto/solana-go/actions?query=branch%3Amain)
[![Lint Status](https://github.com/gagliardetto/solana-go/workflows/lint/badge.svg?branch=main)](https://github.com/gagliardetto/solana-go/actions?query=branch%3Amain+workflow%3Alint)
[![TODOs](https://badgen.net/https/api.tickgit.com/badgen/github.com/gagliardetto/solana-go/main)](https://www.tickgit.com/browse?repo=github.com/gagliardetto/solana-go&branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/gagliardetto/solana-go)](https://goreportcard.com/report/github.com/gagliardetto/solana-go)

Go library to interface with Solana JSON RPC and WebSocket interfaces.

More contracts to come.

**If you're using/developing Solana programs written in [Anchor Framework](https://github.com/coral-xyz/anchor), you can use [anchor-go](https://github.com/gagliardetto/anchor-go) to generate Golang clients**

<div align="center">
    <img src="https://user-images.githubusercontent.com/15271561/128235229-1d2d9116-23bb-464e-b2cc-8fb6355e3b55.png" margin="auto" height="175"/>
</div>

## Contents

- [Solana SDK library for Go](#solana-sdk-library-for-go)
  - [Contents](#contents)
  - [Features](#features)
  - [Current development status](#current-development-status)
  - [Note](#note)
  - [Requirements](#requirements)
  - [Installation](#installation)
  - [Pretty-Print transactions/instructions](#pretty-print-transactionsinstructions)
  - [SendAndConfirmTransaction](#sendandconfirmtransaction)
  - [Address Lookup Tables](#address-lookup-tables)
  - [Parse/decode an instruction from a transaction](#parsedecode-an-instruction-from-a-transaction)
  - [Borsh encoding/decoding](#borsh-encodingdecoding)
  - [ZSTD account data encoding](#zstd-account-data-encoding)
  - [Working with rate-limited RPC providers](#working-with-rate-limited-rpc-providers)
  - [Custom Headers for authenticating with RPC providers](#custom-headers-for-authenticating-with-rpc-providers)
  - [Timeouts and Custom HTTP Clients](#timeouts-and-custom-http-clients)
  - [Examples](#examples)
    - [Create account (wallet)](#create-account-wallet)
    - [Load/parse private and public keys](#loadparse-private-and-public-keys)
    - [Transfer Sol from one wallet to another wallet](#transfer-sol-from-one-wallet-to-another-wallet)
  - [RPC Methods](#rpc-methods)
  - [WebSocket Subscriptions](#websocket-subscriptions)
  - [Contributing](#contributing)
  - [License](#license)
  - [Credits](#credits)

## Features

- [x] [Full JSON RPC API](#rpc-usage-examples)
- [x] Full WebSocket JSON streaming API
- [ ] Wallet, account, and keys management
- [ ] Clients for native programs
  - [x] [system](/programs/system)
  - [ ] config
  - [ ] stake
  - [ ] vote
  - [x] BPF Loader
  - [ ] Secp256k1
- [ ] Clients for Solana Program Library (SPL)
  - [x] [SPL token](/programs/token)
  - [x] [associated-token-account](/programs/associated-token-account)
  - [x] memo
  - [ ] name-service
  - [ ] ...
- [x] [Metaplex](https://github.com/gagliardetto/metaplex-go):
  - [x] auction
  - [x] metaplex
  - [x] token-metadata
  - [x] token-vault
  - [x] nft-candy-machine
- [ ] More programs

## Current development status

There is currently **no stable release**. The SDK is actively developed and latest is `v1.16.0` which is an `alpha` release.

The RPC and WS client implementation is based on the [Solana RPC API documentation](https://solana.com/docs/rpc).

## Note

- solana-go is in active development, so all APIs are subject to change.
- This code is unaudited. Use at your own risk.

## Requirements

- Go 1.24 or later

## Installation

```bash
$ cd my-project
$ go get github.com/gagliardetto/solana-go@v1.16.0
```

## Pretty-Print transactions/instructions

![pretty-printed](https://user-images.githubusercontent.com/15271561/136708519-399c9498-3d20-48d6-89fa-bdf43aac6d83.png)

Instructions can be pretty-printed with the `String()` method on a `Transaction`:

```go
tx, err := solana.NewTransaction(
  []solana.Instruction{
    system.NewTransferInstruction(
      amount,
      accountFrom.PublicKey(),
      accountTo,
    ).Build(),
  },
  recent.Value.Blockhash,
  solana.TransactionPayer(accountFrom.PublicKey()),
)

...

// Pretty print the transaction:
fmt.Println(tx.String())
// OR you can choose a destination and a title:
// tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer SOL"))
```

## SendAndConfirmTransaction

You can wait for a transaction confirmation using the `github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction` package tools (for a complete example: [see here](#transfer-sol-from-one-wallet-to-another-wallet))

```go
// Send transaction, and wait for confirmation:
sig, err := confirm.SendAndConfirmTransaction(
  context.TODO(),
  rpcClient,
  wsClient,
  tx,
)
if err != nil {
  panic(err)
}
spew.Dump(sig)
```

The above command will send the transaction, and wait for its confirmation.

## Address Lookup Tables

Resolve lookups for a transaction:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	lookup "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/time/rate"
)

func main() {
	cluster := rpc.MainNetBeta

	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		cluster.RPC,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))

	version := uint64(0)
	tx, err := rpcClient.GetTransaction(
		context.Background(),
		solana.MustSignatureFromBase58("24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CMC"),
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &version,
			Encoding:                       solana.EncodingBase64,
		},
	)
	if err != nil {
		panic(err)
	}
	parsed, err := tx.Transaction.GetTransaction()
	if err != nil {
		panic(err)
	}
	processTransactionWithAddressLookups(parsed, rpcClient)
}

func processTransactionWithAddressLookups(txx *solana.Transaction, rpcClient *rpc.Client) {
	if !txx.Message.IsVersioned() {
		fmt.Println("tx is not versioned; only versioned transactions can contain lookups")
		return
	}
	tblKeys := txx.Message.GetAddressTableLookups().GetTableIDs()
	if len(tblKeys) == 0 {
		fmt.Println("no lookup tables in versioned transaction")
		return
	}
	numLookups := txx.Message.GetAddressTableLookups().NumLookups()
	if numLookups == 0 {
		fmt.Println("no lookups in versioned transaction")
		return
	}
	fmt.Println("num lookups:", numLookups)
	fmt.Println("num tbl keys:", len(tblKeys))
	resolutions := make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, key := range tblKeys {
		fmt.Println("Getting table", key)

		info, err := rpcClient.GetAccountInfo(
			context.Background(),
			key,
		)
		if err != nil {
			panic(err)
		}
		fmt.Println("got table "+key.String())

		tableContent, err := lookup.DecodeAddressLookupTableState(info.GetBinary())
		if err != nil {
			panic(err)
		}

		fmt.Println("table content:", spew.Sdump(tableContent))
		fmt.Println("isActive", tableContent.IsActive())

		resolutions[key] = tableContent.Addresses
	}

	err := txx.Message.SetAddressTables(resolutions)
	if err != nil {
		panic(err)
	}

	err = txx.Message.ResolveLookups()
	if err != nil {
		panic(err)
	}
	fmt.Println(txx.String())
}
```


## Parse/decode an instruction from a transaction

```go
package main

import (
  "context"
  "encoding/base64"
  "os"
  "reflect"

  "github.com/davecgh/go-spew/spew"
  bin "github.com/gagliardetto/binary"
  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/programs/system"
  "github.com/gagliardetto/solana-go/rpc"
  "github.com/gagliardetto/solana-go/text"
)

func main() {
  exampleFromGetTransaction()
}

func exampleFromBase64() {
  encoded := "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="

  data, err := base64.StdEncoding.DecodeString(encoded)
  if err != nil {
    panic(err)
  }

  // parse transaction:
  tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(data))
  if err != nil {
    panic(err)
  }

  decodeSystemTransfer(tx)
}

func exampleFromGetTransaction() {
  endpoint := rpc.TestNet_RPC
  client := rpc.New(endpoint)

  txSig := solana.MustSignatureFromBase58("3hZorctJtD3QLCRV3zF6JM6FDbFR5kAvsuKEG1RH9rWdz8YgnDzAvMWZFjdJgoL8KSNzZnx7aiExm1JEMC8KHfyy")
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

    tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
    if err != nil {
      panic(err)
    }

    decodeSystemTransfer(tx)
  }
}

func decodeSystemTransfer(tx *solana.Transaction) {
  spew.Dump(tx)

  // Get (for example) the first instruction of this transaction
  // which we know is a `system` program instruction:
  i0 := tx.Message.Instructions[0]

  // Find the program address of this instruction:
  progKey, err := tx.ResolveProgramIDIndex(i0.ProgramIDIndex)
  if err != nil {
    panic(err)
  }

  // Find the accounts of this instruction:
  accounts, err := i0.ResolveInstructionAccounts(&tx.Message)
  if err != nil {
    panic(err)
  }

  // Feed the accounts and data to the system program parser
  // OR see below for alternative parsing when you DON'T know
  // what program the instruction is for / you don't have a parser.
  inst, err := system.DecodeInstruction(accounts, i0.Data)
  if err != nil {
    panic(err)
  }

  // inst.Impl contains the specific instruction type (in this case, `inst.Impl` is a `*system.Transfer`)
  spew.Dump(inst)
  if _, ok := inst.Impl.(*system.Transfer); !ok {
    panic("the instruction is not a *system.Transfer")
  }

  // OR
  {
    // There is a more general instruction decoder: `solana.DecodeInstruction`.
    // But before you can use `solana.DecodeInstruction`,
    // you must register a decoder for each program ID beforehand
    // by using `solana.RegisterInstructionDecoder` (all solana-go program clients do it automatically with the default program IDs).
    decodedInstruction, err := solana.DecodeInstruction(
      progKey,
      accounts,
      i0.Data,
    )
    if err != nil {
      panic(err)
    }
    // The returned `decodedInstruction` is the decoded instruction.
    spew.Dump(decodedInstruction)

    // decodedInstruction == inst
    if !reflect.DeepEqual(inst, decodedInstruction) {
      panic("they are NOT equal (this would never happen)")
    }

    // To register other (not yet registered decoders), you can add them with
    // `solana.RegisterInstructionDecoder` function.
  }

  {
    // pretty-print whole transaction:
    _, err := tx.EncodeTree(text.NewTreeEncoder(os.Stdout, text.Bold("TEST TRANSACTION")))
    if err != nil {
      panic(err)
    }
  }
}

```

## Borsh encoding/decoding

You can use the `github.com/gagliardetto/binary` package for encoding/decoding borsh-encoded data:

Decoder:

```go
  resp, err := client.GetAccountInfo(
    context.TODO(),
    pubKey,
  )
  if err != nil {
    panic(err)
  }

  borshDec := bin.NewBorshDecoder(resp.GetBinary())
  var meta token_metadata.Metadata
  err = borshDec.Decode(&meta)
  if err != nil {
    panic(err)
  }
```

Encoder:

```go
buf := new(bytes.Buffer)
borshEncoder := bin.NewBorshEncoder(buf)
err := borshEncoder.Encode(meta)
if err != nil {
  panic(err)
}
// fmt.Print(buf.Bytes())
```

## ZSTD account data encoding

You can request account data to be encoded with base64+zstd in the `Encoding` parameter:

```go
resp, err := client.GetAccountInfoWithOpts(
  context.TODO(),
  pubKey,
  &rpc.GetAccountInfoOpts{
    Encoding:   solana.EncodingBase64Zstd,
    Commitment: rpc.CommitmentFinalized,
  },
)
if err != nil {
  panic(err)
}
spew.Dump(resp)

var mint token.Mint
err = bin.NewDecoder(resp.GetBinary()).Decode(&mint)
if err != nil {
  panic(err)
}
spew.Dump(mint)
```

## Working with rate-limited RPC providers

```go
package main

import (
  "context"

  "golang.org/x/time/rate"
  "github.com/davecgh/go-spew/spew"
  "github.com/gagliardetto/solana-go/rpc"
)

func main() {
  cluster := rpc.MainNetBeta
  client := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		cluster.RPC,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))

  out, err := client.GetVersion(
    context.TODO(),
  )
  if err != nil {
    panic(err)
  }
  spew.Dump(out)
}
```

## Custom Headers for authenticating with RPC providers

```go
package main

import (
  "context"

  "golang.org/x/time/rate"
  "github.com/davecgh/go-spew/spew"
  "github.com/gagliardetto/solana-go/rpc"
)

func main() {
  cluster := rpc.MainNetBeta
  client := rpc.NewWithHeaders(
    cluster.RPC,
    map[string]string{
      "x-api-key": "...",
    },
  )

  out, err := client.GetVersion(
    context.TODO(),
  )
  if err != nil {
    panic(err)
  }
  spew.Dump(out)
}
```

The data will **AUTOMATICALLY get decoded** and returned (**the right decoder will be used**) when you call the `resp.GetBinary()` method.

## Timeouts and Custom HTTP Clients

You can use a timeout context:

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
defer cancel()
acc, err := rpcClient.GetAccountInfoWithOpts(
  ctx,
  accountID,
  &rpc.GetAccountInfoOpts{
    Commitment: rpc.CommitmentProcessed,
  },
)
```

Or you can initialize the RPC client using a custom HTTP client using `rpc.NewWithCustomRPCClient`:

```go
import (
  "net"
  "net/http"
  "time"

  "github.com/gagliardetto/solana-go/rpc"
  "github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

func NewHTTPTransport(
  timeout time.Duration,
  maxIdleConnsPerHost int,
  keepAlive time.Duration,
) *http.Transport {
  return &http.Transport{
    IdleConnTimeout:     timeout,
    MaxIdleConnsPerHost: maxIdleConnsPerHost,
    Proxy:               http.ProxyFromEnvironment,
    Dial: (&net.Dialer{
      Timeout:   timeout,
      KeepAlive: keepAlive,
    }).Dial,
  }
}

// NewHTTP returns a new Client from the provided config.
func NewHTTP(
  timeout time.Duration,
  maxIdleConnsPerHost int,
  keepAlive time.Duration,
) *http.Client {
  tr := NewHTTPTransport(
    timeout,
    maxIdleConnsPerHost,
    keepAlive,
  )

  return &http.Client{
    Timeout:   timeout,
    Transport: tr,
  }
}

// NewRPC creates a new Solana JSON RPC client.
func NewRPC(rpcEndpoint string) *rpc.Client {
  var (
    defaultMaxIdleConnsPerHost = 10
    defaultTimeout             = 25 * time.Second
    defaultKeepAlive           = 180 * time.Second
  )
  opts := &jsonrpc.RPCClientOpts{
    HTTPClient: NewHTTP(
      defaultTimeout,
      defaultMaxIdleConnsPerHost,
      defaultKeepAlive,
    ),
  }
  rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
  return rpc.NewWithCustomRPCClient(rpcClient)
}
```

## Examples

### Create account (wallet)

```go
package main

import (
  "context"
  "fmt"

  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
)

func main() {
  // Create a new account:
  account := solana.NewWallet()
  fmt.Println("account private key:", account.PrivateKey)
  fmt.Println("account public key:", account.PublicKey())

  // Create a new RPC client:
  client := rpc.New(rpc.TestNet_RPC)

  // Airdrop 1 SOL to the new account:
  out, err := client.RequestAirdrop(
    context.TODO(),
    account.PublicKey(),
    solana.LAMPORTS_PER_SOL*1,
    rpc.CommitmentFinalized,
  )
  if err != nil {
    panic(err)
  }
  fmt.Println("airdrop transaction signature:", out)
}
```

### Load/parse private and public keys

```go
{
  // Load private key from a json file generated with
  // $ solana-keygen new --outfile=standard.solana-keygen.json
  privateKey, err := solana.PrivateKeyFromSolanaKeygenFile("/path/to/standard.solana-keygen.json")
  if err != nil {
    panic(err)
  }
  fmt.Println("private key:", privateKey.String())
  // To get the public key, you need to call the `PublicKey()` method:
  publicKey := privateKey.PublicKey()
  // To get the base58 string of a public key, you can call the `String()` method:
  fmt.Println("public key:", publicKey.String())
}

{
  // Load private key from base58:
  {
    privateKey, err := solana.PrivateKeyFromBase58("66cDvko73yAf8LYvFMM3r8vF5vJtkk7JKMgEKwkmBC86oHdq41C7i1a2vS3zE1yCcdLLk6VUatUb32ZzVjSBXtRs")
    if err != nil {
      panic(err)
    }
    fmt.Println("private key:", privateKey.String())
    fmt.Println("public key:", privateKey.PublicKey().String())
  }
  // OR:
  {
    privateKey := solana.MustPrivateKeyFromBase58("66cDvko73yAf8LYvFMM3r8vF5vJtkk7JKMgEKwkmBC86oHdq41C7i1a2vS3zE1yCcdLLk6VUatUb32ZzVjSBXtRs")
    _ = privateKey
  }
}

{
  // Generate a new key pair:
  {
    privateKey, err := solana.NewRandomPrivateKey()
    if err != nil {
      panic(err)
    }
    _ = privateKey
  }
  {
    { // Generate a new private key (a Wallet struct is just a wrapper around a private key)
      account := solana.NewWallet()
      _ = account
    }
  }
}

{
  // Parse a public key from a base58 string:
  {
    publicKey, err := solana.PublicKeyFromBase58("F8UvVsKnzWyp2nF8aDcqvQ2GVcRpqT91WDsAtvBKCMt9")
    if err != nil {
      panic(err)
    }
    _ = publicKey
  }
  // OR:
  {
    publicKey := solana.MustPublicKeyFromBase58("F8UvVsKnzWyp2nF8aDcqvQ2GVcRpqT91WDsAtvBKCMt9")
    _ = publicKey
  }
}
```

### Transfer Sol from one wallet to another wallet

```go
package main

import (
  "context"
  "fmt"
  "os"
  "time"

  "github.com/davecgh/go-spew/spew"
  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/programs/system"
  "github.com/gagliardetto/solana-go/rpc"
  confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
  "github.com/gagliardetto/solana-go/rpc/jsonrpc"
  "github.com/gagliardetto/solana-go/rpc/ws"
  "github.com/gagliardetto/solana-go/text"
)

func main() {
  // Create a new RPC client:
  rpcClient := rpc.New(rpc.DevNet_RPC)

  // Create a new WS client (used for confirming transactions)
  wsClient, err := ws.Connect(context.Background(), rpc.DevNet_WS)
  if err != nil {
    panic(err)
  }

  // Load the account that you will send funds FROM:
  accountFrom, err := solana.PrivateKeyFromSolanaKeygenFile("/path/to/.config/solana/id.json")
  if err != nil {
    panic(err)
  }
  fmt.Println("accountFrom private key:", accountFrom)
  fmt.Println("accountFrom public key:", accountFrom.PublicKey())

  // The public key of the account that you will send sol TO:
  accountTo := solana.MustPublicKeyFromBase58("TODO")
  // The amount to send (in lamports);
  // 1 sol = 1000000000 lamports
  amount := uint64(3333)

  if true {
    // Airdrop 1 sol to the account so it will have something to transfer:
    out, err := rpcClient.RequestAirdrop(
      context.TODO(),
      accountFrom.PublicKey(),
      solana.LAMPORTS_PER_SOL*1,
      rpc.CommitmentFinalized,
    )
    if err != nil {
      panic(err)
    }
    fmt.Println("airdrop transaction signature:", out)
    time.Sleep(time.Second * 5)
  }
  //---------------

  recent, err := rpcClient.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
  if err != nil {
    panic(err)
  }

  tx, err := solana.NewTransaction(
    []solana.Instruction{
      system.NewTransferInstruction(
        amount,
        accountFrom.PublicKey(),
        accountTo,
      ).Build(),
    },
    recent.Value.Blockhash,
    solana.TransactionPayer(accountFrom.PublicKey()),
  )
  if err != nil {
    panic(err)
  }

  _, err = tx.Sign(
    func(key solana.PublicKey) *solana.PrivateKey {
      if accountFrom.PublicKey().Equals(key) {
        return &accountFrom
      }
      return nil
    },
  )
  if err != nil {
    panic(fmt.Errorf("unable to sign transaction: %w", err))
  }
  spew.Dump(tx)
  // Pretty print the transaction:
  tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer SOL"))

  // Send transaction, and wait for confirmation:
  sig, err := confirm.SendAndConfirmTransaction(
    context.TODO(),
    rpcClient,
    wsClient,
    tx,
  )
  if err != nil {
    panic(err)
  }
  spew.Dump(sig)

  // Or just send the transaction WITHOUT waiting for confirmation:
  // sig, err := rpcClient.SendTransactionWithOpts(
  //   context.TODO(),
  //   tx,
  //   false,
  //   rpc.CommitmentFinalized,
  // )
  // if err != nil {
  //   panic(err)
  // }
  // spew.Dump(sig)
}

```

## RPC Methods

All RPC methods from the [Solana JSON RPC API](https://solana.com/docs/rpc) are supported.
Each method has a testable example in [`rpc/example_test.go`](rpc/example_test.go) that is rendered on
[pkg.go.dev](https://pkg.go.dev/github.com/gagliardetto/solana-go@v1.16.0/rpc#pkg-examples).


## WebSocket Subscriptions

All WebSocket subscriptions from the [Solana WebSocket API](https://solana.com/docs/rpc/websocket) are supported.
Each subscription has a testable example in [`rpc/ws/example_test.go`](rpc/ws/example_test.go) that is rendered on
[pkg.go.dev](https://pkg.go.dev/github.com/gagliardetto/solana-go@v1.16.0/rpc/ws#pkg-examples).


## Contributing

We encourage everyone to contribute, submit issues, PRs, discuss. Every kind of help is welcome.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on commit messages, semver policy, linting, and CI checks.

Before opening a PR, run:

```bash
go test ./... -count=1
golangci-lint run
```

Both must pass in CI.

## License

[Apache 2.0](LICENSE)

## Credits

- Gopher logo was originally created by Takuya Ueda (https://twitter.com/tenntenn). Licensed under the Creative Commons 3.0 Attributions license.
