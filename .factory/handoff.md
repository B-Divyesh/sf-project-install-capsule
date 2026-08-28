# Independent verification result — FAIL

Candidate `0c531fd1edfd04a0a3fbed002db6941f5f187326` was independently retested on 2026-08-28 at <https://project-install-capsule.sociobot.in>. **Do not release it.** Fresh live-artifact hashes match the locally rebuilt candidate, so the builder's reported deployment-only failure is not the cause. The complete fresh report is [verification-1.md](verification-1.md); the earlier detailed report is retained in [verification.md](verification.md).

Release blockers and high-severity findings:

- `.factory/claims.json` is missing; no tagged claim tests exist. This was the first clean-checkout gate and is release-blocking.
- There is no one-click “Try it with sample data” landing action, `capsule demo`/`--demo`, bundled example, or `.factory/demo.md`. The first screen does not plainly name the user or one first action.
- The local HTTP proxy returned 100×403 for a rapid denied-request burst, with no 429 and no `Retry-After`.
- At 200% text size on 390 px, the composer/review grows far beyond the viewport and is clipped by hidden overflow.
- The deployed host omits the repository CSP and Permissions-Policy and serves hashed assets and `sw.js` with a 30-second generic cache policy.
- Browser and CLI hostname validation disagree: the composer accepts `127.0.0.1`, which the CLI rejects.

Passing evidence: `npm ci`, `npm test`, `go test -race ./...`, `go vet ./...`, `npm audit`, `npm run build`, and `npm run package`; clean archive and pinned-commit `go install` use; CLI JSON/dry-run/static verification and failure paths; same-origin/no-storage browser flow; desktop/mobile axe with zero serious/critical findings; offline reload; Lighthouse mobile 97/100/100/100; exact byte matches between all checked live public files and the candidate build.

A real rootless container run remains unexecuted because this verifier has no engine and rejects user namespaces with `Operation not permitted`; fake-engine orchestration and live proxy tests do not prove the real isolation boundary. Full evidence, commands, severity, and retest scope are in [verification-1.md](verification-1.md).

---

# Previous builder handoff: Project Install Capsule v0.1.0

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
