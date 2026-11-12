---
name: go-program-instruction
description: Add or change an instruction in a program client package under programs/. Use when adding an instruction to system, stake, vote, token, token-2022 or another program package, or when fixing how an existing instruction is built, encoded or decoded.
argument-hint: "[program] [instruction]"
---

# Add an Instruction to a Program Client

Every instruction in `programs/<name>` has the same shape, and a new one has to be wired in several places. A missing piece still compiles: the builder works, but decoding or the name lookup quietly does not know the instruction.

## Key Rules

- Read the Rust definition of the instruction first: its tag, its data layout and its account list. Then copy the shape of a neighboring instruction in the same package.
- Register a new instruction everywhere in `instructions.go`: the id constant, `InstructionIDToName` and the `InstructionImplDef` variant list. Never insert an id in the middle of an `iota` block.
- The account order and the signer and writable flags must match the Rust docs exactly.
- Test the exact bytes of `Data()` and decode those bytes back with `DecodeInstruction`. The round trip test is not enough.
- On `main`, never change the name, the type or the signature of an existing exported identifier of an instruction.

## Guidelines

### Files

- `programs/<name>/<Instruction>.go` and `programs/<name>/<Instruction>_test.go`. The file is named after the Rust enum variant, in the same case: `WithdrawExcessLamports.go`.
- `programs/<name>/instructions.go` holds the registry.

### Packages Differ

Do not carry the shape of one package into another. Read a neighbor in the same package first.

| | `system`, `stake`, `vote` | `token`, `token-2022` |
|---|---|---|
| Instruction tag | `u32`, little endian: `ag_binary.TypeIDFromUint32(id, binary.LittleEndian)` | `u8`: `ag_binary.TypeIDFromUint8(id)` |
| `Build` stores | the value: `Impl: inst` | a pointer: `Impl: &inst` |

- **What `Build` stores** matters to callers, who type-assert on `Impl`. `DecodeInstruction` has to produce the same kind. In `programs/token`, `TestBuilderImplIsPointer_MatchesDecode` pins this.
- **Accounts** come in two shapes. `programs/system/Assign.go` embeds one `AccountMetaSlice` of fixed length. `programs/token/WithdrawExcessLamports.go` has `Accounts` and `Signers`, with `SetAccounts` and `GetAccounts`, because the multisig signers follow the fixed accounts.
- **Other packages** have their own details. Check them in their own files.

### The Instruction File

Write the parts in the order that the neighbors use:

1. **The struct.**
    - The doc comment comes from the Rust docs of the variant.
    - Parameters are pointer fields, so that `Validate` can tell a missing value from a zero.
    - The account table is a comment in the struct: the index, the flags and the role of each account.
    - The account slices carry the tags `bin:"-" borsh_skip:"true"`, so that they are not encoded as data.
2. **`New<Name>InstructionBuilder`** allocates the account slice with the right length.
3. **Setters and getters**: `Set<Param>`, `Set<Role>Account` and `Get<Role>Account`. An account setter sets the flags: `ag_solanago.Meta(key).WRITE().SIGNER()`.
4. **`Validate`, `Build` and `ValidateAndBuild`.** `Validate` checks every required parameter and account. For a multisig authority, either the authority signs or signers are given.
5. **`EncodeToTree`** prints the instruction for humans.
6. **`MarshalWithEncoder` and `UnmarshalWithDecoder`** write and read the data after the tag, field by field, in the order of the Rust definition.
7. **`New<Name>Instruction`** is the shortcut that takes the parameters first and the accounts after them.

### Register the Instruction

In `instructions.go`:

1. **The id constant**, `Instruction_<Name>`, with a doc comment.
    - In an `iota` block, the position is the id. Only append. An id that is inserted in the middle shifts every id after it.
    - When the id is not the next number, give it an explicit value in its own `const` block, as `programs/token` does for 38, 45 and 255.
2. **`InstructionIDToName`**: add the case.
3. **`InstructionImplDef`**: add `{Name: "<Name>", Type: (*<Name>)(nil)}` at the position that equals the id. This list is indexed by the id, so it only works for ids without gaps. An id after a gap needs its own decode path, such as `pTokenInstructionMap` in `programs/token`.
4. **`DecodeInstruction`**: confirm with a test that it returns the new type.

A Token-2022 extension instruction has two tags, the extension and the sub-instruction. Follow an existing extension in that package, for example `TransferFeeExtension.go`.

### Tests

- **Round trip**: `TestEncodeDecode_<Name>`, with `encodeT` and `decodeT` from the package's `testing_utils.go`. For an instruction without parameters this compares two empty structs, so it proves nothing alone.
- **Exact bytes**: build the instruction, call `Build().Data()` and compare with bytes that come from upstream. See the `upstream-parity` skill.
- **Decode**: pass the same bytes and accounts to `DecodeInstruction`, assert the concrete type, and compare the parameters and the accounts.
- **Accounts**: assert the key, `IsSigner` and `IsWritable` of every account, with a different key for each role.
- **Validate**: one case for each missing parameter or account.
- **Malformed data**: short data returns an error and does not panic.

Then check that the tests can fail, as the `go-test-mutation` skill describes: swap two accounts, drop a flag, skip a field in the encoder.

### Compatibility

- A change that corrects the bytes of an existing instruction is a `fix`.
- When the fix needs a new value, add a field and a setter. Do not add a parameter to `New<Name>Instruction`, because that breaks its signature. Pull request #518 fixed an instruction this way, with the new field `TransferFeeExtension.NumTokenAccounts`.
- Check the result with the `go-api-compat` skill.

### Docs

- Every exported identifier gets a doc comment. See the `go-doc-comments` skill.
- A new program package gets an entry in the `Features` list of the README. See the `go-readme` skill.

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `[program]`: the package under `programs/`, such as `token-2022`.
- `[instruction]`: the name of the Rust enum variant, such as `WithdrawExcessLamports`.

### Steps

1. Find the Rust definition at a release tag. Write down the tag, the data layout and the account list.
2. Read `instructions.go` of the package and one neighboring instruction with a similar shape.
3. Write the instruction file.
4. Register the instruction.
5. Write the tests.
6. Run `go test ./programs/<name>/ -race -count=1` and `golangci-lint run ./programs/<name>/...`.
7. Check that the tests can fail.
8. Report the upstream file and tag, the layout, what the tests cover, and what was not checked against a validator.
