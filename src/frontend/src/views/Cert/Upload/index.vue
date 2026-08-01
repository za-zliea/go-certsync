<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadRawFile, UploadUserFile, FormInstance, FormRules } from 'element-plus'
import { listCerts, uploadCert, type CertListItem } from '@/api/cert'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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

// Reactive rules so validation messages follow the active locale.
const rules = computed<FormRules>(() => ({
  alias: [{ required: true, message: t('cert.upload.ruleAlias'), trigger: 'change' }],
  fullchain: [
    {
      validator: (_r, value, cb) => {
        if (!value || value.length === 0) cb(new Error(t('cert.upload.ruleFullchain')))
        else cb()
      },
      trigger: 'change'
    }
  ],
  privkey: [
    {
      validator: (_r, value, cb) => {
        if (!value || value.length === 0) cb(new Error(t('cert.upload.rulePrivkey')))
        else cb()
      },
      trigger: 'change'
    }
  ]
}))

// Restrict uploads to a single PEM file per field; reject extras gently.
function handleExceed() {
  ElMessage.warning(t('cert.upload.warnMultipleFile'))
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
    ElMessage.success(t('cert.upload.success'))
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
        <span class="card-title">{{ t('cert.upload.title') }}</span>
      </template>

      <el-alert
        v-if="!loading && uploadableCerts.length === 0"
        type="info"
        :closable="false"
        show-icon
        :title="t('cert.upload.alertTitle')"
        :description="t('cert.upload.alertDesc')"
      />

      <el-form
        v-else
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 640px"
      >
        <el-form-item :label="t('cert.upload.alias')" prop="alias">
          <el-select
            v-model="form.alias"
            :placeholder="t('cert.upload.placeholderAlias')"
            style="width: 100%"
          >
            <el-option
              v-for="c in uploadableCerts"
              :key="c.alias"
              :label="c.alias"
              :value="c.alias"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('cert.upload.fullchain')" prop="fullchain">
          <el-upload
            v-model:file-list="form.fullchain"
            :limit="1"
            :auto-upload="false"
            :on-exceed="handleExceed"
            accept=".pem,.crt,.cer,.txt"
          >
            <el-button type="primary" plain>{{ t('cert.upload.selectFile') }}</el-button>
            <template #tip>
              <div class="form-tip">{{ t('cert.upload.fullchainTip') }}</div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item :label="t('cert.upload.privkey')" prop="privkey">
          <el-upload
            v-model:file-list="form.privkey"
            :limit="1"
            :auto-upload="false"
            :on-exceed="handleExceed"
            accept=".pem,.key,.txt"
          >
            <el-button type="primary" plain>{{ t('cert.upload.selectFile') }}</el-button>
            <template #tip>
              <div class="form-tip">{{ t('cert.upload.privkeyTip') }}</div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ t('cert.upload.submit') }}
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
