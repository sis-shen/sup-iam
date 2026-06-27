<template>
  <el-dialog
    :model-value="visible"
    :title="mode === 'create' ? '创建策略' : '编辑策略'"
    width="650px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @open="handleOpen"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="left">
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入策略名称" />
      </el-form-item>
      <el-form-item label="描述" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入策略描述" />
      </el-form-item>
      <el-form-item label="用户名" prop="user_name" v-if="mode === 'create'">
        <el-input v-model="form.user_name" placeholder="关联的用户名" />
      </el-form-item>
      <el-form-item label="策略内容" prop="content">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="10"
          placeholder='请输入策略 JSON 内容，例如：{"statement": [{"effect": "allow", "action": ["*"], "resource": ["*"]}]}'
          style="font-family: 'Courier New', Courier, monospace; font-size: 13px;"
        />
        <div style="margin-top: 8px; text-align: right;">
          <el-button size="small" @click="formatContent">格式化 JSON</el-button>
          <el-button size="small" @click="showContentHelp = !showContentHelp">帮助</el-button>
        </div>
        <el-alert v-if="showContentHelp" type="info" :closable="false" show-icon style="margin-top: 8px">
          <template #default>
            <p style="margin: 4px 0;">策略内容为 JSON 格式，示例：</p>
            <pre style="background: #f5f7fa; padding: 8px; border-radius: 4px; font-size: 12px;">{
  "statement": [
    {
      "effect": "allow",
      "action": ["iam:*"],
      "resource": ["*"]
    }
  ]
}</pre>
          </template>
        </el-alert>
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
import { createPolicy, updatePolicy } from '@/api/policies'

const props = defineProps({
  visible: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  policy: { type: Object, default: null },
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitting = ref(false)
const showContentHelp = ref(false)

const form = reactive({
  name: '',
  description: '',
  user_name: '',
  content: '',
})

const rules = {
  name: [
    { required: true, message: '请输入策略名称', trigger: 'blur' },
    { max: 64, message: '名称不能超过64个字符', trigger: 'blur' },
  ],
  description: [{ max: 255, message: '描述不能超过255个字符', trigger: 'blur' }],
  user_name: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  content: [
    { required: true, message: '请输入策略内容', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value) return callback()
        try {
          JSON.parse(value)
          callback()
        } catch (e) {
          callback(new Error('JSON 格式不正确: ' + e.message))
        }
      },
      trigger: 'blur',
    },
  ],
}

watch(() => props.visible, (val) => {
  if (!val) {
    resetForm()
  }
})

function resetForm() {
  form.name = ''
  form.description = ''
  form.user_name = ''
  form.content = ''
}

function handleOpen() {
  if (props.mode === 'edit' && props.policy) {
    form.name = props.policy.name || ''
    form.description = props.policy.description || ''
    form.content = props.policy.content || ''
    form.user_name = props.policy.user_name || ''
    // Pretty-print JSON content for editing
    formatContent()
  } else {
    resetForm()
  }
}

function formatContent() {
  if (!form.content) return
  try {
    const parsed = JSON.parse(form.content)
    form.content = JSON.stringify(parsed, null, 2)
  } catch {
    ElMessage.warning('JSON 格式不正确，无法格式化')
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
    const payload = {
      name: form.name,
      description: form.description || undefined,
      content: form.content,
      user_name: form.user_name,
    }

    if (props.mode === 'create') {
      await createPolicy(payload)
      ElMessage.success('策略创建成功')
    } else {
      await updatePolicy(props.policy.id, {
        name: form.name,
        description: form.description || undefined,
        content: form.content,
      })
      ElMessage.success('策略更新成功')
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
