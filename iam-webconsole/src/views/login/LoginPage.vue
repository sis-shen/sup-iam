<template>
  <div class="login-container">
    <el-card class="login-card" shadow="always">
      <div class="login-header">
        <h2>IAM 管理控制台</h2>
        <p class="login-subtitle">身份与访问管理系统</p>
      </div>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="0"
        size="large"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名 / 邮箱 / 手机号"
            :prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { login as loginApi } from '@/api/auth'
import { setTokens } from '@/utils/auth'

const router = useRouter()

const formRef = ref(null)
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [
    { required: true, message: '请输入用户名/邮箱/手机号', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
}

async function handleLogin() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const res = await loginApi({
      username: form.username,
      password: form.password,
    })
    setTokens(res.access_token, res.refresh_token, res.expires_in)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (err) {
    // Error message already shown by interceptor
    console.error('Login failed:', err)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-card {
  width: 400px;
  border-radius: 12px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h2 {
  margin: 0 0 8px;
  font-size: 24px;
  color: #303133;
}

.login-subtitle {
  margin: 0;
  font-size: 14px;
  color: #909399;
}
</style>
