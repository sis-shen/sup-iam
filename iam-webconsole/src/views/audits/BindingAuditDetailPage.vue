<template>
  <div class="page-container">
    <div class="page-header">
      <h2>绑定审计详情</h2>
      <el-button @click="$router.push('/audits/bindings')">
        <el-icon><ArrowLeft /></el-icon> 返回列表
      </el-button>
    </div>

    <div class="detail-card" v-loading="loading">
      <template v-if="audit">
        <div class="detail-row">
          <div class="detail-label">审计 ID</div>
          <div class="detail-value">{{ audit.binding_audit_id }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">密钥 ID</div>
          <div class="detail-value">{{ audit.secret_id }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">策略 ID</div>
          <div class="detail-value">{{ audit.policy_id }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">用户名</div>
          <div class="detail-value">{{ audit.username }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">操作内容</div>
          <div class="detail-value">{{ audit.action_content }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">操作时间</div>
          <div class="detail-value">{{ formatDateTime(audit.create_time) }}</div>
        </div>
      </template>
      <el-empty v-else description="未找到审计记录" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getBindingAudit } from '@/api/audits'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const loading = ref(false)
const audit = ref(null)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getBindingAudit(route.params.id)
    audit.value = res
  } catch {
    audit.value = null
  } finally {
    loading.value = false
  }
})
</script>
