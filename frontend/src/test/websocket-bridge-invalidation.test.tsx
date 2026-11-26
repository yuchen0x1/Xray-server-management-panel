import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';

import { useWebSocketBridge } from '@/api/websocketBridge';
import { keys } from '@/api/queryKeys';

const { listeners } = vi.hoisted(() => ({
  listeners: new Map<string, Set<(payload: unknown) => void>>(),
}));

vi.mock('@/api/websocket', () => ({
  getSharedWebSocketClient: () => ({
    connect() {},
    on(event: string, handler: (payload: unknown) => void) {
      const handlers = listeners.get(event) ?? new Set();
      handlers.add(handler);
      listeners.set(event, handlers);
    },
    off(event: string, handler: (payload: unknown) => void) {
      listeners.get(event)?.delete(handler);
    },
  }),
}));

vi.mock('@/api/invalidationTracker', () => ({
  isRecentLocalInvalidate: () => false,
}));

function emitInvalidate(type: string) {
  act(() => {
    for (const handler of listeners.get('invalidate') ?? []) handler({ type });
  });
}

function mountBridge(queryClient: QueryClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return renderHook(() => useWebSocketBridge(), { wrapper });
}

describe('websocket bridge invalidation', () => {
  const clients: QueryClient[] = [];

  beforeEach(() => {
    listeners.clear();
  });

  afterEach(() => {
    for (const client of clients) client.clear();
    clients.length = 0;
  });

  function createClient() {
    const client = new QueryClient();
    client.setQueryData(keys.inbounds.slim(), []);
    client.setQueryData(keys.clients.onlines(), []);
    clients.push(client);
    return client;
  }

  it('refreshes every resource in a burst of different invalidations', async () => {
    const queryClient = createClient();
    mountBridge(queryClient);

    emitInvalidate('inbounds');
    emitInvalidate('clients');

    await waitFor(() => {
      expect(queryClient.getQueryState(keys.inbounds.slim())?.isInvalidated).toBe(true);
      expect(queryClient.getQueryState(keys.clients.onlines())?.isInvalidated).toBe(true);
    });
  });

  it('coalesces repeated notifications for the same resource', async () => {
    const queryClient = createClient();
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries');
    mountBridge(queryClient);

    emitInvalidate('inbounds');
    emitInvalidate('inbounds');

    await waitFor(() => expect(invalidate).toHaveBeenCalledExactlyOnceWith({ queryKey: keys.inbounds.root() }));
    expect(queryClient.getQueryState(keys.clients.onlines())?.isInvalidated).toBe(false);
  });

  it('keeps pending invalidations independent between bridge instances', async () => {
    const first = createClient();
    const second = createClient();
    const firstBridge = mountBridge(first);
    mountBridge(second);

    emitInvalidate('inbounds');
    firstBridge.unmount();

    await waitFor(() => expect(second.getQueryState(keys.inbounds.slim())?.isInvalidated).toBe(true));
    expect(first.getQueryState(keys.inbounds.slim())?.isInvalidated).toBe(false);
  });
});
