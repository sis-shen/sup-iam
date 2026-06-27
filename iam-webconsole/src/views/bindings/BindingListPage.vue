<template>
  <div class="page-container">
    <div class="page-header">
      <h2>绑定关系</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 创建绑定
      </el-button>
    </div>

    <div class="content-card">
      <el-table :data="bindings" v-loading="loading" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="secret_id" label="密钥 ID" width="80" />
        <el-table-column label="Access Key" min-width="160">
          <template #default="{ row }">
            <code style="background: #f5f7fa; padding: 2px 6px; border-radius: 3px; font-size: 12px;">
              {{ row.secret_access_key ? maskKey(row.secret_access_key) : row.secret_id }}
            </code>
          </template>
        </el-table-column>
        <el-table-column prop="policy_id" label="策略 ID" width="80" />
        <el-table-column prop="policy_name" label="策略名称" min-width="140" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="created_at" label="创建时间" width="170" :formatter="dateFormatter" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
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
          @size-change="fetchBindings"
          @current-change="fetchBindings"
        />
      </div>
    </div>

    <BindingFormDialog
      v-model:visible="dialogVisible"
      @success="fetchBindings"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getBindings, deleteBinding } from '@/api/bindings'
import { formatDateTime, maskKey } from '@/utils/format'
import BindingFormDialog from './BindingFormDialog.vue'

const loading = ref(false)
const bindings = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)

onMounted(() => {
  fetchBindings()
})

function dateFormatter(row, column, cellValue) {
  return formatDateTime(cellValue)
}

async function fetchBindings() {
  loading.value = true
  try {
    const res = await getBindings({ page: page.value, page_size: pageSize.value })
    bindings.value = res.items || []
    total.value = res.total || 0
  } catch {
    // Handled by interceptor
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  dialogVisible.value = true
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除此绑定关系吗？（密钥 ID: ${row.secret_id}, 策略 ID: ${row.policy_id}）`,
      '确认删除',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await deleteBinding(row.id)
    ElMessage.success('删除成功')
    fetchBindings()
  } catch {
    // Cancelled or error
  }
}
</script>
