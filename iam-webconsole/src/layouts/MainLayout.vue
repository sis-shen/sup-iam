<template>
  <el-container style="height: 100vh">
    <el-aside :width="sidebarWidth" class="sidebar-container">
      <SidebarMenu :collapsed="isCollapsed" />
    </el-aside>
    <el-container>
      <HeaderBar @toggle-sidebar="toggleSidebar" :collapsed="isCollapsed" />
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, provide } from 'vue'
import SidebarMenu from './SidebarMenu.vue'
import HeaderBar from './HeaderBar.vue'

const isCollapsed = ref(false)
const sidebarWidth = ref('220px')

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value
  sidebarWidth.value = isCollapsed.value ? '64px' : '220px'
}
</script>

<style scoped>
.sidebar-container {
  background-color: #304156;
  transition: width 0.3s;
  overflow: hidden;
}

.main-content {
  background-color: #f5f7fa;
  padding: 0;
  overflow-y: auto;
}
</style>
