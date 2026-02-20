<template>
  <div class="inventory-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>库存查询</span>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="仓库">
          <el-select v-model="searchForm.warehouse_id" placeholder="全部仓库" clearable style="width: 150px">
            <el-option
              v-for="wh in warehouses"
              :key="wh.id"
              :label="wh.name"
              :value="wh.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="搜索">
          <el-input v-model="searchForm.keyword" placeholder="商品编码/名称/条码" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="searchForm.low_stock">只看库存预警</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 库存表格 -->
      <el-table :data="inventoryList" v-loading="loading" stripe>
        <el-table-column prop="product?.sku_code" label="商品编码" width="120" />
        <el-table-column prop="product?.name" label="商品名称" min-width="150">
          <template #default="{ row }">
            <div class="product-info">
              <span>{{ row.product?.name }}</span>
              <el-tag v-if="row.quantity <= row.alert_qty && row.alert_qty > 0" type="danger" size="small">
                库存预警
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="product?.category" label="分类" width="100" />
        <el-table-column prop="product?.brand" label="品牌" width="100" />
        <el-table-column prop="warehouse?.name" label="仓库" width="100" />
        <el-table-column prop="quantity" label="可用库存" width="100" align="right">
          <template #default="{ row }">
            <span :class="{ 'low-stock': row.quantity <= row.alert_qty && row.alert_qty > 0 }">
              {{ row.quantity }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="locked_qty" label="锁定库存" width="100" align="right" />
        <el-table-column prop="total_qty" label="总库存" width="100" align="right" />
        <el-table-column prop="alert_qty" label="预警值" width="80" align="right" />
        <el-table-column prop="location" label="库位" width="80" />
        <el-table-column prop="product?.sale_price" label="销售价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ row.product?.sale_price?.toFixed(2) || '0.00' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="showAdjustDialog(row)">
              调整
            </el-button>
            <el-button type="info" size="small" @click="showLogsDialog(row)">
              流水
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchInventory"
        @current-change="fetchInventory"
        style="margin-top: 20px; justify-content: flex-end;"
      />
    </el-card>

    <!-- 库存调整对话框 -->
    <el-dialog v-model="adjustDialogVisible" title="库存调整" width="400px">
      <el-form :model="adjustForm" label-width="80px">
        <el-form-item label="商品">
          <span>{{ currentInventory?.product?.name }}</span>
        </el-form-item>
        <el-form-item label="当前库存">
          <span>{{ currentInventory?.quantity }}</span>
        </el-form-item>
        <el-form-item label="调整数量">
          <el-input-number v-model="adjustForm.change_qty" :min="-9999" :max="9999" />
          <div class="adjust-tip">正数增加库存，负数减少库存</div>
        </el-form-item>
        <el-form-item label="调整后">
          <span :class="{ 'low-stock': (currentInventory?.quantity || 0) + adjustForm.change_qty < 0 }">
            {{ (currentInventory?.quantity || 0) + adjustForm.change_qty }}
          </span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="adjustForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAdjust" :loading="adjusting">确认调整</el-button>
      </template>
    </el-dialog>

    <!-- 库存流水对话框 -->
    <el-dialog v-model="logsDialogVisible" title="库存流水" width="800px">
      <el-table :data="inventoryLogs" v-loading="logsLoading" max-height="400">
        <el-table-column prop="created_at" label="时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="change_qty" label="变动数量" width="100" align="right">
          <template #default="{ row }">
            <span :class="row.change_qty > 0 ? 'increase' : 'decrease'">
              {{ row.change_qty > 0 ? '+' : '' }}{{ row.change_qty }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="before_qty" label="变动前" width="80" align="right" />
        <el-table-column prop="after_qty" label="变动后" width="80" align="right" />
        <el-table-column prop="ref_type" label="类型" width="80">
          <template #default="{ row }">
            {{ getRefTypeLabel(row.ref_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="ref_no" label="关联单号" width="150" />
        <el-table-column prop="remark" label="备注" />
      </el-table>
      <el-pagination
        v-model:current-page="logsPagination.page"
        :total="logsPagination.total"
        :page-size="10"
        layout="total, prev, pager, next"
        @current-change="fetchLogs"
        style="margin-top: 15px; justify-content: flex-end;"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { inventoryApi, type Inventory, type InventoryLog } from '@/api/inventory'
import { warehouseApi, type Warehouse } from '@/api/warehouse'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const inventoryList = ref<Inventory[]>([])
const warehouses = ref<Warehouse[]>([])

const searchForm = reactive({
  warehouse_id: undefined as number | undefined,
  keyword: '',
  low_stock: false
})

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 库存调整
const adjustDialogVisible = ref(false)
const currentInventory = ref<Inventory | null>(null)
const adjustForm = reactive({
  change_qty: 0,
  remark: ''
})
const adjusting = ref(false)

// 库存流水
const logsDialogVisible = ref(false)
const logsLoading = ref(false)
const inventoryLogs = ref<InventoryLog[]>([])
const logsPagination = reactive({
  page: 1,
  total: 0
})

const fetchWarehouses = async () => {
  try {
    const res = await warehouseApi.list({ status: 1 })
    warehouses.value = res.list || []
  } catch (error) {
    console.error('Failed to fetch warehouses:', error)
  }
}

const fetchInventory = async () => {
  loading.value = true
  try {
    const res = await inventoryApi.list({
      page: pagination.page,
      size: pagination.size,
      warehouse_id: searchForm.warehouse_id,
      keyword: searchForm.keyword,
      low_stock: searchForm.low_stock || undefined
    })
    inventoryList.value = res.list || []
    pagination.total = res.pagination?.total || 0
  } catch (error) {
    ElMessage.error('获取库存列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchInventory()
}

const handleReset = () => {
  searchForm.warehouse_id = undefined
  searchForm.keyword = ''
  searchForm.low_stock = false
  handleSearch()
}

const showAdjustDialog = (inventory: Inventory) => {
  currentInventory.value = inventory
  adjustForm.change_qty = 0
  adjustForm.remark = ''
  adjustDialogVisible.value = true
}

const handleAdjust = async () => {
  if (!currentInventory.value) return
  if (adjustForm.change_qty === 0) {
    ElMessage.warning('请输入调整数量')
    return
  }

  adjusting.value = true
  try {
    await inventoryApi.adjust({
      product_id: currentInventory.value.product_id,
      warehouse_id: currentInventory.value.warehouse_id,
      change_qty: adjustForm.change_qty,
      remark: adjustForm.remark
    })
    ElMessage.success('库存调整成功')
    adjustDialogVisible.value = false
    fetchInventory()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '调整失败')
  } finally {
    adjusting.value = false
  }
}

const showLogsDialog = async (inventory: Inventory) => {
  currentInventory.value = inventory
  logsPagination.page = 1
  logsDialogVisible.value = true
  await fetchLogs()
}

const fetchLogs = async () => {
  if (!currentInventory.value) return

  logsLoading.value = true
  try {
    const res = await inventoryApi.logs({
      page: logsPagination.page,
      size: 10,
      product_id: currentInventory.value.product_id,
      warehouse_id: currentInventory.value.warehouse_id
    })
    inventoryLogs.value = res.list || []
    logsPagination.total = res.pagination?.total || 0
  } catch (error) {
    ElMessage.error('获取流水失败')
  } finally {
    logsLoading.value = false
  }
}

const getRefTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    inbound: '入库',
    outbound: '出库',
    stocktake: '盘点',
    transfer: '调拨',
    adjust: '调整'
  }
  return labels[type] || type
}

const formatDate = (date: string): string => {
  return new Date(date).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchWarehouses()
  fetchInventory()
})
</script>

<style scoped>
.inventory-list {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.search-form {
  margin-bottom: 20px;
}
.product-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.low-stock {
  color: #f56c6c;
  font-weight: bold;
}
.increase {
  color: #67c23a;
}
.decrease {
  color: #f56c6c;
}
.adjust-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
