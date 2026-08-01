import { request } from '@/utils/request'

// Matches src/core/server/auth.go LoginResponse.
export interface LoginResult {
  token: string
  // Unix seconds.
  expires_at: number
}

export interface LoginData {
  username: string
  password: string
}

// POST /front/pub/login → { token, expires_at }
// skipErrorHandler: a wrong password yields HTTP 401, which the global
// interceptor would otherwise turn into "session expired" + a forced
// redirect. The caller surfaces a login-specific message instead.
export function loginApi(data: LoginData): Promise<LoginResult> {
  return request<LoginResult>({
    url: '/front/pub/login',
    method: 'POST',
    data,
    skipErrorHandler: true
  })
}
