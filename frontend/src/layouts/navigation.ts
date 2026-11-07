export type NavIconName = 'dashboard' | 'inbound' | 'team' | 'groups' | 'cluster' | 'hosts' | 'outbound' | 'routing' | 'setting' | 'tool' | 'apidocs' | 'logout';

export type NavTab = {
  key: string;
  icon: NavIconName;
  title: string;
};

export const NAV_TABS: { key: string; icon: Exclude<NavIconName, 'logout'> }[] = [
  { key: '/', icon: 'dashboard' },
  { key: '/inbounds', icon: 'inbound' },
  { key: '/clients', icon: 'team' },
  { key: '/groups', icon: 'groups' },
  { key: '/nodes', icon: 'cluster' },
  { key: '/hosts', icon: 'hosts' },
  { key: '/outbound', icon: 'outbound' },
  { key: '/routing', icon: 'routing' },
  { key: '/settings', icon: 'setting' },
  { key: '/xray', icon: 'tool' },
  { key: '/api-docs', icon: 'apidocs' },
];