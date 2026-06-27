<template>
  <el-dialog
    :model-value="visible"
    :title="mode === 'create' ? '创建用户' : '编辑用户'"
    width="500px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @open="handleOpen"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="left">
      <el-form-item label="用户名" prop="username" v-if="mode === 'create'">
        <el-input v-model="form.username" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item label="密码" prop="password" v-if="mode === 'create'">
        <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
      </el-form-item>
      <el-form-item label="昵称" prop="nickname">
        <el-input v-model="form.nickname" placeholder="请输入昵称" />
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model="form.email" placeholder="请输入邮箱" />
      </el-form-item>
      <el-form-item label="手机号" prop="phone">
        <el-input v-model="form.phone" placeholder="请输入手机号" />
      </el-form-item>
      <el-form-item label="管理员">
        <el-switch v-model="form.is_admin" />
      </el-form-item>
      <el-form-item label="启用">
        <el-switch v-model="form.is_enable" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ mode === 'create' ? '创建' : '保存' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createUser, updateUser } from '@/api/users'

const props = defineProps({
  visible: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  user: { type: Object, default: null },
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  is_admin: false,
  is_enable: true,
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 32, message: '用户名长度在3-32位之间', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
  nickname: [{ max: 64, message: '昵称长度不能超过64位', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }],
}

watch(() => props.visible, (val) => {
  if (!val) {
    resetForm()
  }
})

function resetForm() {
  form.username = ''
  form.password = ''
  form.nickname = ''
  form.email = ''
  form.phone = ''
  form.is_admin = false
  form.is_enable = true
}

function handleOpen() {
  if (props.mode === 'edit' && props.user) {
    form.username = props.user.username || ''
    form.nickname = props.user.nickname || ''
    form.email = props.user.email || ''
    form.phone = props.user.phone || ''
    form.is_admin = !!props.user.is_admin
    form.is_enable = props.user.is_enable !== false
  } else {
    resetForm()
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    if (props.mode === 'create') {
      await createUser({
        username: form.username,
        password: form.password,
        nickname: form.nickname || undefined,
        email: form.email || undefined,
        phone: form.phone || undefined,
        is_admin: form.is_admin,
        is_enable: form.is_enable,
      })
      ElMessage.success('用户创建成功')
    } else {
      await updateUser(props.user.id, {
        nickname: form.nickname || undefined,
        email: form.email || undefined,
        phone: form.phone || undefined,
        is_admin: form.is_admin,
        is_enable: form.is_enable,
      })
      ElMessage.success('用户更新成功')
    }
    emit('update:visible', false)
    emit('success')
  } catch {
    // Error shown by interceptor
  } finally {
    submitting.value = false
  }
}
</script>
