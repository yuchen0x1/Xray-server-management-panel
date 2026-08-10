import { describe, expect, it, vi } from 'vitest';

import { WebSocketClient } from '@/api/websocket';

describe('WebSocket base path handling', () => {
  it('builds a valid socket URL for a custom base path', () => {
    const OriginalWebSocket = globalThis.WebSocket;
    const originalWindow = (globalThis as { window?: unknown }).window;
    const addEventListener = vi.fn();
    const FakeWebSocket = class {
      static readonly OPEN = 1;
      static readonly CONNECTING = 0;
      readonly readyState = 0;

      constructor(readonly url: string) {}

      addEventListener = addEventListener;

      close() {}
    } as unknown as typeof WebSocket;

    Object.defineProperty(globalThis, 'WebSocket', { configurable: true, value: FakeWebSocket });
    Object.defineProperty(globalThis, 'window', {
      configurable: true,
      value: { location: { protocol: 'https:', host: 'panel.example.test' } },
    });

    try {
      const client = new WebSocketClient('xui');
      client.connect();

      const socket = (client as unknown as { ws: { url: string } }).ws;
      expect(socket.url).toBe('wss://panel.example.test/xui/ws');
      expect(addEventListener).toHaveBeenCalled();
    } finally {
      Object.defineProperty(globalThis, 'WebSocket', { configurable: true, value: OriginalWebSocket });
      Object.defineProperty(globalThis, 'window', { configurable: true, value: originalWindow });
    }
  });
});
