import request from '@/utils/request'

export interface LoginParams {
  email: string
  password: string
}

export interface RegisterParams {
  email: string
  password: string
  name: string
}

export interface UserInfo {
  id: number
  email: string
  name: string
  is_root: boolean
  is_approved: boolean
  status: number
  created_at: string
  updated_at: string
}

export interface LoginResponse {
  token: string
  user: UserInfo
}

export interface GetMeResponse {
  user: UserInfo
}

export const authApi = {
  login(data: LoginParams): Promise<LoginResponse> {
    return request.post('/auth/login', data)
  },

  register(data: RegisterParams): Promise<{ message: string }> {
    return request.post('/auth/register', data)
  },

  getMe(): Promise<GetMeResponse> {
    return request.get('/auth/me')
  }
}
