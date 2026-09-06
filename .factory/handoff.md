# Verification 2 handoff — live promoted, acceptance blocked

## Independent verification 2

- Verdict: **FAIL**
- Finding count: **5**
- Untested public claim count: **1**
- Implementation reviewed: `48792f344c6598aaa1ed6a9ce297613530376d64`
- Documentation baseline: `1e0fe68b6b4350192c77da6601e66bdb6480c6f8`
- Full report: `.factory/verification-2.md`

The repaired artifact is now live. Its footer reports build `1e0fe68`, and its
public files match a clean production build. All eight declared claim commands,
the 21-test suite, race detector, vet, audit, build, packaging, browser demo,
live accessibility checks, offline reload, and Lighthouse run pass.

Acceptance remains blocked. The real seeded-home and network isolation probe
could not run because this worker forbids user namespaces, so the core runtime
promise remains untested. The CLI demo prints a review but does not run bundled
sample data. Public claims remain outside `.factory/claims.json`. The GitHub
security-model fragment is missing, and the earlier responsive-image finding
remains.

## Implementation

- Implementation commit: `48792f344c6598aaa1ed6a9ce297613530376d64` (`fix: add demo and release hardening`)
- Documentation/handoff baseline: `1e0fe68b6b4350192c77da6601e66bdb6480c6f8`.
- Product URL: <https://project-install-capsule.sociobot.in>

This repair keeps the product scope: a Linux CLI for developers who want to
inspect an unfamiliar GitHub project before it can use their machine. The first
action is **Try it with sample data**; it opens the browser sample at
`/?demo=1`. The CLI equivalent is `capsule demo`.

## What changed

- Added `.factory/claims.json` with eight observable, clean-setup claim tests.
- Added `capsule demo`, embedded and repository-shipped Hello World sample
  input, `examples/hello-world`, and `.factory/demo.md`. The CLI creates only a
  fresh new folder and does not start an engine until the user explicitly runs
  the printed command.
- Added a browser demo banner, `demo:capsule-composer` storage namespace,
  reset action, and start-for-real cleanup. Normal preview edits make no
  network request and use no browser storage.
- Added shared canonical hostname handling: case and one final DNS root dot are
  normalised, while IP literals remain rejected in both CLI and browser.
- Added a sliding one-second proxy burst limit. Rapid requests now receive HTTP
  429 with `Retry-After: 1`.
- Removed the clipped main overflow path; made grid tracks shrinkable; shortened
  the mobile headline; added 200% text, touch-target, and focus-contrast
  browser regressions.
- Added the static-site deployment configuration with CSP,
  Permissions-Policy, referrer/cache policies, immutable asset routes, and a
  designed 404 rewrite. Added canonical/social/Apple-touch metadata, `404.html`,
  sitemap demo URL, footer build ID, and a derived 1200×630 product social card.
- Fixed archives so each platform folder contains `capsule`, `LICENSE`, and
  `README.md`.

## Verification

Run from a clean checkout with Node 22+, Go 1.22+, and the pinned Playwright
browser:

```sh
npm ci
npm test
npm run build
npm run package
node -e "for (const c of require('./.factory/claims.json')) console.log(c.test)"
```

Completed on this implementation:

| Check | Result |
| --- | --- |
| Every command in `.factory/claims.json` | PASS (8/8) |
| `npm test` | PASS (21 tests) |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `npm audit --audit-level=high` | PASS (0 vulnerabilities) |
| `npm run build` | PASS; `dist/site/` and `dist/bin/capsule` created |
| `npm run package` | PASS; amd64 and arm64 archives include binary, MIT license, and README |
| Clean consumer artifact | PASS; extracted amd64 `capsule demo` and `run --dry-run --json` completed |
| Browser/axe | PASS; desktop/mobile routes have no serious or critical violations |
| 390 px at 200% text | PASS; composer and review remain within the viewport |
| Proxy burst | PASS; claim test observes 429 plus `Retry-After: 1` |

Local production Lighthouse output was 100 performance, 100 accessibility, 100
best practices, and 100 SEO; LCP 1.52 s, TBT 0 ms, CLS 0. The launcher reported
a late screenshot-tab crash after it wrote the audit JSON, so treat those scores
as local diagnostic evidence rather than a clean Lighthouse process exit.

Production assets remain within budget: initial JavaScript 4.38 KB (1.99 KB
gzip), CSS 13.47 KB (4.03 KB gzip), self-hosted fonts 52.63 KB, hero 78.17 KB,
and social image 47.97 KB.

## Earlier findings disposition

| Earlier finding | Current disposition |
| --- | --- |
| Missing claims registry/tests | Partly fixed: eight declared claims pass, but public claims remain outside the registry. |
| No one-click web/CLI sample | Web demo fixed. The CLI command only prepares and reviews a sample; it does not run the main job. |
| First screen unclear | Fixed: job, audience, first sample action, immediate outcome, and three facts appear before scrolling. |
| Proxy burst had only 403 | Fixed and regression-tested with 429/Retry-After. |
| 200% mobile clipping | Fixed and browser-tested. |
| Browser/CLI hostname mismatch | Fixed and browser/CLI normalization tests added. |
| Missing deployment policies | Fixed and confirmed on the public origin. |
| No 404/metadata/footer/archive contents | Fixed in the built and live artifacts. |
| Focus/touch minor findings | Fixed and browser-tested. |
| Real rootless isolation | Still not proven: Docker 29 was installed, but this worker forbids user namespaces and no rootless daemon can run. |
| Hero lacks responsive sources | Still present: the 1400 px image has no `srcset` or `sizes`. |

## Live deployment status

The public origin now serves the repaired site and reports build `1e0fe68`.
HTML and `sw.js` use `Cache-Control: no-cache`; tested static assets use
one-year immutable caching. CSP, Permissions-Policy,
`Referrer-Policy: no-referrer`, and `X-Content-Type-Options: nosniff` are
present. Unknown routes return the designed page with HTTP 404. Live browser,
header, cache, metadata, legal-page, and 404 checks were repeated during
verification 2.

No paid offer is advertised or required by the researched brief, so no billing
metadata was created.
