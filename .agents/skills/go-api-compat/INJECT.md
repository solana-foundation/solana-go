---
heading: Go API Compatibility
---

- Every exported identifier outside `internal/` is public API. On `main`, only add: never remove, rename or change the signature or type of an exported identifier.
- Adding a method to an exported interface is a breaking change too.
- To replace an API, add the new one next to the old one and mark the old one with a `// Deprecated:` paragraph that names the replacement.
- Check every change that touches exported identifiers with `gorelease`, as the `go-api-compat` skill describes. CI fails on incompatible changes.
- Breaking changes go to the `v2` branch. Never change the module path.
- Keep new exported surface small: leave helpers unexported or put them in an `internal/` package.
