<template>
  <div class="page-container">
    <div class="page-header">
      <h2>绑定审计</h2>
    </div>

    <div class="content-card">
      <el-table :data="audits" v-loading="loading" border stripe style="width: 100%">
        <el-table-column prop="binding_audit_id" label="审计 ID" width="100" />
        <el-table-column prop="secret_id" label="密钥 ID" width="80" />
        <el-table-column prop="policy_id" label="策略 ID" width="80" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="action_content" label="操作内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="create_time" label="操作时间" width="170" :formatter="dateFormatter" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">
              详情
            </el-button>
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
          @size-change="fetchAudits"
          @current-change="fetchAudits"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getBindingAudits } from '@/api/audits'
import { formatDateTime } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const audits = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

onMounted(() => {
  fetchAudits()
})

function dateFormatter(row, column, cellValue) {
  return formatDateTime(cellValue)
}

async function fetchAudits() {
  loading.value = true
  try {
    const res = await getBindingAudits({ page: page.value, page_size: pageSize.value })
    audits.value = res.items || []
    total.value = res.total || 0
  } catch {
    // Handled by interceptor
  } finally {
    loading.value = false
  }
}

function viewDetail(row) {
  const auditId = row.binding_audit_id || row.id
  router.push(`/audits/bindings/${auditId}`)
}
</script>
