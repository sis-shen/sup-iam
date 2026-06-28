<template>
  <div class="page-container">
    <div class="page-header">
      <h2>仪表盘</h2>
    </div>
    <div class="content-card">
      <div class="welcome-section">
        <h3>欢迎回来，{{ userStore.username || '用户' }}</h3>
        <p class="welcome-text">身份与访问管理系统 (IAM) 管理控制台</p>
      </div>
      <el-row :gutter="20" class="stats-row">
        <el-col :span="6" v-for="stat in stats" :key="stat.label">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value">{{ stat.count }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </el-card>
        </el-col>
      </el-row>
      <el-row :gutter="20" class="quick-links">
        <el-col :span="24">
          <el-card shadow="never">
            <template #header>
              <span>快捷入口</span>
            </template>
            <el-space wrap>
              <el-button type="primary" @click="$router.push('/secrets')">
                <el-icon><Key /></el-icon> 管理 AK/SK
              </el-button>
              <el-button type="success" @click="$router.push('/policies')">
                <el-icon><Document /></el-icon> 管理策略
              </el-button>
              <el-button type="warning" @click="$router.push('/bindings')">
                <el-icon><Link /></el-icon> 管理绑定
              </el-button>
              <el-button v-if="userStore.isAdmin" @click="$router.push('/users')">
                <el-icon><User /></el-icon> 管理用户
              </el-button>
            </el-space>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getUsers } from '@/api/users'
import { getSecrets } from '@/api/secrets'
import { getPolicies } from '@/api/policies'
import { getBindings } from '@/api/bindings'

const router = useRouter()
const userStore = useUserStore()

const stats = ref([
  { label: '用户总数', count: '-' },
  { label: '密钥总数', count: '-' },
  { label: '策略总数', count: '-' },
  { label: '绑定总数', count: '-' },
])

onMounted(async () => {
  await userStore.fetchUserInfo()

  // Fetch stats (using page_size=1 to get total count)
  try {
    const [usersRes, secretsRes, policiesRes, bindingsRes] = await Promise.allSettled([
      getUsers({ page: 1, page_size: 1 }),
      getSecrets({ page: 1, page_size: 1 }),
      getPolicies({ page: 1, page_size: 1 }),
      getBindings({ page: 1, page_size: 1 }),
    ])

    if (usersRes.status === 'fulfilled') stats.value[0].count = usersRes.value.total ?? '-'
    if (secretsRes.status === 'fulfilled') stats.value[1].count = secretsRes.value.total ?? '-'
    if (policiesRes.status === 'fulfilled') stats.value[2].count = policiesRes.value.total ?? '-'
    if (bindingsRes.status === 'fulfilled') stats.value[3].count = bindingsRes.value.total ?? '-'
  } catch {
    // Ignore stat fetch errors
  }
})
</script>

<style scoped>
.welcome-section {
  margin-bottom: 24px;
}

.welcome-section h3 {
  margin: 0 0 8px;
  font-size: 20px;
  color: #303133;
}

.welcome-text {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
  cursor: default;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: #409eff;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}

.quick-links {
  margin-top: 16px;
}
</style>
