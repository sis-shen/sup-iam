<template>
  <div class="page-container">
    <div class="page-header">
      <h2>AK/SK 管理</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 创建密钥
      </el-button>
    </div>

    <div class="content-card">
      <el-table :data="secrets" v-loading="loading" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="access_key" label="Access Key" min-width="160">
          <template #default="{ row }">
            <code style="background: #f5f7fa; padding: 2px 6px; border-radius: 3px; font-size: 12px;">{{ maskKey(row.access_key) }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="160" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="expires" label="过期时间" width="170" :formatter="dateFormatter" />
        <el-table-column prop="created_at" label="创建时间" width="170" :formatter="dateFormatter" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="handleRotate(row)">轮转</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchSecrets"
          @current-change="fetchSecrets"
        />
      </div>
    </div>

    <SecretFormDialog
      v-model:visible="dialogVisible"
      :mode="dialogMode"
      :secret="selectedSecret"
      @success="onDialogSuccess"
    />

    <!-- Secret key display dialog (one-time after create/rotate) -->
    <el-dialog v-model="secretKeyVisible" title="密钥信息" width="500px" :close-on-click-modal="false">
      <el-alert title="请立即保存此密钥，关闭后将无法再次查看。" type="warning" :closable="false" show-icon style="margin-bottom: 16px" />
      <div class="secret-key-display">
        <div class="secret-key-label">Access Key:</div>
        <div class="secret-key-value">{{ secretKeyResult?.access_key }}</div>
      </div>
      <div class="secret-key-display">
        <div class="secret-key-label">Secret Key:</div>
        <div class="secret-key-value">{{ secretKeyResult?.secret_key }}</div>
      </div>
      <div class="secret-key-display" v-if="secretKeyResult?.description">
        <div class="secret-key-label">描述:</div>
        <div class="secret-key-value">{{ secretKeyResult.description }}</div>
      </div>
      <template #footer>
        <el-button type="primary" @click="copySecretKey">复制密钥</el-button>
        <el-button @click="secretKeyVisible = false">我已保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSecrets, deleteSecret, rotateSecret } from '@/api/secrets'
import { formatDateTime, maskKey } from '@/utils/format'
import SecretFormDialog from './SecretFormDialog.vue'

const loading = ref(false)
const secrets = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const dialogMode = ref('create')
const selectedSecret = ref(null)

const secretKeyVisible = ref(false)
const secretKeyResult = ref(null)

onMounted(() => {
  fetchSecrets()
})

function dateFormatter(row, column, cellValue) {
  return formatDateTime(cellValue)
}

async function fetchSecrets() {
  loading.value = true
  try {
    const res = await getSecrets({ page: page.value, page_size: pageSize.value })
    secrets.value = res.items || []
    total.value = res.total || 0
  } catch {
    // Handled by interceptor
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  selectedSecret.value = null
  dialogVisible.value = true
}

function openEditDialog(secret) {
  dialogMode.value = 'edit'
  selectedSecret.value = { ...secret }
  dialogVisible.value = true
}

function onDialogSuccess(result) {
  if (result && result.secret_key) {
    secretKeyResult.value = result
    secretKeyVisible.value = true
  }
  fetchSecrets()
}

async function handleRotate(row) {
  try {
    await ElMessageBox.confirm(
      `确定要轮转密钥「${row.access_key}」吗？轮转后旧密钥将立即失效。`,
      '确认轮转',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await rotateSecret(row.id)
    secretKeyResult.value = { access_key: row.access_key, secret_key: res.secret_key }
    secretKeyVisible.value = true
    ElMessage.success('密钥轮转成功')
    fetchSecrets()
  } catch {
    // Cancelled or error
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除密钥「${row.access_key}」吗？`,
      '确认删除',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await deleteSecret(row.id)
    ElMessage.success('删除成功')
    fetchSecrets()
  } catch {
    // Cancelled or error
  }
}

async function copySecretKey() {
  if (!secretKeyResult.value) return
  const text = `Access Key: ${secretKeyResult.value.access_key}\nSecret Key: ${secretKeyResult.value.secret_key}`
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
}
</script>

<style scoped>
.secret-key-display {
  display: flex;
  padding: 10px 0;
  border-bottom: 1px solid #ebeef5;
}

.secret-key-label {
  width: 110px;
  flex-shrink: 0;
  color: #909399;
  font-weight: 500;
}

.secret-key-value {
  flex: 1;
  font-family: 'Courier New', Courier, monospace;
  word-break: break-all;
  color: #303133;
}
</style>
