import fs from 'node:fs';
import { chromium } from 'playwright';
import { AxeBuilder } from '@axe-core/playwright';

const base = 'https://project-install-capsule.sociobot.in';
const browser = await chromium.launch({ headless: true });
const report = { pages: [], offline: {}, reducedMotion: {}, keyboard: {}, privacy: {} };

for (const viewport of [{ name: 'desktop', width: 1440, height: 1000 }, { name: 'mobile', width: 390, height: 844 }]) {
  for (const path of ['/', '/privacy/', '/terms/']) {
    const context = await browser.newContext({ viewport });
    const page = await context.newPage();
    const consoleErrors = [];
    const pageErrors = [];
    page.on('console', m => { if (m.type() === 'error') consoleErrors.push(m.text()); });
    page.on('pageerror', e => pageErrors.push(e.message));
    const response = await page.goto(base + path, { waitUntil: 'networkidle' });
    const axe = await new AxeBuilder({ page }).analyze();
    const touchTargets = await page.locator('a,button,input,summary').evaluateAll(nodes => nodes.map(node => {
      const r = node.getBoundingClientRect();
      return { tag: node.tagName, text: (node.textContent || node.getAttribute('aria-label') || node.getAttribute('name') || '').trim().replace(/\s+/g, ' ').slice(0, 80), width: Math.round(r.width), height: Math.round(r.height) };
    }).filter(x => x.width < 44 || x.height < 44));
    report.pages.push({ viewport: viewport.name, path, status: response.status(), title: await page.title(), lang: await page.locator('html').getAttribute('lang'), h1Count: await page.locator('h1').count(), landmarks: { header: await page.locator('header').count(), nav: await page.locator('nav').count(), main: await page.locator('main').count(), footer: await page.locator('footer').count() }, overflow: await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), consoleErrors, pageErrors, axeSeriousCritical: axe.violations.filter(v => ['serious','critical'].includes(v.impact)).map(v => ({ id: v.id, impact: v.impact, targets: v.nodes.map(n => n.target) })), touchTargets });
    await context.close();
  }
}

{
  const context = await browser.newContext({ viewport: { width: 390, height: 844 } });
  const page = await context.newPage();
  await page.goto(base, { waitUntil: 'networkidle' });
  const sequence = [];
  for (let i = 0; i < 10; i++) {
    await page.keyboard.press('Tab');
    sequence.push(await page.evaluate(() => { const e = document.activeElement; const s = getComputedStyle(e); return { tag: e?.tagName, text: (e?.textContent || e?.getAttribute?.('aria-label') || '').trim().replace(/\s+/g,' ').slice(0,60), outline: `${s.outlineWidth} ${s.outlineStyle} ${s.outlineColor}`, rect: e ? { width: Math.round(e.getBoundingClientRect().width), height: Math.round(e.getBoundingClientRect().height) } : null }; }));
  }
  report.keyboard = { sequence };
  await context.close();
}

for (const reducedMotion of ['no-preference', 'reduce']) {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, reducedMotion });
  const page = await context.newPage();
  await page.goto(base, { waitUntil: 'domcontentloaded' });
  report.reducedMotion[reducedMotion] = await page.locator('.rise').first().evaluate(e => { const s = getComputedStyle(e); return { animationName: s.animationName, animationDuration: s.animationDuration, transitionDuration: s.transitionDuration, transform: s.transform }; });
  await context.close();
}

{
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, serviceWorkers: 'allow' });
  const page = await context.newPage();
  const requests = [];
  page.on('request', r => requests.push(r.url()));
  await page.goto(base, { waitUntil: 'networkidle' });
  await page.locator('#install-command').fill('git clone https://github.com/example/sample.git .');
  await page.locator('#allowed-hosts').fill('github.com, registry.npmjs.org');
  await page.locator('#allowed-ports').fill('4173');
  await page.locator('#copy-command').click();
  await page.waitForTimeout(200);
  report.privacy = { requests: [...new Set(requests)], localStorage: await page.evaluate(() => ({...localStorage})), sessionStorage: await page.evaluate(() => ({...sessionStorage})), cookies: await context.cookies() };
  await page.evaluate(() => navigator.serviceWorker.ready);
  await page.reload({ waitUntil: 'networkidle' });
  const controlled = await page.evaluate(() => Boolean(navigator.serviceWorker.controller));
  await context.setOffline(true);
  let rootOffline = false;
  let termsOffline = false;
  try { const r = await page.reload({ waitUntil: 'domcontentloaded' }); rootOffline = Boolean(r) && (await page.locator('h1').count()) === 1; } catch {}
  try { const r = await page.goto(base + '/terms/', { waitUntil: 'domcontentloaded' }); termsOffline = Boolean(r) && (await page.locator('h1').textContent()).trim() === 'Terms'; } catch {}
  report.offline = { controlled, rootOffline, termsOffline };
  await context.close();
}

fs.writeFileSync('.factory/qa-evidence/live-browser-qa.json', JSON.stringify(report, null, 2));
await browser.close();
