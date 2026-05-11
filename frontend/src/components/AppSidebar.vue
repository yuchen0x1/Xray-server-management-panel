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

import { theme, currentTheme, toggleTheme, toggleUltra, pauseAnimationsUntilLeave } from '@/composables/useTheme.js';

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

/* 3-state theme cycle driven by the brand-row icon button.
 *   Light  → Dark   (turn dark on, ensure ultra off)
 *   Dark   → Ultra  (turn ultra on)
 *   Ultra  → Light  (turn ultra off, turn dark off)
 * Using a single button keeps the sider header clean — the old
 * ThemeSwitch a-sub-menu plus its expandable items lived here. */
function cycleTheme() {
  pauseAnimationsUntilLeave('theme-cycle');
  if (!theme.isDark) {
    toggleTheme();
    if (theme.isUltra) toggleUltra();
  } else if (!theme.isUltra) {
    toggleUltra();
  } else {
    toggleUltra();
    toggleTheme();
  }
}
</script>

<template>
  <div class="ant-sidebar">
    <a-layout-sider :theme="currentTheme" collapsible :collapsed="collapsed" breakpoint="md" @collapse="onCollapse">
      <div class="sider-brand" :class="{ 'sider-brand-collapsed': collapsed }">
        <span class="brand-text">{{ collapsed ? '3X' : '3X-UI' }}</span>
        <button v-if="!collapsed" id="theme-cycle" type="button" class="theme-cycle" :aria-label="t('menu.theme')"
          :title="t('menu.theme')" @click="cycleTheme">
          <svg v-if="!theme.isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="4" />
            <path
              d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
          </svg>
          <svg v-else-if="!theme.isUltra" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            <path fill="none" d="M19 3l0.7 1.4 1.4 0.7-1.4 0.7L19 7.2l-0.7-1.4-1.4-0.7 1.4-0.7z" />
          </svg>
        </button>
      </div>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="sider-nav"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in navTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="sider-utility"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in utilTabs" :key="tab.key">
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
        <div class="drawer-header-actions">
          <button id="theme-cycle-drawer" type="button" class="theme-cycle" :aria-label="t('menu.theme')"
            :title="t('menu.theme')" @click="cycleTheme">
            <svg v-if="!theme.isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="12" cy="12" r="4" />
              <path
                d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
            </svg>
            <svg v-else-if="!theme.isUltra" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
              <path fill="none" d="M19 3l0.7 1.4 1.4 0.7-1.4 0.7L19 7.2l-0.7-1.4-1.4-0.7 1.4-0.7z" />
            </svg>
          </button>
          <button class="drawer-close" type="button" :aria-label="t('close')" @click="closeDrawer">
            <CloseOutlined />
          </button>
        </div>
      </div>
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
/* Pin the desktop sider to the viewport. Without this, AD-Vue's
 * `<a-layout-sider>` stretches to match the flex row's height — which
 * equals the page height on tall dashboards (cards stack into one
 * column below `lg` = 992px), so the bottom-anchored
 * `.ant-layout-sider-trigger` (and Logout right above it) slide off
 * the screen. Sticky + 100vh keeps the sider exactly viewport-tall;
 * `align-self: flex-start` stops the flex row from re-stretching it. */
.ant-sidebar>.ant-layout-sider {
  position: sticky;
  top: 0;
  height: 100vh;
  align-self: flex-start;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 14px 14px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.15);
  user-select: none;
}

/* Collapsed sider only has room for the '3X' brand — center it and
 * hide the theme cycle button (which is `v-if`-ed out in template). */
.sider-brand-collapsed {
  justify-content: center;
  font-size: 16px;
  padding: 14px 4px;
  letter-spacing: 0;
}

.brand-text {
  flex: 1 1 auto;
}

.sider-brand-collapsed .brand-text {
  flex: 0 0 auto;
}

.theme-cycle {
  background: transparent;
  border: none;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: rgba(0, 0, 0, 0.75);
  padding: 0;
  flex-shrink: 0;
  transition: background-color 0.2s, transform 0.15s, color 0.2s;
}

.theme-cycle:hover,
.theme-cycle:focus-visible {
  background-color: rgba(64, 150, 255, 0.1);
  color: #4096ff;
  transform: scale(1.08);
  outline: none;
}

.theme-cycle svg {
  width: 16px;
  height: 16px;
}

.drawer-header-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
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

/* Pin Logout exactly above AD-Vue's `.ant-layout-sider-trigger` (the
 * collapse bar at the bottom, position: absolute; height: 48px). The
 * old `margin-top: auto` approach only pushed the utility down when the
 * content was shorter than the container — on short viewports the
 * Logout got hidden behind the trigger. Switching to a flex layout
 * where `.sider-nav` consumes all spare space (flex: 1) and
 * `.sider-utility` stays at content height pins it consistently. The
 * padding-bottom: 48px on the parent reserves the trigger's strip so
 * Logout sits directly above it.
 *
 * The mobile @media rule below still hides the whole sider on phones;
 * this block only kicks in once that override no longer matches. */
.ant-sidebar>.ant-layout-sider :deep(.ant-layout-sider-children) {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding-bottom: 48px;
}

.sider-brand {
  flex: 0 0 auto;
}

.sider-nav {
  flex: 1 1 auto;
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 0;
}

.sider-utility {
  flex: 0 0 auto;
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

/* Force a visible icon colour on the theme cycle button across themes.
 * The scoped `color: inherit` previously relied on parent chain to
 * cascade — fine on the desktop sider where `.sider-brand` is themed,
 * but inside the teleported drawer body the cascade didn't reach and
 * the icon merged into the dark background on mobile. */
body.dark .theme-cycle {
  color: rgba(255, 255, 255, 0.85);
}

html[data-theme='ultra-dark'] .theme-cycle {
  color: rgba(255, 255, 255, 0.92);
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

/* Force the same light-blue tint on selected + hover/active across
 * all three themes. AD-Vue's defaults read too subtle on the dark
 * sider, and the light-theme variant looked inconsistent vs. dark —
 * applying the same RGBA tint over all backgrounds gives the active
 * page the same visual weight everywhere. `!important` is required to
 * beat AD-Vue's CSS-in-JS specificity; scoped to .sider-nav /
 * .sider-utility / .drawer-menu so only the navigation menus pick up
 * the override (other a-menu instances keep AD-Vue defaults). */
.sider-nav .ant-menu-item-selected,
.sider-utility .ant-menu-item-selected,
.drawer-menu .ant-menu-item-selected {
  background-color: rgba(64, 150, 255, 0.2) !important;
  color: #4096ff !important;
}

.sider-nav .ant-menu-item-active:not(.ant-menu-item-selected),
.sider-utility .ant-menu-item-active:not(.ant-menu-item-selected),
.drawer-menu .ant-menu-item-active:not(.ant-menu-item-selected),
.sider-nav .ant-menu-item:not(.ant-menu-item-selected):not(.ant-menu-item-disabled):hover,
.sider-utility .ant-menu-item:not(.ant-menu-item-selected):not(.ant-menu-item-disabled):hover,
.drawer-menu .ant-menu-item:not(.ant-menu-item-selected):not(.ant-menu-item-disabled):hover {
  background-color: rgba(64, 150, 255, 0.1) !important;
  color: #4096ff !important;
}
</style>
