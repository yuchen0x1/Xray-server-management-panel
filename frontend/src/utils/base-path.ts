export function normalizeBasePath(basePath: string | null | undefined): string {
  const value = (basePath ?? '').trim();
  if (!value || value === '/') return '';
  return `/${value.replace(/^\/+|\/+$/g, '')}`;
}

export function withBasePath(path: string, basePath?: string | null): string {
  const prefix = normalizeBasePath(
    basePath ?? (typeof window !== 'undefined' ? window.X_UI_BASE_PATH : ''),
  );
  const suffix = `/${path.replace(/^\/+/, '')}`;
  return `${prefix}${suffix}`;
}
