import { createRouter, createWebHistory } from 'vue-router'
import pinia, { useUserStore } from '../stores'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '设备看板' }
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('../views/Users.vue'),
        meta: { title: '用户管理', adminOnly: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 登录守卫：未登录跳转登录页，已登录访问登录页则回首页；管理员专属页面拦截普通用户
router.beforeEach((to) => {
  const userStore = useUserStore(pinia)
  if (to.path !== '/login' && !userStore.isLoggedIn) {
    return '/login'
  }
  if (to.path === '/login' && userStore.isLoggedIn) {
    return '/'
  }
  if (to.meta.adminOnly && !userStore.isAdmin) {
    return '/dashboard'
  }
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} - GV-Admin-Atom` : 'GV-Admin-Atom'
})

export default router
