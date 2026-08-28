# Independent product verification — FAIL

Verified on 2026-08-28 UTC for candidate `0c531fd1edfd04a0a3fbed002db6941f5f187326` at <https://project-install-capsule.sociobot.in>.

## Verdict

**FAIL. Do not release this candidate.** The live deployment is byte-for-byte consistent with the candidate build, so this is not a stale-deployment result. The candidate fails mandatory acceptance gates: `.factory/claims.json` is absent, the landing page and CLI have no one-click sample-data demo, and the first screen does not plainly identify the user or a single first action. A local server endpoint also has no required rate limit. High-severity accessibility and deployed-policy defects remain.

## Mandatory claims gate

This check was performed first, before dependency installation or any other product test.

- `.factory/claims.json`: **missing** at the tested commit. There were therefore no claim commands to run. The contract defines the missing file as release-blocking.
- `rg -n '@claim'`: no tagged claim tests exist.
- Claim-like copy is nevertheless present and unregistered, including “No telemetry,” “This browser preview makes no requests and stores nothing,” “Direct container networking is disabled,” and the home-secret/network isolation promises in the README and landing page.
- The available unit/browser tests cannot substitute for the required registry and one-test-per-claim mapping.

## Cold first-read and demo gate

Cold desktop and 390×844 browser contexts loaded the live `/` with no existing state.

- What it does: the body copy says it gives an unfamiliar project an empty workspace and named network destinations, then prints a capability diff before a rootless container starts.
- Who it is for: **not stated in plain words on the first screen**. A developer audience can only be inferred from “project,” “container,” and the product context.
- What to click first: **ambiguous**. “Install Capsule” and “Preview a review” are presented as peer actions with no adjacent outcome explanation.
- Required sample action: **absent**. There is no “Try it with sample data” action.
- `GET /demo` returns the same 9,404-byte landing document with HTTP 200, not a demo.
- `capsule demo` exits 2 with `unknown command "demo"`; no `--demo`, `examples/`, `.factory/demo.md`, persistent demo banner, reset action, or start-for-real action exists.
- The landing “terminal” is static documentation, not a recording of the real CLI on bundled sample input.

Evidence: `.factory/qa-evidence/first-read-desktop.json`, `first-read-mobile.json`, and their screenshots.

## Clean checkout, tests, build, and packaging

The clone began clean on `main` at the exact candidate. `npm ci` installed 21 packages and reported zero vulnerabilities. The supplied image did not include Go, so a checksum-verified Go 1.27.0 toolchain was installed under `/tmp`; the product requires Go 1.22 or newer.

| Check | Result |
| --- | --- |
| `npm ci` | PASS |
| `npm test` | PASS: 8 Go tests and 10 Node/browser tests |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `npm audit --audit-level=high` | PASS: 0 vulnerabilities |
| Type/lint scripts | Not present in the repository |
| `npm run build` | PASS; produced `dist/bin/capsule` and `dist/site/` |
| `npm run package` | PASS; amd64/arm64 archives and checksums produced |

Production site budgets pass: JS 2,980 B (1.44 KB gzip), CSS 12,404 B (3.83 KB gzip), fonts 52,628 B total, and hero WebP 78,166 B. Lighthouse 12.8.2 mobile measured performance 97, accessibility 100, best practices 100, SEO 100, LCP 1.50 s, TBT 195.5 ms, and CLS 0. Lighthouse 13 could not use the supplied Chromium because of its stable-version gate.

## CLI and core workflow

Both release checksums verified. A clean extracted amd64 archive and a clean `go install github.com/B-Divyesh/sf-project-install-capsule/cmd/capsule@latest` produced version 0.1.0 from pseudo-version `0c531fd`.

- `--help` and `version`: PASS; commands, options, risk boundary, and exit codes are clear.
- Normal `init --json` → `inspect --json` → `run --dry-run --json` → `verify --static --json`: PASS.
- Generated runtime arguments contain `--network=none`, `--read-only`, `--cap-drop=ALL`, `no-new-privileges`, PID limit 256, an empty tmpfs workspace, loopback port forwarding, and only the executable/ephemeral bridge mounts.
- Duplicate hosts/ports normalize; ports 1 and 65535 pass. Ports 0 and 65536, IP allowlists, malformed/unknown JSON, blank install commands, whitespace-bearing image references, missing configs, and accidental overwrite all fail with exit 2 and actionable errors. A valid edit recovers normally.
- No engine and a simulated rootful engine fail closed with exit 3.
- A controlled fake-engine run confirmed machine-readable success, workload failure propagation (engine exit 7 produces CLI exit 4), automatic cleanup call, and 0600 JSON receipt creation. This tests orchestration, not container isolation.
- Live proxy smoke: allowed `example.com` returned 200; undeclared `example.org` returned 403; IP literal `127.0.0.1` returned 400.
- **Rate-limit failure:** a burst of 100 concurrent requests to the local HTTP proxy produced 100×403, zero 429 responses, and zero `Retry-After` headers. No threshold was observed. The deployed site has no remote API, and no sign-in is present.

A genuine rootless container run and `capsule verify` could not execute in this verifier because no engine is installed and user-namespace creation fails with `Operation not permitted`. This environment limitation is not itself logged as a product defect, but static/fake-engine checks do not independently prove the brief's live home-secret and direct-egress success measure. The mandatory demo would have supplied a repeatable acceptance path and is absent.

Packaging defect: each release archive contains only a platform-suffixed executable. It omits `LICENSE` even though the MIT text requires the notice in distributed copies, and the README instruction to put a binary named `capsule` on `PATH` omits the required rename.

## Live deployment and browser behavior

Every public candidate artifact checked—HTML routes, JS, CSS, fonts, image, favicon, service worker, robots, and sitemap—matched the locally rebuilt file byte for byte. The live deployment therefore matches candidate `0c531fd` by reproducible artifact comparison. The page itself exposes no build identifier.

- `/`, `/privacy/`, and `/terms/`: HTTP 200; correct route titles, `lang="en"`, one `h1`, header/nav/main/footer landmarks, no horizontal overflow at normal zoom, and no console or page errors.
- Axe on all three routes at 1440 px and 390 px: zero serious/critical findings.
- Keyboard smoke: skip link and controls are reachable; sampled controls show a 3 px focus outline and no trap.
- Reduced motion: the 420 ms entrance becomes 0.01 ms with no transform.
- Privacy: a complete composer interaction made same-origin requests only; cookies, localStorage, and sessionStorage stayed empty. No analytics or third-party scripts loaded.
- Service worker: active and controlling; `registration.update()` completed; cache `capsule-shell-v1` held all shell assets; offline reload of `/` and `/terms/` passed.
- Normal form failures and recovery work for empty install commands, malformed hosts, and invalid ports.
- All discovered links returned 200 after redirects.

Accessibility failures not caught by the aggregate axe score:

- At simulated 200% text size on a 390 px viewport, the composer expands to 744 px and the review ticket to 1,081 px while `<main>` hides overflow. Fields, headings, and most review content are clipped with no horizontal recovery. Evidence: `.factory/qa-evidence/mobile-text-200.png`.
- The universal amber focus outline has only 1.63:1 contrast against the paper-colored limits section; the requirement is at least 3:1. Evidence: `.factory/qa-evidence/focus-light-section.png`.
- Several mobile targets are below 44 px in one dimension, including the 38 px-tall wordmark and the 41 px-wide “Home” link. Inline legal links are 19 px tall.
- Lighthouse separately reports a visible-label/accessibility-name mismatch on the home wordmark because its accessible name omits the visible decorative “C”.

The live capability composer contradicts its claim that the CLI “validates the same hostname and port shapes.” It accepts `127.0.0.1` and enables “Copy command,” while the CLI rejects that IP with exit 2. Conversely, the CLI normalizes and accepts `EXAMPLE.COM.`, while the composer rejects the trailing dot.

## Response policies, routing, and metadata

The repository ships an `_headers` file, but the live host does not apply it:

| Policy | Repository intent | Live response |
| --- | --- | --- |
| CSP | restrictive self-only policy with `frame-ancestors 'none'` | missing |
| Permissions-Policy | camera, microphone, geolocation disabled | missing |
| Referrer-Policy | `no-referrer` | `strict-origin-when-cross-origin` |
| Hashed assets | one year, immutable | `public, must-revalidate, max-age=30` |
| `/sw.js` | `no-cache` | `public, must-revalidate, max-age=30` |

HSTS and `X-Content-Type-Options: nosniff` are present. Unknown paths return the landing page with HTTP 200, so there is no real designed 404. Required canonical, Open Graph, Twitter card, 1200×630 social image, apple-touch icon, and `staticwebapp.config.json` are absent. The footer has no version/build ID and does not use the required “Built by Param Factory” attribution.

## Defects by severity

### Critical / release-blocking

1. `.factory/claims.json` and all tagged claim tests are missing despite numerous user-facing claims.
2. The mandatory one-click sample-data demo is absent from both landing page and CLI; the first screen also fails the explicit who/first-action test.

### High

1. The local HTTP proxy has no required rate limit: 100 rapid requests yielded no 429 and no `Retry-After`.
2. Text resized to 200% is clipped across the core composer/review UI on a 390 px viewport.
3. Live deployment omits the candidate's CSP and Permissions-Policy and does not honor required cache policies.
4. The principal isolation success measure could not be exercised end to end, and no bundled demo path exists to make it independently reproducible.

### Medium

1. Browser composer and CLI hostname validation disagree; the composer generates an IP allowlist command the CLI rejects.
2. Unknown routes return the home page with HTTP 200; no designed 404 exists.
3. Release archives omit the MIT license and do not contain a binary named as the install instructions state.
4. Required canonical/social/apple-touch metadata, deployment routing config, footer build ID, `.factory/demo.md`, and `.factory/copy-audit.md` are missing.
5. Focus contrast and several mobile touch targets miss the supplied accessibility thresholds.

### Low

1. The 1400 px hero lacks responsive sources; Lighthouse estimates 61–71 KiB avoidable mobile transfer, though the explicit 300 KB image and performance budgets pass.

## Release decision

Keep the candidate blocked. Reverification must begin with a real `.factory/claims.json` and its demo-based tests, then validate the new one-click CLI demo, proxy rate limiting, 200% text behavior, live headers/caching, archive contents, and live rootless isolation.
