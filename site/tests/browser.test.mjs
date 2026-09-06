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

test('core composer content stays reachable at 200% text on a 390px phone', async () => {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 } });
  const page = await context.newPage();
  await page.goto(baseURL + '/', { waitUntil: 'networkidle' });
  await page.addStyleTag({ content: 'html { font-size: 200%; }' });
  const layout = await page.evaluate(() => {
    const bounds = (selector) => document.querySelector(selector).getBoundingClientRect().toJSON();
    return {
      viewport: window.innerWidth,
      scrollWidth: document.documentElement.scrollWidth,
      form: bounds('#capsule-composer'),
      review: bounds('.review-ticket')
    };
  });
  assert.equal(layout.scrollWidth <= layout.viewport, true, JSON.stringify(layout));
  assert.equal(layout.form.right <= layout.viewport, true, JSON.stringify(layout));
  assert.equal(layout.review.right <= layout.viewport, true, JSON.stringify(layout));
  await context.close();
});

test('phone controls meet target size and the light section focus ring has contrast', async () => {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 } });
  const page = await context.newPage();
  await page.goto(baseURL + '/', { waitUntil: 'networkidle' });
  const controls = await page.locator('a, button, input, summary').evaluateAll((elements) => elements.map((element) => {
    const box = element.getBoundingClientRect();
    return { text: element.textContent?.trim() || element.getAttribute('aria-label') || element.id, width: box.width, height: box.height, visible: Boolean(box.width && box.height) };
  }));
  assert.deepEqual(controls.filter((control) => control.visible && (control.width < 44 || control.height < 44)), []);
  await page.locator('.limits .text-link').focus();
  const contrast = await page.evaluate(() => {
    const parse = (value) => value.match(/\d+/g).slice(0, 3).map(Number).map((component) => {
      const channel = component / 255;
      return channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
    });
    const luminance = (value) => {
      const [red, green, blue] = parse(value);
      return red * 0.2126 + green * 0.7152 + blue * 0.0722;
    };
    const link = document.querySelector('.limits .text-link');
    const section = document.querySelector('.limits');
    const outline = getComputedStyle(link).outlineColor;
    const background = getComputedStyle(section).backgroundColor;
    const one = luminance(outline);
    const two = luminance(background);
    return (Math.max(one, two) + 0.05) / (Math.min(one, two) + 0.05);
  });
  assert.equal(contrast >= 3, true, `focus contrast was ${contrast}`);
  await context.close();
});
