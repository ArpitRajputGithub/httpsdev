<div align="center">

# httpsdev

**Localhost with a working padlock.**

Point it at a port. Get a green padlock.

[![Release](https://img.shields.io/github/v/release/ArpitRajputGithub/httpsdev?style=flat-square&color=cba6f7&labelColor=1e1e2e)](https://github.com/ArpitRajputGithub/httpsdev/releases)
[![Go Reference](https://img.shields.io/badge/go-reference-cba6f7?style=flat-square&labelColor=1e1e2e)](https://pkg.go.dev/github.com/ArpitRajputGithub/httpsdev)
[![License](https://img.shields.io/badge/license-MIT-cba6f7?style=flat-square&labelColor=1e1e2e)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/ArpitRajputGithub/httpsdev/ci.yml?style=flat-square&labelColor=1e1e2e&color=a6e3a1)](https://github.com/ArpitRajputGithub/httpsdev/actions)

<!-- Swap this for a real GIF after `vhs demo.tape` runs. See demo.tape. -->
<!-- <img src=".github/demo.gif" alt="httpsdev demo" width="720" /> -->

</div>

---

```
$ mkcert -install       # one time, system-wide
$ httpsdev 5173         # https://localhost:3443  →  http://localhost:5173
```

Real browser-trusted cert. Works with Vite, Next, Rails, Django, Flask — any HTTP dev server. WebSocket + SSE pass through untouched, so HMR keeps working.

## Why

Every framework has its own broken HTTPS story. Some ship self-signed certs that trip the browser. Some hide it behind an `HTTPS=true` env var. Some just don't. `httpsdev` sits in front of any dev server and hands it a real, browser-trusted cert — the kind that comes from mkcert's local CA, so Chrome and Safari already trust it.

Specifically fixes:
- Service workers refusing to register over `http://`
- OAuth callbacks rejecting non-HTTPS redirect URIs (Google, GitHub, Auth0)
- `Secure` / `SameSite=None` / `__Host-` cookies not surviving the round-trip
- WebAuthn, Web Crypto, clipboard, camera/mic APIs gated behind secure context

## Install

Requires [`mkcert`](https://github.com/FiloSottile/mkcert) for the local CA:

```bash
brew install mkcert
mkcert -install
```

Then:

```bash
# Go install (Go users)
go install github.com/ArpitRajputGithub/httpsdev@latest

# Homebrew (coming in v0.2)
# brew install ArpitRajputGithub/tap/httpsdev

# Or grab a prebuilt binary from Releases:
#   https://github.com/ArpitRajputGithub/httpsdev/releases
```

## Usage

```
httpsdev <target-port> [flags]

  --listen <port>   HTTPS listen port          (default 3443)
  --host <name>     Additional SAN for cert    (default: none)
  --lan             Bind on 0.0.0.0 for LAN    (default: 127.0.0.1)
  --tui             Full-screen dashboard mode (default: false)
  --version         Print version
```

### Log mode (default)

```
$ httpsdev 5173

  ▲ httpsdev v0.1.0

  ➜  Local:    https://localhost:3443
  ➜  Upstream: http://localhost:5173
  ➜  Cert:     mkcert · valid 30 days

  press Ctrl+C to quit

GET   /                     200   12ms
GET   /@vite/client         200    3ms
POST  /api/login            401   45ms
```

### Dashboard mode

```bash
httpsdev --tui 5173
```

Full-screen Catppuccin Mocha TUI with live request feed, latency stats, and connection info. The wordmark alone justifies the download.

### Test on your phone

```bash
httpsdev --lan 5173
```

Binds on `0.0.0.0`. Your phone (on the same wifi) can hit `https://<mac-ip>:3443` — install the mkcert root CA on the phone via `mkcert -CAROOT` to make the padlock green.

## What's out of scope for v0.1

- Wrapping the dev-server process (`httpsdev npm run dev`) — you run your dev server yourself.
- Multi-service reverse proxy (`api.dev.local`, `web.dev.local`) — use [Caddy](https://caddyserver.com) for that.
- Framework detection.

All tracked as [issues](https://github.com/ArpitRajputGithub/httpsdev/issues) for v0.2+.

## Website

<https://httpsdev.pages.dev> — landing page with a live TUI preview.

## License

MIT.
