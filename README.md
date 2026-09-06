# Project Install Capsule

Run an unfamiliar GitHub project in a rootless container with an empty
workspace, named network hosts, loopback-only ports, and a local receipt.

Project Install Capsule is for developers who want to inspect a repository
before it can use their computer. It is a Linux CLI. It is free software under
the MIT License.

> Containers reduce exposure. They are not complete security boundaries. Use a
> disposable VM on a separate account for hostile code or high-value
> credentials.

## Try the bundled sample

The sample needs no account and does not start a container until you explicitly
run the printed command.

```sh
capsule demo
```

`capsule demo` makes a fresh folder below your system temporary directory. It
adds a bundled sample note and a reviewable `capsule.json`, then prints the
capability review and the exact command to run next. It does not read or change
an existing project. Pass a new path when you want to keep the sample:

```sh
capsule demo --dir ./capsule-sample
capsule run --config ./capsule-sample/capsule.json
```

The reviewable sample input is also in [examples/hello-world](examples/hello-world).

## Install

You need a Linux host, Go 1.22 or newer, and rootless Podman 5 or newer
(recommended) or rootless Docker 27 or newer. Build from source:

```sh
CGO_ENABLED=0 go install github.com/B-Divyesh/sf-project-install-capsule/cmd/capsule@latest
capsule demo
```

A release archive contains a `capsule` binary, `LICENSE`, and `README.md` in
its platform folder. Put that `capsule` binary on your `PATH`.

## Create a capsule for your project

Create a directory for the declaration. The install command runs inside a new,
empty `/workspace`; clone or download the project there.

```sh
mkdir try-project && cd try-project
capsule init \
  --install 'git clone https://github.com/owner/project.git .' \
  --run 'npm install && npm run dev -- --host 127.0.0.1' \
  --allow-host github.com \
  --allow-host registry.npmjs.org \
  --port 3000

capsule inspect
capsule run
```

`capsule init` writes `capsule.json`. `capsule inspect` prints the capability
review without starting an engine. `capsule run` requires a rootless engine.

## Review the boundary

Run this before starting a project when you want machine-readable output:

```sh
capsule inspect --json
capsule run --dry-run --json
capsule verify --static --json
```

The sample dry run shows an empty workspace, no host-home mount, no direct
container network, and loopback-only ports. A completed run writes a local
receipt with owner-only permissions. `capsule verify` can make live checks when
a rootless engine is available; use `--static` when an engine is unavailable.

The proxy allows only named HTTP and HTTPS hosts from `allow_hosts`. It rejects
IP addresses and unsafe resolved destinations. A rapid local proxy burst is
throttled with HTTP 429 and `Retry-After`.

## Security limits

- The container gets an empty tmpfs at `/workspace`, a read-only image root,
  dropped Linux capabilities, `no-new-privileges`, a PID limit, and no direct
  container network.
- Do not mount the host project, home directory, SSH agent, Docker socket, or
  cloud credential paths into a capsule.
- Approved ports listen only on `127.0.0.1` through local Unix bridges.
- HTTP and HTTPS installers must honor standard proxy variables. SSH, `git://`,
  UDP, and arbitrary protocols stay blocked by design.

## Develop, test, package, and deploy

From a clean checkout with Node 22 or newer, Go 1.22 or newer, and the pinned
Playwright Chromium browser:

```sh
npm ci
npm test
npm run build
npm run package
```

Run every public claim from the same clean checkout:

```sh
node -e "for (const c of require('./.factory/claims.json')) console.log(c.test)"
```

Run the site locally with `npm run dev`. `npm run build` creates the deployable
site in `dist/site/` and the static Linux binary in `dist/bin/`. `npm run
package` creates Linux amd64 and arm64 release archives in `dist/packages/`.
Deploy `dist/site/` with its `staticwebapp.config.json` intact. That file keeps
the security, cache, route, and 404 policies with the static site.

The documentation site is <https://project-install-capsule.sociobot.in>.

## License

MIT. See [LICENSE](LICENSE).
