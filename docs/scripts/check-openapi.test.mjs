import assert from 'node:assert/strict';
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { buildSpec, syncOpenApiPages, syncOpenApiSpec } from '../../frontend/scripts/build-openapi.mjs';

const spec = {
  openapi: '3.0.3',
  info: { title: 'Panel API', version: '3.x' },
  paths: { '/panel/api/server/getUpdateStatus': { get: { responses: {} } } },
};

test('generated OpenAPI contains HTTP operations and preserves the WebSocket handshake', () => {
  const generated = buildSpec();
  const methods = new Set(['get', 'put', 'post', 'delete', 'options', 'head', 'patch', 'trace']);

  for (const [path, operations] of Object.entries(generated.paths)) {
    assert.ok(path.startsWith('/'), `Invalid HTTP path: ${path}`);
    for (const method of Object.keys(operations)) assert.ok(methods.has(method), `Invalid HTTP method: ${method}`);
  }
  assert.equal(generated.paths['/ws'].get.operationId, 'get_ws');
});

test('generation works in the frontend-only Docker build stage', (t) => {
  const root = mkdtempSync(join(tmpdir(), 'xui-frontend-stage-'));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const sourceRoot = fileURLToPath(new URL('../../', import.meta.url));
  for (const path of [
    'frontend/scripts/build-openapi.mjs',
    'frontend/src/pages/api-docs/endpoints.ts',
    'frontend/src/generated/examples.ts',
    'frontend/src/generated/schemas.ts',
  ]) {
    const target = join(root, path);
    mkdirSync(dirname(target), { recursive: true });
    copyFileSync(join(sourceRoot, path), target);
  }
  mkdirSync(join(root, 'frontend', 'public'));
  const result = spawnSync(process.execPath, [join(root, 'frontend', 'scripts', 'build-openapi.mjs')], {
    encoding: 'utf8',
  });

  assert.equal(result.status, 0, result.stderr);
  const generated = JSON.parse(readFileSync(join(root, 'frontend', 'public', 'openapi.json'), 'utf8'));
  assert.equal(generated.paths['/ws'].get.operationId, 'get_ws');
});

function temporaryPaths(t) {
  const directory = mkdtempSync(join(tmpdir(), 'xui-openapi-'));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  return [join(directory, 'frontend.json'), join(directory, 'docs.json')];
}

test('generation writes the current specification to both consumers', (t) => {
  const paths = temporaryPaths(t);
  syncOpenApiSpec(spec, paths);

  const contents = paths.map((path) => readFileSync(path, 'utf8'));
  assert.equal(contents[0], contents[1]);
  assert.deepEqual(JSON.parse(contents[1]), spec);
});

test('freshness checks reject a stale docs copy without rewriting either copy', (t) => {
  const paths = temporaryPaths(t);
  syncOpenApiSpec(spec, paths);
  writeFileSync(paths[1], JSON.stringify({ ...spec, paths: {} }));
  const before = paths.map((path) => readFileSync(path, 'utf8'));

  assert.throws(
    () => syncOpenApiSpec(spec, paths, { check: true }),
    (error) => error.message.includes(paths[1]) && error.message.includes('gen:api'),
  );
  assert.deepEqual(paths.map((path) => readFileSync(path, 'utf8')), before);
});

test('freshness checks compare structured content independently of formatting', (t) => {
  const paths = temporaryPaths(t);
  syncOpenApiSpec(spec, paths);
  writeFileSync(paths[1], JSON.stringify(spec));

  assert.doesNotThrow(() => syncOpenApiSpec(spec, paths, { check: true }));
});

test('freshness checks reject missing and malformed specifications', (t) => {
  const paths = temporaryPaths(t);
  writeFileSync(paths[0], '{');

  assert.throws(
    () => syncOpenApiSpec(spec, paths, { check: true }),
    (error) => paths.every((path) => error.message.includes(path)),
  );
});

const taggedSpec = {
  tags: [{ name: 'Hosts' }, { name: 'API Tokens' }],
  paths: {
    '/panel/api/hosts/get/{groupId}': { get: { tags: ['Hosts'] } },
    '/panel/api/hosts/bulk/add': { post: { tags: ['Hosts'] } },
    '/panel/api/setting/apiTokens': { get: { tags: ['API Tokens'] } },
  },
};

function temporaryPages(t) {
  const root = mkdtempSync(join(tmpdir(), 'xui-api-pages-'));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const pages = [];
  for (const locale of ['en', 'zh']) {
    const directory = join(root, locale, 'reference', 'api');
    mkdirSync(directory, { recursive: true });
    for (const slug of ['hosts', 'api-tokens']) {
      const path = join(directory, `${slug}.mdx`);
      const content = `---\ntitle: ${locale} ${slug}\n---\nUnchanged introduction.\n<Comp operations={[]} showTitle />\n`;
      writeFileSync(path, content);
      pages.push({ path, content });
    }
  }
  return { root, pages };
}

test('page generation includes every tagged operation in every existing locale', (t) => {
  const { root, pages } = temporaryPages(t);
  syncOpenApiPages(taggedSpec, root);

  for (const page of pages) {
    const content = readFileSync(page.path, 'utf8');
    const match = content.match(/operations=\{(\[[^\r\n]*\])\}/);
    const operations = JSON.parse(match[1]);
    const tag = page.path.endsWith('hosts.mdx') ? 'Hosts' : 'API Tokens';
    const expected = Object.entries(taggedSpec.paths).flatMap(([path, methods]) =>
      Object.entries(methods)
        .filter(([, operation]) => operation.tags.includes(tag))
        .map(([method]) => ({ path, method })),
    );
    assert.deepEqual(operations, expected);
    assert.equal(content.replace(match[1], '[]'), page.content);
  }
  assert.doesNotThrow(() => syncOpenApiPages(taggedSpec, root, { check: true }));
});

test('page freshness checks reject stale operations without rewriting content', (t) => {
  const { root, pages } = temporaryPages(t);

  assert.throws(
    () => syncOpenApiPages(taggedSpec, root, { check: true }),
    (error) => pages.every(({ path }) => error.message.includes(path)),
  );
  for (const { path, content } of pages) assert.equal(readFileSync(path, 'utf8'), content);
});

test('page generation fails before writing if an expected tag page is missing', (t) => {
  const { root, pages } = temporaryPages(t);
  rmSync(pages[3].path);

  assert.throws(() => syncOpenApiPages(taggedSpec, root), { code: 'ENOENT' });
  for (const { path, content } of pages.slice(0, 3)) assert.equal(readFileSync(path, 'utf8'), content);
});

test('page generation rejects operations assigned to an undeclared tag', (t) => {
  const { root } = temporaryPages(t);
  const invalid = { ...taggedSpec, tags: [{ name: 'Hosts' }] };

  assert.throws(() => syncOpenApiPages(invalid, root), /undeclared tag: API Tokens/);
});
