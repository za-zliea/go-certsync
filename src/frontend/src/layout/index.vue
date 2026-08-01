<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Upload, Download, SwitchButton } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// Single source of truth for the sidebar: the children of the `/` route.
// Only routes that declare `meta.title` are shown — this filters out any
// future helpers without touching the menu code.
const ICONS = { Upload, Download } as const

interface MenuEntry {
  path: string
  title: string
  icon: keyof typeof ICONS | undefined
}

const menus = computed<MenuEntry[]>(() => {
  const root = router.getRoutes().find((r) => r.path === '/')
  if (!root || !root.children) return []
  return root.children
    .filter((c) => !!c.meta?.title)
    .map((c) => ({
      path: c.path,
      title: String(c.meta?.title ?? ''),
      icon: (c.meta?.icon as keyof typeof ICONS) ?? undefined
    }))
})

// el-menu uses the route path as the active value.
const activeMenu = computed(() => route.path)

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消'
    })
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
        <span>CertSync</span>
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
          <span>{{ m.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <div class="header-title">证书管理后台</div>
        <div class="header-right">
          <el-button :icon="SwitchButton" text @click="handleLogout">退出登录</el-button>
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
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
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

.layout-main {
  background-color: #f0f2f5;
  padding: 24px;
}
</style>
