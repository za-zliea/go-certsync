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
  // Use locale-aware formatting (YYYY-MM-DD HH:mm:ss in zh-CN).
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
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
    ElMessage.success('开始下载')
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
          <span class="card-title">证书下载</span>
          <el-button text type="primary" @click="fetchList">刷新</el-button>
        </div>
      </template>

      <el-table :data="list" border stripe empty-text="暂无证书">
        <el-table-column prop="alias" label="别名" min-width="180" />
        <el-table-column prop="domain" label="主域名" min-width="180" />
        <el-table-column label="DNS 提供商" width="130">
          <template #default="{ row }">
            <el-tag :type="providerTagType(row.provider)" effect="plain">
              {{ row.provider || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="续期模式" width="110">
          <template #default="{ row }">
            <el-tag :type="row.auto_renew ? 'success' : 'info'" effect="dark">
              {{ row.auto_renew ? '自动' : '手动' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="domain_check" label="检查地址" min-width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              size="small"
              :icon="View"
              :loading="busyAlias === row.alias && detailLoading"
              @click="handleDetail(row)"
            >
              详情
            </el-button>
            <el-button
              size="small"
              type="primary"
              :icon="DownloadIcon"
              :loading="busyAlias === row.alias && !detailLoading"
              :disabled="!row.alias"
              @click="handleDownload(row)"
            >
              下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="detailVisible"
      title="证书详情"
      width="560px"
      destroy-on-close
    >
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="1" border>
          <el-descriptions-item label="别名">{{ detail.alias }}</el-descriptions-item>
          <el-descriptions-item label="通用名 (CN)">{{ detail.cn || '-' }}</el-descriptions-item>
          <el-descriptions-item label="SAN 域名">
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
          <el-descriptions-item label="生效时间">
            {{ formatTime(detail.not_before) }}
          </el-descriptions-item>
          <el-descriptions-item label="过期时间">
            {{ formatTime(detail.not_after) }}
          </el-descriptions-item>
        </el-descriptions>
        <el-empty v-else-if="!detailLoading" description="暂无数据" />
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
