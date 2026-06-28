<template>
  <div class="page-container">
    <div class="page-header">
      <h2>策略管理</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 创建策略
      </el-button>
    </div>

    <div class="content-card">
      <el-table :data="policies" v-loading="loading" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="created_at" label="创建时间" width="170" :formatter="dateFormatter" />
        <el-table-column prop="updated_at" label="更新时间" width="170" :formatter="dateFormatter" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
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
          @size-change="fetchPolicies"
          @current-change="fetchPolicies"
        />
      </div>
    </div>

    <PolicyFormDialog
      v-model:visible="dialogVisible"
      :mode="dialogMode"
      :policy="selectedPolicy"
      @success="fetchPolicies"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPolicies, deletePolicy } from '@/api/policies'
import { formatDateTime } from '@/utils/format'
import PolicyFormDialog from './PolicyFormDialog.vue'

const loading = ref(false)
const policies = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const dialogMode = ref('create')
const selectedPolicy = ref(null)

onMounted(() => {
  fetchPolicies()
})

function dateFormatter(row, column, cellValue) {
  return formatDateTime(cellValue)
}

async function fetchPolicies() {
  loading.value = true
  try {
    const res = await getPolicies({ page: page.value, page_size: pageSize.value })
    policies.value = res.items || []
    total.value = res.total || 0
    // If current page is empty and not on page 1, go back one page
    if (policies.value.length === 0 && page.value > 1) {
      page.value--
      return fetchPolicies()
    }
  } catch {
    // Handled by interceptor
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  selectedPolicy.value = null
  dialogVisible.value = true
}

function openEditDialog(policy) {
  dialogMode.value = 'edit'
  selectedPolicy.value = { ...policy }
  dialogVisible.value = true
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除策略「${row.name}」吗？`,
      '确认删除',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await deletePolicy(row.id)
    ElMessage.success('删除成功')
    fetchPolicies()
  } catch {
    // Cancelled or error
  }
}
</script>
