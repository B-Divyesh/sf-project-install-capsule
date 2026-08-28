import { mkdir, rm } from 'node:fs/promises';
import { spawnSync } from 'node:child_process';

const version = '0.1.0';
const targets = [['linux', 'amd64'], ['linux', 'arm64'], ['darwin', 'amd64'], ['darwin', 'arm64']];
await rm('dist/packages', { recursive: true, force: true });
await mkdir('dist/packages', { recursive: true });

for (const [goos, goarch] of targets) {
  const name = `capsule-${version}-${goos}-${goarch}`;
  const env = { ...process.env, CGO_ENABLED: '0', GOOS: goos, GOARCH: goarch };
  const build = spawnSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', `dist/packages/${name}`, './cmd/capsule'], { env, stdio: 'inherit' });
  if (build.status !== 0) process.exit(build.status ?? 1);
  const archive = spawnSync('tar', ['-C', 'dist/packages', '-czf', `dist/packages/${name}.tar.gz`, name], { stdio: 'inherit' });
  if (archive.status !== 0) process.exit(archive.status ?? 1);
}
