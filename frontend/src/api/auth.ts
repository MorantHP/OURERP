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

export const authApi = {
  login(data: LoginParams) {
    return request.post('/auth/login', data)
  },
  
  register(data: RegisterParams) {
    return request.post('/auth/register', data)
  },
  
  getMe() {
    return request.get('/auth/me')
  }
}