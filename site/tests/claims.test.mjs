import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import { access, mkdtemp, readFile, stat, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import test, { after, before } from 'node:test';
import { chromium } from 'playwright';
import { createServer } from 'vite';

let browser;
let server;
let baseURL;
let workDir;
let binary;

function runBinary(args, options = {}) {
  const result = spawnSync(binary, args, { encoding: 'utf8', ...options });
  assert.equal(result.status, 0, `${result.stdout}\n${result.stderr}`);
  return result.stdout;
}

function createSample(name) {
  const directory = join(workDir, name);
  const output = runBinary(['demo', '--dir', directory, '--json']);
  return JSON.parse(output);
}

before(async () => {
  workDir = await mkdtemp(join(tmpdir(), 'capsule-claims-'));
  binary = join(workDir, 'capsule');
  execFileSync('go', ['build', '-trimpath', '-o', binary, './cmd/capsule'], { stdio: 'inherit' });
  server = await createServer({ configFile: 'site/vite.config.ts', server: { host: '127.0.0.1', port: 0 } });
  await server.listen();
  const address = server.httpServer.address();
  baseURL = `http://127.0.0.1:${address.port}`;
  browser = await chromium.launch({ args: ['--no-sandbox'] });
});

after(async () => {
  await browser?.close();
  await server?.close();
});

test('@claim:cli-demo creates a fresh sample folder and prints its review', async () => {
  const sample = createSample('cli-demo');
  await access(sample.config);
  await access(sample.sample_input);
  assert.equal(sample.directory, join(workDir, 'cli-demo'));
  assert.match(sample.review.filesystem.join('\n'), /host home and project: not mounted/);
  assert.match(sample.review.network.join('\n'), /direct network: denied/);
});

test('@claim:runtime-boundary shows the sample runtime boundary before execution', () => {
  const sample = createSample('runtime-boundary');
  const preview = JSON.parse(runBinary(['run', '--dry-run', '--json', '--config', sample.config]));
  const invocation = preview.arguments.join('\u0000');
  assert.match(invocation, /--network=none/);
  assert.match(invocation, /--read-only/);
  assert.match(invocation, /--cap-drop=ALL/);
  assert.match(invocation, /\/workspace:rw,nosuid,nodev,exec,size=1g/);
  assert.doesNotMatch(invocation, /--network=host|--privileged|\/home\//);
  assert.deepEqual(preview.capabilities.ports, ['127.0.0.1:3000 → capsule:3000']);
});

test('@claim:teardown-receipt writes a protected receipt after a sample run', async () => {
  const sample = createSample('receipt');
  const fakeEngine = join(workDir, 'fake-podman');
  await writeFile(fakeEngine, '#!/bin/sh\nif [ "$1" = "info" ]; then echo true; exit 0; fi\nif [ "$1" = "run" ] || [ "$1" = "rm" ]; then exit 0; fi\nexit 1\n', { mode: 0o755 });
  const output = runBinary(['run', '--json', '--config', sample.config], { cwd: sample.directory, env: { ...process.env, CAPSULE_ENGINE: fakeEngine } });
  const result = JSON.parse(output);
  const receipt = join(sample.directory, result.receipt);
  const details = JSON.parse(await readFile(receipt, 'utf8'));
  assert.equal(details.outcome, 'completed');
  assert.equal((await stat(receipt)).mode & 0o777, 0o600);
});

test('@claim:demo-sandbox keeps sample edits separate and can reset them', async () => {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 } });
  const page = await context.newPage();
  await page.goto(`${baseURL}/?demo=1`, { waitUntil: 'networkidle' });
  await assert.doesNotReject(() => page.getByText('Demo — sample data, nothing is saved').waitFor());
  await page.locator('#install-command').fill('echo changed only in the sample');
  assert.deepEqual(await page.evaluate(() => Object.keys(localStorage)), ['demo:capsule-composer']);
  await page.getByRole('button', { name: 'Reset demo' }).click();
  assert.match(await page.locator('#install-command').inputValue(), /octocat\/Hello-World/);
  await Promise.all([page.waitForURL(`${baseURL}/`), page.getByRole('link', { name: 'Start for real' }).click()]);
  await page.locator('#demo-banner').waitFor({ state: 'hidden' });
  assert.equal(await page.locator('#demo-banner').isHidden(), true);
  assert.deepEqual(await page.evaluate(() => Object.keys(localStorage)), []);
  await context.close();
});

test('@claim:preview-private changes the normal preview without requests or storage', async () => {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(baseURL, { waitUntil: 'networkidle' });
  const requests = [];
  page.on('request', (request) => requests.push(request.url()));
  await page.locator('#allowed-hosts').fill('example.com');
  await page.locator('#allowed-ports').fill('4173');
  await page.waitForTimeout(100);
  assert.deepEqual(requests, []);
  assert.deepEqual(await page.evaluate(() => ({ local: Object.keys(localStorage), session: Object.keys(sessionStorage) })), { local: [], session: [] });
  await context.close();
});

test('@claim:preview-host-rule accepts a DNS root dot and rejects an IP address', async () => {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(`${baseURL}/?demo=1`);
  await page.locator('#allowed-hosts').fill('EXAMPLE.COM.');
  assert.match(await page.locator('#review-output').textContent(), /--allow-host example\.com/);
  assert.equal(await page.locator('#copy-command').isEnabled(), true);
  await page.locator('#allowed-hosts').fill('127.0.0.1');
  assert.match(await page.locator('#form-error').textContent(), /IP addresses/);
  assert.equal(await page.locator('#copy-command').isEnabled(), false);
  await context.close();
});

test('@claim:site-privacy loads without third-party requests, cookies, or normal storage', async () => {
  const context = await browser.newContext();
  const page = await context.newPage();
  const requests = [];
  page.on('request', (request) => requests.push(request.url()));
  await page.goto(`${baseURL}/privacy/`, { waitUntil: 'networkidle' });
  assert.equal(requests.every((url) => new URL(url).origin === baseURL), true);
  assert.deepEqual(await context.cookies(), []);
  assert.deepEqual(await page.evaluate(() => ({ local: Object.keys(localStorage), session: Object.keys(sessionStorage) })), { local: [], session: [] });
  await context.close();
});
