import { copyFile, readFile, readdir, writeFile, rm, mkdir } from 'node:fs/promises';
import { spawnSync } from 'node:child_process';

const siteOnly = process.argv.includes('--site-only');
await rm(siteOnly ? 'dist/site' : 'dist', { recursive: true, force: true });
if (!siteOnly) {
  await mkdir('dist/bin', { recursive: true });
  const go = spawnSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', 'dist/bin/capsule', './cmd/capsule'], { stdio: 'inherit', env: { ...process.env, CGO_ENABLED: '0' } });
  if (go.status !== 0) process.exit(go.status ?? 1);
}

const vite = spawnSync(process.platform === 'win32' ? 'npx.cmd' : 'npx', ['vite', 'build', '--config', 'site/vite.config.ts'], { stdio: 'inherit' });
if (vite.status !== 0) process.exit(vite.status ?? 1);
await copyFile('node_modules/@fontsource-variable/league-spartan/LICENSE', 'dist/site/font-license.txt');
const assets = (await readdir('dist/site/assets')).filter((name) => /\.(?:css|js|woff2)$/.test(name)).map((name) => `'/assets/${name}'`).join(', ');
const serviceWorkerPath = 'dist/site/sw.js';
const serviceWorker = await readFile(serviceWorkerPath, 'utf8');
await writeFile(serviceWorkerPath, serviceWorker.replace('/* ASSET_MANIFEST */', assets ? `${assets},` : ''));
