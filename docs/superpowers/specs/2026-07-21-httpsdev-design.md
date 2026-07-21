# httpsdev v0.1 — Design Spec

**Date:** 2026-07-21
**Status:** Approved for planning
**Owner:** @ArpitRajputGithub

## Purpose

One command that gives any local dev server a browser-trusted HTTPS front end. Solves the recurring "how do I run HTTPS locally" pain that every framework re-invents differently, and does it without config files, framework detection, or reverse-proxy DSLs.

## Pitch

```
$ mkcert -install       # one time, system-wide
$ httpsdev 5173         # https://localhost:3443  →  http://localhost:5173
```

Two commands. Works with any HTTP dev server (Vite, Next, Rails, Django, Flask, Rust, Go, whatever).

## Non-Goals (v0.1)

Explicitly out of scope. Each becomes a GitHub issue for future consideration; none block v0.1.

- Child-process wrapping (`httpsdev npm run dev`).
- Multi-service reverse proxy with hostnames (`api.dev.local`, `web.dev.local`) — Caddy already covers this space.
- Framework auto-detection (reading `package.json`, parsing stdout for ports).
- Windows-first testing (build works via `GOOS=windows`; correctness testing is macOS + Linux only).
- Systemd / launchd daemon mode.
- Metrics endpoint (`/metrics`, Prometheus).
- Web UI dashboard.
- Custom CA — we shell out to mkcert instead of embedding one.

## Architecture

Single Go binary. Three responsibilities, executed in order at startup:

### 1. Cert acquisition

- Cert cache location: `~/.config/httpsdev/certs/localhost.pem` and `localhost.key` (respects `XDG_CONFIG_HOME`).
- On startup: if cert files are missing OR their mtime is older than 30 days, shell out to `mkcert` to regenerate:
  ```
  mkcert -cert-file <cache>/localhost.pem \
         -key-file  <cache>/localhost.key \
         localhost 127.0.0.1 ::1
  ```
- If `--host <name>` is passed, add it to the mkcert SAN list.
- If `mkcert` is not on `PATH`: exit 1 with message `mkcert not found. Install: brew install mkcert && mkcert -install`.

### 2. TLS server

- `net/http.Server` with `crypto/tls.Config` loading the cached PEM + key.
- Default listen port: `3443` (override with `--listen <port>`).
- `MinVersion: tls.VersionTLS12`.
- Bind address: `127.0.0.1` by default (safe — nothing on the LAN can reach your dev server). Pass `--lan` to bind on `0.0.0.0` for on-device mobile testing over wifi.
- If the listen port is already bound: exit 1 with `port <n> already in use — try --listen <other>`.

### 3. Reverse proxy

- `httputil.NewSingleHostReverseProxy` targeting `http://127.0.0.1:<target-port>`.
- WebSocket + SSE work out of the box via stdlib's built-in `Hijacker` path (Go 1.22+).
- Preserves `Host` header from the client.
- On upstream connection error: return HTTP 502 with a small plain-text body (`upstream unreachable at localhost:<n>`).
- Do NOT retry — the dev server may just not be up yet; retry loops mask real errors.

## CLI Surface

```
httpsdev <target-port> [flags]

Flags:
  --listen <port>   HTTPS listen port          (default: 3443)
  --host <name>     Additional SAN for cert    (default: none, cert is for localhost only)
  --lan             Bind on 0.0.0.0 for LAN    (default: false, binds 127.0.0.1)
  --tui             Full-screen dashboard mode (default: false, plain log mode)
  --version         Print version and exit
  --help            Print help and exit
```

Four functional flags. That's the entire surface.

## Output Modes

### Default: plain log mode

Prints a startup banner to stderr, then one line per request to stdout. Fully pipeable (`httpsdev 5173 | grep 500`).

**Startup banner (stderr):**
```
  ▲ httpsdev v0.1.0

  ➜  Local:    https://localhost:3443
  ➜  Upstream: http://localhost:5173
  ➜  Cert:     mkcert · valid 30 days

  press Ctrl+C to quit
```

**Per-request line (stdout), colorized when isatty:**
```
GET   /                     200   12ms
POST  /api/login            401   45ms
GET   /favicon.svg          404    1ms
```

Colors: green for 2xx, yellow for 3xx, red for 4xx/5xx. Suppressed automatically when stdout is piped.

**On Ctrl+C (stderr):**
```
served 47 requests · 2 errors · avg 8ms · uptime 3m21s
```

### Optional: `--tui` dashboard mode

Full-screen TUI via `bubbletea` + `lipgloss`. Alt-buffer, restored on exit.

Layout:
- Top panel: live stats (req/s, p50/p95 latency, error rate, uptime).
- Middle panel: scrolling request feed, last 20 requests, colored by status.
- Bottom bar: hotkeys (`q` quit, `c` clear feed, `1`/`2` tabs).
- Tabs: `1` = requests, `2` = cert info (subject, issuer, expiry).
- Refresh at 10 FPS (adjustable via internal constant, not exposed as flag).

**Purpose:** README hero GIF. Not a user-facing default because TUIs break pipes and copy-paste.

## Data Flow

```
browser ──HTTPS──▶ httpsdev:3443 ──HTTP──▶ dev-server:5173
                        │
                        └── mkcert-signed cert for localhost (+ optional --host SAN)
```

Bidirectional streaming preserved end-to-end (WebSocket for Vite HMR, SSE for Next dev overlay, etc.).

## Error Handling

| Condition | Behavior |
|-----------|----------|
| `mkcert` not on PATH | Exit 1, one-line install instruction. |
| mkcert exec fails | Exit 1, print mkcert's stderr. |
| Cert file unreadable | Exit 1, print path + errno. |
| Listen port already bound | Exit 1, suggest `--listen`. |
| Target port not reachable at startup | Warn to stderr, keep running. Dev server may start later. |
| Upstream connection refused during a request | HTTP 502 with plain-text body. |
| Upstream timeout | Default Go proxy behavior (504-ish). No custom timeout logic in v0.1. |
| SIGINT / SIGTERM | Graceful shutdown: stop accepting, drain in-flight, print summary line, exit 0. |

No retries, no circuit breakers, no logging library. Stdlib `log` to stderr.

## File Layout

```
httpsdev/
├── main.go                        # cli parsing, startup, wiring
├── proxy.go                       # TLS server + reverse proxy + request logger
├── cert.go                        # mkcert shell-out + cache check
├── ui_log.go                      # default log-mode banner + per-request output
├── ui_tui.go                      # bubbletea TUI (--tui mode)
├── proxy_test.go                  # integration test: dummy upstream → proxy → https client
├── cert_test.go                   # cert cache + expiry logic (unit)
├── go.mod
├── go.sum
├── README.md
├── LICENSE                        # MIT
├── .goreleaser.yaml
└── .github/workflows/release.yml  # GoReleaser on tag push
```

Six `.go` files at ~100-200 LOC each. No packages, no interfaces, no config struct. Split when a file crosses 300 LOC — not before.

## Testing

**Two tests total, both stdlib-only.**

1. **`proxy_test.go` — integration.** Start a dummy `httptest.NewServer` upstream. Generate a self-signed cert inline (bypasses mkcert dependency in CI). Start `httpsdev`'s proxy against the dummy. Make an HTTPS request through it with `InsecureSkipVerify` (test-only). Assert body round-trips and status code passes through. Covers cert loading, TLS handshake, and proxy forwarding in one go.

2. **`cert_test.go` — unit.** Test the "should we regenerate?" logic: missing files → yes; files younger than 30 days → no; files older than 30 days → yes. Uses a temp dir, no shell-out.

No test framework, no fixtures, no mocks. `go test ./...` runs both.

## Release / Distribution

- **GoReleaser** on git tag `v*` → GitHub Release with prebuilt binaries:
  - macOS: arm64, amd64
  - Linux: arm64, amd64
  - Windows: amd64 (untested but built)
- **`go install github.com/ArpitRajputGithub/httpsdev@latest`** works from day 1.
- **Homebrew tap** deferred to v0.2 (adds release-pipeline complexity for weekend v0.1).

## Dependencies

Direct, all pinned in `go.mod`:

| Dep | Why | Alternative considered |
|-----|-----|------------------------|
| `github.com/fatih/color` | Colorized log-mode output | Raw ANSI codes — rejected: not worth the pixel-fiddling for one file. |
| `github.com/charmbracelet/bubbletea` | TUI framework for `--tui` | `tview` — bubbletea has better ecosystem, better docs, more stars = signal for recruiters. |
| `github.com/charmbracelet/lipgloss` | Styling for the TUI | Ships with bubbletea ecosystem. |

Total binary size target: **< 15 MB** stripped. Bubbletea is the biggest contributor.

Stdlib only: `net/http`, `net/http/httputil`, `crypto/tls`, `os/exec`, `flag`, `context`.

## Success Criteria for v0.1

- `httpsdev 5173` proxies a running Vite dev server, browser shows a green padlock, HMR (WebSocket) works.
- `--tui` mode produces a screenshot-worthy dashboard that renders correctly in `asciinema` recordings.
- Both tests pass in CI (GitHub Actions) on macOS + Linux.
- Release binaries downloadable from GitHub Releases page.
- README has: one-paragraph pitch, animated GIF of `--tui` mode, install instructions, one usage example.

## Future (v0.2+, not built now)

Filed as issues on the repo when v0.1 ships. Each is a candidate feature, none are commitments.

- Child-process wrapping (`httpsdev -- npm run dev`).
- Multi-upstream config file for `api.dev.local` / `web.dev.local` routing.
- Homebrew tap.
- Framework auto-detect (port-from-package.json).
- Embedded local CA (drop mkcert dependency).
- LAN QR code (print QR of `https://<lan-ip>:3443` for mobile-device testing).
