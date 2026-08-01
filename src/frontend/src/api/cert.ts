import service, { request, type ResponseDTO } from '@/utils/request'
import { t } from '@/locales'

// Matches src/core/server/front.go CertListItem.
// `not_before` / `not_after` / `days_remaining` / `expected_renew_at` are
// populated only when a local cert exists; otherwise undefined → table
// renders `-`.
export interface CertListItem {
  alias: string
  email: string
  domain: string
  domain_cn: string[]
  domain_check: string
  provider: string
  auto_renew: boolean
  disable_cname?: boolean | null
  not_before?: string
  not_after?: string
  days_remaining?: number
  expected_renew_at?: string
}

// Matches src/core/server/front.go CertInfo.
// `not_before` / `not_after` are RFC3339 time strings. When the local cert is
// missing, the backend still emits `not_before` / `not_after` as Go's
// `0001-01-01T00:00:00Z` zero value (rendered as `-` by formatTime); the
// `cn` / `dns_names` fields are then backfilled from the remote cert.
export interface CertInfo {
  alias: string
  cn: string
  dns_names: string[]
  not_before: string
  not_after: string
  days_remaining: number
  expected_renew_at: string | null
  // Remote cert info. Present only when CertConfig.DomainCheck is configured
  // AND the TLS dial succeeds; undefined otherwise → render as '-'.
  remote_not_before?: string
  remote_not_after?: string
  remote_days_remaining?: number
}

// GET /front/api/list → CertListItem[]
export function listCerts(): Promise<CertListItem[]> {
  return request<CertListItem[]>({
    url: '/front/api/list',
    method: 'GET'
  })
}

// GET /front/api/{alias}/info → CertInfo
export function certInfo(alias: string): Promise<CertInfo> {
  return request<CertInfo>({
    url: `/front/api/${encodeURIComponent(alias)}/info`,
    method: 'GET'
  })
}

// POST /front/api/{alias}/upload
// Multipart fields MUST be named `fullchain` and `privkey` (see FrontUploadHandler).
export function uploadCert(alias: string, form: FormData): Promise<void> {
  return request<void>({
    url: `/front/api/${encodeURIComponent(alias)}/upload`,
    method: 'POST',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// GET /front/api/{alias}/download → binary ZIP.
//
// The backend's error path for download is unusual: it returns HTTP 200 with
// a JSON envelope body (`{code:1, message:...}`) instead of a ZIP when the
// cert files are incomplete or unreadable. Since we requested `responseType:
// 'blob'`, axios bypasses the JSON-envelope unwrap, so a naive caller would
// save the JSON bytes as `alias.zip`. We detect that here by inspecting the
// response Content-Type and re-parsing as an envelope, then reject with the
// server's `message`. Resolves with a Blob that is genuinely a ZIP archive.
export async function downloadCert(alias: string): Promise<Blob> {
  const resp = await service.get<Blob>(
    `/front/api/${encodeURIComponent(alias)}/download`,
    { responseType: 'blob' }
  )
  const contentType = String(resp.headers?.['content-type'] ?? '')
  if (contentType.includes('application/json')) {
    // Server returned an error envelope disguised as a 200.
    const text = await resp.data.text()
    let message = t('cert.download.invalidPackage')
    try {
      const env = JSON.parse(text) as ResponseDTO
      if (env && typeof env.message === 'string' && env.message) message = env.message
    } catch {
      // keep default message
    }
    throw new Error(message)
  }
  return resp.data
}
