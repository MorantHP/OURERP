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
          path: 'shops',
          name: 'shops',
          component: () => import('@/views/shops/ShopListView.vue'),
          meta: { title: '店铺管理' }
        },
        {
          path: 'inventory',
          name: 'inventory',
          component: () => import('@/views/inventory/InventoryListView.vue'),
          meta: { title: '库存查询' }
        },
        {
          path: 'tenants',
          name: 'tenants',
          component: () => import('@/views/tenants/TenantListView.vue'),
          meta: { title: '账套管理' }
        },
        {
          path: 'finance',
          name: 'finance',
          redirect: '/finance/income-expense',
          meta: { title: '财务管理' },
          children: [
            {
              path: 'income-expense',
              name: 'income-expense',
              component: () => import('@/views/finance/IncomeExpenseView.vue'),
              meta: { title: '收支管理' }
            },
            {
              path: 'suppliers',
              name: 'suppliers',
              component: () => import('@/views/finance/SupplierView.vue'),
              meta: { title: '供应商管理' }
            },
            {
              path: 'product-costs',
              name: 'product-costs',
              component: () => import('@/views/finance/ProductCostView.vue'),
              meta: { title: '商品成本' }
            },
            {
              path: 'order-costs',
              name: 'order-costs',
              component: () => import('@/views/finance/OrderCostView.vue'),
              meta: { title: '订单成本' }
            },
            {
              path: 'monthly-settlements',
              name: 'monthly-settlements',
              component: () => import('@/views/finance/MonthlySettlementView.vue'),
              meta: { title: '月度结算' }
            }
          ]
        },
        {
          path: 'datacenter',
          name: 'datacenter',
          redirect: '/datacenter/realtime',
          meta: { title: '数据中心' },
          children: [
            {
              path: 'realtime',
              name: 'realtime',
              component: () => import('@/views/datacenter/RealtimeMonitorView.vue'),
              meta: { title: '实时监控' }
            },
            {
              path: 'customer',
              name: 'customer-analysis',
              component: () => import('@/views/datacenter/CustomerAnalysisView.vue'),
              meta: { title: '客户分析' }
            },
            {
              path: 'product',
              name: 'product-analysis',
              component: () => import('@/views/datacenter/ProductAnalysisView.vue'),
              meta: { title: '商品分析' }
            },
            {
              path: 'compare',
              name: 'compare-analysis',
              component: () => import('@/views/datacenter/CompareAnalysisView.vue'),
              meta: { title: '对比分析' }
            },
            {
              path: 'alerts',
              name: 'alerts',
              component: () => import('@/views/datacenter/AlertManagementView.vue'),
              meta: { title: '预警管理' }
            }
          ]
        },
        {
          path: 'settings',
          name: 'settings',
          redirect: '/settings/roles',
          meta: { title: '系统设置' },
          children: [
            {
              path: 'users',
              name: 'users',
              component: () => import('@/views/settings/UserManagementView.vue'),
              meta: { title: '用户管理' }
            },
            {
              path: 'roles',
              name: 'roles',
              component: () => import('@/views/settings/RoleView.vue'),
              meta: { title: '角色管理' }
            },
            {
              path: 'permissions',
              name: 'permissions',
              component: () => import('@/views/settings/UserPermissionView.vue'),
              meta: { title: '用户权限' }
            }
          ]
        }
      ]
    }
  ]
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  // 公开页面直接放行
  if (to.meta.public) {
    // 已登录用户访问登录页，重定向到首页
    if (to.path === '/login' && userStore.token) {
      next('/')
      return
    }
    next()
    return
  }

  // 需要认证的页面
  if (!userStore.token) {
    // 没有 token，跳转登录
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }

  // 如果还未初始化用户信息，尝试获取
  if (!userStore.isInitialized) {
    try {
      await userStore.init()
    } catch {
      // token 无效，跳转登录
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }
  }

  // 检查用户是否已审核
  if (!userStore.isApproved && to.path !== '/tenants') {
    // 未审核用户只能访问账套管理页
    // 可以根据需要添加提示
  }

  next()
})

export default router