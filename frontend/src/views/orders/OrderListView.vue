<template>
  <div class="order-list">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="平台">
          <el-select v-model="searchForm.platform" placeholder="全部平台" clearable>
            <el-option label="淘宝" value="taobao" />
            <el-option label="京东" value="jd" />
            <el-option label="抖音" value="douyin" />
            <el-option label="拼多多" value="pdd" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable>
            <el-option label="待付款" :value="100" />
            <el-option label="待审核" :value="200" />
            <el-option label="待发货" :value="300" />
            <el-option label="已发货" :value="400" />
            <el-option label="已完成" :value="600" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="订单号/买家/收件人" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 操作栏 -->
    <div class="toolbar">
      <el-button type="primary" :icon="Plus" @click="handleCreate">新建订单</el-button>
      <el-button :icon="Download">导出</el-button>
    </div>

    <!-- 数据表格 -->
    <el-table :data="orders" v-loading="loading" border @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="order_no" label="订单号" width="180">
        <template #default="{ row }">
          <el-link type="primary" @click="viewDetail(row)">{{ row.order_no }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="platform" label="平台" width="100">
        <template #default="{ row }">
          <el-tag :type="platformType(row.platform)">
            {{ platformText(row.platform) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="buyer_nick" label="买家" width="120" />
      <el-table-column prop="receiver_name" label="收件人" width="100" />
      <el-table-column prop="total_amount" label="金额" width="120">
        <template #default="{ row }">
          ¥{{ row.total_amount?.toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button 
            v-if="row.status === 200" 
            type="primary" 
            size="small"
            @click="handleAudit(row)"
          >
            审核
          </el-button>
          <el-button 
            v-if="row.status === 300" 
            type="success" 
            size="small"
            @click="handleShip(row)"
          >
            发货
          </el-button>
          <el-button link size="small" @click="viewDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next"
      @change="fetchOrders"
      class="pagination"
    />

    <!-- 发货弹窗 -->
    <el-dialog v-model="shipDialogVisible" title="订单发货" width="400px">
      <el-form :model="shipForm" label-width="100px">
        <el-form-item label="物流公司">
          <el-select v-model="shipForm.logistics_company" placeholder="选择物流">
            <el-option label="顺丰速运" value="sf" />
            <el-option label="中通快递" value="zto" />
            <el-option label="圆通速递" value="yto" />
            <el-option label="韵达速递" value="yd" />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号">
          <el-input v-model="shipForm.logistics_no" placeholder="请输入物流单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmShip">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { orderApi, type Order } from '@/api/order'

const router = useRouter()
const loading = ref(false)
const orders = ref<Order[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const selectedOrders = ref<Order[]>([])

const searchForm = reactive({
  platform: '',
  status: undefined as number | undefined,
  keyword: ''
})

const shipDialogVisible = ref(false)
const currentOrder = ref<Order | null>(null)
const shipForm = reactive({
  logistics_company: '',
  logistics_no: ''
})

const platformType = (platform: string) => {
  const map: Record<string, string> = {
    taobao: 'warning',
    jd: 'danger',
    douyin: 'info',
    pdd: 'success'
  }
  return map[platform] || ''
}

const platformText = (platform: string) => {
  const map: Record<string, string> = {
    taobao: '淘宝',
    jd: '京东',
    douyin: '抖音',
    pdd: '拼多多'
  }
  return map[platform] || platform
}

const statusType = (status: number) => {
  const map: Record<number, string> = {
    100: 'info',
    200: 'warning',
    300: 'success',
    400: '',
    500: 'success',
    600: 'info'
  }
  return map[status] || 'info'
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    100: '待付款',
    200: '待审核',
    300: '待发货',
    400: '已发货',
    500: '已签收',
    600: '已完成'
  }
  return map[status] || '未知'
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const res = await orderApi.getList({
      page: page.value,
      size: pageSize.value,
      platform: searchForm.platform,
      status: searchForm.status,
      keyword: searchForm.keyword
    })
    orders.value = res.list
    total.value = res.pagination.total
  } catch (error) {
    ElMessage.error('获取订单列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchOrders()
}

const resetSearch = () => {
  searchForm.platform = ''
  searchForm.status = undefined
  searchForm.keyword = ''
  handleSearch()
}

const handleSelectionChange = (val: Order[]) => {
  selectedOrders.value = val
}

const handleCreate = () => {
  // TODO: 打开创建订单弹窗
  ElMessage.info('创建订单功能开发中')
}

const viewDetail = (row: Order) => {
  router.push(`/orders/${row.order_no}`)
}

const handleAudit = async (row: Order) => {
  try {
    await ElMessageBox.confirm(`确认审核订单 ${row.order_no}？`)
    await orderApi.audit(row.order_no)
    ElMessage.success('审核成功')
    fetchOrders()
  } catch (error) {
    // 取消操作
  }
}

const handleShip = (row: Order) => {
  currentOrder.value = row
  shipForm.logistics_company = ''
  shipForm.logistics_no = ''
  shipDialogVisible.value = true
}

const confirmShip = async () => {
  if (!shipForm.logistics_company || !shipForm.logistics_no) {
    ElMessage.warning('请填写完整物流信息')
    return
  }
  
  try {
    await orderApi.ship(currentOrder.value!.order_no, shipForm)
    ElMessage.success('发货成功')
    shipDialogVisible.value = false
    fetchOrders()
  } catch (error) {
    ElMessage.error('发货失败')
  }
}

onMounted(fetchOrders)
</script>

<style scoped>
.order-list {
  padding: 20px;
}

.search-card {
  margin-bottom: 20px;
}

.toolbar {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
// 在 toolbar 中添加
<el-button type="warning" :icon="Plus" @click="generateMockData">生成测试数据</el-button>

// 在 script 中添加
import { mockApi } from '@/api/mock'

const generateMockData = async () => {
  try {
    await ElMessageBox.confirm('生成100个测试订单？', '提示')
    await mockApi.generateOrders(100, 'taobao', 1)
    ElMessage.success('生成成功')
    fetchOrders()
  } catch (error) {
    // 取消
  }
}