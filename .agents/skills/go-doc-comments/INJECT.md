---
heading: Go Doc Comments
---

- Every exported function, type, method, constant, variable and package that you add or change MUST have a doc comment.
- Start the comment with the name it documents and write full sentences: `// Transfer moves lamports ...`, `// Package rpc ...`.
- Say what the function returns and when it fails. Name the sentinel or typed errors that callers can match.
- State units and encodings where they matter: lamports or SOL, slot or block height, base58 or base64.
- Mark a deprecated identifier with a `// Deprecated:` paragraph that names the replacement. Link related identifiers with `[Name]` or `[pkg.Name]`.
- Add a testable `Example` function when the usage is not clear from the signature.
