import { useCallback, useEffect, useMemo, useState } from 'react';
import { Outlet } from 'react-router';
import { useTranslation } from 'react-i18next';

import CommandPalette from '@/components/feedback/CommandPalette';
import { useWebSocketBridge } from '@/api/websocketBridge';
import { usePageTitle } from '@/hooks/usePageTitle';
import { NAV_TABS } from './navigation';

const MENU_KEY_BY_ICON: Record<string, string> = {
  dashboard: 'dashboard',
  inbound: 'inbounds',
  team: 'clients',
  groups: 'groups',
  cluster: 'nodes',
  hosts: 'hosts',
  outbound: 'outbounds',
  routing: 'routing',
  setting: 'settings',
  tool: 'xray',
  apidocs: 'apiDocs',
};

export default function PanelLayout() {
  useWebSocketBridge();
  usePageTitle();
  const { t } = useTranslation();
  const [paletteOpen, setPaletteOpen] = useState(false);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setPaletteOpen((open) => !open);
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  const paletteItems = useMemo(
    () =>
      NAV_TABS.map((tab) => {
        const menuKey = MENU_KEY_BY_ICON[tab.icon];
        return {
          key: tab.key,
          icon: tab.icon,
          title: t(`menu.${menuKey}`),
          keywords: t(`menu.${menuKey}`),
        };
      }),
    [t],
  );

  const closePalette = useCallback(() => setPaletteOpen(false), []);

  return (
    <>
      <Outlet />
      <CommandPalette items={paletteItems} open={paletteOpen} onClose={closePalette} />
    </>
  );
}