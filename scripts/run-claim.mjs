import { spawnSync } from 'node:child_process';

const tag = process.argv[2];
if (!/^@claim:[a-z0-9-]+$/.test(tag ?? '')) {
  console.error('Usage: npm run test:claim -- @claim:<claim-id>');
  process.exit(2);
}

const result = spawnSync(process.execPath, ['--test', '--test-name-pattern', tag, 'site/tests/claims.test.mjs'], { stdio: 'inherit' });
process.exit(result.status ?? 1);
