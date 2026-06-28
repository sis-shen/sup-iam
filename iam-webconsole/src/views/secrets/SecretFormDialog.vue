<template>
  <el-dialog
    :model-value="visible"
    :title="mode === 'create' ? '创建密钥' : '编辑密钥'"
    width="500px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @open="handleOpen"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="left">
      <el-form-item label="描述" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入密钥描述" />
      </el-form-item>
      <el-form-item label="用户名" prop="username" v-if="mode === 'create'">
        <el-input v-model="form.username" placeholder="关联的用户名" />
      </el-form-item>
      <el-form-item label="过期时间" prop="expires">
        <el-date-picker
          v-model="form.expires"
          type="datetime"
          placeholder="选择过期时间"
          value-format="x"
          style="width: 100%"
        />
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
import { createSecret, updateSecret } from '@/api/secrets'

const props = defineProps({
  visible: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  secret: { type: Object, default: null },
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  description: '',
  username: '',
  expires: null,
})

const rules = {
  description: [{ max: 255, message: '描述不能超过255个字符', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  expires: [{ required: true, message: '请选择过期时间', trigger: 'change' }],
}

watch(() => props.visible, (val) => {
  if (!val) {
    resetForm()
  }
})

function resetForm() {
  form.description = ''
  form.username = ''
  form.expires = null
}

function handleOpen() {
  if (props.mode === 'edit' && props.secret) {
    form.description = props.secret.description || ''
    form.username = props.secret.username || ''
    // API returns expires as seconds (int64), date-picker needs milliseconds
    form.expires = props.secret.expires ? props.secret.expires * 1000 : null
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
    // expires: date-picker returns ms timestamp, API expects seconds (int64)
    const expiresSec = Math.floor(form.expires / 1000)
    if (props.mode === 'create') {
      const res = await createSecret({
        description: form.description || undefined,
        expires: expiresSec,
        user_name: form.username,
      })
      ElMessage.success('密钥创建成功')
      emit('update:visible', false)
      emit('success', res)
    } else {
      await updateSecret(props.secret.id, {
        description: form.description || undefined,
        expires: expiresSec,
      })
      ElMessage.success('密钥更新成功')
      emit('update:visible', false)
      emit('success')
    }
  } catch {
    // Error shown by interceptor
  } finally {
    submitting.value = false
  }
}
</script>
