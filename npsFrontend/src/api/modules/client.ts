import api from '../index'

export default {
  // 获取
  get: () => api.post('/list/client', {}),
  // 添加 and 编辑
  edit: (data: any) => api.post('/edit/client', data),
  // 删除
  del: (data: any) => api.get(`/del/client/${data}`),
}
