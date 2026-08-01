<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Upload, Download, SwitchButton } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { useI18n } from 'vue-i18n'
import LangSwitch from '@/components/LangSwitch.vue'
import logo from '@/assets/logo.svg'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

// Single source of truth for the sidebar: the children of the `/` route.
// Only routes that declare `meta.titleKey` are shown — this filters out any
// future helpers without touching the menu code.
const ICONS = { Upload, Download } as const

interface MenuEntry {
  path: string
  titleKey: string
  icon: keyof typeof ICONS | undefined
}

const menus = computed<MenuEntry[]>(() => {
  const root = router.getRoutes().find((r) => r.path === '/')
  if (!root || !root.children) return []
  return root.children
    .filter((c) => !!c.meta?.titleKey)
    .map((c) => ({
      path: c.path,
      titleKey: String(c.meta?.titleKey ?? ''),
      icon: (c.meta?.icon as keyof typeof ICONS) ?? undefined
    }))
})

// el-menu uses the route path as the active value.
const activeMenu = computed(() => route.path)

async function handleLogout() {
  try {
    await ElMessageBox.confirm(
      t('layout.logoutConfirmPrompt'),
      t('layout.logoutConfirmTitle'),
      {
        type: 'warning',
        confirmButtonText: t('layout.logoutButton'),
        cancelButtonText: t('common.cancel')
      }
    )
  } catch {
    return // cancelled
  }
  userStore.logout()
  router.replace({ name: 'Login' })
}
</script>

<template>
  <el-container class="layout-root">
    <el-aside width="220px" class="layout-aside">
      <div class="logo">
        <img :src="logo" alt="" class="logo-icon" />
        <span>{{ t('brand.name') }}</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#001529"
        text-color="#c9d1d9"
        active-text-color="#ffffff"
      >
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon v-if="m.icon && ICONS[m.icon]">
            <component :is="ICONS[m.icon]" />
          </el-icon>
          <span>{{ t(m.titleKey) }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <div class="header-title">{{ t('console.title') }}</div>
        <div class="header-right">
          <LangSwitch class="header-lang" />
          <el-button :icon="SwitchButton" text @click="handleLogout">
            {{ t('layout.logout') }}
          </el-button>
        </div>
      </el-header>
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <component :is="Component" />
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-root {
  height: 100%;
}

.layout-aside {
  background-color: #001529;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.logo-icon {
  width: 28px;
  height: 28px;
}

.layout-aside :deep(.el-menu) {
  border-right: none;
}

.layout-header {
  background-color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e6e8eb;
  padding: 0 24px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-lang {
  color: #303133;
}

.layout-main {
  background-color: #f0f2f5;
  padding: 24px;
}
</style>
