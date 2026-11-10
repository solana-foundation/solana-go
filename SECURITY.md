# Security Policy

1. [Reporting security problems](#reporting)
2. [Incident Response Process](#process)

<a name="reporting"></a>

## Reporting security problems in the Solana Go SDK

**DO NOT CREATE A GITHUB ISSUE OR PULL REQUEST** to report a security problem.

Instead please use this [Report a Vulnerability](https://github.com/solana-foundation/solana-go/security/advisories/new) link. Provide a helpful title, detailed description of the vulnerability and an exploit proof-of-concept. Speculative submissions without proof-of-concept will be closed with no further consideration.

This policy covers both Go modules published from this repository:

- `github.com/gagliardetto/solana-go` (v1, `main` branch)
- `github.com/solana-foundation/solana-go/v2` (v2, `v2` branch)

Please include in the advisory:

- the affected module path and version (for example, from `go list -m all | grep solana-go` in your project)
- the affected packages and exported functions or methods
- your Go version (`go version`) and GOOS/GOARCH, if relevant
- a proof-of-concept that runs against a released version, ideally a minimal `go test`

If you haven't done so already, please **enable two-factor auth** in your GitHub account.

---

If you do not receive a response in the advisory, send an email to the Solana Foundation at disclosures@solana.org with the full URL of the advisory you have created. DO NOT include attachments or provide detail sufficient for exploitation regarding the security issue in this email. **Only provide such details in the advisory**.

<a name="process"></a>

## Incident Response Process

In case an incident is discovered or reported, the following process will be followed to contain, respond and remediate:

### 1. Accept the new report

In response to a newly reported security problem, a maintainer of `solana-foundation/solana-go` will accept the report to turn it into a draft advisory. The Solana Foundation security team should be added to the draft security advisory and should create a temporary private fork of the repository (grey button towards the bottom of the page) if necessary.

If the advisory is the result of an audit finding, follow the same process as above but add the auditor's GitHub user(s) and begin the title with "[Audit]".

If the report is out of scope, a maintainer will comment as such and then close the report.

### 2. Triage

Within the draft security advisory, discuss and determine the severity of the issue and which module lines and versions are affected. The `v2` branch contains everything on `main`, so most issues affect both. Record the affected packages and exported functions or methods in the advisory: the Go vulnerability database lists them so that `govulncheck` only flags applications that actually call the vulnerable code. If necessary, the Solana Foundation security team may add other GitHub users to the advisory to assist. If the report turns out not to be a security vulnerability, close the advisory and, if more follow-up is required, open a normal public GitHub issue. A confirmed vulnerability of any severity stays in the private advisory through the fix, patch and release steps below.

If the root cause is in a dependency rather than in solana-go, report it to that project (for the Go standard library and `golang.org/x` modules, follow the [Go security policy](https://go.dev/doc/security/policy)) and ship the solana-go side as a dependency upgrade.

### 3. Prepare Fixes

Prepare a fix for the issue and push it to the private fork associated with the draft security advisory, with one branch per affected line (based on `main` for v1 and on `v2` for v2). There is no CI available in the private fork, so you must run the checks from `.github/workflows` locally on each branch, at minimum:

```sh
go test ./... -race -count=1
golangci-lint run
gorelease   # go install golang.org/x/exp/cmd/gorelease@latest
```

Prefer a fix that does not change the exported API, so it can ship as a patch release. Code review from the reporter is ideal, as well as from multiple maintainers.

Do not push the fix, or any commit or tag that reveals it, to this repository or to any public fork before disclosure. Any public commit can be fetched through the Go module mirror (`proxy.golang.org`), which may keep serving it even after it is deleted, and its checksum is recorded permanently in the checksum database (`sum.golang.org`).

### 4. Ship the patch

Once the fix is accepted, a member of the Solana Foundation security team should export it with `git format-patch` for the latest release of each affected line, check that each patch applies, builds and passes the tests on that release, and name that release with the patch. Within a major version, newer releases keep the public API, so an application on an older release can usually build against the patched one. When an important application cannot, prepare and test a separate patch for its release.

Share the patches privately with as many vulnerable applications as possible, only through channels that control who can read them: add the application maintainers as collaborators on the draft security advisory, or send the patches encrypted to known maintainer contacts. Do not post them anywhere that anyone with the link can read, such as a secret (unlisted) GitHub Gist, before the public disclosure in step 5.

Send these instructions with each patch. They build the application against the patched release through a [`replace`](https://go.dev/ref/mod#go-mod-file-replace) directive:

```sh
# Check out the release named with the patch, then apply the patch
git clone https://github.com/solana-foundation/solana-go.git ../solana-go-patched
git -C ../solana-go-patched checkout vX.Y.Z   # the release named with the patch
git -C ../solana-go-patched apply /path/to/fix.patch

# Build the application against the patched copy
# (on v2, replace github.com/solana-foundation/solana-go/v2 instead)
go mod edit -replace github.com/gagliardetto/solana-go=../solana-go-patched
go build ./...
```

A `replace` directive only applies in the main module (the application being built), so a library that depends on solana-go cannot pass the patch on to its users. Applications that vendor their dependencies must also re-run `go mod vendor`.

### 5. Public Disclosure and Release

Once the fix has been deployed to as large an affected application set as practical, the patches from the security advisory may be merged into `main` and `v2`. A new patch release should be shipped on each affected line (v1 from `main` via the Release Please PR, v2 from the `v2` branch), and the affected releases retracted:

1. Add a [`retract`](https://go.dev/ref/mod#go-mod-file-retract) directive to `go.mod` on each affected line, include it in the patched release, and keep it in later releases. `go get` and `go mod tidy` will no longer upgrade to a retracted version, users still on one are warned by `go get` and `go list -m -u all`, and pkg.go.dev marks it as retracted:

   ```
   retract [v1.20.0, v1.26.0] // Security vulnerability: GHSA-xxxx-xxxx-xxxx
   ```

2. Publish the GitHub security advisory with each affected module path listed under the Go ecosystem, along with its affected and patched versions, and request a CVE if warranted.

3. Check that the advisory reaches the [Go vulnerability database](https://go.dev/doc/security/vuln/database), so that [`govulncheck`](https://go.dev/doc/security/vuln/) flags affected applications. GitHub advisories are usually imported automatically; if it is missing, report it at https://go.dev/s/vulndb-report-new.

Applications that applied the patch can then drop it with `go mod edit -dropreplace github.com/gagliardetto/solana-go` and upgrade with `go get github.com/gagliardetto/solana-go@vX.Y.Z`.
