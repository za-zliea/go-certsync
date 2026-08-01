import { defineStore } from 'pinia'
import { ref } from 'vue'
import i18n, { type AppLocale } from '@/locales'

// localStorage key for the chosen UI locale. Shared with locales/index.ts so
// the i18n bootstrap can read it before Pinia is mounted.
export const LOCALE_KEY = 'certsync_locale'

export const useSettingsStore = defineStore('settings', () => {
  const locale = ref<AppLocale>(i18n.global.locale.value as AppLocale)

  function setLocale(l: AppLocale) {
    locale.value = l
    i18n.global.locale.value = l
    localStorage.setItem(LOCALE_KEY, l)
    // Keep the <html lang="..."> attribute in sync for accessibility / SEO.
    document.documentElement.lang = l === 'zh' ? 'zh-CN' : 'en'
  }

  return { locale, setLocale }
})
