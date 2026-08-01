<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">订单管理</span>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="tab-card" @tab-change="handleTabChange">
      <el-tab-pane label="全部订单" name="all" />
      <el-tab-pane label="大额验证失败" name="large_failed" />
    </el-tabs>

    <!-- 搜索栏 -->
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
        <el-form-item label="用户手机号">
          <el-input
            v-model="searchForm.userPhone"
            placeholder="请输入手机号"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="打手昵称">
          <el-input
            v-model="searchForm.dispatcherNickname"
            placeholder="请输入打手昵称"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="订单状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部"
            multiple
            clearable
            collapse-tags
            collapse-tags-tooltip
            style="width: 200px"
          >
            <el-option
              v-for="s in statusOptions"
              :key="s.value"
              :label="s.label"
              :value="s.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="是否大额验证失败">
          <el-select v-model="searchForm.largeVerifyFailed" placeholder="全部" clearable style="width: 140px">
            <el-option label="是" :value="true" />
            <el-option label="否" :value="false" />
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

    <!-- 表格 -->
    <el-card class="table-card">
      <div class="table-toolbar">
        <el-button
          type="warning"
          :disabled="selectedRows.length === 0"
          @click="handleBatchOperation"
        >
          批量操作（已选 {{ selectedRows.length }} 条）
        </el-button>
      </div>
      <el-table
        ref="tableRef"
        :data="tableData"
        v-loading="loading"
        stripe
        border
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="orderNo" label="订单号" width="180" show-overflow-tooltip />
        <el-table-column label="下单用户" min-width="130">
          <template #default="{ row }">
            <div>{{ row.userNickname }}</div>
            <div style="color: #909399; font-size: 12px;">{{ maskPhone(row.userPhone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="打手" min-width="130">
          <template #default="{ row }">
            <div>{{ row.dispatcherNickname || '-' }}</div>
            <div style="color: #909399; font-size: 12px;">{{ maskPhone(row.dispatcherPhone) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="serviceType" label="服务类型" width="120" align="center" />
        <el-table-column prop="amount" label="金额(元)" width="100" align="center">
          <template #default="{ row }">
            <span style="color: #f56c6c; font-weight: 600;">¥{{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="orderStatusTag(row.status)" size="small">
              {{ orderStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
            <el-button type="warning" link size="small" @click="handleForceStatus(row)">强制扭转</el-button>
            <el-button type="danger" link size="small" @click="handleRefund(row)">退款</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
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

    <!-- 强制扭转状态弹窗 -->
    <el-dialog
      v-model="forceStatusDialogVisible"
      title="强制扭转状态"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="forceStatusFormRef" :model="forceStatusForm" :rules="forceStatusRules" label-width="90px">
        <el-form-item label="当前状态">
          <el-tag :type="orderStatusTag(currentForceRow?.status)" size="default">
            {{ orderStatusLabel(currentForceRow?.status) }}
          </el-tag>
        </el-form-item>
        <el-form-item label="目标状态" prop="targetStatus">
          <el-select v-model="forceStatusForm.targetStatus" placeholder="请选择目标状态" style="width: 100%">
            <el-option
              v-for="s in statusOptions"
              :key="s.value"
              :label="s.label"
              :value="s.value"
              :disabled="s.value === currentForceRow?.status"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="变更原因" prop="reason">
          <el-input
            v-model="forceStatusForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入变更原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="forceStatusDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="forceStatusLoading" @click="handleForceStatusSubmit">
          确认变更
        </el-button>
      </template>
    </el-dialog>

    <!-- 退款弹窗 -->
    <el-dialog
      v-model="refundDialogVisible"
      title="订单退款"
      width="450px"
      :close-on-click-modal="false"
    >
      <el-form ref="refundFormRef" :model="refundForm" :rules="refundRules" label-width="90px">
        <el-form-item label="订单号">
          <span>{{ currentRefundRow?.orderNo }}</span>
        </el-form-item>
        <el-form-item label="订单金额">
          <span style="color: #f56c6c; font-weight: 600;">¥{{ currentRefundRow?.amount }}</span>
        </el-form-item>
        <el-form-item label="退款金额" prop="refundAmount">
          <el-input-number
            v-model="refundForm.refundAmount"
            :min="0"
            :max="currentRefundRow?.amount"
            :precision="2"
            style="width: 100%"
            placeholder="请输入退款金额"
          />
        </el-form-item>
        <el-form-item label="退款原因" prop="reason">
          <el-input
            v-model="refundForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入退款原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="refundLoading" @click="handleRefundSubmit">
          确认退款
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量操作确认弹窗 -->
    <BatchConfirm
      v-model="batchConfirmVisible"
      :operation-type="batchOperationType"
      :affected-count="selectedRows.length"
      :order-ids="selectedOrderIds"
      @confirmed="handleBatchConfirmed"
    />
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import BatchConfirm from './components/BatchConfirm.vue'

export default {
  name: 'OrderList',
  components: {
    BatchConfirm
  },
  data() {
    return {
      Search,
      Refresh,
      activeTab: 'all',
      searchForm: {
        orderNo: '',
        userPhone: '',
        dispatcherNickname: '',
        status: [],
        largeVerifyFailed: '',
        dateRange: []
      },
      statusOptions: [
        { value: 'created', label: '已创建' },
        { value: 'accepted', label: '已接单' },
        { value: 'processing', label: '进行中' },
        { value: 'completed', label: '已完成' },
        { value: 'settled', label: '已结算' },
        { value: 'refunded', label: '已退款' },
        { value: 'cancelled', label: '已取消' }
      ],
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      selectedRows: [],
      forceStatusDialogVisible: false,
      currentForceRow: null,
      forceStatusForm: {
        targetStatus: '',
        reason: ''
      },
      forceStatusRules: {
        targetStatus: [{ required: true, message: '请选择目标状态', trigger: 'change' }],
        reason: [{ required: true, message: '请输入变更原因', trigger: 'blur' }]
      },
      forceStatusLoading: false,
      refundDialogVisible: false,
      currentRefundRow: null,
      refundForm: {
        refundAmount: null,
        reason: ''
      },
      refundRules: {
        refundAmount: [{ required: true, message: '请输入退款金额', trigger: 'blur' }],
        reason: [{ required: true, message: '请输入退款原因', trigger: 'blur' }]
      },
      refundLoading: false,
      batchConfirmVisible: false,
      batchOperationType: ''
    }
  },
  computed: {
    selectedOrderIds() {
      return this.selectedRows.map(r => r.id)
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
          orderNo: this.searchForm.orderNo || undefined,
          userPhone: this.searchForm.userPhone || undefined,
          dispatcherNickname: this.searchForm.dispatcherNickname || undefined,
          status: this.searchForm.status.length > 0 ? this.searchForm.status.join(',') : undefined,
          largeVerifyFailed: this.searchForm.largeVerifyFailed === '' ? undefined : this.searchForm.largeVerifyFailed,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined,
          tab: this.activeTab === 'large_failed' ? 'large_failed' : undefined
        }
        const res = await request.get('/orders', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取订单列表失败:', err)
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
        userPhone: '',
        dispatcherNickname: '',
        status: [],
        largeVerifyFailed: '',
        dateRange: []
      }
      this.handleSearch()
    },
    handleTabChange() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleSelectionChange(rows) {
      this.selectedRows = rows
    },
    handleDetail(row) {
      this.$router.push(`/order/detail/${row.id}`)
    },
    handleForceStatus(row) {
      this.currentForceRow = row
      this.forceStatusForm = {
        targetStatus: '',
        reason: ''
      }
      this.forceStatusDialogVisible = true
    },
    async handleForceStatusSubmit() {
      const valid = await this.$refs.forceStatusFormRef.validate().catch(() => false)
      if (!valid) return
      this.forceStatusLoading = true
      try {
        await request.post(`/orders/${this.currentForceRow.id}/force-status`, {
          targetStatus: this.forceStatusForm.targetStatus,
          reason: this.forceStatusForm.reason
        })
        ElMessage.success('状态变更成功')
        this.forceStatusDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('状态变更失败:', err)
      } finally {
        this.forceStatusLoading = false
      }
    },
    handleRefund(row) {
      this.currentRefundRow = row
      this.refundForm = {
        refundAmount: row.amount,
        reason: ''
      }
      this.refundDialogVisible = true
    },
    async handleRefundSubmit() {
      const valid = await this.$refs.refundFormRef.validate().catch(() => false)
      if (!valid) return
      this.refundLoading = true
      try {
        await request.post(`/orders/${this.currentRefundRow.id}/refund`, {
          refundAmount: this.refundForm.refundAmount,
          reason: this.refundForm.reason
        })
        ElMessage.success('退款成功')
        this.refundDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('退款失败:', err)
      } finally {
        this.refundLoading = false
      }
    },
    handleBatchOperation() {
      this.batchOperationType = 'batch_status_change'
      this.batchConfirmVisible = true
    },
    handleBatchConfirmed() {
      this.batchConfirmVisible = false
      this.selectedRows = []
      this.$refs.tableRef?.clearSelection()
      this.fetchList()
    },
    maskPhone(phone) {
      if (!phone) return '-'
      return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    },
    orderStatusTag(status) {
      const map = {
        created: 'info',
        accepted: 'warning',
        processing: 'warning',
        completed: 'success',
        settled: '',
        refunded: 'danger',
        cancelled: 'info'
      }
      return map[status] || 'info'
    },
    orderStatusLabel(status) {
      const map = {
        created: '已创建',
        accepted: '已接单',
        processing: '进行中',
        completed: '已完成',
        settled: '已结算',
        refunded: '已退款',
        cancelled: '已取消'
      }
      return map[status] || '未知'
    }
  }
}
</script>

<style lang="scss" scoped>
.tab-card {
  margin-bottom: 16px;
  background: #fff;
  padding: 0 20px;
}

.search-card {
  margin-bottom: 16px;
}

.search-form-inline {
  display: flex;
  flex-wrap: wrap;
}

.table-card {
  .table-toolbar {
    margin-bottom: 16px;
  }
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>