import api from '../index'

export default {
  // 登录
  login: (data: any) => api.post('/login', data, { responseType: 'text' }),

  // 获取权限
  permission: () => api.get('user/permission', {
    baseURL: '/mock/',
  }),

  // 修改密码
  passwordEdit: (data: {
    password: string
    newpassword: string
  }) => api.post('user/password/edit', data, {
    baseURL: '/mock/',
  }),
}
