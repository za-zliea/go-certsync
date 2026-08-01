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
export function loginApi(data: LoginData): Promise<LoginResult> {
  return request<LoginResult>({
    url: '/front/pub/login',
    method: 'POST',
    data
  })
}
