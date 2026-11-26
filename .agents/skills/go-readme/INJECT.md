---
heading: Go README and Examples
---

- When adding a user-facing feature, add or update its section in `README.md` and its entry in the `Contents` list.
- Every section needs a short intro and a Go snippet that compiles as it is, with imports and error handling.
- Snippets use the import path of the current branch: `github.com/gagliardetto/solana-go` on `main`.
- A new RPC method or subscription also gets a function in `example_test.go` and a program in `examples/`.
- Show the quickest path to success first. Put details and edge cases after it.
