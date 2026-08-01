<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { View, Download as DownloadIcon } from '@element-plus/icons-vue'
import {
  listCerts,
  certInfo,
  downloadCert,
  type CertListItem,
  type CertInfo
} from '@/api/cert'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const list = ref<CertListItem[]>([])
const loading = ref(false)

// Detail dialog state.
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<CertInfo | null>(null)

// Track which row is busy (downloading / fetching detail) so we can show a
// per-button loading indicator instead of a global overlay.
const busyAlias = ref<string>('')

function providerTagType(provider: string) {
  // Stable colors per provider — purely cosmetic.
  const map: Record<string, 'success' | 'warning' | 'info' | 'primary' | 'danger'> = {
    TENCENT: 'success',
    ALIYUN: 'warning',
    GODADDY: 'info',
    GOOGLE: 'primary',
    CLOUDFLARE: 'danger'
  }
  return map[provider] ?? 'info'
}

async function fetchList() {
  loading.value = true
  try {
    list.value = await listCerts()
  } catch {
    // surfaced by interceptor
  } finally {
    loading.value = false
  }
}

function formatTime(rfc3339: string): string {
  if (!rfc3339) return '-'
  const d = new Date(rfc3339)
  if (Number.isNaN(d.getTime())) return rfc3339
  if (d.getUTCFullYear() <= 1) return '-' // Go time.Time zero value (local cert missing)
  // Use locale-aware formatting (YYYY-MM-DD HH:mm:ss in zh-CN).
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// Local cert missing signal: when absent, backend returns
// not_after=0001-01-01T00:00:00Z and days_remaining=0.
function isLocalMissing(info: CertInfo): boolean {
  if (!info.not_after) return true
  const d = new Date(info.not_after)
  return !Number.isNaN(d.getTime()) && d.getUTCFullYear() <= 1
}

// Date-only formatting for `expected_renew_at` (the scheduler's next renewal
// tick; time-of-day is an implementation detail users don't need).
function formatDate(rfc3339: string): string {
  if (!rfc3339) return '-'
  const d = new Date(rfc3339)
  if (Number.isNaN(d.getTime())) return rfc3339
  if (d.getUTCFullYear() <= 1) return '-'
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

async function handleDetail(row: CertListItem) {
  busyAlias.value = row.alias
  detailLoading.value = true
  detailVisible.value = true
  detail.value = null
  try {
    detail.value = await certInfo(row.alias)
  } catch {
    // interceptor shows the message; keep the dialog open but empty.
  } finally {
    detailLoading.value = false
    busyAlias.value = ''
  }
}

async function handleDownload(row: CertListItem) {
  busyAlias.value = row.alias
  try {
    const blob = await downloadCert(row.alias)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${row.alias}.zip`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success(t('cert.download.startDownload'))
  } catch {
    // surfaced by interceptor / downloadCert
  } finally {
    busyAlias.value = ''
  }
}

onMounted(fetchList)
</script>

<template>
  <div v-loading="loading" class="download-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="card-title">{{ t('cert.download.title') }}</span>
          <el-button text type="primary" @click="fetchList">
            {{ t('cert.download.refresh') }}
          </el-button>
        </div>
      </template>

      <el-table :data="list" border stripe :empty-text="t('cert.download.empty')">
        <el-table-column prop="alias" :label="t('cert.download.columnAlias')" min-width="180" />
        <el-table-column prop="domain" :label="t('cert.download.columnDomain')" min-width="180" />
        <el-table-column :label="t('cert.download.columnProvider')" width="130">
          <template #default="{ row }">
            <el-tag :type="providerTagType(row.provider)" effect="plain">
              {{ row.provider || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('cert.download.columnRenewMode')" width="110">
          <template #default="{ row }">
            <el-tag :type="row.auto_renew ? 'success' : 'info'" effect="dark">
              {{ row.auto_renew ? t('cert.download.tagAuto') : t('cert.download.tagManual') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('cert.download.columnNotBefore')" min-width="180">
          <template #default="{ row }">{{ formatTime(row.not_before) }}</template>
        </el-table-column>
        <el-table-column :label="t('cert.download.columnNotAfter')" min-width="180">
          <template #default="{ row }">{{ formatTime(row.not_after) }}</template>
        </el-table-column>
        <el-table-column :label="t('cert.download.columnDaysRemaining')" width="120">
          <template #default="{ row }">
            <template v-if="row.days_remaining == null">-</template>
            <el-tag v-else-if="row.days_remaining < 0" type="danger" effect="dark">
              {{ t('cert.download.tagExpired') }}
            </el-tag>
            <el-tag v-else-if="row.days_remaining < 10" type="danger" effect="plain">
              {{ row.days_remaining }}
            </el-tag>
            <el-tag v-else-if="row.days_remaining < 30" type="warning" effect="plain">
              {{ row.days_remaining }}
            </el-tag>
            <span v-else>{{ row.days_remaining }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('cert.download.columnExpectedRenewAt')" min-width="180">
          <template #default="{ row }">{{ formatDate(row.expected_renew_at) }}</template>
        </el-table-column>
        <el-table-column :label="t('cert.download.columnOperation')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              size="small"
              :icon="View"
              :loading="busyAlias === row.alias && detailLoading"
              @click="handleDetail(row)"
            >
              {{ t('cert.download.btnDetail') }}
            </el-button>
            <el-button
              size="small"
              type="primary"
              :icon="DownloadIcon"
              :loading="busyAlias === row.alias && !detailLoading"
              :disabled="!row.alias"
              @click="handleDownload(row)"
            >
              {{ t('cert.download.btnDownload') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="detailVisible"
      :title="t('cert.download.detail.title')"
      width="560px"
      destroy-on-close
    >
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="1" border>
          <el-descriptions-item :label="t('cert.download.detail.alias')">
            {{ detail.alias }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.cn')">
            {{ detail.cn || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.san')">
            <div v-if="detail.dns_names && detail.dns_names.length">
              <el-tag
                v-for="name in detail.dns_names"
                :key="name"
                type="info"
                effect="plain"
                style="margin: 2px 4px 2px 0"
              >
                {{ name }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.notBefore')">
            {{ formatTime(detail.not_before) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.notAfter')">
            {{ formatTime(detail.not_after) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.daysRemaining')">
            <template v-if="isLocalMissing(detail)">
              -
            </template>
            <el-tag
              v-else-if="detail.days_remaining < 0"
              type="danger"
              effect="dark"
            >{{ t('cert.download.detail.expired') }}</el-tag>
            <span v-else>{{ detail.days_remaining }}</span>
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.expectedRenewAt')">
            {{ detail.expected_renew_at ? formatDate(detail.expected_renew_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.remoteNotBefore')">
            {{ detail.remote_not_before ? formatTime(detail.remote_not_before) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.remoteNotAfter')">
            {{ detail.remote_not_after ? formatTime(detail.remote_not_after) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('cert.download.detail.remoteDaysRemaining')">
            <template v-if="detail.remote_days_remaining == null">-</template>
            <el-tag v-else-if="detail.remote_days_remaining < 0" type="danger" effect="dark">
              {{ t('cert.download.detail.expired') }}
            </el-tag>
            <span v-else>{{ detail.remote_days_remaining }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <el-empty
          v-else-if="!detailLoading"
          :description="t('cert.download.detail.empty')"
        />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.download-page {
  width: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-title {
  font-weight: 600;
}
</style>
