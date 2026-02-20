<template>
  <div class="shop-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>店铺管理</span>
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            添加店铺
          </el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <el-form :inline="true" class="filter-form">
        <el-form-item label="平台">
          <el-select v-model="filters.platform" placeholder="全部平台" clearable @change="loadShops">
            <el-option v-for="p in platforms" :key="p.code" :label="p.name" :value="p.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部状态" clearable @change="loadShops">
            <el-option label="已启用" :value="1" />
            <el-option label="已禁用" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>

      <!-- 店铺列表 -->
      <el-table :data="shops" v-loading="loading">
        <el-table-column label="店铺名称" prop="name" />
        <el-table-column label="平台" width="120">
          <template #default="{ row }">
            <el-tag :type="getPlatformTagType(row.platform)">
              {{ getPlatformName(row.platform) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '已启用' : '已禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="授权状态" width="120">
          <template #default="{ row }">
            <el-tag v-if="isTokenExpired(row)" type="danger">已过期</el-tag>
            <el-tag v-else-if="isTokenExpiring(row)" type="warning">即将过期</el-tag>
            <el-tag v-else type="success">正常</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="同步间隔" width="100">
          <template #default="{ row }">
            {{ row.sync_interval }}分钟
          </template>
        </el-table-column>
        <el-table-column label="最后同步" width="180">
          <template #default="{ row }">
            {{ row.last_sync_at ? formatTime(row.last_sync_at) : '从未同步' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleSync(row)" :disabled="row.status !== 1">同步</el-button>
            <el-button size="small" type="warning" @click="handleAuthorize(row)"
                       v-if="needAuth(row)">授权</el-button>
            <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        @current-change="loadShops"
        layout="total, prev, pager, next"
        class="pagination"
      />
    </el-card>

    <!-- 添加/编辑店铺对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingShop ? '编辑店铺' : '添加店铺'" width="500px">
      <el-form :model="shopForm" label-width="100px">
        <el-form-item label="平台" required v-if="!editingShop">
          <el-select v-model="shopForm.platform" placeholder="选择平台" @change="onPlatformChange">
            <el-option v-for="p in platforms" :key="p.code" :label="p.name" :value="p.code">
              <span>{{ p.name }}</span>
              <span style="color: #999; margin-left: 8px; font-size: 12px;">{{ p.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="店铺名称" required>
          <el-input v-model="shopForm.name" placeholder="输入店铺名称" />
        </el-form-item>
        <el-form-item label="平台店铺ID">
          <el-input v-model="shopForm.platform_shop_id" placeholder="平台店铺ID（可选）" />
        </el-form-item>
        <el-form-item label="App Key" v-if="needAppKey">
          <el-input v-model="shopForm.app_key" placeholder="输入App Key" />
        </el-form-item>
        <el-form-item label="App Secret" v-if="needAppKey">
          <el-input v-model="shopForm.app_secret" placeholder="输入App Secret" show-password />
        </el-form-item>
        <el-form-item label="API地址" v-if="shopForm.platform === 'custom'">
          <el-input v-model="shopForm.api_url" placeholder="自定义API地址" />
        </el-form-item>
        <el-form-item label="同步间隔">
          <el-select v-model="shopForm.sync_interval">
            <el-option label="每5分钟" :value="5" />
            <el-option label="每15分钟" :value="15" />
            <el-option label="每30分钟" :value="30" />
            <el-option label="每小时" :value="60" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" v-if="editingShop">
          <el-switch v-model="shopForm.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { shopApi, platformApi, oauthApi, type Shop, type Platform } from '@/api/shop'

const loading = ref(false)
const shops = ref<Shop[]>([])
const platforms = ref<Platform[]>([])
const showAddDialog = ref(false)
const editingShop = ref<Shop | null>(null)

const filters = ref({
  platform: '',
  status: undefined as number | undefined,
})

const pagination = ref({
  page: 1,
  size: 20,
  total: 0,
})

const shopForm = ref({
  platform: '',
  name: '',
  platform_shop_id: '',
  app_key: '',
  app_secret: '',
  api_url: '',
  sync_interval: 30,
  status: 1,
})

const needAppKey = computed(() => {
  const p = platforms.value.find(x => x.code === shopForm.value.platform)
  return p && ['oauth', 'apikey'].includes(p.auth_type)
})

onMounted(async () => {
  await loadPlatforms()
  await loadShops()
})

async function loadPlatforms() {
  try {
    const res = await platformApi.list()
    platforms.value = res.platforms
  } catch (e) {
    console.error('加载平台列表失败', e)
  }
}

async function loadShops() {
  loading.value = true
  try {
    const res = await shopApi.list({
      platform: filters.value.platform,
      status: filters.value.status,
      page: pagination.value.page,
      size: pagination.value.size,
    })
    shops.value = res.list
    pagination.value.total = res.pagination.total
  } catch (e) {
    ElMessage.error('加载店铺列表失败')
  } finally {
    loading.value = false
  }
}

function getPlatformName(code: string) {
  return platforms.value.find(p => p.code === code)?.name || code
}

function getPlatformTagType(code: string) {
  const types: Record<string, string> = {
    taobao: 'warning',
    tmall: 'warning',
    douyin: '',
    kuaishou: 'danger',
    xiaohongshu: 'danger',
    vip: 'danger',
    '1688': 'warning',
    wechat_video: 'success',
    tiktok: '',
    jingqi: 'danger',
    custom: 'info',
  }
  return types[code] || 'info'
}

function isTokenExpired(shop: Shop) {
  if (!shop.token_expires_at) return true
  return new Date(shop.token_expires_at) < new Date()
}

function isTokenExpiring(shop: Shop) {
  if (!shop.token_expires_at) return false
  const expiresAt = new Date(shop.token_expires_at)
  const warningTime = new Date(Date.now() + 24 * 60 * 60 * 1000)
  return expiresAt < warningTime
}

function needAuth(shop: Shop) {
  const p = platforms.value.find(x => x.code === shop.platform)
  return p?.auth_type === 'oauth' && isTokenExpired(shop)
}

async function handleAuthorize(shop: Shop) {
  try {
    const res = await oauthApi.getAuthUrl(shop.id)
    window.location.href = res.auth_url
  } catch (e) {
    ElMessage.error('获取授权链接失败')
  }
}

async function handleSync(shop: Shop) {
  try {
    await shopApi.triggerSync(shop.id)
    ElMessage.success('同步任务已启动')
  } catch (e) {
    ElMessage.error('启动同步失败')
  }
}

function handleEdit(shop: Shop) {
  editingShop.value = shop
  shopForm.value = {
    platform: shop.platform,
    name: shop.name,
    platform_shop_id: shop.platform_shop_id || '',
    app_key: '',
    app_secret: '',
    api_url: shop.api_url || '',
    sync_interval: shop.sync_interval,
    status: shop.status,
  }
  showAddDialog.value = true
}

async function handleDelete(shop: Shop) {
  try {
    await ElMessageBox.confirm('确定要删除此店铺吗？', '确认删除')
    await shopApi.delete(shop.id)
    ElMessage.success('删除成功')
    loadShops()
  } catch (e) {
    // 取消
  }
}

async function handleSave() {
  if (!shopForm.value.name) {
    ElMessage.warning('请输入店铺名称')
    return
  }

  try {
    if (editingShop.value) {
      await shopApi.update(editingShop.value.id, {
        name: shopForm.value.name,
        platform_shop_id: shopForm.value.platform_shop_id,
        app_key: shopForm.value.app_key || undefined,
        app_secret: shopForm.value.app_secret || undefined,
        api_url: shopForm.value.api_url || undefined,
        sync_interval: shopForm.value.sync_interval,
        status: shopForm.value.status,
      })
      ElMessage.success('更新成功')
    } else {
      if (!shopForm.value.platform) {
        ElMessage.warning('请选择平台')
        return
      }
      await shopApi.create({
        name: shopForm.value.name,
        platform: shopForm.value.platform,
        platform_shop_id: shopForm.value.platform_shop_id,
        app_key: shopForm.value.app_key,
        app_secret: shopForm.value.app_secret,
        api_url: shopForm.value.api_url,
        sync_interval: shopForm.value.sync_interval,
      })
      ElMessage.success('添加成功，请完成授权')
    }
    showAddDialog.value = false
    editingShop.value = null
    resetForm()
    loadShops()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

function onPlatformChange() {
  // 可以根据平台自动填充默认名称
}

function resetForm() {
  shopForm.value = {
    platform: '',
    name: '',
    platform_shop_id: '',
    app_key: '',
    app_secret: '',
    api_url: '',
    sync_interval: 30,
    status: 1,
  }
}

function formatTime(time: string) {
  return new Date(time).toLocaleString()
}
</script>

<style scoped>
.shop-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
