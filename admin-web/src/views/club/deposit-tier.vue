<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">保证金阶梯配置</span>
      <el-button type="primary" @click="openCreateDialog">新增阶梯</el-button>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="俱乐部类型">
          <el-select v-model="searchForm.club_type" placeholder="全部" clearable style="width: 140px">
            <el-option label="个人俱乐部" value="personal" />
            <el-option label="企业俱乐部" value="enterprise" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="俱乐部类型" width="130">
          <template #default="{ row }">
            <el-tag :type="row.club_type === 'enterprise' ? 'primary' : 'success'" size="small">
              {{ row.club_type === 'enterprise' ? '企业俱乐部' : '个人俱乐部' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tier_name" label="阶梯名称" min-width="140" />
        <el-table-column label="月流水阈值(元)" width="160">
          <template #default="{ row }">{{ row.revenue_threshold_yuan }}</template>
        </el-table-column>
        <el-table-column label="保证金额度(元)" width="160">
          <template #default="{ row }">{{ row.deposit_amount_yuan }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        :current-page="page"
        :page-size="limit"
        @current-change="handlePageChange"
        class="pagination"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="formData" label-width="120px">
        <el-form-item label="俱乐部类型">
          <el-select v-model="formData.club_type" style="width: 100%">
            <el-option label="个人俱乐部" value="personal" />
            <el-option label="企业俱乐部" value="enterprise" />
          </el-select>
        </el-form-item>
        <el-form-item label="阶梯名称">
          <el-input v-model="formData.tier_name" placeholder="如：青铜/白银/黄金" />
        </el-form-item>
        <el-form-item label="月流水阈值(元)">
          <el-input-number v-model="formData.revenue_threshold" :min="0" :precision="2" :step="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="保证金额度(元)">
          <el-input-number v-model="formData.deposit_amount" :min="0" :precision="2" :step="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
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
import request from '@/utils/request'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)

const searchForm = reactive({
  club_type: '',
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const isEdit = ref(false)
const formData = reactive({
  id: 0,
  club_type: 'personal',
  tier_name: '',
  revenue_threshold: 0,
  deposit_amount: 0,
  status: 1,
})

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      limit: limit.value,
    }
    if (searchForm.club_type) params.club_type = searchForm.club_type
    const res = await request.get('/club/deposit-tier/list', params)
    tableData.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error(e.message || '获取列表失败')
  } finally {
    loading.value = false
  }
}

const handlePageChange = (p) => {
  page.value = p
  fetchList()
}

const openCreateDialog = () => {
  isEdit.value = false
  dialogTitle.value = '新增阶梯'
  Object.assign(formData, {
    id: 0,
    club_type: 'personal',
    tier_name: '',
    revenue_threshold: 0,
    deposit_amount: 0,
    status: 1,
  })
  dialogVisible.value = true
}

const openEditDialog = (row) => {
  isEdit.value = true
  dialogTitle.value = '编辑阶梯'
  Object.assign(formData, {
    id: row.id,
    club_type: row.club_type,
    tier_name: row.tier_name,
    revenue_threshold: parseFloat(row.revenue_threshold_yuan),
    deposit_amount: parseFloat(row.deposit_amount_yuan),
    status: row.status,
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.tier_name) {
    ElMessage.warning('请输入阶梯名称')
    return
  }
  try {
    if (isEdit.value) {
      await request.post('/club/deposit-tier/update', formData)
      ElMessage.success('更新成功')
    } else {
      await request.post('/club/deposit-tier/create', formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除阶梯「${row.tier_name}」吗？`, '提示', {
    type: 'warning',
  }).then(async () => {
    try {
      await request.post('/club/deposit-tier/delete', { id: row.id })
      ElMessage.success('删除成功')
      fetchList()
    } catch (e) {
      ElMessage.error(e.message || '删除失败')
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
