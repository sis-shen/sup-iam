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
        <div class="register-link">
          还没有账号？
          <el-link type="primary" :underline="false" @click="registerVisible = true">
            立即注册
          </el-link>
        </div>
      </el-form>
    </el-card>

    <!-- 注册弹窗 -->
    <el-dialog v-model="registerVisible" title="注册账号" width="420px" :close-on-click-modal="false">
      <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules" label-width="80px" label-position="left">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="registerForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="registerForm.password" type="password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="registerForm.email" placeholder="请输入邮箱（可选）" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="registerForm.phone" placeholder="请输入手机号（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="registerVisible = false">取消</el-button>
        <el-button type="primary" :loading="registering" @click="handleRegister">注册</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { login as loginApi, register as registerApi } from '@/api/auth'
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

// 注册相关
const registerVisible = ref(false)
const registering = ref(false)
const registerFormRef = ref(null)

const registerForm = reactive({
  username: '',
  password: '',
  email: '',
  phone: '',
})

const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 32, message: '用户名长度在3-32位之间', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
}

async function handleRegister() {
  if (!registerFormRef.value) return
  try {
    await registerFormRef.value.validate()
  } catch {
    return
  }

  registering.value = true
  try {
    await registerApi({
      username: registerForm.username,
      password: registerForm.password,
      email: registerForm.email || undefined,
      phone: registerForm.phone || undefined,
    })
    ElMessage.success('注册成功，请登录')
    registerVisible.value = false
    form.username = registerForm.username
    form.password = ''
    // Reset register form
    registerForm.username = ''
    registerForm.password = ''
    registerForm.email = ''
    registerForm.phone = ''
  } catch {
    // Error shown by interceptor
  } finally {
    registering.value = false
  }
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
  } catch {
    // Error message already shown by interceptor
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

.register-link {
  text-align: center;
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}
</style>
