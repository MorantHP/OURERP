import request from '@/utils/request'

export const mockApi = {
  // 生成模拟订单
  generateOrders(count: number, platform: string, shopID: number) {
    return request.post('/mock/generate', {
      count,
      platform,
      shop_id: shopID
    })
  },
  
  // 查看模拟订单
  listMockOrders(limit: number = 10) {
    return request.get('/mock/orders', { params: { limit } })
  },
  
  // 启动实时生成
  startRealtime() {
    return request.post('/mock/realtime/start')
  }
}