---
heading: Go Test Mutation
---

- A regression test MUST fail on the code without the fix. Check it with `go test -overlay`, as the `go-test-mutation` skill describes. Do not assume it.
- For new code, break it on purpose where it matters and confirm that a test fails each time. A mutant that does not compile proves nothing.
- Never edit, stash, reset or commit files in the working tree to run these checks.
- Use distinct, non-zero values in test data, so that a swapped or dropped field shows up.
