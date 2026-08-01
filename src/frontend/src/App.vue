<script setup lang="ts">
import { computed, watch } from 'vue'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '@/store/settings'

const settings = useSettingsStore()
const { t } = useI18n()

// Element Plus runtime locale follows the user's selection.
const epLocale = computed(() => (settings.locale === 'zh' ? zhCn : en))

// The browser-tab title is fixed to the console title and intentionally
// does not change with the active route.
function applyTitle() {
  document.title = t('console.title')
}

// Sync <html lang> with the active locale (for accessibility / SEO).
function applyHtmlLang() {
  if (typeof document !== 'undefined') {
    document.documentElement.lang = settings.locale === 'zh' ? 'zh-CN' : 'en'
  }
}

watch(() => settings.locale, () => {
  applyHtmlLang()
  applyTitle()
})

applyHtmlLang()
applyTitle()
</script>

<template>
  <el-config-provider :locale="epLocale">
    <router-view />
  </el-config-provider>
</template>
