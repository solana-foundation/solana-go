---
name: go-doc-comments
description: Write Go doc comments for exported identifiers the way pkg.go.dev renders them. Use when adding or changing exported functions, types, methods, constants or packages, or when code is missing documentation.
argument-hint: "[path] [--all]"
---

# Add Missing Doc Comments

Scan the given path, or the Go files that the current branch adds or changes when no path is given, and add the missing doc comments to the exported functions, types, methods, constants, variables and packages. pkg.go.dev renders these comments, so they are the API reference of the SDK.

## Key Rules

- Every exported function, type, method, constant, variable and package that you add or change MUST have a doc comment.
- Start the comment with the name it documents and write full sentences: `// Transfer moves lamports ...`, `// Package rpc ...`.
- Say what the function returns and when it fails. Name the sentinel or typed errors that callers can match.
- State units and encodings where they matter: lamports or SOL, slot or block height, base58 or base64.
- Mark a deprecated identifier with a `// Deprecated:` paragraph that names the replacement. Link related identifiers with `[Name]` or `[pkg.Name]`.
- Add a testable `Example` function when the usage is not clear from the signature.
- Do NOT modify real code outside of comments. Do NOT modify existing doc comments.

## Guidelines

### Style

Follow the official [Go Doc Comments](https://go.dev/doc/comment) guide, with these conventions:

- Use `//` line comments directly above the declaration, with no blank line between them.
- Write full sentences that start with a capital letter and end with a period. Use US spelling.
- The first sentence is the summary. pkg.go.dev shows it in the package index and in search results, so it has to make sense alone.
- Keep lines at about 80 columns.
- Separate paragraphs with an empty `//` line.
- Indent code with a tab after `//`. Write a list as lines that start with `  - `.
- Link identifiers in square brackets: `[Client]`, `[Client.GetBalance]`, `[rpc.Client]`.
- Do not use tags such as `@param` or `@return`. Describe a parameter in the text, and only when its meaning is not clear from its name and type.

### What to Say

- **Functions and methods**: what it does or returns. A function that returns a bool "reports whether" something is true.
- **Errors**: when it fails, and with which sentinel or typed error, so that callers know what to pass to `errors.Is` or `errors.As`.
- **Context**: what happens when the context is done, when that is not simply "it returns `ctx.Err()`".
- **Types**: what a value represents. Say when the zero value is not ready to use, and whether a value is safe for use by several goroutines.
- **Constants and variables**: one comment for the group, and a comment for each member whose meaning is not clear from its name.
- **Units and encodings**: lamports or SOL, slot or block height, seconds or milliseconds, base58 or base64, little endian or big endian.
- **Wire layouts**: for a type that mirrors an on-chain layout, name the Rust type it mirrors.
- **Instructions**: keep the account table on the struct, with the index, the signer and writable flags, and the role of each account, as the program packages already do.

### Packages

- Each package has one package comment that starts with `// Package <name>`.
- Put a long package comment in `doc.go`.

### Deprecation

Add a paragraph that starts with `Deprecated:` and names the replacement. Keep the summary above it.

```go
// TokenGroupState is the extension state for ExtensionTokenGroup.
//
// Deprecated: the Size and MaxSize fields have the wrong width (the on-chain
// layout uses u64, not u32), so this struct cannot represent real account
// data. Use TokenGroup instead.
type TokenGroupState struct {
```

### Example Functions

- Examples go in `example_test.go`, in the external test package (`package rpc_test`).
- The name ties the example to an identifier: `Example` for the package, `ExampleNewClient` for a function, `ExampleClient` for a type, `ExampleClient_GetBalance` for a method. A suffix in lower case separates several examples: `ExampleClient_GetBalance_withCommitment`.
- An example with an `// Output:` comment runs in `go test`. Add that comment only when the example needs no network.

### Examples

These comments are from this repo.

A method that says what it returns in each case, and names the error to match:

```go
// WaitSubscribed blocks until the server confirms the subscription. It returns
// nil once the subscription is confirmed. If the subscription closes first, it
// returns an error that matches ErrSubscriptionClosed and wraps the cause, such
// as the server's error reply. If ctx is done first, it returns ctx.Err().
//
// Notifications are only sent after the server has registered the
// subscription. Wait for it before reading state that a notification would
// otherwise report, so that no change falls between the read and the
// subscription.
func (s *Subscription) WaitSubscribed(ctx context.Context) error {
```

A type with a summary, then a warning in its own paragraph:

```go
// DurableNonce is the 32-byte value stored in an initialized nonce account and
// used as the recent_blockhash field of a durable-nonce transaction.
//
// Although it is used in the recent_blockhash slot, it is NOT a real blockhash:
// it is derived from one via DurableNonceFromBlockhash with a domain prefix.
type DurableNonce Hash
```

A constant that states its unit and the Rust definition it mirrors:

```go
// NonceStateSize is the serialized size, in bytes, of a nonce account's data.
// It is constant regardless of whether the account is initialized: the runtime
// always allocates this many bytes. Mirrors solana_nonce::state::State::size().
const NonceStateSize = 80
```

A sentinel error and a typed error that say who returns them and when:

```go
// ErrNonceUninitialized is returned by Authorize when the nonce account has not
// been initialized.
var ErrNonceUninitialized = errors.New("nonce account is uninitialized")

// MissingRequiredSignatureError is returned by Authorize when the current nonce
// authority is not among the provided signers.
type MissingRequiredSignatureError struct {
	// Authority is the nonce account's current authority whose signature is
	// required.
	Authority PublicKey
}
```

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `[path]` (optional): the path to scan, such as `rpc/ws` or `programs/token`. The default is the Go files that the current branch adds or changes.
- `[--all]` (optional): also scan identifiers that are not exported.

### Steps

1. If a path is given, scan only that path. If not, scan the `.go` files that the current branch adds or changes: `git diff --name-only main... -- '*.go'` and the uncommitted files from `git status --short`. Never default to the whole module: most of its exported identifiers have no doc comment yet, so the result would be too large to review.
2. Look at the `.go` files. Skip `*_test.go` files and files that start with a `// Code generated` line.
3. Find the exported identifiers without a doc comment: functions, methods on exported types, types, constants, variables, and the package clause when no file of the package has a package comment. Without a path, take only the identifiers that the branch adds or changes, not every identifier in those files.
4. If `--all` is passed, also find the identifiers that are not exported.
5. For each one:
    - Read the code to understand what it does. This may take several files.
    - For a type that mirrors an on-chain layout or an RPC response, check the Rust definition before you describe it.
    - Write the comment in the style above.
6. Present all changes clearly, grouped by file. Apply them without asking for further approval, unless they cover more than one package: then list the identifiers first and wait for approval.
7. Run `gofmt -l` on the changed files and `go build ./...` to confirm that nothing but comments changed.
