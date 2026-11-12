---
name: go-readme
description: Guidelines for the solana-go README and its runnable examples. Use when adding a user-facing feature, changing how a public API is used, or updating documentation.
argument-hint: "[section]"
---

# Create or Update README Sections

Add a section to the root `README.md` for a new feature, or update an existing section, together with the runnable examples that go with it. The README is the guide of the SDK. The API reference is the doc comments, which pkg.go.dev renders.

## Key Rules

- When adding a user-facing feature, add or update its section in `README.md` and its entry in the `Contents` list.
- Every section needs a short intro and a Go snippet that compiles as it is, with imports and error handling.
- Snippets use the import path of the current branch: `github.com/gagliardetto/solana-go` on `main`.
- A new RPC method or subscription also gets a function in `example_test.go` and a program in `examples/`.
- Show the quickest path to success first. Put details and edge cases after it.
- Do NOT update any real code outside of the README and the examples. If you find errors in the codebase, warn the user but do not fix them.
- Do NOT change the version numbers or the release notes in the README, unless the user asks for it.

## Guidelines

You need to understand the feature well to write a good section. Read the code, the tests and the doc comments first, and find the usage that most users will need.

### Structure of the README

The module has one README, at the repo root. Its order is:

1. The title, the badges and a one-line summary.
2. `Contents`, a list of links to the sections.
3. `Features`, the development status, `Requirements` and `Installation`.
4. One section per feature, such as `Address Lookup Tables` or `Parse/decode an instruction from a transaction`.
5. `Examples`, with complete programs for common tasks.
6. `RPC Methods` and `WebSocket Subscriptions`, which point to the examples that pkg.go.dev renders.
7. `About v2`, `Contributing`, `License` and `Credits`.

A new feature section goes with the other feature sections, before `Examples`.

### A Feature Section

- **Heading**: the name of the feature as a user would search for it.
- **Intro**: one or two sentences that say what the feature is for. Link the Solana docs or the SIMD when there is one.
- **Snippet**: the shortest complete use of the feature.
- **Details**: more snippets or short paragraphs for options, limits and errors. Each one covers one idea.
- **Contents entry**: a link in the `Contents` list. GitHub builds the anchor from the heading: lower case, spaces become hyphens, and most punctuation is dropped.

### Code Snippets

- A snippet for a whole task is a complete program: `package main`, the imports and `func main`.
- Use the import path of the current branch. On `main` that is `github.com/gagliardetto/solana-go`. On `v2` it is `github.com/solana-foundation/solana-go/v2`.
- Handle every error. The README uses `if err != nil { panic(err) }` in programs. Keep that, and do not drop errors with `_`.
- Use the endpoint constants of the `rpc` package, such as `rpc.DevNet_RPC`, and not literal URLs.
- Use clear variable names, not `foo` and `bar`.
- Never put a real private key or a seed phrase in a snippet.
- Check that a new snippet compiles before you propose it. Build it as a `main` package in a temporary directory inside the module, and remove the directory after that.

### Runnable Examples

- Each RPC method has an `Example` function in `rpc/example_test.go` and a program in `rpc/examples/<method>/`. Each subscription has the same in `rpc/ws/`.
- These examples call public endpoints. They have no `// Output:` comment, so `go test` compiles them but does not run them. Do not add one.
- A feature that needs no network can have an example with an `// Output:` comment. `go test` then checks it, and it cannot go stale.

### Tone

- Friendly and direct. Use the active voice.
- No marketing language.
- Assume a reader who knows Go and wants to get started quickly. Explain Solana terms the first time they appear, or link them.

### What to Avoid

- Long introductions before the first snippet.
- Walls of text without code.
- A complex example as the first example of a section.
- The same information in two sections. Link to the other section.
- Snippets that use an API from another branch or from an unreleased dependency.

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `[section]` (optional): the feature or the section heading to write or update. The default is the feature that the current changes add.

### Steps

1. Find the target. With an argument, look for a section with that heading. Without one, use `git status` and `git diff` to find the user-facing change on the current branch.
2. Read `README.md` to learn its structure and the style of its snippets.
3. Read the code, the doc comments and the tests of the feature. Find its main entry points and the most common usage.
4. Write or update the section, and add or update the entry in the `Contents` list.
5. For a new RPC method or subscription, add the `Example` function and the program in `examples/`.
6. Check that every new snippet compiles.
7. Present the complete section for review before you apply the changes.
