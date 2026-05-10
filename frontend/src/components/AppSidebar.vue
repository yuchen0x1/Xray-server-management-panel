<script setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  DashboardOutlined,
  UserOutlined,
  SettingOutlined,
  ToolOutlined,
  ClusterOutlined,
  LogoutOutlined,
  CloseOutlined,
  MenuOutlined,
} from '@ant-design/icons-vue';

import { currentTheme } from '@/composables/useTheme.js';
import ThemeSwitch from '@/components/ThemeSwitch.vue';

const { t } = useI18n();

const SIDEBAR_COLLAPSED_KEY = 'isSidebarCollapsed';

const props = defineProps({
  // Path prefix (e.g. /custom-base/) the panel is served under. Defaults
  // to '' which means tab keys end up as '/panel/...'. Pages pass the
  // value the Go backend gave them (in production via a meta tag).
  basePath: { type: String, default: '' },
  // Current request URI so the matching menu item highlights.
  requestUri: { type: String, default: '' },
});

// AD-Vue 4 dropped <a-icon :type="x"> in favor of explicit icon
// imports — keep a small name-to-component map so tab definitions stay
// declarative.
const iconByName = {
  dashboard: DashboardOutlined,
  user: UserOutlined,
  setting: SettingOutlined,
  tool: ToolOutlined,
  cluster: ClusterOutlined,
  logout: LogoutOutlined,
};

// basePath comes from Go (`/` by default, `/myprefix/` when configured) so
// these concatenations land on absolute paths. In dev we synthesize the prop
// from a window global which can be empty — force a leading slash so the
// browser doesn't resolve the link relative to the current pathname (which
// would turn /panel/settings + 'panel/...' into /panel/panel/...).
const prefix = props.basePath?.startsWith('/') ? props.basePath : `/${props.basePath || ''}`;

// Labels are i18n-driven so the sidebar matches the locale picked
// in panel settings without a page reload of the sidebar component.
const tabs = computed(() => [
  { key: `${prefix}panel/`, icon: 'dashboard', title: t('menu.dashboard') },
  { key: `${prefix}panel/inbounds`, icon: 'user', title: t('menu.inbounds') },
  { key: `${prefix}panel/nodes`, icon: 'cluster', title: t('menu.nodes') },
  { key: `${prefix}panel/settings`, icon: 'setting', title: t('menu.settings') },
  { key: `${prefix}panel/xray`, icon: 'tool', title: t('menu.xray') },
  { key: `${prefix}logout`, icon: 'logout', title: t('logout') },
]);

// Logout sits in its own pinned-to-bottom block on the drawer; the
// remaining items are the navigation proper. The full-height sider on
// desktop still uses `tabs` as-is so the desktop look is unchanged.
const navTabs = computed(() => tabs.value.filter((tab) => tab.icon !== 'logout'));
const utilTabs = computed(() => tabs.value.filter((tab) => tab.icon === 'logout'));

const activeTab = ref([props.requestUri]);

const drawerOpen = ref(false);
const collapsed = ref(JSON.parse(localStorage.getItem(SIDEBAR_COLLAPSED_KEY) || 'false'));

// Drawer width is capped against the viewport — AD-Vue's default 378px
// overflows on narrow phones (e.g. 360px portrait), leaving the page
// hidden behind the mask. `min()` keeps it sane on both phones and
// tablets while never exceeding 320px on larger displays.
const drawerWidth = 'min(82vw, 320px)';

function openLink(key) {
  if (key.startsWith('http')) {
    window.open(key);
  } else {
    window.location.href = key;
  }
}

function onCollapse(isCollapsed, type) {
  // Only persist explicit toggle clicks, not breakpoint-triggered collapses.
  if (type === 'clickTrigger') {
    localStorage.setItem(SIDEBAR_COLLAPSED_KEY, isCollapsed);
    collapsed.value = isCollapsed;
  }
}

function toggleDrawer() {
  drawerOpen.value = !drawerOpen.value;
}

function closeDrawer() {
  drawerOpen.value = false;
}
</script>

<template>
  <div class="ant-sidebar">
    <a-layout-sider :theme="currentTheme" collapsible :collapsed="collapsed" breakpoint="md" @collapse="onCollapse">
      <div class="sider-brand" :class="{ 'sider-brand-collapsed': collapsed }">
        {{ collapsed ? '3X' : '3X-UI' }}
      </div>
      <ThemeSwitch />
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in tabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-drawer placement="left" :closable="false" :open="drawerOpen" :wrap-class-name="currentTheme"
      :wrap-style="{ padding: 0 }" :width="drawerWidth"
      :body-style="{ padding: 0, display: 'flex', flexDirection: 'column', height: '100%' }"
      :header-style="{ display: 'none' }" @close="closeDrawer">
      <div class="drawer-header">
        <span class="drawer-brand">3X-UI</span>
        <button class="drawer-close" type="button" :aria-label="t('close')" @click="closeDrawer">
          <CloseOutlined />
        </button>
      </div>
      <ThemeSwitch />
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="drawer-menu drawer-nav"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in navTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="drawer-menu drawer-utility"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in utilTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
    </a-drawer>

    <button v-show="!drawerOpen" class="drawer-handle" type="button" :aria-label="t('menu.dashboard')"
      @click="toggleDrawer">
      <MenuOutlined />
    </button>
  </div>
</template>

<style scoped>
.ant-sidebar>.ant-layout-sider {
  height: 100%;
}

/* `.sider-brand` and `.drawer-brand` share the same light-theme colour
 * but differ in layout — the sider one is centered with its own
 * top-of-sidebar padding + border, the drawer one sits inside a flex
 * header next to the close button. Dark/ultra colour overrides live
 * in the non-scoped block at the bottom (theme classes attach to
 * body / html). */
.sider-brand,
.drawer-brand {
  font-weight: 600;
  font-size: 18px;
  letter-spacing: 0.5px;
  color: rgba(0, 0, 0, 0.88);
}

.sider-brand {
  text-align: center;
  padding: 16px 12px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.15);
  user-select: none;
}

.sider-brand-collapsed {
  font-size: 16px;
  padding: 16px 4px;
  letter-spacing: 0;
}

.drawer-handle {
  position: fixed;
  top: 12px;
  left: 12px;
  z-index: 1100;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  border: none;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  display: none;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.15);
}

.drawer-close {
  background: transparent;
  border: none;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  color: rgba(0, 0, 0, 0.65);
}

.drawer-close:hover,
.drawer-close:focus-visible {
  background: rgba(128, 128, 128, 0.18);
}

.drawer-menu :deep(.ant-menu-item) {
  height: 48px;
  line-height: 48px;
  margin: 0;
  border-radius: 0;
}

.drawer-menu :deep(.ant-menu-item .anticon) {
  font-size: 16px;
}

/* Push the utility (Logout) block to the bottom of the flex-column
 * drawer body and separate it from the nav block with a hairline. The
 * border colour is theme-neutral so it reads on both light and dark. */
.drawer-utility {
  margin-top: auto;
  border-top: 1px solid rgba(128, 128, 128, 0.15);
}

@media (max-width: 768px) {
  .drawer-handle {
    display: inline-flex;
  }

  /* On mobile the drawer is the menu — hide the inline sider's content
   * + the collapse trigger so the sider stops taking layout space and
   * leaves no remnant button next to the page. */
  .ant-sidebar>.ant-layout-sider :deep(.ant-layout-sider-children),
  .ant-sidebar>.ant-layout-sider :deep(.ant-layout-sider-trigger) {
    display: none;
  }

  .ant-sidebar>.ant-layout-sider {
    flex: 0 0 0 !important;
    max-width: 0 !important;
    min-width: 0 !important;
    width: 0 !important;
  }
}
</style>

<style>
/* Non-scoped so the rules survive AD-Vue teleporting the drawer body
 * outside the AppSidebar element's scope id. Without this the Vue
 * `:global(body.dark) .drawer-brand` form did not produce the expected
 * `body.dark .drawer-brand[data-v-xxx]` selector reliably, and the
 * drawer brand stayed at the light-theme dark colour on the navy
 * drawer surface. Class names are specific enough that no collision is
 * expected; AppSidebar owns the only drawer in the app. */
body.dark .drawer-brand,
body.dark .sider-brand {
  color: rgba(255, 255, 255, 0.92);
}

html[data-theme='ultra-dark'] .drawer-brand,
html[data-theme='ultra-dark'] .sider-brand {
  color: #ffffff;
}

body.dark .drawer-close {
  color: rgba(255, 255, 255, 0.75);
}

html[data-theme='ultra-dark'] .drawer-close {
  color: rgba(255, 255, 255, 0.85);
}

/* Pin the drawer surface to the same colour the desktop sider uses
 * (Layout.colorBgHeader / Menu.colorItemBg from useTheme.js) so the
 * header, empty body region, and menu items read as one continuous
 * panel. AD-Vue's CSS-in-JS tokens otherwise leave the drawer at
 * colorBgElevated (#2d2d30 in dark) which clashes with the #252526
 * menu rows. `!important` is required to beat the CSS-in-JS rule
 * specificity; AppSidebar owns the only drawer in the app so this
 * doesn't collide with anything else. */
body.dark .ant-drawer .ant-drawer-content,
body.dark .ant-drawer .ant-drawer-body {
  background: #252526 !important;
}

html[data-theme='ultra-dark'] .ant-drawer .ant-drawer-content,
html[data-theme='ultra-dark'] .ant-drawer .ant-drawer-body {
  background: #0a0a0a !important;
}
</style>
