<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">竞价订单列表</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="订单号">
          <el-input
            v-model="searchForm.orderId"
            placeholder="请输入订单ID"
            clearable
            style="width: 180px"
            @keyup.enter="fetchList"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="竞价中" :value="0" />
            <el-option label="已中标" :value="1" />
            <el-option label="已取消" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_id" label="订单ID" width="100" />
        <el-table-column prop="order_sn" label="订单号" width="180" />
        <el-table-column prop="player_user_id" label="打手ID" width="100" />
        <el-table-column prop="player_nickname" label="打手昵称" width="120">
          <template #default="{ row }">
            <div class="player-cell">
              <el-avatar :size="32" :src="row.player_avatar">
                {{ row.player_nickname ? row.player_nickname.charAt(0) : '?' }}
              </el-avatar>
              <span class="player-name">{{ row.player_nickname }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="bid_price" label="出价(元)" width="120">
          <template #default="{ row }">
            <span class="bid-price">¥{{ (row.bid_price / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="bid_time" label="出价时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_winner" label="是否中标" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_winner" type="success" effect="dark">
              <el-icon :size="12"><Trophy /></el-icon>
              中标
            </el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="180" />
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Trophy } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const searchForm = reactive({
  orderId: '',
  status: ''
})

const getStatusText = (status) => {
  const map = {
    0: '竞价中',
    1: '已中标',
    2: '已取消'
  }
  return map[status] || '未知'
}

const getStatusTagType = (status) => {
  const map = {
    0: 'primary',
    1: 'success',
    2: 'info'
  }
  return map[status] || 'info'
}

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }
    if (searchForm.orderId) params.order_id = searchForm.orderId
    if (searchForm.status !== '') params.status = searchForm.status

    const res = await request.get('/order/bid/list', { params })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error('获取竞价列表失败', e)
    ElMessage.error('获取竞价列表失败')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  searchForm.orderId = ''
  searchForm.status = ''
  currentPage.value = 1
  fetchList()
}

onMounted(() => {
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
.player-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.player-name {
  font-size: 14px;
  color: #303133;
}
.bid-price {
  color: #f56c6c;
  font-weight: 600;
}
.text-muted {
  color: #909399;
}
</style>
