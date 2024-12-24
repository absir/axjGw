import api from '../index'

export default {
  // 获取
  get: () => api.post('/list/tcp', {}),
  // 添加 and 编辑
  edit: (data: any) => api.post('/edit/tcp', data),
  // 删除
  del: (data: any) => api.get(`/del/tcp/${data}`),
}
