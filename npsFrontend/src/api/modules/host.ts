import api from '../index'

export default {
  // 获取
  get: () => api.post('/list/host', {}),
  // 添加 and 编辑
  edit: (data: any) => api.post('/edit/host', data),
  // 删除
  del: (data: any) => api.get(`/del/host/${data}`),
}
