<template>
  <el-dialog
    :model-value="visible"
    title="创建绑定关系"
    width="500px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @open="handleOpen"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" label-position="left">
      <el-form-item label="Access Key" prop="secret_id">
        <el-select
          v-model="form.secret_id"
          filterable
          remote
          :remote-method="searchSecrets"
          :loading="loadingSecrets"
          placeholder="搜索并选择密钥"
          style="width: 100%"
        >
          <el-option
            v-for="s in secretOptions"
            :key="s.id"
            :label="`${s.access_key} (${s.username})`"
            :value="s.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="策略" prop="policy_id">
        <el-select
          v-model="form.policy_id"
          filterable
          remote
          :remote-method="searchPolicies"
          :loading="loadingPolicies"
          placeholder="搜索并选择策略"
          style="width: 100%"
        >
          <el-option
            v-for="p in policyOptions"
            :key="p.id"
            :label="`${p.name} (${p.username})`"
            :value="p.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" placeholder="关联的用户名" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">创建</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { createBinding } from '@/api/bindings'
import { getSecrets } from '@/api/secrets'
import { getPolicies } from '@/api/policies'

const props = defineProps({
  visible: { type: Boolean, default: false },
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  secret_id: null,
  policy_id: null,
  username: '',
})

const rules = {
  secret_id: [{ required: true, message: '请选择密钥', trigger: 'change' }],
  policy_id: [{ required: true, message: '请选择策略', trigger: 'change' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
}

const secretOptions = ref([])
const policyOptions = ref([])
const loadingSecrets = ref(false)
const loadingPolicies = ref(false)

async function handleOpen() {
  form.secret_id = null
  form.policy_id = null
  form.username = ''
  // Load initial options
  searchSecrets('')
  searchPolicies('')
}

async function searchSecrets(query) {
  loadingSecrets.value = true
  try {
    const res = await getSecrets({ page: 1, page_size: 50 })
    let items = res.items || []
    if (query) {
      items = items.filter(s =>
        s.access_key.toLowerCase().includes(query.toLowerCase()) ||
        (s.username && s.username.toLowerCase().includes(query.toLowerCase()))
      )
    }
    secretOptions.value = items
  } catch {
    secretOptions.value = []
  } finally {
    loadingSecrets.value = false
  }
}

async function searchPolicies(query) {
  loadingPolicies.value = true
  try {
    const res = await getPolicies({ page: 1, page_size: 50 })
    let items = res.items || []
    if (query) {
      items = items.filter(p =>
        p.name.toLowerCase().includes(query.toLowerCase()) ||
        (p.username && p.username.toLowerCase().includes(query.toLowerCase()))
      )
    }
    policyOptions.value = items
  } catch {
    policyOptions.value = []
  } finally {
    loadingPolicies.value = false
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
    await createBinding({
      secret_id: form.secret_id,
      policy_id: form.policy_id,
      username: form.username,
    })
    ElMessage.success('绑定创建成功')
    emit('update:visible', false)
    emit('success')
  } catch {
    // Error shown by interceptor
  } finally {
    submitting.value = false
  }
}
</script>
