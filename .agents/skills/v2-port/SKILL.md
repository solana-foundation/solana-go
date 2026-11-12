---
name: v2-port
description: Bring changes from main to the v2 branch of solana-go, or prepare a change that only v2 can take. Use when asked to sync v2 with main, to port a commit or a pull request to v2, or to make a change that breaks the v1 API.
argument-hint: "[sync | <commit-or-pr>]"
---

# Port Changes to v2

solana-go ships two modules from two branches. `main` is v1, `github.com/gagliardetto/solana-go`. `v2` is `github.com/solana-foundation/solana-go/v2`. This skill covers the three ways a change reaches `v2`:

- **The sync** of `v2` with `main`, which maintainers do. It brings everything that is on `main`.
- **A port of one change** that is already on `main`, when `v2` needs it before the next sync.
- **A pull request for `v2` only**, for a change that cannot go to `main`.

The commands below call the remote of `solana-foundation/solana-go` `upstream`. Use the name that your clone has.

## Key Rules

- NEVER rebase, force-push or tag `v2` without explicit user approval. A sync rewrites the history of a shared release branch. Only a maintainer starts it, and only a maintainer pushes it.
- Everything lands on `main` first. Only a change that breaks the v1 API, or that touches code that exists only on `v2`, targets `v2` directly.
- A change that is on `main` reaches `v2` with the next sync. Port it alone only when `v2` needs it before that, and name the original commit in the port.
- A sync may drop a port only after you compared it with its original. A change that the port made for the v2 API exists nowhere else.
- On `v2`, every import uses `github.com/solana-foundation/solana-go/v2`. No Go file may import `github.com/gagliardetto/solana-go`.
- After you rewrite import paths, format the files. The new path sorts differently inside an import block.
- Do not stop at the import paths. Code that comes from `main` may not fit the v2 API, so build and test everything.

## Guidelines

### How v2 Relates to main

- `v2` is `main` plus a short stack of commits that only `v2` has. List them with `git log --oneline upstream/main..upstream/v2`.
- At every sync, that stack is rebased onto the tip of `main`. The history of `v2` is rewritten each time, so an older `v2` tag can point at a commit that is no longer on the branch.
- Release Please, `semver-check` and `commit-lint` run only for `main`. On `v2`, maintainers cut the tags by hand. The `tests` and `lint` checks run for every pull request.

### What Is Different on v2

- **The module path**, in `go.mod` and in every import.
- **`rpc/ws`**: the subscription API follows the JSON-RPC spec (pull request #407). Result types differ from `main`.
- **`rpc/sendAndConfirmTransaction`**: adapted to that WebSocket API.
- **`programs/loader-v2`, `programs/loader-v3` and `programs/loader-v4`** exist only on `v2` (pull request #410).
- **`rpc.GetVoteAccounts`**: `EpochCredits` is `[][]uint64` (pull request #486).
- **`README.md`**: the install instructions and the version numbers.

This list goes stale. The stack of commits is the truth, so read it before you start.

### Port One Change From main

Use this for a commit that is already on `main`, when `v2` needs it before the next sync. Creating the branch and applying the change need approval first, as the `shipping-git` skill describes.

**1. Find the commit, and check that v2 does not have it yet.**

```sh
git fetch upstream
gh pr view <number> --json mergeCommit --jq .mergeCommit.oid   # for a pull request: its commit on main
git merge-base --is-ancestor <commit> upstream/v2 && echo "already on v2"
```

**2. Read the change.** Run `git show --stat <commit>`, then read the full diff. Note which files are new, and whether the change touches code that differs on `v2`.

Then check what the change builds on:

```sh
base=$(git merge-base upstream/main upstream/v2)
git log --oneline "$base"..<commit>~1 -- <files of the change>
```

Every commit in this list is on `main`, is not on `v2`, and touches the same files. If the change needs one of them, port that one first, or wait for the sync.

**3. Predict the conflicts without changing anything.**

```sh
git merge-tree --write-tree --name-only --merge-base=<commit>~1 upstream/v2 <commit>
```

Exit code 0 means that the change applies without conflicts. That does not mean that it builds.

**4. Apply the change on a branch from v2.**

```sh
git switch -c <type>/<short-name>-v2 upstream/v2
git cherry-pick --no-commit <commit>
```

- `--no-commit` stages the change and creates no commit, so the result stays open for review.
- In a conflict, keep the v2 API and bring in the behavior of the change. Read both sides before you choose.

**5. Adjust the imports.** The change comes with the v1 path, in the new files and in every import line that it touched:

```sh
git diff --cached --name-only --diff-filter=d -z -- '*.go' |
  xargs -0 perl -pi -e 's{github\.com/gagliardetto/solana-go(?![\w-])}{github.com/solana-foundation/solana-go/v2}g'
git diff --cached --name-only --diff-filter=d -z -- '*.go' | xargs -0 gofmt -w
git grep -n 'github.com/gagliardetto/solana-go' -- '*.go'   # prints nothing
```

Code snippets in the README and in doc comments need the v2 path too.

**6. Fit the change to the v2 API and run all checks.**

```sh
go build ./...
go vet ./...
go test ./... -race -count=1
golangci-lint run
```

Run all tests, not only the tests of the packages that the change touched. Fix what does not fit in the ported code. Do not change the v2 API to make it compile.

**7. Hand over.** The maintainer commits and opens the pull request with `v2` as the base.

- The title is the title of the original commit, without its pull request number.
- The body names the original: `(cherry picked from commit <sha>)`. The next sync needs this line to recognize the port.
- When step 6 changed more than the imports, the body lists those changes. They exist only on `v2`, and the next sync has to keep them.

### A Pull Request for v2 Only

Use this for a change that is not on `main` and cannot go there. A change that is already on `main` is a port: see "Port One Change From main".

1. Confirm that the change cannot go to `main`. Say what it breaks in v1, or which v2-only code it touches. If a compatible form exists, it goes to `main`. See the `go-api-compat` skill.
2. Branch from `upstream/v2`.
3. Use the v2 import path in every file.
4. Run `go test ./... -race -count=1` and `golangci-lint run`.
5. Open the pull request with `v2` as the base. Write the title as a Conventional Commit, although `commit-lint` does not run there.
6. `v2` has releases too. A change that breaks v2 users is a decision for the maintainers, so say clearly in the pull request what breaks.

### Sync v2 With main

Do this only when the user asks for a sync. Every step that creates or rewrites a commit needs approval first, as the `shipping-git` skill describes.

**1. See what is new.**

```sh
git fetch upstream
base=$(git merge-base upstream/main upstream/v2)
git log --oneline "$base"..upstream/main      # new on main since the last sync
git log --oneline upstream/main..upstream/v2  # the v2-only stack
git rev-parse upstream/v2                     # write this down: the old tip
```

**2. Predict the conflicts without changing anything.**

```sh
git merge-tree --write-tree --merge-base="$base" upstream/main upstream/v2
```

Exit code 0 means no conflicts. If there are conflicts, the output names the files. `README.md` conflicts often, because both branches edit the same sections. A file that merges cleanly can still fail to build: look closely at `rpc/ws` and `rpc/sendAndConfirmTransaction`, where the v2 API differs.

**3. Check the ports in the stack.** A commit that was ported from `main` since the last sync is part of `main` now. Find the ports by the line that names their original:

```sh
git log --oneline --grep='cherry picked from commit' upstream/main..upstream/v2
```

Compare each port with its original before you rebase:

```sh
git range-diff <original>~1..<original> <port>~1..<port>
```

- A difference in the import path, or in the order of the imports, is expected. `main` has everything else of such a port.
- Every other difference is a change that the port made for the v2 API. `main` does not have it, and rewriting imports does not bring it back. Write each one down: it has to survive the sync.

**4. Rebase the stack onto main.**

```sh
git rebase --onto upstream/main "$base" v2
```

- Name the branch at the end of the command. The form that ends on a commit leaves `HEAD` detached, and later commits then miss the branch.
- In a conflict, keep the v2 API in the files that the stack changed, and keep the new behavior that `main` added. Read both sides before you choose. Do not take one side for a whole file.
- **A port without v2-specific changes** leaves the stack. Git drops it when it becomes empty. When the rebase stops on it with a conflict, skip it with `git rebase --skip`: `main` has the change, and step 5 fixes the import paths.
- **A port with v2-specific changes** stays. Never skip it. When the rebase stops on it, resolve the conflict: take what `main` has for the parts that the original also changed, and keep every v2-specific change from your list. Then run `git rebase --continue`. The commit that remains holds only the v2-specific part.

**5. Finish the module path.** Files that `main` added since the last sync still import the v1 path:

```sh
git grep -lz 'github.com/gagliardetto/solana-go' -- '*.go' |
  xargs -0 perl -pi -e 's{github\.com/gagliardetto/solana-go(?![\w-])}{github.com/solana-foundation/solana-go/v2}g'
git diff --name-only -z -- '*.go' | xargs -0 gofmt -w
```

- The pattern leaves other modules of the same owner alone, such as `github.com/gagliardetto/binary`.
- `gofmt -w` sorts the import blocks again. Without it the `lint` check fails.
- Check the result: `git grep -n 'github.com/gagliardetto/solana-go' -- '*.go'` prints nothing.
- Look at files that are not Go by hand. The README of `v2` names the v1 path on purpose in a few places.

**6. Fit the new code to the v2 API.**

```sh
go build ./...
go vet ./...
```

`go vet` also compiles the tests and the examples. Fix every error in the code that came from `main`. Do not change the v2 API to make the new code compile.

**7. Run all checks.**

```sh
go test ./... -race -count=1
golangci-lint run
```

**8. Compare before and after.**

```sh
git range-diff "$base"..<old-tip> upstream/main..v2
```

- Every commit of the stack should show up once, with only the changes that the conflicts required.
- A commit that is missing has to be a port without v2-specific changes. Any other missing commit is lost work.
- For every port that had v2-specific changes, check that each change on your list from step 3 is still in the code.

To undo the sync, run `git reset --hard <old-tip>`.

**9. Hand over.** The maintainer commits the changes of steps 5 and 6, pushes and tags. Earlier syncs used these titles:

```
chore: complete /v2 module rename for files added on main since last sync
docs(readme): update version references for v2.3.0
```

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `sync`: sync `v2` with `main`.
- `<commit-or-pr>`: port one commit or one pull request to `v2`.

### Steps

1. Read the v2-only stack with `git log --oneline upstream/main..upstream/v2`, so that you know what differs today.
2. For `sync`, follow "Sync v2 With main". Stop after each step that rewrites history, show the result and wait for approval.
3. For one commit or pull request, first find out where the change is. `git merge-base --is-ancestor <commit> upstream/main` exits with 0 when it is on `main`.
    - **It is on `main`**: follow "Port One Change From main".
    - **It is not on `main`, and it cannot go there**, because it breaks the v1 API or touches code that only `v2` has: follow "A Pull Request for v2 Only".
    - **It is not on `main`, but it can go there**: it goes to `main` first. Say so and stop.
4. Report what was ported, every conflict and how it was resolved, the results of the checks, and, after a sync, the old tip of `v2` for an undo.
