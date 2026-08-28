import assert from 'node:assert/strict';
import test, { after, before } from 'node:test';
import { AxeBuilder } from '@axe-core/playwright';
import { chromium } from 'playwright';
import { createServer } from 'vite';

let server;
let browser;
let baseURL;

before(async () => {
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

for (const path of ['/', '/privacy/', '/terms/']) {
  test(`axe finds no serious issues on ${path}`, async () => {
    const context = await browser.newContext({ viewport: { width: 390, height: 844 } });
    const page = await context.newPage();
    const consoleErrors = [];
    page.on('console', (message) => { if (message.type() === 'error') consoleErrors.push(message.text()); });
    await page.goto(baseURL + path, { waitUntil: 'networkidle' });
    const results = await new AxeBuilder({ page }).analyze();
    const severe = results.violations.filter((item) => ['serious', 'critical'].includes(item.impact));
    assert.deepEqual(severe.map(({ id, impact, nodes }) => ({ id, impact, nodes: nodes.map((node) => ({ target: node.target, summary: node.failureSummary })) })), []);
    assert.equal(consoleErrors.length, 0, consoleErrors.join('\n'));
    assert.equal(await page.locator('h1').count(), 1);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), true, 'page overflows at 390px');
    await context.close();
  });
}

test('capability composer validates and redraws without a request', async () => {
  const page = await browser.newPage();
  await page.goto(baseURL + '/');
  await page.locator('#allowed-hosts').fill('github.com, bad host');
  await assert.rejects(() => page.locator('#copy-command').click({ timeout: 500 }), /disabled|Timeout/);
  assert.match(await page.locator('#form-error').textContent(), /hostnames/);
  await page.locator('#allowed-hosts').fill('github.com');
  await page.locator('#allowed-ports').fill('4173');
  assert.match(await page.locator('#review-output').textContent(), /--port 4173/);
  assert.equal(await page.locator('#copy-command').isEnabled(), true);
  await page.close();
});
