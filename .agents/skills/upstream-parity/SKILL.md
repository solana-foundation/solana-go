---
name: upstream-parity
description: Check that an encoding, a decoder or an RPC type in solana-go matches the upstream Rust source. Use when adding or changing instruction data, an account layout, a transaction or message format, or an RPC request or response type, and when a report says that the chain or a node rejects or misreads something.
argument-hint: "[go-type-or-file]"
---

# Check Parity With the Upstream Rust Source

solana-go implements formats that are defined in Rust. A layout that is one byte off still compiles and still passes a Go encode and decode round trip. It fails only on chain. This skill is the procedure to find the upstream definition, compare it with the Go code field by field, and prove the result with bytes that come from upstream.

## Key Rules

- The Rust source at a release tag decides every byte layout and every RPC type. Do not trust another SDK, one observed response or memory.
- Find the code that really serializes the type before you trust its derives. One type can have serde, borsh and a hand-written `pack`.
- Write the layout down field by field: offset, size and encoding. Then compare each field with the Go code.
- Prove parity with bytes that come from upstream: a vector from the Rust tests, from a pinned Rust program or from a real transaction. A Go encode and decode round trip does not prove it.
- Match the strictness of the Rust decoder: short input, trailing bytes, invalid tags and length limits.
- Name the upstream repo, tag and file in the test or in the pull request.

## Guidelines

### Where the Definitions Are

| What | Repo | Path |
|---|---|---|
| RPC responses and request configs | `anza-xyz/agave` | `rpc-client-types/src/response.rs`, `rpc-client-types/src/config.rs` |
| RPC handlers: what a method really does, and its errors | `anza-xyz/agave` | `rpc/src/rpc.rs`, `rpc/src/rpc_pubsub.rs` |
| Transaction status and parsed types | `anza-xyz/agave` | `transaction-status-client-types/` |
| Transaction, message, short vec, nonce | `anza-xyz/solana-sdk` | `transaction/`, `message/`, `short-vec/`, `nonce/` |
| System, vote, compute budget, lookup table and loader instructions | `anza-xyz/solana-sdk` | `system-interface/`, `vote-interface/`, `compute-budget-interface/`, `address-lookup-table-interface/`, `loader-v3-interface/` |
| SPL Token and p-token | `solana-program/token` | `interface/`, `pinocchio/` |
| Token-2022 | `solana-program/token-2022` | `interface/` |
| Stake | `solana-program/stake` | `interface/` |
| Associated token account, memo | `solana-program/associated-token-account`, `solana-program/memo` | `interface/` |
| ZK ElGamal proof | `solana-program/zk-elgamal-proof` | `interface/`, `zk-sdk/` |

Paths move between releases, so check them at the tag you use. Read a file at a tag with the GitHub CLI:

```sh
gh api "repos/anza-xyz/agave/contents/rpc-client-types/src/response.rs?ref=<tag>" \
  -H "Accept: application/vnd.github.raw"
```

Pick the tag on purpose and write it down. For a program, use the release that is deployed on mainnet, or the release that ships the feature. For RPC, use the Agave release that the change targets. Never read `master` and call it the definition.

### Find the Serializer

The derives on a type do not tell you which format is used on the wire. Find the call that serializes it.

- **Native programs** (system, stake, vote): the instruction enum goes through bincode with serde.
- **SPL Token**: hand-written `pack` and `unpack` functions for instructions, and the `Pack` trait for accounts.
- **Token-2022**: the SPL Token packing for the base instructions. An extension instruction has two tags, the extension and then the sub-instruction, followed by a plain struct of fixed-size fields. Account extensions are TLV entries: a `u16` type, a `u16` length, then the value.
- **Associated token account**: a borsh enum.
- **Transactions and messages**: a custom wire format, with short-vec (compact-u16) lengths.
- **RPC**: JSON through serde. The attributes on the type decide the field names and which fields can be missing.

The formats differ exactly where bugs hide:

| | bincode, as Solana uses it | borsh | SPL Token packing |
|---|---|---|---|
| Integers | fixed width, little endian | fixed width, little endian | fixed width, little endian |
| Enum tag | `u32` | `u8` | `u8` |
| `Vec` and `String` length | `u64` | `u32` | no length, the layout is fixed |
| `Option` | 1-byte tag, then the value if present | 1-byte tag, then the value if present | see below |
| `bool` | 1 byte, must be 0 or 1 | 1 byte, must be 0 or 1 | 1 byte, check `unpack` |
| Trailing bytes | accepted by `bincode::deserialize` and `limited_deserialize` | rejected by `try_from_slice` | check `unpack` |

Optional keys in SPL Token and Token-2022 have three different encodings:

- In SPL Token **instruction data**: a 1-byte tag, then 32 bytes only when the key is present.
- In SPL Token **account data** (`COption`): a 4-byte tag, then 32 bytes that are always there.
- In Token-2022 **extension instructions** (`MaybeNull<Address>`): 32 bytes with no tag. All zeros mean none.

### Write the Layout Down

Make a table before you compare anything. This one is the instruction that pull request #518 fixed, where the Go code sent two bytes and the program expected three:

| Offset | Size | Field | Encoding |
|---|---|---|---|
| 0 | 1 | instruction tag | `26`, `TransferFeeExtension` |
| 1 | 1 | sub-instruction tag | `3`, `WithdrawWithheldTokensFromAccounts` |
| 2 | 1 | `num_token_accounts` | `u8` |

Then read the Go `MarshalWithEncoder` and `UnmarshalWithDecoder` next to the table, one field at a time. Check which encoder of `github.com/gagliardetto/binary` runs (bin, borsh or compact-u16), because the same Go struct encodes differently in each.

For accounts, do the same for the account list: the order, the signer flag and the writable flag of each account.

### Get Bytes From Upstream

Use the first source that fits:

1. **A vector in the Rust tests.** Upstream tests often compare packed data with literal bytes. Copy the bytes, and name the file and the tag in a comment.
2. **A pinned Rust program.** `programs/zk-elgamal-proof/internal/paritygen` is a Cargo project with exact dependency versions and a lock file. Each program in `src/bin/` prints a JSON file of named vectors, and a Go test builds the same instruction for each name and compares the bytes (see `TestConfidentialMintBurnRustParity`). CI builds the vectors again and fails on any difference. Add a vector there when you need many cases or values that only Rust can compute. A new vector file also has to be added to `.github/workflows/parity-vectors.yml`.
3. **A real transaction.** Send or simulate the instruction on a local `solana-test-validator`, or read a confirmed transaction with `getTransaction`. This shows what the program accepts. Write down how to reproduce it.

Give every field of a vector its own non-zero value, so that a swapped or dropped field changes the bytes. `TestVector_MaybeNullPubkeyInstructions` in `programs/token-2022` shows the pattern: a table of cases, one key per role, and the expected bytes for each case.

### Decoder Strictness

A decoder reads bytes that anyone can put on chain. Check each of these against the Rust side:

- **Short input** returns an error. It never panics and never reads out of range.
- **Trailing bytes** are accepted or rejected as the Rust decoder does.
- **Unknown enum tags and bools that are not 0 or 1** return an error where Rust returns one.
- **Element counts** are bounded before memory is allocated for them.
- A `Fuzz*` test feeds random bytes to the decoder and expects no panic.

### RPC Types

- **Field names** come from `#[serde(rename_all = "camelCase")]` and `#[serde(rename = ...)]`.
- **Missing fields**: an `Option<T>`, or a field with `skip_serializing_if`, is a pointer in Go.
- **Integer width and sign** follow the Rust type. A `u64` is a `uint64`.
- **Precision**: decoding a number into `any` or `float64` loses precision above 2^53. Decode into a typed field.
- **Numbers as strings**: some amounts are strings in JSON, such as `UiTokenAmount.amount`. Keep them as the server sends them.
- **Versions**: a field that a later Agave release added is missing on older nodes. Make it optional.
- **Behavior**: read the handler in `rpc/src/rpc.rs` for what the method does, which bank it reads and which error codes it returns. The type alone does not tell you.

### Report What You Checked

Leave a trace that the next reader can follow. The repo already does this in comments, for example `Mirrors solana_nonce::state::State::size()`. In the test or in the pull request, name the repo, the tag, the file and the Rust type.

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `[go-type-or-file]` (optional): the Go type, function or file to check. The default is every encoder, decoder and RPC type that the current branch changes.

### Steps

1. Find the Go code and say what it mirrors: which program, which instruction, account or RPC method.
2. Find the Rust definition. Say which tag you read and why that tag.
3. Find the serializer and name the format.
4. Write the layout table.
5. Compare the Go code with the table field by field. List every difference with its offset.
6. Find upstream bytes, or produce them, and add a test that pins them. Check that the test fails when one field is changed, as the `go-test-mutation` skill describes.
7. Check the strictness of the decoder.
8. Report the differences, the evidence, and what you could not check. A fix that changes a field type or a function signature may break the public API, so check it with the `go-api-compat` skill before you propose it.
