# Independent product verification 1 — FAIL

Verified on 2026-08-28 UTC from a clean detached checkout of candidate `0c531fd1edfd04a0a3fbed002db6941f5f187326`, against <https://project-install-capsule.sociobot.in>. **Do not release this candidate.**

## Release decision

This is not a deployment-only failure. Fresh SHA-256 comparisons found the live `index.html`, privacy and terms pages, hashed JavaScript and CSS, all fonts, hero image, favicon, `robots.txt`, `sitemap.xml`, and service worker identical to a fresh production build of the candidate.

The candidate fails the mandatory claims/demo gate before other tests: `.factory/claims.json` is missing in the clean checkout, `rg -n '@claim:'` finds no tagged claim test, and there is no `.factory/demo.md`, `examples/` sample, `capsule demo`, or `--demo`. The landing page has no visible **“Try it with sample data”** action or demo banner. A cold 390 px load shows the headline “Inspect the route. Then run the code.” and peer actions “Install Capsule” and “Preview a review”; it does not plainly name its developer audience or a single first action. The direct `/demo` URL is the landing page (HTTP 200), not a demo. Per the acceptance contract, either omission is independently release-blocking.

## Fresh mandatory-gate evidence

The first action after creating the clean detached checkout was:

```sh
test -f .factory/claims.json
rg -n '@claim:' .
rg --files -g '.factory/demo.md' -g 'examples/**'
```

Results: the claims file was absent, there were zero claim tags, and neither demo documentation nor examples existed. Consequently there were no claim-test commands to execute; the claims contract explicitly makes the missing registry a failure.

Cold-browser evidence from the live `/` at 390 × 844: HTTP 200, no console/page errors, no sample action (`0` matches), no demo banner (`0` matches), and only same-origin page asset requests. The live browser observations are saved in `.factory/qa-evidence/live-browser-qa.json`; first-read captures are in `.factory/qa-evidence/first-read-desktop.json` and `first-read-mobile.json`.

## Local clean-checkout results

The clean clone was detached at the exact candidate. Node was `v22.23.2`; an official checksum-verified Go `1.27.0` toolchain was used because the base image has no Go executable.

| Check | Result |
| --- | --- |
| `npm ci` | PASS; 21 packages installed, audit reported 0 vulnerabilities |
| `npm test` | PASS; Go unit tests plus 10 Node/browser tests |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| Type/lint scripts | None defined by the repository |
| `npm run build` | PASS; generated `dist/bin/capsule` and `dist/site/` |
| `npm run package` | PASS; generated release archives/checksums |
| `npm audit --audit-level=high` | PASS; 0 vulnerabilities |

The production build is small: JavaScript 2,980 B (1.44 KB gzip), CSS 12,404 B (3.83 KB gzip), and self-hosted fonts 52,628 B total. No third-party browser resources, storage, cookies, or analytics were observed through a complete composer interaction.

## CLI and core workflow

From a fresh consumer directory, the built binary successfully completed `init --json`, `inspect --json`, `run --dry-run --json`, and `verify --static --json`. The reviewed invocation includes `--network=none`, read-only root, dropped capabilities, `no-new-privileges`, PID limit, an empty `/workspace` tmpfs, and loopback-only port forwarding. A clean `go install github.com/B-Divyesh/sf-project-install-capsule/cmd/capsule@0c531fd1edfd04a0a3fbed002db6941f5f187326` produced a working `0.1.0` CLI.

Boundary and recovery checks pass for ports 1 and 65535; ports 0 and 65536 exit 2 with actionable messages. IP allowlists fail, existing config overwrite is refused, and `--force` recovers normally. With no Podman/Docker engine in this verifier, real `run` and live `verify` fail closed with exit 3. A true rootless-container run is therefore not proven here; this is material because the brief's success measure requires showing that a seeded home secret and undeclared host are inaccessible.

The release archive contains only `capsule-0.1.0-linux-amd64`: it does not include the MIT `LICENSE`, and it does not contain a binary literally named `capsule` as the README's archive-install wording implies.

## Live browser, accessibility, and policy checks

Desktop and 390 px checks of `/`, `/privacy/`, and `/terms/` passed basic semantics (English lang, title, one h1, header/nav/main/footer), had no page/console errors, and had zero axe serious/critical violations. Keyboard navigation reaches the skip link and controls with a 3 px visible focus outline. The service worker controls the site, updates, and permits offline reload of the landing and terms pages. Reduced motion reduces the entrance animation to `0.00001s`.

Remaining accessibility defects:

- At simulated 200% text size on a 390 px viewport, the core review is 1,061 px wide while `main` has `overflow-x: hidden`; essential content is clipped with no recovery path. Existing screenshot: `.factory/qa-evidence/mobile-text-200.png`.
- The mobile wordmark is 38 px tall, the mobile Home link is 41 px wide, and inline legal links are 19 px tall, below the supplied 44 px target requirement.
- The amber focus outline has only 1.63:1 contrast against the paper section (`.factory/qa-evidence/focus-light-section.png`), below the required 3:1 focus-indicator contrast.

The browser composer accepts `127.0.0.1` and enables “Copy command”; the CLI rejects that value with exit 2. Thus the site statement that it validates the same hostname shapes is false. The CLI accepts a trailing-dot hostname that the composer rejects.

The checked live response headers have HSTS and `X-Content-Type-Options: nosniff`, but omit the candidate's CSP and Permissions-Policy. The source asks for `Referrer-Policy: no-referrer`; live responses send `strict-origin-when-cross-origin`. All documents, hashed assets, and `sw.js` instead receive `Cache-Control: public, must-revalidate, max-age=30`, not the source's immutable-asset/no-cache service-worker policy. Unknown paths return the landing page with HTTP 200, so no real 404 is deployed. Required canonical/social/Apple-touch metadata, `staticwebapp.config.json`, and `.factory/copy-audit.md` are absent.

## Server endpoint rate limiting

The product has no remote API or sign-in endpoint. Its local Unix-socket HTTP allowlist proxy is a server-side endpoint and contains no rate-limit logic. A fresh 100-request concurrent burst to an undeclared destination yielded:

```text
denied_burst=map[403:100] retry_after_responses=0
```

No 429 was observed at 100 rapid requests and no response had `Retry-After`; the required threshold was therefore not observed. The smoke test also confirms ordinary allowed/blocked behavior: `example.com` returns 200, undeclared `example.org` 403, and IP literal 400.

## Defects by severity

### Critical / release-blocking

1. `.factory/claims.json` and every required tagged claim test are absent despite material product/security claims.
2. There is no one-click, shipped sample-data demo in the landing page or CLI. The first screen fails the plain-language user/first-action acceptance test.

### High

1. The proxy endpoint has no rate limiting: 100 rapid requests produce no 429 or `Retry-After`.
2. At 200% text size on 390 px, core review content is clipped by hidden overflow.
3. The live deployment does not apply the intended CSP, Permissions-Policy, referrer policy, or cache policies.
4. A real rootless engine run could not be independently exercised here; the mandatory shipped demo needed for repeatable proof is absent.

### Medium

1. Site and CLI hostname validation disagree, making a generated browser command invalid in the CLI.
2. The release archive lacks the MIT license and a directly runnable `capsule` filename.
3. Unknown routes return the home page with 200; the required designed 404, metadata, deployment routing config, copy audit, and footer build identity are missing.
4. Touch-target and focus-contrast requirements are missed.

## Required retest scope

Start with a real claim registry and one observable demo-entry-point test per claim. Add a one-command bundled CLI demo (`capsule demo` or `--demo`) with shipped sample input and documented isolation/reset behavior. Then retest proxy 429/`Retry-After`, 200% mobile text, browser/CLI validation parity, archive contents, live headers/caching, and an actual rootless run proving the home-secret and undeclared-host outcomes.
