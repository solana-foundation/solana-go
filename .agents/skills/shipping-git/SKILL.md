---
name: shipping-git
description: Shipping workflow for solana-go with Git and the GitHub CLI. Use before committing, creating a branch, pushing, or opening or updating a pull request.
---

# Ship Code Using Git

Guidance for shipping changes to solana-go with standard Git and the GitHub CLI (`gh`). In this repo the wording of a commit is part of the change: commitlint checks every commit of a pull request, and Release Please builds the version number and `CHANGELOG.md` from the commits on `main`.

## Key Rules

- NEVER commit, push, create branches, or create PRs without explicit user approval.
- Before any git operation that creates or modifies a commit, present a review block containing: commit title, commit/PR description, and release effect (none, patch or minor). ALWAYS wait for approval.
- Commit titles and PR titles follow Conventional Commits: `type(scope): subject`, at most 100 characters, lower-case subject, no period at the end.
- Wrap commit body lines at 100 characters. commitlint rejects longer lines.
- Never use `!` or a `BREAKING CHANGE:` footer on `main`. Breaking changes go to the `v2` branch.
- Never edit `CHANGELOG.md`. Release Please writes it.
- Commits must be signed. Never pass `--no-gpg-sign` or `--no-verify`.
- Force-push only your own branch, only with `--force-with-lease`, and only when the user asks.

## Guidelines

### Commit Titles

```
<type>(<scope>): <subject>
```

Pick the type by what a user of the SDK sees, because the type decides the release:

| Type | Use it for | Changelog section | Release |
|---|---|---|---|
| `feat` | A new feature or a new public API | Features | minor |
| `fix` | A bug fix | Bug Fixes | patch |
| `perf` | A performance improvement | Performance Improvements | patch |
| `docs` | Documentation only | none | none |
| `refactor` | A change that neither fixes a bug nor adds a feature | none | none |
| `test` | Tests only | none | none |
| `chore` | Build, CI, dependencies and tooling | none | none |

- The scope is the package or the area that changed: `ws`, `rpc`, `token-2022`, `message`, `keys`, `base58`, `readme`, `ci`, `deps`. Leave it out when the change has no single area.
- The subject uses the imperative mood and the present tense: "add", not "added" or "adds".
- The subject says what changed for the user, not how the code changed. It becomes a line of the changelog.
- Count the characters before you propose a title. The limit of 100 includes the type and the scope.

```
feat(rpc): add getTransactionsForAddress
fix(token-2022): send num_token_accounts in WithdrawWithheldTokensFromAccounts
docs(readme): document address lookup tables
```

### Commit Bodies

- Say why the change is needed first, then what it does.
- Wrap every line at 100 characters or fewer. A long URL goes on its own line and still has to fit.
- Leave one blank line between the title and the body.

### Creating Commits

```sh
git add path/to/file.go path/to/file_test.go
git commit -m "fix(ws): close pending subscriptions when the read loop exits" -m "The read loop returned without closing the subscriptions that still waited
for the server. Their callers blocked forever. Close them with
ErrSubscriptionClosed."
```

- Stage the files of the change by name. Do not use `git add -A`, because it also takes scratch files.
- One commit holds one logical change.

### Creating Branches

Branch from an up-to-date `main`, with a prefix that matches the commit type:

```sh
git checkout -b feat/short-description main
```

Common prefixes: `feat/`, `fix/`, `docs/`, `refactor/`, `chore/`.

### Before You Propose a Pull Request

Run the checks that CI runs:

```sh
go test ./... -race -count=1
golangci-lint run
gorelease
```

`gorelease` is only needed when the change touches exported identifiers. It reads the last commit, so it needs a clean working tree. See the `go-api-compat` skill.

### Creating Pull Requests

```sh
gh pr create --base main --title "feat(rpc): add getTransactionsForAddress" --body "..."
```

- The base is `main`, unless the change is breaking or exists only on `v2`.
- The title is a Conventional Commit. Pull requests are squash-merged, and the squash commit becomes a line of the changelog. For its title, GitHub proposes the title of the only commit when there is one, and the PR title when there are more. So when a pull request has one commit, give the commit and the pull request the same title.
- There is no pull request template. State the problem, then the changes, then how you tested them. Recent pull requests use the headings `#### Problem` and `#### Summary of Changes`.
- Never describe a security problem that is not fixed yet in a public pull request. See `SECURITY.md`.

### Pushing Changes

```sh
git push -u origin HEAD
```

### When commit-lint Fails

The check names the commit and the rule, for example `header must not be longer than 100 characters`.

- For the last commit, change the message with `git commit --amend`.
- For an older commit, the message has to be reworded in a rebase. Ask the user before you rewrite history.
- Then push with `git push --force-with-lease`, after the user approves it.

### Review Block Format

Before any git operation, present this review block and wait for approval:

1. **Commit title** - a Conventional Commit, with its length when it is close to 100 characters.
2. **Commit/PR description** - why the change is needed and what it does.
3. **Release effect** - none, patch or minor, from the commit type.
