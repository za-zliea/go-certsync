import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/store/user'

// Hash mode: URLs look like `/#/cert/upload`. This requires zero server-side
// cooperation (the Go server only ever serves index.html at `/`), so it works
// flawlessly with the existing atreugo.StaticCustom setup.
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login/index.vue'),
      meta: { title: '登录', requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('@/layout/index.vue'),
      redirect: '/cert/download',
      children: [
        {
          path: '/cert/upload',
          name: 'CertUpload',
          component: () => import('@/views/Cert/Upload/index.vue'),
          meta: { title: '证书上传', icon: 'Upload', requiresAuth: true }
        },
        {
          path: '/cert/download',
          name: 'CertDownload',
          component: () => import('@/views/Cert/Download/index.vue'),
          meta: { title: '证书下载', icon: 'Download', requiresAuth: true }
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/cert/download'
    }
  ]
})

// Global guard: every auth-required route needs a token. Unauthed users are
// sent to /login with the original target as a `redirect` query.
router.beforeEach((to) => {
  const userStore = useUserStore()
  const requiresAuth = to.meta.requiresAuth !== false

  if (requiresAuth && !userStore.isLogin) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'Login' && userStore.isLogin) {
    return { path: '/cert/download' }
  }
  return true
})

export default router
