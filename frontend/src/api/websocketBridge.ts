import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { getSharedWebSocketClient } from '@/api/websocket';
import { keys } from '@/api/queryKeys';
import { isRecentLocalInvalidate } from '@/api/invalidationTracker';

type Handler = (payload: unknown) => void;

export function useWebSocketBridge() {
  const queryClient = useQueryClient();

  useEffect(() => {
    const client = getSharedWebSocketClient();
    let invalidateTimer: number | null = null;
    const pendingInvalidations = new Set<'inbounds' | 'clients'>();

    const onInvalidate: Handler = (payload) => {
      const p = payload as { type?: string } | undefined;
      if (!p || (p.type !== 'inbounds' && p.type !== 'clients')) return;
      pendingInvalidations.add(p.type);
      if (invalidateTimer != null) clearTimeout(invalidateTimer);
      invalidateTimer = window.setTimeout(() => {
        invalidateTimer = null;
        const resources = [...pendingInvalidations];
        pendingInvalidations.clear();
        if (isRecentLocalInvalidate()) return;
        for (const resource of resources) {
          queryClient.invalidateQueries({ queryKey: keys[resource].root() });
        }
      }, 200);
    };

    const onOutbounds: Handler = (payload) => {
      if (!Array.isArray(payload)) return;
      queryClient.setQueryData(keys.xray.outboundsTraffic(), payload);
    };

    const onNodes: Handler = (payload) => {
      if (!Array.isArray(payload)) return;
      queryClient.setQueryData(keys.nodes.list(), payload);
    };

    const onInbounds: Handler = (payload) => {
      if (!Array.isArray(payload)) return;
      queryClient.setQueryData(keys.inbounds.slim(), payload);
    };

    client.on('invalidate', onInvalidate);
    client.on('outbounds', onOutbounds);
    client.on('nodes', onNodes);
    client.on('inbounds', onInbounds);
    client.connect();

    return () => {
      client.off('invalidate', onInvalidate);
      client.off('outbounds', onOutbounds);
      client.off('nodes', onNodes);
      client.off('inbounds', onInbounds);
      if (invalidateTimer != null) {
        clearTimeout(invalidateTimer);
        invalidateTimer = null;
      }
      pendingInvalidations.clear();
    };
  }, [queryClient]);
}
