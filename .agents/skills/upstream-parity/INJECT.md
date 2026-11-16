---
heading: Upstream Parity
---

- Before you write or change an encoding, a decoder or an RPC type, read the Rust definition at a release tag and find the code that really serializes it.
- Write the layout down field by field (offset, size, encoding), then compare each field with the Go code.
- Prove parity with bytes that come from upstream: a vector from the Rust tests, from a pinned Rust program or from a real transaction. A Go encode and decode round trip does not prove it.
- Name the upstream repo, tag and file in the test or in the pull request.
