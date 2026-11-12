---
heading: Shipping (Git)
---

- NEVER commit, push, create branches, or create PRs without explicit user approval.
- Before any git operation that creates or modifies a commit, present a review block containing: commit title, commit/PR description, and release effect (none, patch or minor). ALWAYS wait for approval.
- Commit titles and PR titles follow Conventional Commits: `type(scope): subject`, at most 100 characters, lower-case subject, no period at the end.
- Wrap commit body lines at 100 characters. commitlint rejects longer lines.
- Never use `!` or a `BREAKING CHANGE:` footer on `main`, and never edit `CHANGELOG.md`.
- Commits must be signed. Never pass `--no-gpg-sign` or `--no-verify`.
