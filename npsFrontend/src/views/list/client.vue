<route lang="yaml">
  meta:
    title: 客户端
  </route>

<script setup lang="ts">
import clientApi from '@/api/modules/client'
import { ElMessage, ElMessageBox } from 'element-plus'

const dialogVisible = ref(false)

const dialogEditVisible = ref(false)

const addForm = reactive({
  name: '',
  secret: '',
})

const editForm = reactive({
  id: '',
  name: '',
  secret: '',
})

const listData = ref([])

function getList() {
  clientApi.get().then((res: any) => {
    listData.value = res
  })
}

function add() {
  if (addForm.name === '' || addForm.secret === '') {
    ElMessage.error('请填写完整')
    return
  }
  clientApi.edit(addForm).then((res: any) => {
    if (res === 'ok') {
      ElMessage.success('添加成功')
      dialogVisible.value = false
      getList()
    }
    else {
      ElMessage.error('添加失败')
    }
  })
}

function editSubmit() {
  if (editForm.name === '' || editForm.secret === '' || editForm.id === '') {
    ElMessage.error('请填写完整')
    return
  }
  clientApi.edit(editForm).then((res: any) => {
    if (res === 'ok') {
      ElMessage.success('修改成功')
      dialogEditVisible.value = false
      getList()
    }
    else {
      ElMessage.error('修改失败')
    }
  })
}

function edit(row: any) {
  editForm.id = row.Id
  editForm.name = row.Name
  editForm.secret = row.Secret
  dialogEditVisible.value = true
}

function del(row: any) {
  ElMessageBox.confirm(
    `确定删除${row.Name}客户端？`,
    '删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    },
  )
    .then(() => {
      clientApi.del(row.Id).then(() => {
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

function handleClose(done: () => void) {
  addForm.name = ''
  addForm.secret = ''
  editForm.id = ''
  editForm.name = ''
  editForm.secret = ''
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
          <el-table-column prop="Id" label="Id" />
          <el-table-column prop="Name" label="名称" />
          <el-table-column prop="Secret" label="秘钥" />
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
          <el-form-item label="名称" required>
            <el-input v-model="addForm.name" />
          </el-form-item>
          <el-form-item label="秘钥" required>
            <el-input v-model="addForm.secret" />
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
          <el-form-item label="名称" required>
            <el-input v-model="editForm.name" />
          </el-form-item>
          <el-form-item label="秘钥" required>
            <el-input v-model="editForm.secret" />
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
