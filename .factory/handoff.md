# Project Install Capsule v0.1.0 handoff

## What shipped

- A dependency-free Go CLI with `init`, `inspect`, `run`, `teardown`, `verify`, `--dry-run`, and `--json` paths.
- A rootless-engine gate for Podman and Docker. Workloads run with `--network=none`, a read-only root, dropped capabilities, `no-new-privileges`, a 256 PID limit, and a fresh 1 GB tmpfs workspace.
- A hostname-allowlisted HTTP/HTTPS proxy bridged over an ephemeral Unix socket. It rejects IP literals plus loopback, private, link-local, multicast, and unspecified resolutions. Direct workload egress remains unavailable.
- Approved ports published only on host loopback through per-port Unix bridges. No project, home, credential, SSH-agent, or container-engine socket is mounted.
- JSON teardown receipts in `.capsule/receipts/`, including effective capabilities, engine, timestamps, outcome, and exit code.
- `capsule verify`, which creates a short-lived home-directory sentinel and actively checks inside the capsule that the sentinel is unreadable, direct egress fails, and an undeclared host receives a proxy denial. `--static` checks generated engine arguments where containers cannot start.
- A responsive, original art-deco transit-poster documentation site with a local capability composer, offline cache, privacy and terms pages, self-hosted League Spartan, security headers, and no telemetry.
- Original generated hero at `site/public/capsule-poster.webp` (78,166 bytes). Prompt and provenance are recorded in `.factory/design.md`.

## Build and verify

From a clean checkout with Go 1.22+, Node 22+, and the pinned Playwright browser available:

```sh
npm ci
npm test
npm run build
npm run package
```

- `npm test`: passed — Go unit/security tests plus 10 document, responsive-browser, interaction, console, and axe tests.
- `go test -race ./...`: passed.
- `go vet ./...`: passed.
- `npm audit --audit-level=high`: 0 vulnerabilities.
- `npm run build`: passed; static Linux CLI at `dist/bin/capsule`, deploy root at `dist/site/index.html`.
- `npm run package`: passed; static Linux amd64/arm64 archives and `SHA256SUMS` at `dist/packages/`.
- `ldd dist/bin/capsule`: “not a dynamic executable.”
- Browser smoke: Chromium at 1440 px and 390 px, no horizontal overflow or console errors; keyboard-native form/details and visible focus styles verified in automated tests.
- Offline smoke: `/terms/` loaded successfully after the service worker was installed and the browser was placed offline.

## Lighthouse-class results

Lighthouse 13.4.1, mobile defaults, production preview on 2026-08-28:

| Category / metric | Result |
| --- | ---: |
| Performance | 100 |
| Accessibility | 100 |
| Best practices | 100 |
| SEO | 100 |
| LCP | 1.7 s |
| Total blocking time | 0 ms |
| CLS | 0 |

Asset budgets: initial JS 2.98 KB, CSS 12.40 KB, self-hosted fonts 52.63 KB total, hero WebP 78.17 KB. All are below the product budgets.

## Known gaps and deliberate limits

- The disposable factory container disables unprivileged user namespaces (`cannot clone: Operation not permitted`), so a real rootless Podman container could not start here. The live verifier is shipped for a normal Linux host; this environment was covered with engine-selection tests, proxy denial tests, generated-runtime invariant tests, and `capsule verify --static`.
- v0.1 supports Linux hosts only because the same static Linux executable is mounted into the workload as the bridge. Release packaging intentionally emits Linux amd64 and arm64 artifacts only.
- Outbound access supports HTTP and HTTPS clients that honor standard proxy variables. SSH, `git://`, UDP, and arbitrary protocols remain blocked by design.
- Containers are not complete security boundaries. The README and site direct users handling deliberately hostile code or valuable credentials to a disposable VM.

## Deployment and release

- Deploy exactly `dist/site/` to `https://project-install-capsule.sociobot.in`; `_headers`, `robots.txt`, sitemap, and the generated offline service worker are included.
- Publish the archives from `dist/packages/` after factory signing/release automation. The worker did not publish or touch DNS/infrastructure.
