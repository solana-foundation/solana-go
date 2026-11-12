---
name: go-api-compat
description: Keep the public Go API of solana-go backward compatible. Use when a change adds, changes, deprecates or removes an exported identifier, or when the semver-check (gorelease) check fails.
argument-hint: "[base-version]"
---

# Check API Compatibility

Check a change against the public API of the module with [`gorelease`](https://pkg.go.dev/golang.org/x/exp/cmd/gorelease), and find a compatible way to make it. This is the Go counterpart of choosing a version bump: on `main` the only allowed bumps are patch and minor.

## Key Rules

- Every exported identifier outside `internal/` is public API. On `main`, only add: never remove, rename or change the signature or type of an exported identifier.
- Adding a method to an exported interface is a breaking change too.
- To replace an API, add the new one next to the old one and mark the old one with a `// Deprecated:` paragraph that names the replacement.
- Check every change that touches exported identifiers with `gorelease`, as the `go-api-compat` skill describes. CI fails on incompatible changes.
- Breaking changes go to the `v2` branch. Never change the module path.
- Keep new exported surface small: leave helpers unexported or put them in an `internal/` package.

## Guidelines

### What Breaks

- Removing or renaming an exported function, type, method, constant, variable, struct field or package.
- Changing the signature of an exported function or method. Adding a parameter, a result or a variadic parameter all count.
- Changing the type of an exported struct field, constant or variable.
- Adding a method to an exported interface. Every implementation outside the module stops compiling.
- Changing a method from a value receiver to a pointer receiver. The value type loses the method, so it no longer satisfies the interfaces that need it.
- Making a comparable struct not comparable, for example with a new slice, map or function field.

### What Is Safe

- New exported functions, types, constants, variables and packages.
- New methods on concrete types.
- New fields on structs, when the struct stays comparable.
- Bug fixes that keep every signature.

### How to Grow an API

- **Options**: add a field to the existing `...Opts` struct. When a function has no options yet, add a `...WithOpts` variant next to it and keep the old function as a call to the new one.
- **New behavior**: give it a new name. Do not change what an existing function does for inputs that worked before.
- **Interfaces**: define a new small interface and check for it with a type assertion. Do not add a method to an existing interface.
- **Wrong types**: when an exported field has the wrong type, keep it, add a correct one next to it and deprecate the old one.

A deprecation keeps the old identifier in place and names the replacement. This one is from `programs/token-2022`, where a struct had fields of the wrong width:

```go
// TokenGroupState is the extension state for ExtensionTokenGroup.
//
// Deprecated: the Size and MaxSize fields have the wrong width (the on-chain
// layout uses u64, not u32), so this struct cannot represent real account
// data. Use TokenGroup instead.
type TokenGroupState struct {
```

### Behavior Is API Too

`gorelease` only compares declarations. Check these by hand:

- **Errors**: an error that callers match with `errors.Is` or `errors.As` has to keep matching. Wrapping it in more context is fine.
- **Encoding**: the bytes for a valid input must not change, unless the old bytes were wrong. That is a `fix`, and the commit has to say so.
- **Defaults**: a zero value or an empty option keeps its meaning.
- **Receivers**: changing a pointer receiver to a value receiver compiles, but a call on a nil pointer then panics.
- **Go version**: do not raise the `go` line of `go.mod` as a side effect. It drops users on older toolchains.

### Running gorelease

Inside a git repo, `gorelease` compares the last commit with the latest release. It needs network access, because it downloads that release.

```sh
gorelease
```

It does not read uncommitted changes. When the working tree is not clean, it stops with `repo ... has uncommitted changes`. To check work that is not committed yet, copy the tree to a directory outside the repo and run it there:

```sh
tmp=$(mktemp -d)
git ls-files -co --exclude-standard -z | tar --null -T - -cf - | tar -xf - -C "$tmp"
(cd "$tmp" && gorelease)
rm -rf "$tmp"
```

Do not commit only to make `gorelease` run.

### Reading the Output

`gorelease` prints one block per changed package, then a summary:

```
# github.com/gagliardetto/solana-go/rpc/ws
## compatible changes
(*Subscription).WaitSubscribed: added

# summary
Inferred base version: v1.26.0
Suggested version: v1.27.0
```

- `## compatible changes` are additions. They are fine on `main`.
- `## incompatible changes` fail the CI check. Each line names the identifier and what changed.
- A suggested version with a new minor number means the change is additive. When `gorelease` cannot suggest a version, there are incompatible changes.

The output covers every change since the latest release, not only yours. It also lists what other pull requests merged to `main` since that release. To find your part, run `gorelease` on a copy of `main` too, and compare the two outputs:

```sh
tmp=$(mktemp -d)
git archive upstream/main | tar -x -C "$tmp"
(cd "$tmp" && gorelease > gorelease.txt)
gorelease > "$tmp/branch.txt"   # on a clean tree, or on a copy of the working tree as above
diff "$tmp/gorelease.txt" "$tmp/branch.txt"
rm -rf "$tmp"
```

Use the name that your clone gives the remote of `solana-foundation/solana-go` in place of `upstream`. Lines that are only in the output for the branch come from your change.

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `[base-version]` (optional): the released version to compare against, such as `v1.26.0`. The default is the latest release.

### Steps

1. Check that `gorelease` is installed with `command -v gorelease`. If it is not, install it with `go install golang.org/x/exp/cmd/gorelease@latest`.
2. Run `gorelease` at the repo root, or `gorelease -base=<base-version>` when a version was given. When the working tree has uncommitted changes, run it on a copy of the tree, as described above.
3. List every incompatible change by package. Compare with the output for `main`, as described above, and keep only the lines that come from the change.
4. For each one, propose the compatible way to reach the same goal, using the patterns above.
5. If a change cannot be made compatible, say so and point to the `v2` branch. Do not change any code that the user did not ask you to change.
