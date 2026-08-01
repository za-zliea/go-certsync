<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadRawFile, UploadUserFile, FormInstance, FormRules } from 'element-plus'
import { listCerts, uploadCert, type CertListItem } from '@/api/cert'

// Only `auto_renew: false` certs can be uploaded (FrontUploadHandler rejects
// the other case with 403). We fetch the full list once on mount and filter
// client-side.
const allCerts = ref<CertListItem[]>([])
const loading = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  alias: '' as string,
  fullchain: [] as UploadUserFile[],
  privkey: [] as UploadUserFile[]
})

const uploadableCerts = computed(() => allCerts.value.filter((c) => !c.auto_renew))

const rules: FormRules = {
  alias: [{ required: true, message: '请选择证书别名', trigger: 'change' }],
  fullchain: [
    {
      validator: (_r, value, cb) => {
        if (!value || value.length === 0) cb(new Error('请选择 fullchain 文件'))
        else cb()
      },
      trigger: 'change'
    }
  ],
  privkey: [
    {
      validator: (_r, value, cb) => {
        if (!value || value.length === 0) cb(new Error('请选择 privkey 文件'))
        else cb()
      },
      trigger: 'change'
    }
  ]
}

// Restrict uploads to a single PEM file per field; reject extras gently.
function handleExceed() {
  ElMessage.warning('每个位置只允许上传一个文件，请先删除已有文件')
}

function rawFile(file: UploadUserFile): File {
  // element-plus UploadFile exposes the raw File (we never enable upload-on-select,
  // so the raw file is what we want).
  return (file.raw as UploadRawFile) ?? (file as unknown as File)
}

async function fetchCerts() {
  loading.value = true
  try {
    allCerts.value = await listCerts()
    if (uploadableCerts.value.length > 0) {
      form.alias = uploadableCerts.value[0].alias
    }
  } catch {
    // The interceptor surfaces the error message.
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  const fd = new FormData()
  // Field names must match FrontUploadHandler exactly (`fullchain`, `privkey`).
  fd.append('fullchain', rawFile(form.fullchain[0]))
  fd.append('privkey', rawFile(form.privkey[0]))

  submitting.value = true
  try {
    await uploadCert(form.alias, fd)
    ElMessage.success('证书上传成功')
    // Reset the file inputs; keep the alias selected for convenience.
    form.fullchain = []
    form.privkey = []
  } catch {
    // handled by interceptor
  } finally {
    submitting.value = false
  }
}

onMounted(fetchCerts)
</script>

<template>
  <div v-loading="loading" class="upload-page">
    <el-card shadow="never">
      <template #header>
        <span class="card-title">证书上传</span>
      </template>

      <el-alert
        v-if="!loading && uploadableCerts.length === 0"
        type="info"
        :closable="false"
        show-icon
        title="当前没有可上传的证书"
        description="仅 auto_renew 为 false 的证书允许手动上传；auto_renew 为 true 的证书由 ACME 自动签发与续期。"
      />

      <el-form
        v-else
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 640px"
      >
        <el-form-item label="证书别名" prop="alias">
          <el-select v-model="form.alias" placeholder="请选择证书" style="width: 100%">
            <el-option
              v-for="c in uploadableCerts"
              :key="c.alias"
              :label="c.alias"
              :value="c.alias"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="Fullchain" prop="fullchain">
          <el-upload
            v-model:file-list="form.fullchain"
            :limit="1"
            :auto-upload="false"
            :on-exceed="handleExceed"
            accept=".pem,.crt,.cer,.txt"
          >
            <el-button type="primary" plain>选择文件</el-button>
            <template #tip>
              <div class="form-tip">完整证书链 PEM 文件（fullchain.pem）。</div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item label="Privkey" prop="privkey">
          <el-upload
            v-model:file-list="form.privkey"
            :limit="1"
            :auto-upload="false"
            :on-exceed="handleExceed"
            accept=".pem,.key,.txt"
          >
            <el-button type="primary" plain>选择文件</el-button>
            <template #tip>
              <div class="form-tip">私钥 PEM 文件（privkey.pem）。</div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            提交上传
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.upload-page {
  max-width: 960px;
}

.card-title {
  font-weight: 600;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}
</style>
