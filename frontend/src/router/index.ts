// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      name: 'layout',
      component: () => import('@/views/LayoutView.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: { title: '数据概览' }
        },
        {
          path: 'orders',
          name: 'orders',
          component: () => import('@/views/orders/OrderListView.vue'),
          meta: { title: '订单管理' }
        },
        {
          path: 'inventory',
          name: 'inventory',
          component: () => import('@/views/inventory/InventoryView.vue'),
          meta: { title: '库存管理' }
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (!to.meta.public && !userStore.token) {
    next('/login')
  } else {
    next()
  }
})

export default router