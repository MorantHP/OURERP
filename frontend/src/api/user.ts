import request from '@/utils/request'

export interface User {
  id: number
  email: string
  name: string
  phone: string
  status: number
  is_root: boolean
  is_approved: boolean
  created_at: string
}

export const userApi = {
  // 获取用户列表（仅 root）
  list: (): Promise<{ users: User[] }> =>
    request.get('/users'),

  // 审核用户
  approve: (userId: number, approved: boolean) =>
    request.put(`/users/${userId}/approve`, { approved }),

  // 设置用户状态
  setStatus: (userId: number, status: number) =>
    request.put(`/users/${userId}/status`, { status }),

  // 删除用户
  delete: (userId: number) =>
    request.delete(`/users/${userId}`),
}
