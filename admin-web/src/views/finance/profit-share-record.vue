<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">分账记录</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="订单号">
          <el-input
            v-model="searchForm.orderNo"
            placeholder="请输入订单号"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input
            v-model="searchForm.userId"
            placeholder="请输入用户ID"
            clearable
            style="width: 140px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="全部" clearable style="width: 120px">
            <el-option label="打手" :value="1" />
            <el-option label="俱乐部" :value="2" />
            <el-option label="分销商" :value="3" />
            <el-option label="平台" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待结算" :value="0" />
            <el-option label="已结算" :value="1" />
            <el-option label="已冻结" :value="2" />
            <el-option label="已退款" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <div class="table-toolbar">
        <el-button type="success" :icon="Check" @click="handleBatchSettle" :disabled="selectedIds.length === 0">
          批量结算
        </el-button>
        <span style="margin-left: 16px; color: #909399;">
          已选 {{ selectedIds.length }} 条
        </span>
      </div>
      <el-table
        :data="tableData"
        v-loading="loading"
        stripe
        border
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="orderNo" label="订单号" width="180" show-overflow-tooltip />
        <el-table-column label="用户信息" min-width="130">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '-' }}</div>
            <div style="color: #909399; font-size: 12px;">ID: {{ row.userId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small">{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="分账金额" width="120" align="center">
          <template #default="{ row }">
            <span style="color: #67c23a; font-weight: 600;">¥{{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ratio" label="比例" width="90" align="center">
          <template #default="{ row }">{{ row.ratio }}%</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="shareTime" label="分账时间" width="170" align="center" />
        <el-table-column prop="createTime" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 0"
              type="success"
              link
              size="small"
              @click="handleSettle(row)"
            >结算</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="分账详情" width="500px">
      <el-descriptions :column="2" border v-if="currentRow">
        <el-descriptions-item label="记录ID">{{ currentRow.id }}</el-descriptions-item>
        <el-descriptions-item label="订单号">{{ currentRow.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="用户">
          {{ currentRow.user?.nickname || '-' }} (ID: {{ currentRow.userId }})
        </el-descriptions-item>
        <el-descriptions-item label="角色">{{ roleLabel(currentRow.role) }}</el-descriptions-item>
        <el-descriptions-item label="分账金额">
          <span style="color: #67c23a; font-weight: 600;">¥{{ currentRow.amount }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="分账比例">{{ currentRow.ratio }}%</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(currentRow.status)" size="small">
            {{ statusLabel(currentRow.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="交易号">{{ currentRow.transactionId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="结算批次">{{ currentRow.settleBatchNo || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentRow.createTime }}</el-descriptions-item>
        <el-descriptions-item label="分账时间">{{ currentRow.shareTime || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ currentRow.remark || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Check } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ProfitShareRecord',
  data() {
    return {
      Search,
      Refresh,
      Check,
      searchForm: {
        orderNo: '',
        userId: '',
        role: '',
        status: '',
        dateRange: []
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      selectedIds: [],
      detailVisible: false,
      currentRow: null
    }
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          order_no: this.searchForm.orderNo || undefined,
          user_id: this.searchForm.userId || undefined,
          role: this.searchForm.role || undefined,
          status: this.searchForm.status !== '' ? this.searchForm.status : undefined,
          start_date: this.searchForm.dateRange?.[0] || undefined,
          end_date: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/profit_share/record_list', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取分账记录列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.searchForm = {
        orderNo: '',
        userId: '',
        role: '',
        status: '',
        dateRange: []
      }
      this.handleSearch()
    },
    handleSelectionChange(selection) {
      this.selectedIds = selection.filter(item => item.status === 0).map(item => item.id)
    },
    roleLabel(role) {
      const map = { 1: '打手', 2: '俱乐部', 3: '分销商', 4: '平台' }
      return map[role] || '未知'
    },
    roleTagType(role) {
      const map = { 1: '', 2: 'warning', 3: 'success', 4: 'danger' }
      return map[role] || 'info'
    },
    statusLabel(status) {
      const map = { 0: '待结算', 1: '已结算', 2: '已冻结', 3: '已退款' }
      return map[status] || '未知'
    },
    statusTagType(status) {
      const map = { 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }
      return map[status] || 'info'
    },
    handleDetail(row) {
      this.currentRow = row
      this.detailVisible = true
    },
    handleSettle(row) {
      ElMessageBox.confirm(
        `确定要结算该分账记录吗？金额: ¥${row.amount}`,
        '结算确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
      ).then(async () => {
        try {
          await request.put('/profit_share/record_settle', { id: row.id })
          ElMessage.success('结算成功')
          this.fetchList()
        } catch (err) {
          console.error('结算失败:', err)
        }
      }).catch(() => {})
    },
    handleBatchSettle() {
      if (this.selectedIds.length === 0) return
      ElMessageBox.confirm(
        `确定要批量结算选中的 ${this.selectedIds.length} 条记录吗？`,
        '批量结算确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
      ).then(async () => {
        try {
          await request.post('/profit_share/record_batch_settle', {
            ids: this.selectedIds.join(',')
          })
          ElMessage.success('批量结算完成')
          this.fetchList()
        } catch (err) {
          console.error('批量结算失败:', err)
        }
      }).catch(() => {})
    }
  }
}
</script>

<style lang="scss" scoped>
.search-card {
  margin-bottom: 16px;
}

.search-form-inline {
  display: flex;
  flex-wrap: wrap;
}

.table-toolbar {
  margin-bottom: 16px;
}
</style>
