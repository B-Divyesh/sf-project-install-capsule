import { rm, mkdir } from 'node:fs/promises';
import { spawnSync } from 'node:child_process';

await rm('dist', { recursive: true, force: true });
await mkdir('dist/bin', { recursive: true });

const go = spawnSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', 'dist/bin/capsule', './cmd/capsule'], { stdio: 'inherit' });
if (go.status !== 0) process.exit(go.status ?? 1);

const vite = spawnSync(process.platform === 'win32' ? 'npx.cmd' : 'npx', ['vite', 'build', '--config', 'site/vite.config.ts'], { stdio: 'inherit' });
process.exit(vite.status ?? 1);
