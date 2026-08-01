import { createI18n } from 'vue-i18n'
import zh from './zh'
import en from './en'

export type AppLocale = 'zh' | 'en'

// Persisted locale takes priority; otherwise fall back to the browser language.
function detectInitial(): AppLocale {
  const saved = localStorage.getItem('certsync_locale')
  if (saved === 'zh' || saved === 'en') return saved
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh' : 'en'
}

const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: detectInitial(),
  fallbackLocale: 'zh',
  messages: { zh, en }
})

// Expose the global `t` so non-component modules (utils/request.ts, api/*.ts)
// can translate without useI18n().
export const t = i18n.global.t

export default i18n
