<template>
  <div class="page-container">
    <div class="page-header">
      <h2>策略审计详情</h2>
      <el-button @click="$router.push('/audits/policies')">
        <el-icon><ArrowLeft /></el-icon> 返回列表
      </el-button>
    </div>

    <div class="detail-card" v-loading="loading">
      <template v-if="audit">
        <div class="detail-row">
          <div class="detail-label">审计 ID</div>
          <div class="detail-value">{{ audit.policy_audit_id }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">策略名称</div>
          <div class="detail-value">{{ audit.name }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">描述</div>
          <div class="detail-value">{{ audit.description || '-' }}</div>
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
        <div class="detail-row" v-if="audit.policy_shadow">
          <div class="detail-label">策略快照</div>
          <div class="detail-value">
            <pre class="json-content">{{ formatJson(audit.policy_shadow) }}</pre>
          </div>
        </div>
        <div class="detail-row" v-if="audit.extend_shadow">
          <div class="detail-label">扩展信息</div>
          <div class="detail-value">
            <pre class="json-content">{{ formatJson(audit.extend_shadow) }}</pre>
          </div>
        </div>
      </template>
      <el-empty v-else description="未找到审计记录" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getPolicyAudit } from '@/api/audits'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const loading = ref(false)
const audit = ref(null)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getPolicyAudit(route.params.id)
    audit.value = res
  } catch {
    audit.value = null
  } finally {
    loading.value = false
  }
})

function formatJson(obj) {
  if (!obj) return '-'
  if (typeof obj === 'string') {
    try {
      return JSON.stringify(JSON.parse(obj), null, 2)
    } catch {
      return obj
    }
  }
  return JSON.stringify(obj, null, 2)
}
</script>
