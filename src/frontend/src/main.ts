import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import App from './App.vue'
import router from './router'
import i18n from '@/locales'
import './styles/index.css'

const app = createApp(App)

// Pinia must be installed before the router guard runs (the guard reads the
// user store), so register it first.
app.use(createPinia())
app.use(i18n)
app.use(router)
// Element Plus startup locale — runtime locale switching is handled by the
// <el-config-provider> in App.vue, which overrides this default.
app.use(ElementPlus, { locale: zhCn })

app.mount('#app')
