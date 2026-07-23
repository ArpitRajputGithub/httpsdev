# Deploying the landing page to Cloudflare Pages

The site lives in `web/`. Cloudflare Pages watches the `main` branch and rebuilds automatically on every push. No GitHub Actions workflow required — Cloudflare's own CI does the build.

## One-time setup (~2 minutes)

1. Sign in at [dash.cloudflare.com](https://dash.cloudflare.com/) (free account).
2. **Workers & Pages → Create → Pages → Connect to Git.**
3. Authorize Cloudflare for the `ArpitRajputGithub` account, pick the `httpsdev` repo.
4. **Set up builds and deployments:**

   | Field | Value |
   |-------|-------|
   | Framework preset | Astro |
   | Build command | `npm run build` |
   | Build output directory | `dist` |
   | Root directory | `web` |
   | Node version | `22` (set via env var `NODE_VERSION` → `22`) |

5. **Save and Deploy.** First build takes ~90s.
6. Live at `httpsdev.pages.dev` (or `<some-hash>.httpsdev.pages.dev` on preview branches).

## Custom domain (optional, when ready)

- Buy a domain — [Porkbun](https://porkbun.com/) has `.dev` for ~$10/yr.
- **Pages project → Custom domains → Set up a custom domain.** Cloudflare handles SSL and DNS automatically if the domain is on their nameservers.
- Suggested: `httpsdev.dev`.

## Preview deployments

Every branch and PR gets its own preview URL — great for reviewing landing page changes before merging.

## Local dev

```bash
cd web
npm run dev            # http://localhost:4321
npm run build          # writes dist/
npm run preview        # serves dist/ locally
```

## Why Cloudflare Pages over GitHub Pages

- Unlimited bandwidth (GH Pages soft-caps at 100 GB/month).
- Faster CDN (Cloudflare's edge > GitHub's Fastly).
- Cleaner URL (`httpsdev.pages.dev` vs `arpitrajputgithub.github.io/httpsdev`).
- Automatic preview URLs per branch.

## Emergency fallback: GitHub Pages

If Cloudflare setup blocks you, GH Pages works with a one-liner workflow:

```yaml
# .github/workflows/deploy-web.yml
name: deploy-web
on:
  push:
    branches: [main]
    paths: [web/**]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '22' }
      - run: cd web && npm ci && npm run build
      - uses: actions/upload-pages-artifact@v3
        with: { path: web/dist }
  deploy:
    needs: build
    permissions: { pages: write, id-token: write }
    runs-on: ubuntu-latest
    steps:
      - uses: actions/deploy-pages@v4
```

Then in the repo's Settings → Pages → Source: "GitHub Actions".
