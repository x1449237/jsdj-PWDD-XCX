<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">内部订单监控</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="俱乐部ID">
          <el-input v-model="searchForm.club_id" placeholder="俱乐部ID" clearable style="width: 120px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待接单" :value="1" />
            <el-option label="已接单" :value="2" />
            <el-option label="进行中" :value="3" />
            <el-option label="已完成" :value="4" />
            <el-option label="已取消" :value="5" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="标题/订单号" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="order_no" label="订单号" width="180" />
        <el-table-column prop="club_name" label="俱乐部" min-width="140" />
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column label="酬金" width="110">
          <template #default="{ row }">¥{{ row.reward_yuan }}</template>
        </el-table-column>
        <el-table-column prop="player_nickname" label="接单人" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ row.status_name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="170" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleDetail(row)">详情</el-button>
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

    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <el-descriptions :column="2" border v-if="detailData">
        <el-descriptions-item label="订单号">{{ detailData.order_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detailData.status)" size="small">
            {{ detailData.status_name }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="俱乐部">{{ detailData.club_name }}</el-descriptions-item>
        <el-descriptions-item label="接单人">{{ detailData.player_nickname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="酬金">¥{{ detailData.reward_yuan }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detailData.create_time }}</el-descriptions-item>
        <el-descriptions-item label="订单标题" :span="2">{{ detailData.title }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)

const searchForm = reactive({
  club_id: '',
  status: '',
  keyword: '',
})

const detailVisible = ref(false)
const detailData = ref(null)

const statusTagType = (status) => {
  const map = {
    1: 'warning',
    2: 'primary',
    3: 'info',
    4: 'success',
    5: 'danger',
  }
  return map[status] || 'info'
}

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      limit: limit.value,
      ...searchForm,
    }
    if (!params.club_id) delete params.club_id
    if (params.status === '') delete params.status
    if (!params.keyword) delete params.keyword
    const res = await request.get('/club/internal-order/list', params)
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

const handleDetail = async (row) => {
  try {
    const res = await request.get('/club/internal-order/detail', { id: row.id })
    detailData.value = res
    detailVisible.value = true
  } catch (e) {
    ElMessage.error(e.message || '获取详情失败')
  }
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
