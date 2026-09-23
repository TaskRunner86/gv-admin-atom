import { createRouter, createWebHistory } from 'vue-router'

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

// 当前登录用户信息（登录成功后写入 localStorage）
const currentUser = () => {
  try {
    return JSON.parse(localStorage.getItem('user') || 'null')
  } catch {
    return null
  }
}

// 登录守卫：未登录跳转登录页，已登录访问登录页则回首页；管理员专属页面拦截普通用户
router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    return '/login'
  }
  if (to.path === '/login' && token) {
    return '/'
  }
  if (to.meta.adminOnly && currentUser()?.role !== 'admin') {
    return '/dashboard'
  }
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} - GV Dashboard` : 'GV Dashboard'
})

export default router
