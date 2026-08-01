<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">订单套餐管理</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="游戏">
          <el-select
            v-model="searchForm.gameId"
            placeholder="全部游戏"
            clearable
            style="width: 180px"
          >
            <el-option
              v-for="game in gameList"
              :key="game.id"
              :label="game.name"
              :value="game.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
          <el-button type="success" :icon="Plus" @click="handleAdd">新增套餐</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="套餐名称" min-width="150" />
        <el-table-column prop="game_name" label="所属游戏" width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.type === 'duration'" type="primary">时长套餐</el-tag>
            <el-tag v-else type="success">局数套餐</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration_hours" label="时长(小时)" width="100" />
        <el-table-column prop="games_count" label="局数" width="80" />
        <el-table-column prop="price" label="价格(元)" width="100">
          <template #default="{ row }">
            <span class="price-text">¥{{ (row.price / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="original_price" label="原价(元)" width="100">
          <template #default="{ row }">
            <span class="original-price">¥{{ (row.original_price / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="warning" size="small" @click="handleToggle(row)">
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑套餐' : '新增套餐'" width="500px">
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="套餐名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入套餐名称" />
        </el-form-item>
        <el-form-item label="所属游戏" prop="game_id">
          <el-select v-model="formData.game_id" placeholder="请选择游戏" style="width: 100%">
            <el-option
              v-for="game in gameList"
              :key="game.id"
              :label="game.name"
              :value="game.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="套餐类型" prop="type">
          <el-radio-group v-model="formData.type">
            <el-radio value="duration">时长套餐</el-radio>
            <el-radio value="games">局数套餐</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="formData.type === 'duration'" label="时长(小时)" prop="duration_hours">
          <el-input-number v-model="formData.duration_hours" :min="0.5" :step="0.5" />
        </el-form-item>
        <el-form-item v-if="formData.type === 'games'" label="局数" prop="games_count">
          <el-input-number v-model="formData.games_count" :min="1" :step="1" />
        </el-form-item>
        <el-form-item label="价格(元)" prop="price">
          <el-input-number v-model="formData.price" :min="0" :step="1" :precision="2" />
        </el-form-item>
        <el-form-item label="原价(元)" prop="original_price">
          <el-input-number v-model="formData.original_price" :min="0" :step="1" :precision="2" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="formData.sort" :min="0" :step="1" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const gameList = ref([])

const searchForm = reactive({
  gameId: '',
  status: ''
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const formData = reactive({
  id: 0,
  name: '',
  game_id: null,
  type: 'duration',
  duration_hours: 1,
  games_count: 1,
  price: 0,
  original_price: 0,
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  game_id: [{ required: true, message: '请选择游戏', trigger: 'change' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
}

const fetchGameList = async () => {
  try {
    const res = await request.get('/game/list', { params: { status: 1, page: 1, limit: 100 } })
    gameList.value = res.data.list || []
  } catch (e) {
    console.error('获取游戏列表失败', e)
  }
}

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }
    if (searchForm.gameId) params.game_id = searchForm.gameId
    if (searchForm.status !== '') params.status = searchForm.status

    const res = await request.get('/order/package/list', { params })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error('获取套餐列表失败', e)
    ElMessage.error('获取套餐列表失败')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  searchForm.gameId = ''
  searchForm.status = ''
  currentPage.value = 1
  fetchList()
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(formData, {
    id: 0,
    name: '',
    game_id: null,
    type: 'duration',
    duration_hours: 1,
    games_count: 1,
    price: 0,
    original_price: 0,
    sort: 0,
    status: 1
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(formData, {
    ...row,
    price: row.price / 100,
    original_price: row.original_price / 100
  })
  dialogVisible.value = true
}

const handleToggle = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要${row.status === 1 ? '禁用' : '启用'}该套餐吗？`, '提示', {
      type: 'warning'
    })
    await request.put('/order/package/toggle', { id: row.id })
    ElMessage.success('操作成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该套餐吗？', '提示', { type: 'warning' })
    await request.delete('/order/package/delete', { data: { id: row.id } })
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const data = { ...formData }
      if (isEdit.value) {
        await request.put('/order/package/update', data)
        ElMessage.success('更新成功')
      } else {
        await request.post('/order/package/create', data)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchList()
    } catch (e) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    }
  })
}

onMounted(() => {
  fetchGameList()
  fetchList()
})
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.page-header {
  margin-bottom: 16px;
}
.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}
.search-card {
  margin-bottom: 16px;
}
.table-card {
  margin-bottom: 16px;
}
.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.price-text {
  color: #f56c6c;
  font-weight: 600;
}
.original-price {
  color: #909399;
  text-decoration: line-through;
}
</style>
