import { useCallback, useEffect, useMemo, useState } from 'react';
import type { ComponentType } from 'react';
import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { Modal } from 'antd';
import {
  ApiOutlined,
  ClusterOutlined,
  DashboardOutlined,
  ExportOutlined,
  GlobalOutlined,
  ImportOutlined,
  SettingOutlined,
  SwapOutlined,
  TagsOutlined,
  TeamOutlined,
  ToolOutlined,
} from '@ant-design/icons';

import './CommandPalette.css';

export type CommandPaletteItem = {
  key: string;
  icon: string;
  title: string;
  keywords?: string;
};

const iconByName: Record<string, ComponentType> = {
  dashboard: DashboardOutlined,
  inbound: ImportOutlined,
  team: TeamOutlined,
  groups: TagsOutlined,
  cluster: ClusterOutlined,
  hosts: GlobalOutlined,
  outbound: ExportOutlined,
  routing: SwapOutlined,
  setting: SettingOutlined,
  tool: ToolOutlined,
  apidocs: ApiOutlined,
};

export default function CommandPalette({
  items,
  open,
  onClose,
}: {
  items: CommandPaletteItem[];
  open: boolean;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    if (open) {
      setQuery('');
      setActiveIndex(0);
    }
  }, [open]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return items;
    return items.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        item.key.toLowerCase().includes(q) ||
        item.keywords?.toLowerCase().includes(q),
    );
  }, [items, query]);

  const go = useCallback(
    (key: string) => {
      navigate(key);
      onClose();
    },
    [navigate, onClose],
  );

  const onKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setActiveIndex((i) => (i + 1) % Math.max(filtered.length, 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setActiveIndex((i) => (i - 1 + filtered.length) % Math.max(filtered.length, 1));
      } else if (e.key === 'Enter' && filtered[activeIndex]) {
        e.preventDefault();
        go(filtered[activeIndex].key);
      }
    },
    [filtered, activeIndex, go],
  );

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
      width={520}
      title={t('menu.palette')}
      styles={{ body: { padding: 0 } }}
    >
      <div className="command-palette">
        <input
          autoFocus
          className="command-palette-input"
          placeholder={t('menu.palettePlaceholder')}
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setActiveIndex(0);
          }}
          onKeyDown={onKeyDown}
        />
        <ul className="command-palette-list">
          {filtered.length === 0 && <li className="command-palette-empty">{t('menu.paletteEmpty')}</li>}
          {filtered.map((item, i) => {
            const Icon = iconByName[item.icon];
            return (
              <li key={item.key}>
                <button
                  type="button"
                  className={`command-palette-item${i === activeIndex ? ' is-active' : ''}`}
                  onMouseEnter={() => setActiveIndex(i)}
                  onClick={() => go(item.key)}
                >
                  <Icon />
                  <span className="command-palette-title">{item.title}</span>
                  <span className="command-palette-key">{item.key}</span>
                </button>
              </li>
            );
          })}
        </ul>
      </div>
    </Modal>
  );
}