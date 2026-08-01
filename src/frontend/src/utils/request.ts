import axios, {
  type AxiosInstance,
  type AxiosRequestConfig,
  type AxiosResponse,
  type InternalAxiosRequestConfig
} from 'axios'
import { ElMessage } from 'element-plus'
import { TOKEN_KEY } from '@/store/user'
import { t } from '@/locales'

// When `skipErrorHandler: true` is set on a request, the response error
// interceptor bypasses its global handling (no ElMessage, no auto-redirect
// on 401) and simply rejects. Used by the login call so a 401 from
// `/front/pub/login` is reported by the caller with a login-specific
// message instead of "session expired".
declare module 'axios' {
  export interface AxiosRequestConfig {
    skipErrorHandler?: boolean
  }
}

// Server-side ResponseDTO (see src/core/server/base.go):
//   { code: 0|1, message: string, data: any }
// `code === 0` (SUCCESS) means the business-level request succeeded.
// We unwrap `data` on success and reject with `message` otherwise.
export interface ResponseDTO<T = unknown> {
  code: 0 | 1
  message: string
  data: T
}

const service: AxiosInstance = axios.create({
  // Same-origin: the SPA is served from the Go server's root, so all API
  // requests go to the same host. In dev, vite proxies /front + /api.
  baseURL: '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor: attach JWT if present.
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      config.headers.set('Authorization', `Bearer ${token}`)
    }
    return config
  },
  (error) => Promise.reject(error)
)

// On 401/unauthorized, clear token once and bounce to the login page.
// We use hash-mode routing, so the login URL is `/#/login`.
function handleUnauthorized() {
  localStorage.removeItem(TOKEN_KEY)
  if (!location.hash.startsWith('#/login')) {
    location.hash = '#/login'
  }
}

// Extract a human-readable message from an axios error response. Returns the
// server-provided envelope `message` when present, otherwise a status hint.
function extractErrorMessage(response: AxiosResponse | undefined): string {
  if (!response) return t('request.failed')
  const body = response.data
  if (body && typeof body === 'object' && 'message' in body) {
    const m = (body as ResponseDTO).message
    if (typeof m === 'string' && m) return m
  }
  return t('request.failedWithStatus', { status: response.status })
}

// Response interceptor: unwrap the standard envelope.
//
// Note on typing: axios's `interceptors.response.use` expects the fulfilled
// handler to return an `AxiosResponse`. Our envelope-unwrapping returns the
// bare payload instead, so we deliberately type the handler parameters as
// `unknown` and cast through `unknown` — this is the documented escape hatch.
// The caller-facing `request<T>` helper below is where the generic surface
// lives, so end-to-end typing is preserved for callers.
service.interceptors.response.use(
  (response: AxiosResponse) => {
    // Binary/streaming responses (file downloads) bypass the envelope.
    if (response.config.responseType === 'blob') {
      return response
    }

    const body = response.data as ResponseDTO
    // Non-enveloped edge responses — pass through untouched.
    if (body === null || typeof body !== 'object' || !('code' in body)) {
      return response
    }

    if (body.code === 0) {
      // Returning `body.data` here means callers see the unwrapped payload.
      // We cast through `unknown` because axios's typing insists the
      // interceptor return an `AxiosResponse`; at runtime axios forwards
      // whatever value we return, so callers see the unwrapped `data`.
      return body.data as unknown as AxiosResponse
    }

    // Business error: surface the server message and reject.
    const msg = body.message || t('request.failed')
    ElMessage.error(msg)
    return Promise.reject(new Error(msg))
  },
  (error) => {
    if (error.config?.skipErrorHandler) {
      return Promise.reject(error)
    }
    if (error.response?.status === 401) {
      handleUnauthorized()
      ElMessage.error(t('request.expired'))
    } else if (error.response) {
      ElMessage.error(extractErrorMessage(error.response))
    } else if (error.request) {
      ElMessage.error(t('request.network'))
    } else {
      ElMessage.error(error.message || t('request.unknown'))
    }
    return Promise.reject(error)
  }
)

// Typed request helper. Because the response interceptor unwraps the
// envelope, the promise resolves with the inner `data` payload typed as `T`.
//
// For binary downloads, use `service.get(url, { responseType: 'blob' })`
// (re-exported as default) instead — that path bypasses unwrapping.
export function request<T = unknown>(config: AxiosRequestConfig): Promise<T> {
  return service.request(config) as unknown as Promise<T>
}

export default service
