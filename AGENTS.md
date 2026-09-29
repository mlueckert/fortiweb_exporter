# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project

Prometheus exporter for FortiWeb (WAF) appliances, written in Go (see `go.mod` for the Go version).
Module path: `github.com/mlueckert/fortiweb_exporter`. The design follows `fortigate_exporter`: a multi-target `/probe` endpoint plus the exporter's own `/metrics`.

Go style rules live in `.claude/rules/go.instructions.md`; follow them for all `.go` changes.

## Basic Information

- **Language:** Written in Go.
- **Supported Platforms:** Linux, Windows, FreeBSD, and macOS.

## Layout

- `main.go` – entrypoint, HTTP server, IP restriction middleware, build info (`Version`, `GitHash` injected via `-ldflags`).
- `internal/config` – CLI flags and the auth map file (`fortiweb-key.yaml`, see `fortiweb-key.yaml.example`).
- `pkg/http` – FortiWeb API client (`FortiHTTP` interface). Authenticates with `Authorization: <base64 {"username","password","vdom"}>` or a pre-encoded token.
- `pkg/probe` – `probe.go` holds the probe registry and include/exclude logic; one file per probe (`<area>_<name>.go`) with a matching `_test.go`.
- `pkg/probe/testdata/*.jsonnet` – API response fixtures, named after the endpoint (e.g. `system_status_systemstatus.jsonnet` for `api/v2.0/system/status.systemstatus`).
- `docs/` – FortiWeb 8.0 Monitor API specs (JSON). Use these to find endpoints and response fields.
- `.github/workflows` – `ci.yml` (branches/PRs), `release.yml` (semantic-release on `main`).

## Build and test

```bash
make build        # binary in target/fortiweb-exporter
make build-release # cross-compiled binaries in target/
make clean        # remove target/
make vulncheck    # govulncheck on dependencies and stdlib
make hooks        # enable the betterleaks pre-commit hook (.githooks/)
make secrets-check # betterleaks scan of the full git history
make test         # gofmt check, go vet, go test -race ./...
gofmt -w .        # fix formatting
go mod tidy       # after dependency changes
```

Always run `make test` before finishing a change.

## Adding or changing a probe

1. Look up the endpoint and response schema in `docs/`.
2. Add `pkg/probe/<area>_<name>.go` with `func probeXxx(c http.FortiHTTP) ([]prometheus.Metric, bool)`. Call `c.Get("api/v2.0/...", "", &resp)`, log errors with `log.Printf("Error: ...")` and return `false` on failure.
3. Register it in the probe list in `pkg/probe/probe.go` with a category name (e.g. `System/Status`, `Policy/Status`).
4. Look for a jsonnet fixture in `pkg/probe/testdata/` and create tests that use `newFakeClient()`, `c.prepare(...)`, `testProbe(...)` and `testutil.GatherAndCompare`, plus an error case using `&brokenClient{}`. If not jsonnet files are there, ask the dev to get it from a real device.
5. Update the "Metrics" section of `README.md` and the probe lists in `fortiweb-key.yaml.example`.

Metric rules:
- Prefix with `fortiweb_`, follow Prometheus naming (base units such as `_seconds`/`_bytes`, `_total` for counters, `_ratio` for 0–1 values, `_info` for info metrics with value 1).
- Never rename or remove existing metrics or labels once released; that is a breaking change (`feat!:`).
- Category names must not be a prefix of another category name; include/exclude uses prefix matching.

## Commits and releases

- Use Conventional Commits (`fix:`, `feat:`, `feat!:` / `BREAKING CHANGE:`, `chore:`, `docs:`, `test:` …). The release version is derived from them.
- Work on a branch and open a PR to `main`. Pushes to `main` run `release.yml`: semantic-release creates the tag, GitHub release notes and cross-compiled binaries (`make build-release`, configured in `.releaserc.yml`).
- Do not create release tags manually.
- Build version is injected via `make build VERSION=... GIT_HASH=...`.

## Don'ts

- Never commit `fortiweb-key.yaml` or any real credentials/tokens; only edit `fortiweb-key.yaml.example`. Do not bypass the betterleaks hook with `--no-verify` unless the finding is a confirmed false positive.
- Do not commit build artifacts (`target/`, `fortiweb_exporter`, `cover.out`, `coverage.html`).
- Do not edit `go.sum` by hand.
