<template>
  <div class="sidebar-header">
    <span v-if="!collapsed" class="sidebar-title">IAM 控制台</span>
    <span v-else class="sidebar-title-short">IAM</span>
  </div>
  <el-menu
    :default-active="activeMenu"
    :collapse="collapsed"
    :collapse-transition="false"
    background-color="#304156"
    text-color="#bfcbd9"
    active-text-color="#409eff"
    router
    class="sidebar-menu"
  >
    <template v-for="item in menuItems" :key="item.path">
      <el-menu-item v-if="!item.children" :index="item.path">
        <el-icon><component :is="item.icon" /></el-icon>
        <template #title>{{ item.title }}</template>
      </el-menu-item>
      <el-sub-menu v-else :index="item.path">
        <template #title>
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </template>
        <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
          {{ child.title }}
        </el-menu-item>
      </el-sub-menu>
    </template>
  </el-menu>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const props = defineProps({
  collapsed: {
    type: Boolean,
    default: false,
  },
})

const route = useRoute()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)

const allMenuItems = [
  { path: '/dashboard', title: '仪表盘', icon: 'Odometer', adminOnly: false },
  { path: '/users', title: '用户管理', icon: 'User', adminOnly: true },
  { path: '/secrets', title: 'AK/SK 管理', icon: 'Key', adminOnly: false },
  { path: '/policies', title: '策略管理', icon: 'Document', adminOnly: false },
  { path: '/bindings', title: '绑定关系', icon: 'Link', adminOnly: false },
  {
    path: '/audits',
    title: '审计日志',
    icon: 'List',
    adminOnly: true,
    children: [
      { path: '/audits/policies', title: '策略审计' },
      { path: '/audits/bindings', title: '绑定审计' },
    ],
  },
  { path: '/profile', title: '个人中心', icon: 'Setting', adminOnly: false },
]

const menuItems = computed(() => {
  if (userStore.isAdmin) return allMenuItems
  return allMenuItems.filter((item) => !item.adminOnly)
})
</script>

<style scoped>
.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-title-short {
  font-size: 16px;
}

.sidebar-menu {
  border-right: none;
}
</style>
