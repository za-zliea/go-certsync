<script setup lang="ts">
import { ArrowDown } from '@element-plus/icons-vue'
import { useSettingsStore } from '@/store/settings'
import type { AppLocale } from '@/locales'

const settings = useSettingsStore()

function onCommand(c: string) {
  if (c === 'zh' || c === 'en') settings.setLocale(c as AppLocale)
}
</script>

<template>
  <el-dropdown trigger="click" @command="onCommand">
    <span class="lang-trigger">
      {{ settings.locale === 'zh' ? '中文' : 'English' }}
      <el-icon class="lang-caret"><ArrowDown /></el-icon>
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item command="zh">中文</el-dropdown-item>
        <el-dropdown-item command="en">English</el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<style scoped>
.lang-trigger {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  color: inherit;
  outline: none;
  font-size: 14px;
}

.lang-caret {
  font-size: 12px;
}
</style>
