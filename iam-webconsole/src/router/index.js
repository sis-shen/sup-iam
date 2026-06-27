import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '@/utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginPage.vue'),
    meta: { requiresAuth: false, title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardPage.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/UserListPage.vue'),
        meta: { title: '用户管理', icon: 'User', requiresAdmin: true },
      },
      {
        path: 'secrets',
        name: 'Secrets',
        component: () => import('@/views/secrets/SecretListPage.vue'),
        meta: { title: 'AK/SK 管理', icon: 'Key' },
      },
      {
        path: 'policies',
        name: 'Policies',
        component: () => import('@/views/policies/PolicyListPage.vue'),
        meta: { title: '策略管理', icon: 'Document' },
      },
      {
        path: 'bindings',
        name: 'Bindings',
        component: () => import('@/views/bindings/BindingListPage.vue'),
        meta: { title: '绑定关系', icon: 'Link' },
      },
      {
        path: 'audits/policies',
        name: 'PolicyAudits',
        component: () => import('@/views/audits/PolicyAuditListPage.vue'),
        meta: { title: '策略审计', icon: 'List', requiresAdmin: true },
      },
      {
        path: 'audits/policies/:id',
        name: 'PolicyAuditDetail',
        component: () => import('@/views/audits/PolicyAuditDetailPage.vue'),
        meta: { title: '策略审计详情', hidden: true },
      },
      {
        path: 'audits/bindings',
        name: 'BindingAudits',
        component: () => import('@/views/audits/BindingAuditListPage.vue'),
        meta: { title: '绑定审计', icon: 'List', requiresAdmin: true },
      },
      {
        path: 'audits/bindings/:id',
        name: 'BindingAuditDetail',
        component: () => import('@/views/audits/BindingAuditDetailPage.vue'),
        meta: { title: '绑定审计详情', hidden: true },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/ProfilePage.vue'),
        meta: { title: '个人中心', icon: 'Setting' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard for authentication
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    next('/login')
  } else if (to.path === '/login' && isAuthenticated()) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
