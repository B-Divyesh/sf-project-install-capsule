import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const pages = ['site/index.html', 'site/privacy/index.html', 'site/terms/index.html'];

for (const page of pages) {
  test(`${page} has the required document landmarks`, async () => {
    const html = await readFile(page, 'utf8');
    assert.match(html, /<html lang="en">/);
    assert.match(html, /<title>[^<]+<\/title>/);
    assert.equal((html.match(/<h1(?:\s|>)/g) ?? []).length, 1);
    assert.equal((html.match(/<main(?:\s|>)/g) ?? []).length, 1);
    assert.match(html, /class="skip-link"/);
  });
}

test('hero image has dimensions, useful alt text, and priority', async () => {
  const html = await readFile('site/index.html', 'utf8');
  assert.match(html, /<img[^>]+alt="[^"]+"[^>]+width="1400"[^>]+height="933"[^>]+fetchpriority="high"/);
});

test('runtime has no third-party resources or analytics', async () => {
  for (const page of pages) {
    const html = await readFile(page, 'utf8');
    assert.doesNotMatch(html, /(?:src|href)="https:\/\/(?!github\.com)/);
    assert.doesNotMatch(html, /<script[^>]+(?:analytics|gtag|segment|posthog)/i);
  }
});

test('motion and keyboard focus policies are explicit', async () => {
  const css = await readFile('site/src/styles.css', 'utf8');
  assert.match(css, /:focus-visible/);
  assert.match(css, /prefers-reduced-motion/);
  assert.match(css, /min-height:\s*44px/);
});
