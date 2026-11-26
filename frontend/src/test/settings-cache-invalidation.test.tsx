import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { keys } from '@/api/queryKeys';
import { useAllSettings } from '@/api/queries/useAllSettings';
import { HttpUtil, Msg } from '@/utils';

describe('settings cache invalidation', () => {
  afterEach(() => {
    vi.mocked(HttpUtil.post).mockReset();
  });

  it.each([true, false])('invalidates dependent defaults only when save success is %s', async (success) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const initialDefaults = { subDomain: 'old.example.com', pageSize: 25 };
    queryClient.setQueryData(keys.settings.all(), initialDefaults);
    queryClient.setQueryData(keys.settings.defaults(), initialDefaults);
    queryClient.setQueryData(keys.settings.factoryDefaults(), { pageSize: '25' });
    vi.mocked(HttpUtil.post).mockImplementation(async (url) => (
      url === '/panel/api/setting/update'
        ? new Msg(success, '', null)
        : new Msg(true, '', { subDomain: 'new.example.com', pageSize: 25 })
    ));
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const hook = renderHook(() => useAllSettings(), { wrapper });

    await act(async () => {
      await hook.result.current.savePayload({ subDomain: 'new.example.com' });
    });

    expect(HttpUtil.post).toHaveBeenCalledWith('/panel/api/setting/update', { subDomain: 'new.example.com' });
    expect(queryClient.getQueryState(keys.settings.defaults())?.isInvalidated).toBe(success);
    expect(queryClient.getQueryData(keys.settings.defaults())).toEqual(initialDefaults);
    expect(queryClient.getQueryState(keys.settings.factoryDefaults())?.isInvalidated).toBe(false);
    hook.unmount();
    queryClient.clear();
  });
});
