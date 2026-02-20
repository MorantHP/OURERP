<template>
  <el-container class="layout-container">
    <el-aside width="200px" class="sidebar">
      <div class="logo">OURERP</div>
      <el-menu :default-active="$route.path" router background-color="#304156" text-color="#fff" active-text-color="#409EFF">
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>数据概览</span>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><Document /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
        <el-menu-item index="/shops">
          <el-icon><Shop /></el-icon>
          <span>店铺管理</span>
        </el-menu-item>
        <el-menu-item index="/inventory">
          <el-icon><Box /></el-icon>
          <span>库存管理</span>
        </el-menu-item>
        <el-sub-menu index="/finance">
          <template #title>
            <el-icon><Wallet /></el-icon>
            <span>财务管理</span>
          </template>
          <el-menu-item index="/finance/income-expense">
            <el-icon><Coin /></el-icon>
            <span>收支管理</span>
          </el-menu-item>
          <el-menu-item index="/finance/suppliers">
            <el-icon><Avatar /></el-icon>
            <span>供应商管理</span>
          </el-menu-item>
          <el-menu-item index="/finance/product-costs">
            <el-icon><PriceTag /></el-icon>
            <span>商品成本</span>
          </el-menu-item>
          <el-menu-item index="/finance/order-costs">
            <el-icon><TrendCharts /></el-icon>
            <span>订单成本</span>
          </el-menu-item>
          <el-menu-item index="/finance/monthly-settlements">
            <el-icon><Calendar /></el-icon>
            <span>月度结算</span>
          </el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="/datacenter">
          <template #title>
            <el-icon><DataAnalysis /></el-icon>
            <span>数据中心</span>
          </template>
          <el-menu-item index="/datacenter/realtime">
            <el-icon><Monitor /></el-icon>
            <span>实时监控</span>
          </el-menu-item>
          <el-menu-item index="/datacenter/customer">
            <el-icon><UserFilled /></el-icon>
            <span>客户分析</span>
          </el-menu-item>
          <el-menu-item index="/datacenter/product">
            <el-icon><Histogram /></el-icon>
            <span>商品分析</span>
          </el-menu-item>
          <el-menu-item index="/datacenter/compare">
            <el-icon><TrendCharts /></el-icon>
            <span>对比分析</span>
          </el-menu-item>
          <el-menu-item index="/datacenter/alerts">
            <el-icon><BellFilled /></el-icon>
            <span>预警管理</span>
          </el-menu-item>
        </el-sub-menu>
        <el-menu-item index="/tenants">
          <el-icon><OfficeBuilding /></el-icon>
          <span>账套管理</span>
        </el-menu-item>
        <el-sub-menu index="/settings">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item v-if="userStore.userInfo?.is_root" index="/settings/users">
            <el-icon><UserFilled /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="/settings/roles">
            <el-icon><UserFilled /></el-icon>
            <span>角色管理</span>
          </el-menu-item>
          <el-menu-item index="/settings/permissions">
            <el-icon><Lock /></el-icon>
            <span>用户权限</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-select
            :model-value="tenantStore.currentTenantId"
            placeholder="选择账套"
            class="tenant-select"
            @change="handleTenantChange"
            :disabled="tenantStore.tenants.length === 0"
          >
            <el-option
              v-for="tenant in tenantStore.tenants"
              :key="tenant.id"
              :label="tenant.name"
              :value="tenant.id"
            >
              <div class="tenant-option">
                <span class="tenant-name">{{ tenant.name }}</span>
                <el-tag size="small" :type="getPlatformTagType(tenant.platform)">
                  {{ getPlatformLabel(tenant.platform) }}
                </el-tag>
              </div>
            </el-option>
          </el-select>
          <el-button
            v-if="!tenantStore.hasTenant && tenantStore.tenants.length === 0"
            type="primary"
            size="small"
            @click="router.push('/tenants')"
          >
            创建账套
          </el-button>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              {{ userStore.userInfo?.name || 'User' }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useTenantStore } from '@/stores/tenant'
import { Odometer, Document, Box, ArrowDown, Shop, OfficeBuilding, Setting, UserFilled, Lock, Wallet, Coin, Avatar, PriceTag, TrendCharts, Calendar, DataAnalysis, Monitor, Histogram, BellFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const tenantStore = useTenantStore()

// 平台标签类型映射
const platformTagTypes: Record<string, string> = {
  taobao: 'warning',
  tmall: 'danger',
  douyin: '',
  kuaishou: 'success',
  wechat_video: 'success',
  jd: 'danger',
  xiaohongshu: 'danger',
  custom: 'info'
}

// 平台标签文字映射
const platformLabels: Record<string, string> = {
  taobao: '淘宝',
  tmall: '天猫',
  douyin: '抖音',
  kuaishou: '快手',
  wechat_video: '微信视频号',
  jd: '京东',
  xiaohongshu: '小红书',
  vip: '唯品会',
  custom: '自定义'
}

const getPlatformTagType = (platform: string): string => {
  return platformTagTypes[platform] || 'info'
}

const getPlatformLabel = (platform: string): string => {
  return platformLabels[platform] || platform
}

// 切换租户
const handleTenantChange = async (tenantId: number) => {
  if (tenantId === tenantStore.currentTenantId) return
  try {
    console.log('切换租户:', tenantId)
    await tenantStore.switchTenant(tenantId)
    ElMessage.success('账套切换成功')
  } catch (error: any) {
    console.error('切换租户失败:', error)
    ElMessage.error(error.response?.data?.error || error.message || '账套切换失败')
  }
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    tenantStore.clearTenant()
    userStore.logout()
    router.push('/login')
  }
}

// 加载租户列表
onMounted(async () => {
  try {
    await tenantStore.fetchTenants()
  } catch (error) {
    console.error('Failed to load tenants:', error)
  }
})
</script>

<style scoped>
.layout-container { height: 100vh; }
.sidebar { background-color: #304156; }
.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  border-bottom: 1px solid #1f2d3d;
}
.header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0,21,41,.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tenant-select {
  width: 200px;
}
.tenant-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.tenant-name {
  margin-right: 8px;
}
.header-right {
  display: flex;
  align-items: center;
}
.user-info { cursor: pointer; color: #606266; }
.main-content { background-color: #f0f2f5; padding: 20px; }
</style>
