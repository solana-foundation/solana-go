---
name: go-test-mutation
description: Prove that Go tests can fail, by running them against the old code or against small deliberate bugs with go test -overlay, without touching the working tree. Use after writing a regression test or the tests of a new feature, and when reviewing whether tests really cover a change.
argument-hint: "[package-or-file]"
---

# Prove That a Test Can Fail

A passing test proves little until you have seen it fail for the right reason. This skill has two checks:

1. A regression test fails on the code before the fix.
2. A small deliberate bug in the new code, a mutant, makes at least one test fail.

Both run through `go test -overlay`, which compiles replacement files in place of the files on disk. Nothing in the working tree changes, and nothing has to be committed, stashed or reverted.

## Key Rules

- A regression test MUST fail on the code without the fix. Check it. Do not assume it.
- For new code, break it on purpose where it matters and confirm that a test fails each time.
- Run both checks with `go test -overlay`. Never edit, stash, reset or commit files in the working tree to run them.
- A mutant that does not compile proves nothing. Only a failing assertion counts.
- When a mutant survives, add or strengthen a test, or explain why the change cannot be observed.
- Use distinct, non-zero values in test data, so that a swapped or dropped field shows up.

## Guidelines

### How the Overlay Works

The overlay is a JSON file with one field, `Replace`:

```json
{
  "Replace": {
    "rpc/ws/subscription.go": "/tmp/mut/subscription.go",
    "rpc/ws/new_file.go": ""
  }
}
```

- The key is the file on disk, absolute or relative to the directory where `go` runs.
- The value is the file to compile in its place. An empty value means that the file does not exist.
- Only the compiler sees the overlay. A file that a test reads when it runs, such as a file in `testdata/`, is still the real one.
- Keep the replacement files outside the repo, in a directory from `mktemp -d`.
- Always pass `-count=1`.

### Check 1: The Test Fails Without the Fix

Take the old version of every fixed file from git and compile the new tests against it:

```sh
tmp=$(mktemp -d)
git show HEAD:rpc/ws/subscription.go > "$tmp/subscription.go"
cat > "$tmp/overlay.json" <<EOF
{"Replace": {"rpc/ws/subscription.go": "$tmp/subscription.go"}}
EOF
go test -overlay "$tmp/overlay.json" ./rpc/ws/ -run 'TestName' -count=1
```

The test has to fail, and the message has to describe the bug.

- `HEAD:` is right while the fix is not committed. When it is committed, name the commit before the fix: `git show <commit>~1:<path>`.
- Add one entry for every file that the fix changed. Map a file that the fix added to `""`.
- Do not put test files in the overlay. The new tests run against the old code.
- If the old code does not compile with the new tests, for example because a test calls a function that the fix added, this check proves nothing. Use check 2: put the bug back into the fixed code by hand.

### Check 2: Mutants

Copy the current file, make one small change in the copy, and run the tests of the package:

```sh
tmp=$(mktemp -d)
cp rpc/ws/subscription.go "$tmp/subscription.go"
# Edit "$tmp/subscription.go": make one small change.
cat > "$tmp/overlay.json" <<EOF
{"Replace": {"rpc/ws/subscription.go": "$tmp/subscription.go"}}
EOF
go test -overlay "$tmp/overlay.json" ./rpc/ws/ -count=1 -timeout 120s
```

- Make one change for each run. With two changes you do not know which one was caught.
- Run all tests of the package, not only the new ones. When other packages depend on the behavior, run their tests too.
- Pass `-race` for a change to locking or to goroutines.

An example from this repo. The mutant stops wrapping the cause of a closed subscription:

```diff
-	return fmt.Errorf("%w: %w", ErrSubscriptionClosed, s.closeErr)
+	return fmt.Errorf("%w: %v", ErrSubscriptionClosed, s.closeErr)
```

```
--- FAIL: TestTypedRecv_ReturnsBufferedNotificationsBeforeCloseError (0.00s)
FAIL	github.com/gagliardetto/solana-go/rpc/ws	0.288s
```

A test caught it. A first try, `return s.closeErr`, left the `fmt` import unused. The result was `[build failed]`, which is not a catch.

### What to Mutate

Take the mutants from the diff of the change. Every branch and every constant that the change added deserves one.

- **Boundaries**: `<` to `<=`, a `+1` removed, a length or an offset that is one off.
- **Conditions**: an inverted `if`, one side of an `&&` or `||` dropped.
- **Errors**: `nil` returned in place of the error, `%w` changed to `%v`, another sentinel returned.
- **Encoders and decoders**: a field skipped, two fields swapped, a width or the byte order changed, a length check removed.
- **Instructions**: two accounts swapped, `SIGNER()` or `WRITE()` dropped, the tag changed.
- **State**: a map delete, a close or an unlock skipped, a `ctx.Done()` case removed.
- **Defaults**: a timeout or a limit changed.

### Read the Result

| Result | Meaning | What to do |
|---|---|---|
| `--- FAIL` with an assertion | The mutant was caught. | Check that the message is about the mutated behavior. |
| `[build failed]` | The mutant is not valid. | Change it so that it compiles, and run it again. |
| `ok` | The mutant survived. | Add or strengthen a test. If no test can see the change, say why. |
| The test hangs until the timeout | A test waits without a deadline. | Give the test its own deadline, so that it fails fast. |

### Why a Test Cannot Fail

When a mutant survives, the cause is often one of these:

- **A zero value as the expected value.** `solana.SystemProgramID` is 32 zero bytes, so it equals a key that was never set. The same holds for `0`, `""` and `nil`.
- **A round trip as the only check.** Encoding and then decoding passes when both sides have the same wrong layout. Pin the bytes as well. See the `upstream-parity` skill.
- **The same value in several fields.** A swap of two fields changes nothing.
- **Only `require.NoError`.** The call did not fail, but nothing checks what it did.
- **An instruction without parameters.** Its round trip compares two empty structs.
- **A sleep as the proof that nothing happened.** Prove it with a later event that can be observed.

### Limits

- The overlay cannot replace a file that is read when the test runs.
- You choose the mutants. This is not a coverage tool, and a clean result covers only what you tried.

## Command Process

When invoked as a command, follow these steps:

### Arguments

- `[package-or-file]` (optional): the package or the file to check. The default is every Go file, other than tests, that the current branch changes, committed or not.

### Steps

1. List the changed files with `git diff --name-only main... -- '*.go'` and `git status --short`.
2. For a bug fix, run check 1 on the new tests.
3. Choose the mutants from the diff, and list them before you run anything.
4. Run each mutant through the overlay. Record the result: caught, survived or not valid.
5. For every mutant that survived, propose a test. Do not change the code to make a mutant fail.
6. Report a table: the mutant, the file and line, the result, and the test that caught it.
7. Remove the temporary directory, and confirm with `git status` that the working tree did not change.
