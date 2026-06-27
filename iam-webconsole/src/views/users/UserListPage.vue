<template>
  <div class="page-container">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 创建用户
      </el-button>
    </div>

    <div class="content-card">
      <div class="search-bar">
        <el-input v-model="searchQuery" placeholder="搜索用户名/邮箱" clearable @clear="fetchUsers" @keyup.enter="fetchUsers" />
        <el-button type="primary" @click="fetchUsers">搜索</el-button>
      </div>

      <el-table :data="users" v-loading="loading" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_enable ? 'success' : 'danger'" size="small">
              {{ row.is_enable ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="管理员" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'warning' : 'info'" size="small">
              {{ row.is_admin ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" :formatter="dateFormatter" />
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
          @size-change="fetchUsers"
          @current-change="fetchUsers"
        />
      </div>
    </div>

    <UserFormDialog
      v-model:visible="dialogVisible"
      :mode="dialogMode"
      :user="selectedUser"
      @success="fetchUsers"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUsers, deleteUser } from '@/api/users'
import { formatDateTime } from '@/utils/format'
import UserFormDialog from './UserFormDialog.vue'

const loading = ref(false)
const users = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchQuery = ref('')

const dialogVisible = ref(false)
const dialogMode = ref('create')
const selectedUser = ref(null)

onMounted(() => {
  fetchUsers()
})

function dateFormatter(row, column, cellValue) {
  return formatDateTime(cellValue)
}

async function fetchUsers() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (searchQuery.value.trim()) {
      params.keyword = searchQuery.value.trim()
    }
    const res = await getUsers(params)
    users.value = res.items || []
    total.value = res.total || 0
  } catch {
    // Error handled by interceptor
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  selectedUser.value = null
  dialogVisible.value = true
}

function openEditDialog(user) {
  dialogMode.value = 'edit'
  selectedUser.value = { ...user }
  dialogVisible.value = true
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除用户「${row.username}」吗？此操作不可恢复。`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteUser(row.id)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch {
    // Cancelled or error
  }
}
</script>
