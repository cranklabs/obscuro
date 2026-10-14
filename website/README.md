# Obscuro website

Next.js and Fumadocs documentation site, statically exported to GitHub Pages.

## Development

```bash
pnpm install --frozen-lockfile
pnpm dev
```

Open <http://localhost:3000>. Documentation lives in `content/docs/` and the
application lives in `src/app/`.

## Build and preview

```bash
pnpm lint
pnpm format:check
pnpm preview
```

`pnpm build` exports the site to `out/`. `pnpm start` serves an existing export.
Search indexes are generated at build time and queried in the browser; no
application server or deployment secrets are needed.

## GitHub Pages deployment

In the repository's **Settings → Pages**, select **GitHub Actions** as the build
and deployment source. The `.github/workflows/website.yml` workflow checks and
builds website changes, then deploys `out/` on `main`. It can also be run manually
from the Actions tab.

The site is hosted at <https://cranklabs.github.io/obscuro/>. The workflow gets the
base path and site URL from `actions/configure-pages`, so asset paths and search
requests work under the repository subpath. Next.js links automatically use the
same base path.

To reproduce the GitHub Pages build locally:

```bash
NEXT_PUBLIC_BASE_PATH=/obscuro SITE_URL=https://cranklabs.github.io/obscuro/ pnpm build
```

The base path is fixed at build time. Use `pnpm preview` without these environment
variables to preview at the local server root.
