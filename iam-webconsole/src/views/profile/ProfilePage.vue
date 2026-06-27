<template>
  <div class="page-container">
    <div class="page-header">
      <h2>个人中心</h2>
    </div>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card shadow="never" class="content-card">
          <template #header>
            <span>个人信息</span>
          </template>
          <div class="detail-row" v-for="item in userFields" :key="item.label">
            <div class="detail-label">{{ item.label }}</div>
            <div class="detail-value">{{ item.value }}</div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="never" class="content-card">
          <template #header>
            <span>修改密码</span>
          </template>
          <el-form
            ref="formRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-width="120px"
            label-position="left"
          >
            <el-form-item label="当前密码" prop="old_password">
              <el-input
                v-model="passwordForm.old_password"
                type="password"
                show-password
              />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input
                v-model="passwordForm.new_password"
                type="password"
                show-password
              />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input
                v-model="passwordForm.confirm_password"
                type="password"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="changing" @click="handleChangePassword">
                修改密码
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getMe, changePassword } from '@/api/auth'
import { formatDateTime } from '@/utils/format'

const formRef = ref(null)
const changing = ref(false)
const userInfo = ref(null)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const passwordRules = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.new_password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

const userFields = computed(() => [
  { label: '用户ID', value: userInfo.value?.id ?? '-' },
  { label: '用户名', value: userInfo.value?.username ?? '-' },
  { label: '昵称', value: userInfo.value?.nickname || '-' },
  { label: '邮箱', value: userInfo.value?.email || '-' },
  { label: '手机号', value: userInfo.value?.phone || '-' },
  { label: '角色', value: userInfo.value?.is_admin ? '管理员' : '普通用户' },
  { label: '状态', value: userInfo.value?.is_enable ? '启用' : '禁用' },
  { label: '登录时间', value: formatDateTime(userInfo.value?.logged_at) },
])

onMounted(async () => {
  try {
    const res = await getMe()
    userInfo.value = res
  } catch {
    // Ignore
  }
})

async function handleChangePassword() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  changing.value = true
  try {
    await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    ElMessage.success('密码修改成功')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
  } catch {
    // Error already shown by interceptor
  } finally {
    changing.value = false
  }
}
</script>

<style scoped>
.content-card {
  margin-bottom: 20px;
}
</style>
