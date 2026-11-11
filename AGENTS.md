# AGENTS.md

**Generated:** 2026-05-09 · **Commit:** ba5b65a · **Branch:** main

## What this is

Go 1.25 CLI (`cobra`) that encrypts secrets (Argon2id + AES-256-GCM) into `.obscuro/` at the **git repo root**. Positioned around use by AI coding agents: names are listable without a password, and `obscuro run` / `obscuro inject` hand values to processes or deploy tools without printing them (agent guide: `website/content/docs/agents.mdx`, served at `/docs/agents` and `/agents.md`). `inject` also works as a Helm post-renderer replacing `__KEY__` placeholders via stdin/stdout. Self-upgrades from GitHub Releases. Bundled Next.js docs site lives in `website/`.

## Structure

```
main.go              → 7-line entrypoint → cmd.Execute()
cmd/                 → cobra commands, one file per verb + cmd_test.go (see cmd/AGENTS.md)
internal/crypto/     → Argon2id KDF + AES-256-GCM (see internal/crypto/AGENTS.md)
internal/store/      → .obscuro/ I/O + git-repo-root detection (see internal/store/AGENTS.md)
internal/keychain/   → 28-line wrapper around zalando/go-keyring; key=salt, service="obscuro"
internal/version/    → 1 var, set at link time via -ldflags (default "dev")
.github/workflows/   → ci.yml (fmt/build/vet/test/actionlint) · release.yml (release-please + cross-compile + cosign) · website.yml
release-please-config.json, .release-please-manifest.json → release-please config; manifest holds the current version
install.sh           → curl|sh installer; downloads release asset, verifies sha256, installs to ~/.local/bin
website/             → Next.js 16 + fumadocs docs site (its own AGENTS.md applies inside)
```

## Where to look

| Task                               | Location                                           |
| ---------------------------------- | -------------------------------------------------- |
| Add a new CLI command              | `cmd/<verb>.go` + register in its `init()`         |
| Change password resolution         | `cmd/root.go::getPassword`                         |
| Change on-disk format              | `internal/store/store.go` (`Config`, `Secrets`)    |
| Tune Argon2 params                 | `internal/crypto/crypto.go` constants              |
| Add a release asset / target       | `.github/workflows/release.yml` `targets=(...)`    |
| Modify update-check / self-upgrade | `cmd/upgrade.go` + `cmd/root.go::PersistentPreRun` |
| Docs content                       | `website/content/docs/*.mdx`                       |
| Agent-facing guide (`/agents.md`)  | `website/content/docs/agents.mdx`                  |

## Build

```bash
go build -ldflags "-X github.com/janklabs/obscuro/internal/version.Version=$(git describe --tags --always)" -o obscuro .
```

The `-ldflags` injection is **required** — without it `obscuro version` prints empty AND `PersistentPreRun` skips the update check (treats `version=="dev"` as a dev build).

## Test

```bash
go test ./...
```

Every test calls `t.TempDir()` → `os.Chdir` → `git init` → `store.ResetRoot()`. **`git` must be on PATH** — `store.RepoRoot()` shells out to `git rev-parse --show-toplevel` and caches via `sync.Once`. Forgetting `ResetRoot()` between tests leaks state across cases.

## CI checks (all must pass)

1. `gofmt -l .` — must be empty (run `gofmt -w .` before commit)
2. `go build ./...` and `go vet ./...`
3. `go test ./...`
4. `actionlint` on `.github/workflows/`

Website is built by a separate workflow. release-please ignores commits that only touch `website/` (`exclude-paths`), so they never bump the CLI version. It cannot exclude a single file, so commits that only change `.github/workflows/website.yml` must use a non-releasing type such as `ci:`.

## Release

[release-please](https://github.com/googleapis/release-please) (`release.yml`, on every push to `main`) keeps one release PR open from Conventional Commits (`feat:` minor, `fix:`/`perf:` patch, `!`/`BREAKING CHANGE` major; `docs:`, `ci:`, `chore:`, `test:`, `refactor:` don't release). Merging a PR to `main` never releases by itself; **merging the release PR** does: it commits `CHANGELOG.md` and `.release-please-manifest.json`, creates the `vX.Y.Z` tag and a **draft** GitHub Release. The workflow then runs the full `ci.yml` (`workflow_call`) on that exact commit, cross-compiles 6 targets (linux/darwin/windows × amd64/arm64, `CGO_ENABLED=0`, `-trimpath`, version via `-ldflags`) from it, writes `checksums.txt`, cosign-signs every file (keyless, identity `release.yml@refs/heads/main`), uploads them, and only then publishes the release and marks it latest. `install.sh` and `cmd/upgrade.go` resolve `releases/latest` and depend on the asset naming `obscuro-${VERSION}-${GOOS}-${GOARCH}[.exe]`. Don't edit `CHANGELOG.md` or the manifest by hand, and don't push tags. No manual trigger exists: recover a failed release with "Re-run failed jobs". Full process: README "Releases".

## Code map (top symbols)

| Symbol                       | File                         | Role                                                   |
| ---------------------------- | ---------------------------- | ------------------------------------------------------ |
| `cmd.Execute`                | cmd/root.go                  | Sole exported entrypoint from `main`                   |
| `cmd.Stdout` (`io.Writer`)   | cmd/root.go:25               | **Always write payload here, not `os.Stdout`** (tests) |
| `cmd.authenticate`           | cmd/root.go                  | Resolves key via the password order below              |
| `store.RepoRoot`             | internal/store/store.go:30   | `git rev-parse`-cached; **all paths derive from this** |
| `store.ResetRoot`            | internal/store/store.go:43   | Test-only; clears the `sync.Once`                      |
| `crypto.DeriveKey`           | internal/crypto/crypto.go:36 | Argon2id(t=3, mem=64MiB, lanes=4) → 32-byte key        |
| `crypto.Encrypt` / `Decrypt` | internal/crypto/crypto.go    | AES-256-GCM; output is `base64(nonce‖ct)`              |
| `crypto.VerifyKey`           | internal/crypto/crypto.go    | Decrypts the verification token to gate password       |

## Conventions

- Cobra: one command per file in `cmd/`, registered in that file's `init()` via `rootCmd.AddCommand`.
- Payload (the data the user asked for, e.g. `obscuro get`) → `cmd.Stdout`. Human messages → `os.Stderr`. Never mix.
- Password resolution order (do not reorder):
  `--password` flag → `--password-file` → OS keychain (keyed by salt) → `OBSCURO_PASSWORD` env → interactive TTY prompt.
- `.obscuro/` is always at the git repo root; never resolve relative to CWD.
- Keychain entries are keyed by the **base64 salt string** (not a fixed key), so re-`init` invalidates old entries automatically.
- Errors returned from `RunE` are surfaced by cobra; do not call `os.Exit` from commands.

## Anti-patterns (forbidden here)

- **Do not** commit plaintext secrets or decrypted `obscuro inject` output (see `website/content/docs/guides/{docker-compose,kubernetes}.mdx`).
- **Do not** make `fetchChangelog` or the update-check fatal — they must always degrade silently (`cmd/upgrade.go:204`, `cmd/root.go` 2 s timeout).
- **Do not** weaken Argon2 parameters or shrink `SaltSize`/`NonceSize` without a migration path — old vaults must still decrypt.
- **Do not** read raw `os.Stdout` in commands; use the `cmd.Stdout` indirection.
- **Do not** assume CWD == repo root; always go through `store.RepoRoot()` / `store.Dir()`.
- **Do not** add CGO. Release pipeline builds with `CGO_ENABLED=0` for static cross-platform binaries.

## Commands

```bash
gofmt -w .                         # required pre-commit
go test ./...                      # full test suite
go build -ldflags "-X github.com/janklabs/obscuro/internal/version.Version=$(git describe --tags --always)" -o obscuro .
OBSCURO_NO_UPDATE_CHECK=1 ./obscuro …   # silence the background self-update probe
pnpm --dir website dev             # docs site (Next.js 16 — see website/AGENTS.md)
```

## Notes

- `cmd/root.go` fires a goroutine in `PersistentPreRun` to check GitHub for a newer tag; result is consumed in `PersistentPostRun` with a 2 s timeout. Set `OBSCURO_NO_UPDATE_CHECK=1` in CI/scripts to skip.
- `inject` decrypts **all** secrets up-front (single password prompt) then does string replacement; placeholder format is exactly `__NAME__` (double underscore, no padding).
- Module path is `github.com/janklabs/obscuro` — keep `-ldflags -X` target in sync if it ever moves.
