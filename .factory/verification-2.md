# Run new code with less access — verification 2

Verified on 2026-09-06 UTC against:

- Implementation candidate: `48792f344c6598aaa1ed6a9ce297613530376d64`
- Documentation baseline: `1e0fe68b6b4350192c77da6601e66bdb6480c6f8`
- Live URL: <https://project-install-capsule.sociobot.in>

## Verdict

**FAIL — 5 findings, including 1 untested public claim.**

The repaired site is live and most earlier defects are fixed. All eight declared
claim commands pass. Acceptance still fails because the real rootless isolation
promise was not exercised, the CLI demo does not run the product's main job,
and public claims remain outside the claim registry. Two minor site defects also
remain.

No product code was changed during this verification.

## First screen before scrolling

Fresh Chromium contexts were used at 1440 × 1000 and 390 × 844.

- Job: **Run new code with less access**.
- Audience: developers testing an unfamiliar GitHub project without exposing
  their home folder, credentials, or unrestricted network.
- First action: **Try it with sample data**.
- The action and its outcome text were visible before scrolling on both sizes.
- Both views had one `h1`, one `main`, `lang="en"`, no horizontal overflow, and
  no console or page errors.

## Live demo

The first action opened `/?demo=1` and changed the title to
`Demo — Project Install Capsule`. The page immediately showed the Alpine and
GitHub Hello World values, three allowed hosts, port 3000, and the generated
`capsule init` command.

The persistent label read **Demo — sample data, nothing is saved**. An edited
sample value survived reload under the only key `demo:capsule-composer`.
**Reset demo** restored the shipped values and focused the install field.
**Start for real** removed the demo key, hid the banner, and restored the normal
preview. No third-party request or normal storage key was observed.

The browser sample therefore passes its isolation and reset contract. The CLI
sample has a separate end-to-end defect in V2-02.

## Declared claims

Every command was run separately from a fresh remote clone at documentation
commit `1e0fe68`. Each claim has exactly one matching `@claim:<id>` test.

| Claim | Command | Result |
| --- | --- | --- |
| `cli-demo` | `npm run test:claim -- @claim:cli-demo` | PASS |
| `runtime-boundary` | `npm run test:claim -- @claim:runtime-boundary` | PASS |
| `teardown-receipt` | `npm run test:claim -- @claim:teardown-receipt` | PASS |
| `proxy-rate-limit` | `go test ./internal/capsule -run '^TestClaimProxyRateLimit$'` | PASS |
| `demo-sandbox` | `npm run test:claim -- @claim:demo-sandbox` | PASS |
| `preview-private` | `npm run test:claim -- @claim:preview-private` | PASS |
| `preview-host-rule` | `npm run test:claim -- @claim:preview-host-rule` | PASS |
| `site-privacy` | `npm run test:claim -- @claim:site-privacy` | PASS |

Declared claims: **8 passed, 0 failed, 0 untested**. The separate public-claim
audit found one untested core claim and other unregistered claims; see V2-01
and V2-03.

## Clean checkout and installed artifact

The clean clone had no changes before or after the claim runs.

| Check | Result |
| --- | --- |
| `npm ci` | PASS; 21 packages, 0 vulnerabilities |
| `npm test` | PASS; 21 tests |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `npm audit --audit-level=high` | PASS; 0 vulnerabilities |
| `npm run build` | PASS; created `dist/site/` and `dist/bin/capsule` |
| `npm run package` | PASS; amd64 and arm64 archives created |
| `sha256sum -c dist/packages/SHA256SUMS` | PASS for both archives |
| Archive contents | PASS; each has `capsule`, `LICENSE`, and `README.md` |
| `go install .../cmd/capsule@latest` | PASS; resolved documentation commit `1e0fe68` |

The extracted amd64 binary passed `--help`, `version`, `demo --json`,
`inspect --json`, `run --dry-run --json`, and `verify --static --json` from a
new consumer directory. The dry run contains `--network=none`, read-only root,
all capabilities dropped, `no-new-privileges`, PID limit 256, an empty tmpfs
workspace, and loopback port review.

Ports 1 and 65535 and `EXAMPLE.COM.` passed and normalized. Empty install,
IP hosts, ports 0 and 65536, an image containing whitespace, a missing config,
an unknown command, and overwrite without `--force` returned code 2 with useful
errors. `--force` recovered. A simulated rootful engine returned code 3. A
simulated workload failure returned code 4. A simulated successful run wrote a
`0600` receipt.

The local proxy smoke returned 200 for `example.com`, 403 for undeclared
`example.org`, and 400 for `127.0.0.1`. A 100-request burst returned 21×403 and
79×429; all 79 throttled responses had `Retry-After`.

## Live site, accessibility, privacy, and performance

- The live footer reports build `1e0fe68`. A fresh local production build at
  that commit matches the live HTML, JavaScript, CSS, fonts, images, service
  worker, legal pages, 404 page, robots file, and sitemap byte for byte.
  Relative to implementation `48792f3`, the only repository change is the
  prior handoff report.
- `/`, `/?demo=1`, `/demo/`, `/privacy/`, and `/terms/` return working pages.
  `/demo/` reaches the populated demo. An unknown URL returns the designed 404
  document with HTTP 404; that deliberate status is not a defect.
- The live origin sends CSP, Permissions-Policy, `no-referrer`, and `nosniff`.
  HTML and `sw.js` use `no-cache`; tested static assets use one-year immutable
  caching.
- Live Axe scans on home, demo, privacy, terms, and 404 found no serious or
  critical issue. The only console message was Chromium's expected resource
  message for the deliberate 404 navigation.
- Keyboard order starts with the skip link, then moves through all controls.
  Activating the skip link makes the next Tab enter main content. Focus rings
  are visible; the light-section ring measured 11.52:1. No visible control was
  below 44 × 44 CSS pixels.
- At 390 px with 200% text, the form and review remained inside the 390 px
  viewport. Reduced motion changed the entrance to `0.01 ms`, no delay, no
  transform, and no smooth scrolling.
- Privacy-page traffic was same-origin only. There were no cookies or normal
  local/session storage keys.
- The service worker controlled a fresh session, updated successfully, kept
  only `capsule-shell-v2`, and reloaded the populated demo offline with the
  offline status text.
- `/opt/fleet/lib/verify-url.sh` passed: title, English language, one `h1`, a
  main landmark, image alt text, labeled buttons, and no console errors.
- Lighthouse 12.8.2 on the live mobile page completed with performance 100,
  accessibility 100, best practices 100, and SEO 100. LCP was 1.43 s, total
  blocking time 48 ms, CLS 0, and transfer size 113 KB.
- Production sizes pass the budgets: JavaScript 4.38 KB (1.99 KB gzip), CSS
  13.47 KB (4.03 KB gzip), fonts 52.63 KB, hero image 78.17 KB, and social image
  47.97 KB.

## Findings

### V2-01 — High — Actual rootless isolation remains untested

The page and README say the running project cannot read the host home and has no
direct or undeclared network access. The brief requires an active seeded-home
and undeclared-host check. The declared `runtime-boundary` claim checks only the
generated dry-run arguments, and the receipt claim uses a fake engine.

Docker 29.1.3 was installed as documented, but this worker cannot create a user
namespace: `unshare -Ur true` returns `Operation not permitted`. No rootless
daemon can run under that host restriction. The installed artifact's real
`run` and `verify` commands therefore both returned code 3 with `no rootless
container engine found`. Static arguments and fake-engine orchestration are not
evidence that the live isolation boundary holds.

Disposition: unresolved verification blocker. Run `capsule verify --json` and
the sample workload on a Linux host that supports rootless Podman 5+ or Docker
27+, and record the four live checks. Untested public claim count: **1**.

### V2-02 — High — The CLI demo does not run the main job

The CLI demo contract requires `capsule demo` to run the main job on bundled
sample data in a temporary directory. Here it only creates a config and a local
README, prints a review, and tells the user to run a second command. The bundled
README is explicitly never mounted. The later run downloads a different sample
from GitHub and needs both a rootless engine and external network access.

The browser preview is useful and passes, but the installed CLI still lacks a
one-command, bundled-data execution of its real job.

### V2-03 — High — Public claims are missing from the claim registry

`.factory/claims.json` covers eight claims, but public copy also claims:

- the title says “Run projects safely” without defining or testing “safely”;
- the CLI has no telemetry or remote service and keeps configuration and
  receipts local;
- the runtime proxy rejects undeclared and unsafe resolved destinations;
- the real workspace disappears and direct container networking is disabled.

Some lower-level unit and manual checks support parts of these statements, but
the claims contract requires each public claim to have its own registry entry
and one tagged sandbox test. The registry is therefore incomplete.

### V2-04 — Low — The security-model link has no target

“Read the security model on GitHub” links to repository fragment
`#security-model`. The README has `#review-the-boundary` and `#security-limits`,
but no `#security-model`. The HTTP request returns 200, yet the promised section
is not selected.

### V2-05 — Low — The earlier responsive-image finding remains

The 1400 × 933 hero uses one 78 KB WebP with no `srcset` or `sizes`. It is below
the byte budget and the live performance score is 100, but it still fails the
explicit responsive-image rule and does not resolve the earlier minor finding.

## Earlier findings disposition

| Earlier finding | Verification 2 disposition |
| --- | --- |
| Claims registry and tagged tests absent | Partly fixed: 8/8 declared commands pass, but public claims remain unregistered (V2-03). |
| No web or CLI sample | Web sample fixed; CLI command exists but does not run the main job (V2-02). |
| First screen unclear | Fixed on fresh desktop and phone. |
| Proxy had no 429/`Retry-After` | Fixed; live smoke observed both. |
| 200% phone clipping | Fixed. |
| Browser/CLI hostname mismatch | Fixed. |
| Missing live response policies | Fixed on the public origin. |
| Missing 404, metadata, footer build, archive files | Fixed. |
| Focus contrast and small touch targets | Fixed. |
| Real rootless isolation not proven | Still untested (V2-01). |
| Hero lacks responsive sources | Still present (V2-05). |

## Release decision

Do not declare this candidate accepted. Reverify after the CLI demo runs a
bundled sample end to end, every public claim is registered, the live rootless
probe is recorded on a capable Linux host, the broken section link is fixed,
and the hero has responsive sources.
