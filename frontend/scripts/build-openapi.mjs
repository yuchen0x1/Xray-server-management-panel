#!/usr/bin/env node
import { existsSync, readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { isDeepStrictEqual } from 'node:util';

import { sections } from '../src/pages/api-docs/endpoints.ts';
import { EXAMPLES } from '../src/generated/examples.ts';
import { SCHEMAS } from '../src/generated/schemas.ts';

const __dirname = dirname(fileURLToPath(import.meta.url));
const outPath = join(__dirname, '..', 'public', 'openapi.json');
const docsRoot = join(__dirname, '..', '..', 'docs');

const PANEL_VERSION = process.env.X_UI_VERSION || '3.x';

const SECURITY_SCHEMES = {
  bearerAuth: {
    type: 'http',
    scheme: 'bearer',
    description: 'API token from Settings → Security → API Token. Send as `Authorization: Bearer <token>`.',
  },
  cookieAuth: {
    type: 'apiKey',
    in: 'cookie',
    name: '3x-ui',
    description: 'Session cookie set by POST /login. Browser-only.',
  },
};

function ginPathToOpenApi(path) {
  return path.replace(/:([A-Za-z_][A-Za-z0-9_]*)/g, '{$1}');
}

function extractPathParams(openApiPath) {
  const params = [];
  const re = /\{([A-Za-z_][A-Za-z0-9_]*)\}/g;
  let m;
  while ((m = re.exec(openApiPath)) !== null) params.push(m[1]);
  return params;
}

function mapType(t) {
  const v = String(t || '').toLowerCase();
  if (v === 'number' || v === 'integer' || v === 'int') return 'integer';
  if (v === 'float' || v === 'double') return 'number';
  if (v === 'boolean' || v === 'bool') return 'boolean';
  if (v === 'array') return 'array';
  if (v === 'object') return 'object';
  return 'string';
}

function tryParseJson(raw) {
  if (typeof raw !== 'string') return undefined;
  try {
    return JSON.parse(raw);
  } catch {
    return undefined;
  }
}

function paramToOpenApi(p) {
  const out = {
    name: p.name,
    in: p.in,
    required: p.in === 'path' ? true : !p.optional,
    description: p.desc || '',
    schema: { type: mapType(p.type) },
  };
  if (p.defaultValue !== undefined) out.schema.default = p.defaultValue;
  return out;
}

function buildOperation(ep, tag) {
  const op = {
    tags: [tag],
    summary: ep.summary || '',
    operationId: `${ep.method.toLowerCase()}_${ep.path.replace(/[^A-Za-z0-9]+/g, '_').replace(/^_|_$/g, '')}`,
  };
  if (ep.description) op.description = ep.description;
  if (ep.deprecated) op.deprecated = true;

  const params = [];
  const bodyParams = [];
  for (const p of ep.params || []) {
    if (p.in === 'body') {
      bodyParams.push(p);
    } else if (p.in === 'path' || p.in === 'query' || p.in === 'header') {
      params.push(paramToOpenApi(p));
    }
  }

  const openApiPath = ginPathToOpenApi(ep.path);
  const declared = new Set(params.filter((x) => x.in === 'path').map((x) => x.name));
  for (const name of extractPathParams(openApiPath)) {
    if (declared.has(name)) continue;
    params.push({
      name,
      in: 'path',
      required: true,
      description: '',
      schema: { type: 'string' },
    });
  }

  if (params.length > 0) op.parameters = params;

  if (ep.body || bodyParams.length > 0) {
    const example = tryParseJson(ep.body);
    const properties = {};
    const required = [];
    for (const bp of bodyParams) {
      properties[bp.name] = {
        type: mapType(bp.type),
        description: bp.desc || '',
      };
      if (!bp.optional) required.push(bp.name);
    }
    const schema = bodyParams.length > 0
      ? { type: 'object', properties, ...(required.length > 0 ? { required } : {}) }
      : { type: 'object' };

    op.requestBody = {
      required: required.length > 0 || bodyParams.length === 0,
      content: {
        'application/json': {
          schema,
          ...(example !== undefined ? { example } : {}),
        },
      },
    };
  }

  const responses = {};
  let successExample = tryParseJson(ep.response);
  let objSchema = {};
  if (ep.responseSchema) {
    const obj = EXAMPLES[ep.responseSchema];
    if (obj === undefined) {
      throw new Error(`${ep.method} ${ep.path}: responseSchema "${ep.responseSchema}" has no generated example`);
    }
    if (SCHEMAS[ep.responseSchema] === undefined) {
      throw new Error(`${ep.method} ${ep.path}: responseSchema "${ep.responseSchema}" has no generated schema`);
    }
    const ref = { $ref: `#/components/schemas/${ep.responseSchema}` };
    objSchema = ep.responseSchemaArray ? { type: 'array', items: ref } : ref;
    if (successExample === undefined) {
      successExample = { success: true, obj: ep.responseSchemaArray ? [obj] : obj };
    }
  }
  const responseContentType = ep.responseContentType || 'application/json';
  const responseContent = responseContentType === 'application/json'
    ? {
        schema: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            msg: { type: 'string' },
            obj: objSchema,
          },
        },
        ...(successExample !== undefined ? { example: successExample } : {}),
      }
    : {
        schema: { type: 'string' },
      };
  responses['200'] = {
    description: 'Successful response',
    content: {
      [responseContentType]: responseContent,
    },
  };

  const errExample = tryParseJson(ep.errorResponse);
  if (errExample !== undefined || ep.errorStatus) {
    const code = String(ep.errorStatus || 400);
    responses[code] = {
      description: 'Error response',
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              success: { type: 'boolean' },
              msg: { type: 'string' },
            },
          },
          ...(errExample !== undefined ? { example: errExample } : {}),
        },
      },
    };
  }

  op.responses = responses;
  return op;
}

export function buildSpec() {
  const paths = {};
  for (const section of sections) {
    const tag = section.title;
    for (const ep of section.endpoints) {
      if (ep.method === 'WS') continue;
      const openApiPath = ginPathToOpenApi(ep.path);
      if (!paths[openApiPath]) paths[openApiPath] = {};
      paths[openApiPath][ep.method.toLowerCase()] = buildOperation(ep, tag);
    }
  }

  const tags = sections.map((s) => ({
    name: s.title,
    description: s.description || '',
  }));

  return {
    openapi: '3.0.3',
    info: {
      title: '3X-UI Panel API',
      version: PANEL_VERSION,
      description:
        'Programmatic interface to a 3X-UI panel. Authenticate either by logging in (cookie) or with an API token from Settings → Security → API Token (Bearer). All endpoints under /panel/api/* honour both modes — an API token is a full-admin credential, so treat it like the panel password.',
    },
    servers: [
      { url: '/', description: 'Current panel (basePath aware)' },
    ],
    components: {
      securitySchemes: SECURITY_SCHEMES,
      schemas: SCHEMAS,
    },
    security: [{ bearerAuth: [] }, { cookieAuth: [] }],
    tags,
    paths,
  };
}

export function syncOpenApiSpec(spec, paths, { check = false } = {}) {
  if (check) {
    const stale = paths.filter((path) => {
      try {
        return !isDeepStrictEqual(JSON.parse(readFileSync(path, 'utf8')), spec);
      } catch (error) {
        if (error.code === 'ENOENT' || error instanceof SyntaxError) return true;
        throw error;
      }
    });
    if (stale.length > 0) {
      throw new Error(`[openapi] stale copies: ${stale.join(', ')}. Run npm --prefix frontend run gen:api.`);
    }
    return;
  }

  const content = JSON.stringify(spec, null, 2) + '\n';
  for (const path of paths) writeFileSync(path, content);
}

export function syncOpenApiPages(spec, contentRoot, { check = false } = {}) {
  const operationsByTag = new Map((spec.tags ?? []).map(({ name }) => [name, []]));
  for (const [path, operations] of Object.entries(spec.paths)) {
    for (const [method, operation] of Object.entries(operations)) {
      for (const tag of operation.tags ?? []) {
        const taggedOperations = operationsByTag.get(tag);
        if (!taggedOperations) throw new Error(`[openapi] undeclared tag: ${tag}`);
        taggedOperations.push({ path, method });
      }
    }
  }

  const updates = [];
  const locales = readdirSync(contentRoot, { withFileTypes: true }).filter((entry) => entry.isDirectory());
  if (locales.length === 0) throw new Error(`[openapi] no documentation locales in ${contentRoot}`);
  for (const locale of locales) {
    for (const [tag, operations] of operationsByTag) {
      const slug = tag.toLowerCase().replace(/\s+/g, '-');
      const path = join(contentRoot, locale.name, 'reference', 'api', `${slug}.mdx`);
      const content = readFileSync(path, 'utf8');
      const matches = [...content.matchAll(/\boperations=\{(\[[^\r\n]*\])\}/g)];
      if (matches.length !== 1) throw new Error(`[openapi] expected one operations array in ${path}`);
      const match = matches[0];
      if (isDeepStrictEqual(JSON.parse(match[1]), operations)) continue;
      const start = match.index + match[0].indexOf(match[1]);
      const updated = content.slice(0, start) + JSON.stringify(operations) + content.slice(start + match[1].length);
      updates.push({ path, content: updated });
    }
  }

  if (check && updates.length > 0) {
    throw new Error(`[openapi] stale documentation operations: ${updates.map(({ path }) => path).join(', ')}. Run npm --prefix frontend run gen:api.`);
  }
  if (!check) {
    for (const { path, content } of updates) writeFileSync(path, content);
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const spec = buildSpec();
  const check = process.argv.includes('--check');
  const includeDocs = check || existsSync(docsRoot);
  const outPaths = includeDocs ? [outPath, join(docsRoot, 'public', 'openapi.json')] : [outPath];
  try {
    syncOpenApiSpec(spec, outPaths, { check });
    if (includeDocs) syncOpenApiPages(spec, join(docsRoot, 'content', 'docs'), { check });
    const pathCount = Object.keys(spec.paths).length;
    let opCount = 0;
    for (const ops of Object.values(spec.paths)) opCount += Object.keys(ops).length;
    for (const path of outPaths) console.log(`[openapi] ${check ? 'checked' : 'wrote'} ${path}`);
    console.log(`[openapi] paths: ${pathCount}, operations: ${opCount}, tags: ${spec.tags.length}`);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
