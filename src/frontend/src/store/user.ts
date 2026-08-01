import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { loginApi, type LoginData, type LoginResult } from '@/api/login'

// localStorage key for the JWT. Shared with utils/request.ts so the axios
// interceptor can read the token without importing the store (avoids a
// circular dependency: request ← store ← api ← request).
export const TOKEN_KEY = 'certsync_token'
const EXPIRES_KEY = 'certsync_token_expires_at'

interface PersistedLogin {
  token: string
  expiresAt: number
}

function loadPersisted(): PersistedLogin | null {
  const token = localStorage.getItem(TOKEN_KEY)
  if (!token) return null
  const expiresAt = Number(localStorage.getItem(EXPIRES_KEY) ?? 0)
  // Treat expired tokens as absent.
  if (expiresAt && Date.now() > expiresAt * 1000) {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(EXPIRES_KEY)
    return null
  }
  return { token, expiresAt }
}

export const useUserStore = defineStore('user', () => {
  const persisted = loadPersisted()
  const token = ref<string>(persisted?.token ?? '')
  const expiresAt = ref<number>(persisted?.expiresAt ?? 0)

  const isLogin = computed(() => !!token.value)

  async function login(data: LoginData): Promise<LoginResult> {
    const result = await loginApi(data)
    token.value = result.token
    expiresAt.value = result.expires_at
    localStorage.setItem(TOKEN_KEY, result.token)
    localStorage.setItem(EXPIRES_KEY, String(result.expires_at))
    return result
  }

  function logout(): void {
    token.value = ''
    expiresAt.value = 0
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(EXPIRES_KEY)
  }

  return { token, expiresAt, isLogin, login, logout }
})
