<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { useI18n } from 'vue-i18n'
import LangSwitch from '@/components/LangSwitch.vue'
import logo from '@/assets/logo.svg'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

// Reactive rules so the validation messages follow the active locale.
const rules = computed<FormRules>(() => ({
  username: [{ required: true, message: t('login.ruleUsername'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.rulePassword'), trigger: 'blur' }]
}))

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return // validation errors are shown by the form
  }

  loading.value = true
  try {
    await userStore.login({ username: form.username, password: form.password })
    ElMessage.success(t('login.success'))
    const redirect = (route.query.redirect as string) || '/cert/download'
    router.replace(redirect)
  } catch {
    ElMessage.error(t('login.failed'))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <LangSwitch class="login-lang" />
    <div class="login-card">
      <div class="login-header">
        <img :src="logo" alt="" class="login-logo" />
        <h1>{{ t('brand.name') }}</h1>
        <p>{{ t('console.title') }}</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        size="large"
        label-position="top"
        @submit.prevent="handleSubmit"
      >
        <el-form-item :label="t('login.username')" prop="username">
          <el-input
            v-model="form.username"
            :placeholder="t('login.placeholderUsername')"
            :prefix-icon="User"
            autocomplete="username"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <el-form-item :label="t('login.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="t('login.placeholderPassword')"
            :prefix-icon="Lock"
            autocomplete="current-password"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            class="login-submit"
            :loading="loading"
            @click="handleSubmit"
          >
            {{ t('login.button') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  position: relative;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f2a44 0%, #0b1020 100%);
}

.login-lang {
  position: absolute;
  top: 20px;
  right: 24px;
  color: #ffffff;
}

.login-card {
  width: 380px;
  padding: 40px 36px 28px;
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.25);
}

.login-header {
  text-align: center;
  margin-bottom: 28px;
}

.login-logo {
  width: 48px;
  height: 48px;
  margin-bottom: 8px;
}

.login-header h1 {
  margin: 0 0 6px;
  font-size: 26px;
  letter-spacing: 2px;
  color: #1f2a44;
}

.login-header p {
  margin: 0;
  font-size: 13px;
  color: #909399;
}

.login-submit {
  width: 100%;
}
</style>
