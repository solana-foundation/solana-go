# Agent Instructions

## Project Overview

solana-go is the Go SDK for Solana: keys, transactions, the JSON RPC and WebSocket clients, and clients for the native and SPL programs. The Solana Foundation maintains it in the `solana-foundation/solana-go` GitHub repo. The API reference is on [pkg.go.dev](https://pkg.go.dev/github.com/gagliardetto/solana-go) and the guides are in `README.md`.

The repo is one Go module, with no workspace and no Makefile. It ships from two branches:

- **`main` is v1**, with the module path `github.com/gagliardetto/solana-go`. It takes backward-compatible changes only, and it is the default target for pull requests.
- **`v2`** has the module path `github.com/solana-foundation/solana-go/v2`. It holds everything on `main` plus a short stack of commits with breaking API cleanups. At every sync, maintainers rebase that stack onto `main`, so the history of `v2` is rewritten. Never rebase, force-push or tag `v2` yourself. The `v2-port` skill describes the sync.

Never change the module path or the import paths of the branch you are on, even though the repo now lives under `solana-foundation`.

## Getting Started

```shell
go build ./...                    # Compile every package
go test ./... -race -count=1      # Run all tests the way CI does
golangci-lint run                 # Lint and format checks (CI pins the version in .github/workflows/lint.yml)
golangci-lint run --fix           # Auto-fix gofmt, goimports, misspell and errorlint findings
gorelease                         # Compare the public API with the latest release
```

Install `gorelease` with `go install golang.org/x/exp/cmd/gorelease@latest`. It reads the last commit and stops when the working tree has uncommitted changes. The `go-api-compat` skill shows how to check work that is not committed yet.

The minimum Go version is the `go` line of `go.mod`. The tests need no network and no validator.

For one package or one test:

```shell
go test ./rpc/ws/ -race -count=1
go test ./rpc/ws/ -race -count=1 -run 'TestName'
```

## Architecture

- **Root package `solana`** (the repo root): keys and native types (`PublicKey`, `PrivateKey`, `Signature`, `Hash`), `Transaction`, `Message` (legacy, v0 with address lookup tables, and v1 from SIMD-0385 in `message_v1.go`), sanitize checks, nonce helpers, program ids, and the instruction decoder registry (`RegisterInstructionDecoder`, `DecodeInstruction`).
- **`rpc`**: the JSON RPC client, with one file per method, named after the method (`getBalance.go`). `rpc/jsonrpc` is the HTTP transport. `rpc/ws` is the WebSocket client, with one file per subscription. `rpc/sendAndConfirmTransaction` sends a transaction and waits for it.
- **`programs/<name>`**: clients for the native and SPL programs, such as `system`, `stake`, `vote`, `token`, `token-2022` and `associated-token-account`.
- **`programs/zk-elgamal-proof`**: the client for the ZK ElGamal proof program and the proofs for confidential transfers. It embeds a wasm blob that is built from pinned Rust sources.
- **`introspection`** resolves every top-level and inner instruction of a confirmed transaction. **`sysvar`** holds the sysvar ids and account types.
- **`base58`** is the codec, with AVX2 assembly on amd64. The `purego` build tag turns the assembly off, so test a change there with and without `-tags purego`.
- **`bip39`, `vault`, `cli`, `cmd/slnc`**: mnemonics, the key vault and the `slnc` command line tool. **`text`, `diff`, `zap-box`**: pretty-printing, diff and logging helpers.

The root package imports only `base58`, `bip39` and `text` from this module. `rpc` and every program package import the root package, so the root package can never import them.

Binary encoding goes through [`github.com/gagliardetto/binary`](https://github.com/gagliardetto/binary), which has the bin, borsh and compact-u16 encoders and decoders. Types implement `MarshalWithEncoder` and `UnmarshalWithDecoder`.

## Testing

- **Framework**: the standard `testing` package with testify. Use `require` when the test cannot go on after a failure, and `assert` otherwise.
- **Layout**: tests sit next to the code in `*_test.go` files. Fixtures go in a `testdata/` directory.
- **No network**: a test must never call a live cluster.
    - RPC tests use the `mockJSONRPC` server in `rpc/rpc_test.go`.
    - WebSocket tests use `mockWSServer` in `rpc/ws/fixes_test.go`.
    - The examples in `rpc/example_test.go` and `rpc/ws/example_test.go` call public endpoints. They have no `// Output:` comment, so `go test` compiles them but does not run them. Keep it that way.
- **CI matrix**: the tests run with `-race` on Linux, macOS and Windows. Do not rely on POSIX-only behavior, fixed ports or sleeps. For code that waits only on timers and channels, use `testing/synctest` and not real time.
- **A test must be able to fail**: check that a regression test fails without the fix. The `go-test-mutation` skill shows how, without touching the working tree. Do not use a zero value as the expected value. `solana.SystemProgramID` is 32 zero bytes, so it equals a `PublicKey` that was never set.
- **Instruction tests**: an instruction has an encode and decode round trip (`TestEncodeDecode_<Name>`). Add fixed byte vectors from the real program when the layout is not trivial.
- **Fuzz tests**: a decoder of untrusted bytes should have a `Fuzz*` test. Run one with `go test -run '^$' -fuzz FuzzName -fuzztime 30s ./path/to/pkg`.
- **Rust parity vectors**: the `*_rust_parity.json` files in `testdata/` come from the Rust programs in `programs/zk-elgamal-proof/internal/paritygen`. CI builds them again and fails on any difference. Never edit them by hand. Each program prints its vectors to stdout, so redirect the output into the file: `cargo run --locked --bin <name> > <file>`, run in that directory. `.github/workflows/parity-vectors.yml` shows which program writes which file.
- **Benchmarks**: `go test -run '^$' -bench . -benchmem ./path/to/pkg`. Compare several runs with `benchstat`, not single numbers.

## Wire Formats

Everything this SDK encodes or decodes has to match what the validator and the on-chain programs do. The `upstream-parity` skill has the full procedure.

- **The Rust source is the source of truth.** Read the definition at a release tag before you write or change a layout:
    - [`anza-xyz/agave`](https://github.com/anza-xyz/agave) for the RPC types and the native programs;
    - [`anza-xyz/solana-sdk`](https://github.com/anza-xyz/solana-sdk) for the transaction, message and nonce types;
    - [`solana-program/token`](https://github.com/solana-program/token) and [`solana-program/token-2022`](https://github.com/solana-program/token-2022) for SPL Token.

    Do not infer a layout from one observed response or from another SDK.
- **Check the encoding of every field.** Bincode, borsh and the hand-written packing of SPL Token use different widths for lengths, enum tags and options. One program can mix them: SPL Token packs an optional key with a 4-byte tag in account data and with a 1-byte tag in instruction data.
- **Match the strictness of the Rust decoder.** For example, `bincode::deserialize` accepts trailing bytes but rejects a bool that is not 0 or 1.
- **Integer width follows the Rust type.** A `u64` is a `uint64` in Go, never an `int` or a `float64`, and JSON decoding must not lose precision. A Rust `Option<T>`, or a field that serde skips when it is empty, is a pointer in Go.
- **Decoders read untrusted bytes.** Bound every element count before you allocate, and return an error on short input instead of panicking.
- **A local validator is a sanity check, not the definition.** Use `solana-test-validator` to confirm the real behavior, but let the Rust definition decide.

## Program Clients

The `go-program-instruction` skill has the full checklist for a new instruction.

- **One file per instruction**, named after it (`Assign.go` and `Assign_test.go`). The package's `instructions.go` (`instruction.go` in `compute-budget` and `tokenregistry`) holds the instruction ids, `InstructionIDToName`, the `InstructionImplDef` variant list and `DecodeInstruction`. A new instruction has to be added to all of them.
- **Every instruction type has the same surface**: `New<Name>InstructionBuilder`, `Set...` and `Get...` methods for the parameters and accounts, `Validate`, `Build`, `ValidateAndBuild`, `EncodeToTree`, `MarshalWithEncoder`, `UnmarshalWithDecoder`, and the `New<Name>Instruction` shortcut.
- **The account order and the signer and writable flags** must match the instruction docs of the Rust program. Keep the account table in the comment on the struct.
- **Import names follow the package.** `system`, `token` and most of `token-2022` import with aliases (`ag_binary`, `ag_solanago`, `ag_format`, `ag_treeout`). `stake`, `vote` and many confidential files in `token-2022` import the root package without an alias. A new file follows its neighbors.
- **Each package registers its decoder** in `init` with `solana.MustRegisterInstructionDecoder`. Importing the package is what makes `solana.DecodeInstruction` work for that program.

## RPC and WebSocket

- A new RPC method gets its own file in `rpc/`, named after the method, an example in `rpc/example_test.go` and a runnable program in `rpc/examples/<method>/`. A new subscription gets the same in `rpc/ws/`.
- Every method takes `context.Context` first. Optional parameters go in an `...Opts` struct or a `...WithOpts` variant, so that they can grow without a breaking change.

## Public API and Compatibility

- Every exported identifier outside an `internal/` directory is public API. CI runs `gorelease` on each pull request to `main` and fails on an incompatible change.
- On `main`, only add. To replace an API, add the new one next to the old one and mark the old one with a `// Deprecated:` paragraph that names the replacement. A change that has to break goes to `v2`.
- Keep the new exported surface small. Leave helpers unexported, or put shared ones in an `internal/` package.
- Behavior is part of the contract too. An error that callers match with `errors.Is` or `errors.As` has to keep matching.

## Coding Conventions

- **Formatting**: `gofmt` and `goimports`, checked by `golangci-lint`. Comments and strings use US spelling (`misspell`).
- **Errors**: wrap with `%w`, and compare with `errors.Is` and `errors.As` (`errorlint`). Export a sentinel (`Err...`) or a typed error for every condition that a caller may branch on.
- **Context**: anything that does I/O takes `context.Context` as its first parameter and stops when it is canceled.
- **Established names stay**: exported names with underscores (`Instruction_Assign`) and file names such as `getBalance.go` and `Assign.go` are the house style. Do not rename them to satisfy a linter.
- **`//nolint`** needs the linter name and a reason: `//nolint:errcheck // reason`.
- **Dependencies**: prefer the standard library. Ask before you add a module dependency, because it lands in the build of every user. Do not raise the `go` line of `go.mod` as a side effect.
- **License header**: most files start with the Apache 2.0 header. Copy it from a neighboring file when the package uses it.
- **Generated and pinned files**: never edit `CHANGELOG.md`, the `*_rust_parity.json` vectors or the embedded wasm blob by hand.

## Commits & Releases

- Commits follow [Conventional Commits](https://www.conventionalcommits.org/). commitlint checks them with its default rules: the header and every body line have at most 100 characters, the subject starts in lower case and has no period at the end.
- The scope is the package or the area: `ws`, `rpc`, `token-2022`, `message`, `keys`, `base58`, `readme`, `ci`, `deps`.
- Pull requests are squash-merged. For the title of the squash commit on `main`, GitHub proposes the title of the only commit when the pull request has one, and the PR title when it has more. So a pull request with one commit needs the same title as that commit, and both must be Conventional Commits.
- [Release Please](https://github.com/googleapis/release-please) reads those commits: `fix` gives a patch release and `feat` gives a minor release. It also writes `CHANGELOG.md`.
- All commits need a verified signature, so leave the commit to the user unless they ask you to make it.

## CI

- **tests**: `go test ./... -race -count=1` on Linux, macOS and Windows, on the Go versions in `.github/workflows/tests.yml`. It runs only when Go files, `go.mod` or `go.sum` change.
- **lint**: `golangci-lint` with `.golangci.yml`.
- **semver-check**: `gorelease`, on pull requests to `main`.
- **commit-lint**: commitlint on every commit of a pull request to `main`.
- **rust parity vectors** and **zk wasm**: build the vectors and the wasm blob again from pinned Rust sources and compare them with the committed files.
- **release**: Release Please, on pushes to `main`.

## Docs

- pkg.go.dev renders the doc comments and the `Example` functions, so they are the API reference.
- `README.md` is the guide. A new user-facing feature gets a section there and an entry in its `Contents` list.
- Security problems are never discussed in public issues or pull requests. See `SECURITY.md`.

## Maintaining These Instructions

- `CLAUDE.md` is the source. `AGENTS.md` is a symlink to it, so every agent reads the same text.
- Skills live in `.agents/skills/<name>/`. `SKILL.md` holds the full instructions, which an agent loads on demand. `INJECT.md` holds the short summary that is always in context, and it is optional: `v2-port` has none, because it is only used on request. `.claude/skills/<name>` is a symlink to the same directory.
- The block between the `skills-inject` markers below is generated from the `INJECT.md` files. Do not edit it by hand. Change the `INJECT.md` file, then run `npx skills-inject@latest`.

<!-- skills-inject:start -->

## Skill Instructions

The following instructions come from installed skills (autogenerated, do not edit manually) and should always be followed.

### Go API Compatibility

- Every exported identifier outside `internal/` is public API. On `main`, only add: never remove, rename or change the signature or type of an exported identifier.
- Adding a method to an exported interface is a breaking change too.
- To replace an API, add the new one next to the old one and mark the old one with a `// Deprecated:` paragraph that names the replacement.
- Check every change that touches exported identifiers with `gorelease`, as the `go-api-compat` skill describes. CI fails on incompatible changes.
- Breaking changes go to the `v2` branch. Never change the module path.
- Keep new exported surface small: leave helpers unexported or put them in an `internal/` package.

### Go Doc Comments

- Every exported function, type, method, constant, variable and package that you add or change MUST have a doc comment.
- Start the comment with the name it documents and write full sentences: `// Transfer moves lamports ...`, `// Package rpc ...`.
- Say what the function returns and when it fails. Name the sentinel or typed errors that callers can match.
- State units and encodings where they matter: lamports or SOL, slot or block height, base58 or base64.
- Mark a deprecated identifier with a `// Deprecated:` paragraph that names the replacement. Link related identifiers with `[Name]` or `[pkg.Name]`.
- Add a testable `Example` function when the usage is not clear from the signature.

### Go Program Instructions

- Read the Rust definition of the instruction first: its tag, its data layout and its account list. Then copy the shape of a neighboring instruction in the same package.
- Register a new instruction everywhere in the package's registry file (`instructions.go`, or `instruction.go` in `compute-budget` and `tokenregistry`): the id constant, `InstructionIDToName` and the `InstructionImplDef` variant list. Never insert an id in the middle of an `iota` block.
- The account order and the signer and writable flags must match the Rust docs exactly.
- Test the exact bytes of `Data()` and decode those bytes back with `DecodeInstruction`. The round trip test is not enough.

### Go README and Examples

- When adding a user-facing feature, add or update its section in `README.md` and its entry in the `Contents` list.
- Every section needs a short intro and a Go snippet that compiles as it is, with imports and error handling.
- Snippets use the import path of the current branch: `github.com/gagliardetto/solana-go` on `main`.
- A new RPC method or subscription also gets a function in `example_test.go` and a program in `examples/`.
- Show the quickest path to success first. Put details and edge cases after it.

### Go Test Mutation

- A regression test MUST fail on the code without the fix. Check it with `go test -overlay`, as the `go-test-mutation` skill describes. Do not assume it.
- For new code, break it on purpose where it matters and confirm that a test fails each time. A mutant that does not compile proves nothing.
- Never edit, stash, reset or commit files in the working tree to run these checks.
- Use distinct, non-zero values in test data, so that a swapped or dropped field shows up.

### Shipping (Git)

- NEVER commit, push, create branches, or create PRs without explicit user approval.
- Before any git operation that creates or modifies a commit, present a review block containing: commit title, commit/PR description, and release effect (none, patch or minor). ALWAYS wait for approval.
- Commit titles and PR titles follow Conventional Commits: `type(scope): subject`, at most 100 characters, lower-case subject, no period at the end.
- Wrap commit body lines at 100 characters. commitlint rejects longer lines.
- Never use `!` or a `BREAKING CHANGE:` footer on `main`, and never edit `CHANGELOG.md`.
- Commits must be signed. Never pass `--no-gpg-sign` or `--no-verify`.

### Upstream Parity

- Before you write or change an encoding, a decoder or an RPC type, read the Rust definition at a release tag and find the code that really serializes it.
- Write the layout down field by field (offset, size, encoding), then compare each field with the Go code.
- Prove parity with bytes that come from upstream: a vector from the Rust tests, from a pinned Rust program or from a real transaction. A Go encode and decode round trip does not prove it.
- Name the upstream repo, tag and file in the test or in the pull request.

<!-- skills-inject:end -->
