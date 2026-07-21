# httpsdev

**Zero-config HTTPS for any local dev server.**

```
$ mkcert -install       # one time, system-wide
$ httpsdev 5173         # https://localhost:3443  →  http://localhost:5173
```

Real browser-trusted cert. Works with Vite, Next, Rails, Django, Flask — any HTTP dev server. WebSocket + SSE pass through untouched, so HMR keeps working.

## Why

Every framework has its own HTTPS story (`--https`, `HTTPS=true`, experimental configs) and most default to self-signed certs that trip browser warnings, break service workers, and blow up OAuth callbacks. `httpsdev` sits in front of your existing dev server and gives it real HTTPS in one command.

## Install

Requires [`mkcert`](https://github.com/FiloSottile/mkcert) for the local CA:

```bash
brew install mkcert
mkcert -install
```

Then:

```bash
# Homebrew (coming in v0.2)

# Go install
go install github.com/ArpitRajputGithub/httpsdev@latest

# Or grab a binary from Releases:
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

Full-screen Catppuccin Mocha TUI with live request feed, latency stats, and cert info tab.

*(GIF placeholder — record with asciinema after v0.1 tag)*

### Test on your phone

```bash
httpsdev --lan 5173
```

Binds on `0.0.0.0`. Your phone (on the same wifi) can hit `https://<mac-ip>:3443` — install the mkcert root CA on the phone via `mkcert -CAROOT` to make the padlock green.

## What's out of scope for v0.1

- Wrapping the dev-server process (`httpsdev npm run dev`) — you run your dev server yourself.
- Multi-service reverse proxy (`api.dev.local`, `web.dev.local`) — use [Caddy](https://caddyserver.com) for that.
- Framework detection.

All tracked as issues for v0.2+.

## License

MIT.
