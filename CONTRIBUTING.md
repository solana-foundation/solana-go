# Contributing to solana-go

We encourage everyone to contribute — submit issues, PRs, and discuss. Every kind of help is welcome.

## Getting Started

1. Fork the repository and clone your fork
2. Make sure you have Go 1.24+ installed
3. Run `go mod download` to fetch dependencies
4. Run the tests to verify everything works: `go test ./... -count=1`
5. Install [`golangci-lint`](https://golangci-lint.run/welcome/install/) (v2.11+) and run `golangci-lint run` to verify the linters pass

## Development Workflow

1. Create a branch from `main` for your changes
2. Make your changes and write tests
3. Ensure all tests pass locally
4. Commit using [Conventional Commits](#commit-messages) format
5. Open a pull request against `main`

## Linting

This project uses [golangci-lint](https://golangci-lint.run/) with the config at [`.golangci.yml`](.golangci.yml). The enabled linters are:

| Linter         | Purpose                                                       |
| -------------- | ------------------------------------------------------------- |
| `errcheck`     | Flags unchecked errors                                        |
| `govet`        | Go's standard static analyzer (composites, unusedresult, etc.) |
| `ineffassign`  | Catches ineffectual assignments                               |
| `staticcheck`  | Detects bugs, performance issues, and style violations        |
| `unused`       | Finds unused constants, variables, functions, and types       |
| `errorlint`    | Enforces `%w` wrapping and `errors.Is` / `errors.As`          |
| `misspell`     | Catches common misspellings (US English)                      |
| `gofmt`        | Standard Go formatting                                        |
| `goimports`    | Import ordering and grouping                                  |

Run locally:

```bash
# Run all linters
golangci-lint run

# Auto-fix where possible (errorlint, gofmt, goimports, misspell)
golangci-lint run --fix
```

The `lint` CI check must pass before a PR can merge. If you need to disable a rule for a specific line, use `//nolint:<linter> // reason` and include a reason — unjustified nolint directives will be flagged in review.

A small number of deprecation warnings (`SA1019`), stylistic checks (`ST1003` naming, `ST1001` dot imports), and known-refactor items are disabled in config and tracked as follow-up work — not as accepted debt. See [`.golangci.yml`](.golangci.yml) for the current exclusion list.

## Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/). All commit messages must follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type       | When to use                                           |
| ---------- | ----------------------------------------------------- |
| `feat`     | A new feature or public API addition                  |
| `fix`      | A bug fix                                             |
| `docs`     | Documentation-only changes                            |
| `refactor` | Code changes that neither fix a bug nor add a feature |
| `test`     | Adding or updating tests                              |
| `chore`    | Build process, CI, or auxiliary tool changes          |
| `perf`     | Performance improvements                              |

### Breaking Changes

If your change breaks the public API, you **must** indicate it in the commit:

```
feat!: remove deprecated GetRecentBlockhash method

BREAKING CHANGE: GetRecentBlockhash has been removed. Use GetLatestBlockhash instead.
```

The `!` after the type or a `BREAKING CHANGE:` footer will signal a major version bump.

### Examples

```
feat: add priority fee support to SendTransaction
fix: handle nil response in GetAccountInfo
docs: update RPC methods table in README
refactor(ws): simplify subscription reconnect logic
test: add coverage for address lookup table resolution
```

Commit messages are enforced in CI via [commitlint](https://commitlint.js.org/).

## Semver and API Compatibility

This project follows [Semantic Versioning](https://semver.org/):

- **Patch** (`v1.0.x`): bug fixes, no API changes
- **Minor** (`v1.x.0`): new features, backwards-compatible API additions
- **Major** (`vX.0.0`): breaking changes to the public API

CI runs [`gorelease`](https://pkg.go.dev/golang.org/x/exp/cmd/gorelease) on every pull request to detect whether your changes are backwards-compatible. If you modify or remove any exported type, function, method, or interface, `gorelease` will flag it and the check will fail unless the version is bumped accordingly.

**What counts as a breaking change:**

- Removing or renaming an exported function, type, method, or constant
- Changing the signature of an exported function or method
- Changing the type of an exported struct field
- Removing or renaming a package

**What is safe (minor/patch):**

- Adding new exported functions, types, or methods
- Adding new fields to structs (unless they affect interface satisfaction)
- Bug fixes that don't change API signatures

## CI Checks

Every pull request runs the following checks:

| Check            | What it does                                                                                                                                             |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **tests**        | Runs `go test ./...` across Go 1.24.x on ubuntu, macos, and windows. Only triggered when `.go`, `go.mod`, `go.sum`, or the workflow file itself changes. |
| **lint**         | Runs `golangci-lint` on ubuntu-latest with the config at `.golangci.yml`. See [Linting](#linting) above.                                                 |
| **semver-check** | Runs `gorelease` to compare the public API against the latest release tag. Fails if breaking changes are detected without a major version bump.          |
| **commit-lint**  | Validates that all commits in the PR follow Conventional Commits format.                                                                                 |

## Releases

Releases are automated via [Release Please](https://github.com/googleapis/release-please). When commits land on `main`, Release Please opens (or updates) a release PR with:

- A version bump based on conventional commit types (`fix:` = patch, `feat:` = minor, `feat!:` = major)
- An auto-generated `CHANGELOG.md`

When the release PR is merged, a new GitHub Release and Git tag are created automatically.

## Running Tests

```bash
# Run all tests
go test ./... -count=1

# Run tests for a specific package
go test ./rpc/... -count=1

# Run tests with race detection
go test -race ./... -count=1
```

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include Go version, OS, and a minimal reproducer when reporting bugs
- Check existing issues before opening a new one
