<template>
  <div class="product-cost-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>商品成本管理</span>
          <el-button type="primary" @click="showBatchDialog">批量设置成本</el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="商品SKU">
          <el-input v-model="searchForm.product_sku" placeholder="请输入SKU" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchCosts">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 数据表格 -->
      <el-table :data="costs" v-loading="loading" stripe>
        <el-table-column prop="product_sku" label="商品SKU" width="150" />
        <el-table-column prop="purchase_cost" label="采购成本" width="120" align="right">
          <template #default="{ row }">¥{{ row.purchase_cost?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="shipping_cost" label="运费成本" width="120" align="right">
          <template #default="{ row }">¥{{ row.shipping_cost?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="package_cost" label="包装成本" width="120" align="right">
          <template #default="{ row }">¥{{ row.package_cost?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="other_cost" label="其他成本" width="120" align="right">
          <template #default="{ row }">¥{{ row.other_cost?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="total_cost" label="总成本" width="120" align="right">
          <template #default="{ row }">
            <span class="text-primary">¥{{ row.total_cost?.toFixed(2) || '0.00' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="成本方法" width="100">
          <template #default="{ row }">
            <el-tag>{{ getCostMethodLabel(row.cost_method) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="stock_qty" label="库存数量" width="100" align="right" />
        <el-table-column prop="stock_value" label="库存金额" width="120" align="right">
          <template #default="{ row }">¥{{ row.stock_value?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="showEditDialog(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchCosts"
        @current-change="fetchCosts"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑商品成本" width="500px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="商品SKU">
          <span>{{ editForm.product_sku }}</span>
        </el-form-item>
        <el-form-item label="采购成本">
          <el-input-number v-model="editForm.purchase_cost" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="运费成本">
          <el-input-number v-model="editForm.shipping_cost" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="包装成本">
          <el-input-number v-model="editForm.package_cost" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="其他成本">
          <el-input-number v-model="editForm.other_cost" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="成本方法">
          <el-select v-model="editForm.cost_method" style="width: 100%">
            <el-option label="加权平均" value="weighted" />
            <el-option label="先进先出" value="fifo" />
            <el-option label="标准成本" value="standard" />
          </el-select>
        </el-form-item>
        <el-form-item label="总成本">
          <span class="text-primary">¥{{ totalCost.toFixed(2) }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 批量设置对话框 -->
    <el-dialog v-model="batchDialogVisible" title="批量设置商品成本" width="600px">
      <el-alert type="info" :closable="false" style="margin-bottom: 20px">
        通过商品SKU批量设置成本，每行一个商品，格式：SKU,采购成本,运费成本,包装成本,其他成本
      </el-alert>
      <el-input
        v-model="batchText"
        type="textarea"
        :rows="10"
        placeholder="SKU001,10.00,2.00,1.00,0.50&#10;SKU002,15.00,2.00,1.00,0.50"
      />
      <template #footer>
        <el-button @click="batchDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchSave" :loading="batchSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { productCostApi, type ProductCost } from '@/api/finance'

const loading = ref(false)
const costs = ref<ProductCost[]>([])
const editDialogVisible = ref(false)
const batchDialogVisible = ref(false)
const saving = ref(false)
const batchSaving = ref(false)
const currentId = ref<number | null>(null)
const batchText = ref('')

const searchForm = reactive({
  product_sku: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const editForm = reactive({
  product_sku: '',
  purchase_cost: 0,
  shipping_cost: 0,
  package_cost: 0,
  other_cost: 0,
  cost_method: 'weighted'
})

// 计算总成本
const totalCost = computed(() => {
  return editForm.purchase_cost + editForm.shipping_cost + editForm.package_cost + editForm.other_cost
})

// 获取成本方法标签
const getCostMethodLabel = (method: string) => {
  const labels: Record<string, string> = {
    weighted: '加权平均',
    fifo: '先进先出',
    standard: '标准成本'
  }
  return labels[method] || method
}

// 获取成本列表
const fetchCosts = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (searchForm.product_sku) params.product_sku = searchForm.product_sku

    const res = await productCostApi.list(params) as any
    costs.value = res.costs || []
    pagination.total = res.total || 0
  } catch (error) {
    ElMessage.error('获取成本列表失败')
  } finally {
    loading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.product_sku = ''
  fetchCosts()
}

// 显示编辑对话框
const showEditDialog = (cost: ProductCost) => {
  currentId.value = cost.id
  editForm.product_sku = cost.product_sku
  editForm.purchase_cost = cost.purchase_cost || 0
  editForm.shipping_cost = cost.shipping_cost || 0
  editForm.package_cost = cost.package_cost || 0
  editForm.other_cost = cost.other_cost || 0
  editForm.cost_method = cost.cost_method || 'weighted'
  editDialogVisible.value = true
}

// 显示批量设置对话框
const showBatchDialog = () => {
  batchText.value = ''
  batchDialogVisible.value = true
}

// 保存成本
const handleSave = async () => {
  if (!currentId.value) return

  saving.value = true
  try {
    await productCostApi.update(currentId.value, {
      purchase_cost: editForm.purchase_cost,
      shipping_cost: editForm.shipping_cost,
      package_cost: editForm.package_cost,
      other_cost: editForm.other_cost,
      cost_method: editForm.cost_method
    })
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    fetchCosts()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

// 批量保存
const handleBatchSave = async () => {
  if (!batchText.value.trim()) {
    ElMessage.warning('请输入成本数据')
    return
  }

  batchSaving.value = true
  try {
    const lines = batchText.value.trim().split('\n')
    const costs: any[] = []

    for (const line of lines) {
      const parts = line.split(',')
      if (parts.length >= 5) {
        costs.push({
          product_sku: parts[0].trim(),
          purchase_cost: parseFloat(parts[1]) || 0,
          shipping_cost: parseFloat(parts[2]) || 0,
          package_cost: parseFloat(parts[3]) || 0,
          other_cost: parseFloat(parts[4]) || 0
        })
      }
    }

    if (costs.length === 0) {
      ElMessage.warning('没有有效的成本数据')
      return
    }

    await productCostApi.batchUpdate(costs)
    ElMessage.success('批量设置成功')
    batchDialogVisible.value = false
    fetchCosts()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '批量设置失败')
  } finally {
    batchSaving.value = false
  }
}

onMounted(() => {
  fetchCosts()
})
</script>

<style scoped>
.product-cost-view {
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

.text-primary {
  color: #409eff;
  font-weight: bold;
}
</style>
