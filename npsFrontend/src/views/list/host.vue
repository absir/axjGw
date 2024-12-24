<route lang="yaml">
  meta:
    title: 域名
  </route>

<script setup lang="ts">
import hostApi from '@/api/modules/host'
import { ElMessage, ElMessageBox } from 'element-plus'

const dialogVisible = ref(false)

const dialogEditVisible = ref(false)

const addForm = reactive({
  Domains: '',
  ClientId: '',
  PAddr: '',
})

const editForm = reactive({
  Id: '',
  Domains: '',
  ClientId: '',
  PAddr: '',
})

const listData = ref([])

function getList() {
  hostApi.get().then((res: any) => {
    listData.value = res
  })
}

function add() {
  if (addForm.Domains === '' || addForm.ClientId === '' || addForm.PAddr === '') {
    ElMessage.error('请填写完整')
    return
  }
  hostApi.edit(addForm).then((res: any) => {
    if (res === 'ok') {
      ElMessage.success('添加成功')
      dialogVisible.value = false
      clearForm()
      getList()
    }
    else {
      ElMessage.error('添加失败')
    }
  })
}

function editSubmit() {
  if (editForm.Domains === '' || editForm.ClientId === '' || editForm.PAddr === '' || editForm.Id === '') {
    ElMessage.error('请填写完整')
    return
  }
  hostApi.edit(editForm).then((res: any) => {
    if (res === 'ok') {
      ElMessage.success('修改成功')
      dialogEditVisible.value = false
      clearForm()
      getList()
    }
    else {
      ElMessage.error('修改失败')
    }
  })
}

function edit(row: any) {
  editForm.Id = row.Id
  editForm.Domains = row.Domains
  editForm.ClientId = row.ClientId
  editForm.PAddr = row.PAddr
  dialogEditVisible.value = true
}

function del(row: any) {
  ElMessageBox.confirm(
    `确定删除${row.Domains}域名？`,
    '删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    },
  )
    .then(() => {
      hostApi.del(row.Id).then(() => {
        getList()
        ElMessage({
          type: 'success',
          message: '删除成功',
        })
      })
    })
    .catch(() => {
      ElMessage({
        type: 'info',
        message: '取消删除',
      })
    })
}

function clearForm() {
  addForm.Domains = ''
  addForm.ClientId = ''
  addForm.PAddr = ''
  editForm.Id = ''
  editForm.Domains = ''
  editForm.ClientId = ''
  editForm.PAddr = ''
}

function handleClose(done: () => void) {
  clearForm()
  done()
}

onMounted(() => {
  getList()
})
</script>

<template>
  <div>
    <PageMain>
      <div class="mb-2 flex items-center">
        <el-button type="primary" @click="dialogVisible = true">
          新增
        </el-button>
      </div>
      <div>
        <el-table :data="listData" class="w-full">
          <el-table-column prop="Domains" label="域名" />
          <el-table-column prop="ClientId" label="客户端Id" />
          <el-table-column prop="PAddr" label="代理地址" />
          <el-table-column prop="Cid" label="是否在线">
            <template #default="{ row }">
              <span v-if="row.Cid">在线</span>
              <span v-else>不在线</span>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button type="text" size="small" @click="edit(row)">
                编辑
              </el-button>
              <el-button type="text" size="small" @click="del(row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </PageMain>
    <el-dialog v-model="dialogVisible" title="新增" width="500" :before-close="handleClose">
      <div>
        <el-form :model="addForm" label-width="auto">
          <el-form-item label="域名" required>
            <el-input v-model="addForm.Domains" />
          </el-form-item>
          <el-form-item label="客户端Id" required>
            <el-input v-model.number="addForm.ClientId" />
          </el-form-item>
          <el-form-item label="代理地址" required>
            <el-input v-model="addForm.PAddr" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">
            取消
          </el-button>
          <el-button type="primary" @click="add">
            提交
          </el-button>
        </div>
      </template>
    </el-dialog>
    <el-dialog v-model="dialogEditVisible" title="编辑" width="500" :before-close="handleClose">
      <div>
        <el-form :model="editForm" label-width="auto">
          <el-form-item label="域名" required>
            <el-input v-model="editForm.Domains" />
          </el-form-item>
          <el-form-item label="客户端Id" required>
            <el-input v-model.number="editForm.ClientId" />
          </el-form-item>
          <el-form-item label="代理地址" required>
            <el-input v-model="editForm.PAddr" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogEditVisible = false">
            取消
          </el-button>
          <el-button type="primary" @click="editSubmit">
            提交
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
