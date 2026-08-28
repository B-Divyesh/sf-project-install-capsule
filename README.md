# Project Install Capsule

Run an unfamiliar project inside a rootless, network-denied container without exposing your home directory or credentials. Project Install Capsule gives declared hosts a narrow HTTP/HTTPS proxy, forwards only approved localhost ports, prints the exact capability diff, and writes a teardown receipt.

It is for developers who want to try a repository without turning a broad Docker command or a disposable VM checklist into an act of faith.

> Containers reduce exposure; they are not complete security boundaries. Capsule requires a rootless Podman or Docker engine, but it does not detect malware or defend against a kernel/runtime escape. Use a disposable VM for hostile code or high-value credentials.

## Install

Requires Go 1.22+ to build, and rootless Podman 5+ (recommended) or rootless Docker 27+ to run.

```sh
go install github.com/B-Divyesh/sf-project-install-capsule/cmd/capsule@latest
capsule --help
```

From a release archive, put the single `capsule` binary on your `PATH`. The package starts at `0.1.0`; release archives are produced by the factory, not this repository.

## Usage

Create a directory for the capsule declaration. The install command runs inside a fresh, empty `/workspace`; clone or download the project there.

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

`capsule init` writes `capsule.json`. `capsule inspect` prints the capability diff without running anything. `capsule run` refuses a non-rootless engine, starts with `--network=none`, uses an empty tmpfs workspace, mounts only the capsule binary and ephemeral bridge sockets, and listens for approved ports on `127.0.0.1`.

```json
{
  "version": 1,
  "image": "docker.io/library/node:22-bookworm",
  "install": "git clone https://github.com/owner/project.git .",
  "run": "npm install && npm run dev -- --host 127.0.0.1",
  "allow_hosts": ["github.com", "registry.npmjs.org"],
  "ports": [3000]
}
```

Automation can use JSON and dry-run output without prompts:

```sh
capsule inspect --json
capsule run --dry-run --json
capsule verify --json
capsule teardown --json
```

`capsule verify` performs live probes in the configured image: it seeds a temporary host-home secret, confirms the container cannot read it, confirms direct egress fails, and confirms the proxy rejects an undeclared host. Use `capsule verify --static` when an engine is intentionally unavailable (for example in a config-lint job).

Exit codes are `0` success, `2` invalid arguments/configuration, `3` engine unavailable or not rootless, and `4` runtime failure. `CAPSULE_ENGINE` may select an explicit `podman` or `docker` executable. No telemetry is collected.

## Security model

- No host project directory, home directory, environment, SSH agent, Docker socket, or cloud credential path is mounted.
- The workload gets a new tmpfs at `/workspace`, a read-only root filesystem, dropped Linux capabilities, `no-new-privileges`, a PID limit, and an isolated network namespace.
- Approved HTTP and HTTPS destinations travel through a hostname-checking proxy over an ephemeral Unix socket. Direct IP and undeclared-host requests are refused. Resolved loopback, private, link-local, multicast, and unspecified destinations are always refused.
- Approved ports are forwarded back to loopback through separate Unix sockets; they are never bound to the LAN.
- Redirects and application-specific protocols still need care. Keep allowlists narrow and review the printed diff.

## Develop and verify

```sh
npm install
npm test
npm run build
npm run package
```

`npm run build` creates the deployable docs site in `dist/site/` and the CLI in `dist/bin/`. `npm run package` creates release archives in `dist/packages/`. Run the site with `npm run dev`.

The live documentation is at <https://project-install-capsule.sociobot.in>.

## License

MIT. See [LICENSE](LICENSE).
