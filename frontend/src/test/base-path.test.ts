import { describe, expect, it } from 'vitest';

import { normalizeBasePath, withBasePath } from '@/utils/base-path';

describe('base path helpers', () => {
  it('normalizes root and repeated slashes', () => {
    expect(normalizeBasePath(undefined)).toBe('');
    expect(normalizeBasePath('/')).toBe('');
    expect(normalizeBasePath('///xui///')).toBe('/xui');
    expect(normalizeBasePath('http://localhost')).toBe('');
    expect(normalizeBasePath('http://localhost/xui/')).toBe('/xui');
  });

  it('joins panel paths without losing or duplicating separators', () => {
    expect(withBasePath('/panel/', '/xui')).toBe('/xui/panel/');
    expect(withBasePath('panel/api/openapi.json', '/xui/')).toBe('/xui/panel/api/openapi.json');
    expect(withBasePath('/panel/api/server/getDb', '/')).toBe('/panel/api/server/getDb');
  });
});
