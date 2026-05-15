<template>
  <a-layout class="layout" :class="{ mobile: appStore.hideMenu }">
    <div v-if="navbar" class="layout-navbar">
      <NavBar />
    </div>
    <a-layout>
      <a-layout>
        <a-layout-sider
          v-if="renderMenu"
          v-show="!hideMenu"
          class="layout-sider"
          breakpoint="xl"
          :collapsed="collapsed"
          :collapsible="true"
          :width="menuWidth"
          :style="{ paddingTop: navbar ? '60px' : '' }"
          :hide-trigger="true"
          @collapse="setCollapsed"
        >
          <div class="menu-wrapper">
            <div class="sider-brand" :class="{ collapsed }">
              <div class="brand-mark">观</div>
              <div v-if="!collapsed" class="brand-copy">
                <div class="brand-title">观潮</div>
                <div class="brand-subtitle">竞品分析工作台</div>
              </div>
            </div>
            <div class="menu-content">
              <Menu />
            </div>
          </div>
        </a-layout-sider>
        <a-drawer
          v-if="hideMenu"
          :visible="drawerVisible"
          placement="left"
          :footer="false"
          mask-closable
          :closable="false"
          @cancel="drawerCancel"
        >
          <Menu />
        </a-drawer>
        <a-layout class="layout-content" :style="paddingStyle">
          <TabBar v-if="appStore.tabBar" />
          <a-layout-content>
            <PageLayout />
          </a-layout-content>
          <Footer v-if="footer" />
        </a-layout>
      </a-layout>
    </a-layout>
    <GlobalAssistant ref="globalAssistantRef" />
  </a-layout>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, provide, onMounted } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { useAppStore, useUserStore } from '@/store';
  import NavBar from '@/components/navbar/index.vue';
  import Menu from '@/components/menu/index.vue';
  import Footer from '@/components/footer/index.vue';
  import TabBar from '@/components/tab-bar/index.vue';
  import GlobalAssistant from '@/components/global-assistant/index.vue';
  import usePermission from '@/hooks/permission';
  import useResponsive from '@/hooks/responsive';
  import PageLayout from './page-layout.vue';

  const isInit = ref(false);
  const appStore = useAppStore();
  const userStore = useUserStore();
  const router = useRouter();
  const route = useRoute();
  const permission = usePermission();
  useResponsive(true);
  const navbarHeight = `60px`;
  const navbar = computed(() => appStore.navbar);
  const renderMenu = computed(() => appStore.menu && !appStore.topMenu);
  const hideMenu = computed(() => appStore.hideMenu);
  const footer = computed(() => appStore.footer);
  const menuWidth = computed(() => {
    return appStore.menuCollapse ? 48 : appStore.menuWidth;
  });
  const collapsed = computed(() => {
    return appStore.menuCollapse;
  });
  const paddingStyle = computed(() => {
    const paddingLeft =
      renderMenu.value && !hideMenu.value
        ? { paddingLeft: `${menuWidth.value}px` }
        : {};
    const paddingTop = navbar.value ? { paddingTop: navbarHeight } : {};
    return { ...paddingLeft, ...paddingTop };
  });
  const setCollapsed = (val: boolean) => {
    if (!isInit.value) return; // for page initialization menu state problem
    appStore.updateSettings({ menuCollapse: val });
  };
  watch(
    () => userStore.role,
    (roleValue) => {
      if (roleValue && !permission.accessRouter(route))
        router.push({ name: 'notFound' });
    }
  );
  const drawerVisible = ref(false);
  const globalAssistantRef = ref<InstanceType<typeof GlobalAssistant> | null>(null);
  const drawerCancel = () => {
    drawerVisible.value = false;
  };
  provide('toggleDrawerMenu', () => {
    drawerVisible.value = !drawerVisible.value;
  });
  provide('openGlobalAssistant', () => {
    globalAssistantRef.value?.open();
  });
  onMounted(() => {
    isInit.value = true;
  });
</script>

<style scoped lang="less">
  @nav-size-height: 60px;
  @layout-max-width: 1100px;

  .layout {
    width: 100%;
    height: 100%;
  }

  .layout-navbar {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 100;
    width: 100%;
    height: @nav-size-height;
  }

  .layout-sider {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 99;
    height: 100%;
    background: linear-gradient(
      180deg,
      rgba(6, 20, 46, 0.98) 0%,
      rgba(9, 30, 67, 0.98) 48%,
      rgba(11, 49, 98, 0.98) 100%
    );
    box-shadow: 8px 0 28px rgba(3, 10, 24, 0.18);
    transition: all 0.2s cubic-bezier(0.34, 0.69, 0.1, 1);
    &::after {
      position: absolute;
      top: 0;
      right: -1px;
      display: block;
      width: 1px;
      height: 100%;
      background-color: var(--color-border);
      content: '';
    }

    > :deep(.arco-layout-sider-children) {
      overflow-y: hidden;
    }
  }

  .menu-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 18px 12px 14px;
    overflow: hidden;
  }

  .sider-brand {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
    padding: 14px 12px;
    border: 1px solid rgba(150, 211, 255, 0.14);
    border-radius: 16px;
    background: linear-gradient(
      180deg,
      rgba(255, 255, 255, 0.08),
      rgba(255, 255, 255, 0.02)
    );
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.06),
      0 8px 20px rgba(2, 9, 22, 0.18);
  }

  .sider-brand.collapsed {
    justify-content: center;
    padding: 12px 0;
  }

  .brand-mark {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 12px;
    background: linear-gradient(180deg, #61b8ff 0%, #2c7cff 100%);
    color: #f7fbff;
    font-size: 20px;
    font-weight: 700;
    box-shadow:
      0 8px 18px rgba(37, 123, 255, 0.28),
      0 0 22px rgba(97, 184, 255, 0.18);
  }

  .brand-copy {
    overflow: hidden;
  }

  .brand-title {
    color: rgba(247, 251, 255, 0.96);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 2px;
  }

  .brand-subtitle {
    margin-top: 4px;
    color: rgba(214, 232, 255, 0.54);
    font-size: 12px;
    letter-spacing: 1px;
  }

  .menu-content {
    flex: 1;
    overflow: auto;
    overflow-x: hidden;
    padding: 8px 4px 0;
    :deep(.arco-menu) {
      background: transparent;
      color: rgba(228, 239, 255, 0.72);

      ::-webkit-scrollbar {
        width: 12px;
        height: 4px;
      }

      ::-webkit-scrollbar-thumb {
        border: 4px solid transparent;
        background-clip: padding-box;
        border-radius: 7px;
        background-color: var(--color-text-4);
      }

      ::-webkit-scrollbar-thumb:hover {
        background-color: var(--color-text-3);
      }

      .arco-menu-inner {
        padding: 0;
      }

      .arco-menu-item,
      .arco-menu-inline-header {
        margin-bottom: 6px;
        border-radius: 12px;
      }

      .arco-menu-inline-header,
      .arco-menu-item {
        padding-left: 12px !important;
        padding-right: 12px !important;
        color: rgba(228, 239, 255, 0.76);
        transition: all 0.2s ease;
      }

      .arco-menu-item:hover,
      .arco-menu-inline-header:hover {
        background: rgba(255, 255, 255, 0.08);
        color: #f7fbff;
      }

      .arco-menu-selected,
      .arco-menu-selected:hover {
        background: linear-gradient(
          90deg,
          rgba(67, 150, 255, 0.28),
          rgba(92, 196, 255, 0.18)
        );
        color: #f7fbff;
        box-shadow:
          inset 0 0 0 1px rgba(163, 227, 255, 0.12),
          0 0 24px rgba(67, 150, 255, 0.12);
      }

      .arco-menu-icon {
        color: inherit;
      }

      .arco-menu-pop-header,
      .arco-menu-collapse-button {
        background: transparent;
      }
    }
  }

  .layout-content {
    min-height: 100vh;
    overflow-y: hidden;
    background-color: var(--color-fill-2);
    transition: padding 0.2s cubic-bezier(0.34, 0.69, 0.1, 1);
  }
</style>
