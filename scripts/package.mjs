import { mkdir, readFile, readdir, rm, unlink, writeFile } from 'node:fs/promises';
import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';

const version = '0.1.0';
const targets = [['linux', 'amd64'], ['linux', 'arm64']];
await rm('dist/packages', { recursive: true, force: true });
await mkdir('dist/packages', { recursive: true });

for (const [goos, goarch] of targets) {
  const name = `capsule-${version}-${goos}-${goarch}`;
  const env = { ...process.env, CGO_ENABLED: '0', GOOS: goos, GOARCH: goarch };
  const build = spawnSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', `dist/packages/${name}`, './cmd/capsule'], { env, stdio: 'inherit' });
  if (build.status !== 0) process.exit(build.status ?? 1);
  const archive = spawnSync('tar', ['-C', 'dist/packages', '-czf', `dist/packages/${name}.tar.gz`, name], { stdio: 'inherit' });
  if (archive.status !== 0) process.exit(archive.status ?? 1);
  await unlink(`dist/packages/${name}`);
}

const checksums = [];
for (const filename of (await readdir('dist/packages')).filter((name) => name.endsWith('.tar.gz')).sort()) {
  const digest = createHash('sha256').update(await readFile(`dist/packages/${filename}`)).digest('hex');
  checksums.push(`${digest}  ${filename}`);
}
await writeFile('dist/packages/SHA256SUMS', `${checksums.join('\n')}\n`);
