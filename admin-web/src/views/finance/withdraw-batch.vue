<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">批量提现</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="批次号">
          <el-input
            v-model="searchForm.batchNo"
            placeholder="请输入批次号"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="提现渠道">
          <el-select v-model="searchForm.channel" placeholder="全部" clearable style="width: 140px">
            <el-option label="微信" :value="1" />
            <el-option label="支付宝" :value="2" />
            <el-option label="银行卡" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已完成" :value="2" />
            <el-option label="部分失败" :value="3" />
            <el-option label="全部失败" :value="4" />
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
        <el-button type="primary" :icon="Plus" @click="handleCreateBatch">创建批次</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="batchNo" label="批次号" width="200" show-overflow-tooltip />
        <el-table-column label="提现渠道" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="channelTagType(row.channel)" size="small">
              {{ channelLabel(row.channel) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="totalCount" label="总笔数" width="100" align="center" />
        <el-table-column prop="totalAmount" label="总金额" width="130" align="center">
          <template #default="{ row }">
            <span style="font-weight: 600;">¥{{ row.totalAmount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="successCount" label="成功笔数" width="100" align="center" />
        <el-table-column prop="successAmount" label="成功金额" width="120" align="center">
          <template #default="{ row }">
            <span style="color: #67c23a;">¥{{ row.successAmount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="failCount" label="失败笔数" width="100" align="center" />
        <el-table-column prop="failAmount" label="失败金额" width="120" align="center">
          <template #default="{ row }">
            <span style="color: #f56c6c;">¥{{ row.failAmount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operatorName" label="操作人" width="100" align="center" />
        <el-table-column prop="createTime" label="创建时间" width="170" align="center" />
        <el-table-column prop="completeTime" label="完成时间" width="170" align="center" />
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 0"
              type="success"
              link
              size="small"
              @click="handleProcess(row)"
            >处理</el-button>
            <el-button
              v-if="row.status === 1"
              type="warning"
              link
              size="small"
              @click="handleComplete(row)"
            >完成</el-button>
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

    <el-dialog v-model="createDialogVisible" title="创建提现批次" width="450px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="提现渠道">
          <el-select v-model="createForm.channel" style="width: 100%">
            <el-option label="微信" :value="1" />
            <el-option label="支付宝" :value="2" />
            <el-option label="银行卡" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="createForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注（选填）"
          />
        </el-form-item>
        <el-form-item label="待处理数量">
          <el-tag type="warning">{{ pendingCount }} 笔</el-tag>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreateSubmit">
          创建批次
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="completeDialogVisible" title="完成提现批次" width="500px">
      <el-form :model="completeForm" label-width="100px">
        <el-form-item label="成功笔数">
          <el-input-number v-model="completeForm.successCount" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="成功金额(元)">
          <el-input-number
            v-model="completeForm.successAmount"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="失败笔数">
          <el-input-number v-model="completeForm.failCount" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="失败金额(元)">
          <el-input-number
            v-model="completeForm.failAmount"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="completeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="completeLoading" @click="handleCompleteSubmit">
          确认完成
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="批次详情" width="500px">
      <el-descriptions v-if="currentRow" :column="2" border>
        <el-descriptions-item label="批次号">{{ currentRow.batchNo }}</el-descriptions-item>
        <el-descriptions-item label="渠道">{{ channelLabel(currentRow.channel) }}</el-descriptions-item>
        <el-descriptions-item label="总笔数">{{ currentRow.totalCount }}</el-descriptions-item>
        <el-descriptions-item label="总金额">¥{{ currentRow.totalAmount }}</el-descriptions-item>
        <el-descriptions-item label="成功笔数">{{ currentRow.successCount }}</el-descriptions-item>
        <el-descriptions-item label="成功金额">
          <span style="color: #67c23a;">¥{{ currentRow.successAmount }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="失败笔数">{{ currentRow.failCount }}</el-descriptions-item>
        <el-descriptions-item label="失败金额">
          <span style="color: #f56c6c;">¥{{ currentRow.failAmount }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(currentRow.status)" size="small">
            {{ statusLabel(currentRow.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作人">{{ currentRow.operatorName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentRow.createTime }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ currentRow.completeTime || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ currentRow.remark || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'WithdrawBatch',
  data() {
    return {
      Search,
      Refresh,
      Plus,
      searchForm: {
        batchNo: '',
        channel: '',
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
      createDialogVisible: false,
      createForm: {
        channel: 1,
        remark: ''
      },
      createLoading: false,
      pendingCount: 0,
      detailVisible: false,
      currentRow: null,
      completeDialogVisible: false,
      completeRow: null,
      completeForm: {
        successCount: 0,
        successAmount: 0,
        failCount: 0,
        failAmount: 0
      },
      completeLoading: false
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
          batch_no: this.searchForm.batchNo || undefined,
          channel: this.searchForm.channel || undefined,
          status: this.searchForm.status !== '' ? this.searchForm.status : undefined,
          start_date: this.searchForm.dateRange?.[0] || undefined,
          end_date: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/finance/withdraw_batch_list', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取提现批次列表失败:', err)
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
        batchNo: '',
        channel: '',
        status: '',
        dateRange: []
      }
      this.handleSearch()
    },
    channelLabel(channel) {
      const map = { 1: '微信', 2: '支付宝', 3: '银行卡' }
      return map[channel] || '未知'
    },
    channelTagType(channel) {
      const map = { 1: 'success', 2: 'warning', 3: '' }
      return map[channel] || 'info'
    },
    statusLabel(status) {
      const map = { 0: '待处理', 1: '处理中', 2: '已完成', 3: '部分失败', 4: '全部失败' }
      return map[status] || '未知'
    },
    statusTagType(status) {
      const map = { 0: 'warning', 1: '', 2: 'success', 3: 'warning', 4: 'danger' }
      return map[status] || 'info'
    },
    handleCreateBatch() {
      this.createForm = { channel: 1, remark: '' }
      this.pendingCount = 0
      this.createDialogVisible = true
    },
    async handleCreateSubmit() {
      this.createLoading = true
      try {
        const res = await request.post('/finance/withdraw_batch_create', {
          channel: this.createForm.channel,
          remark: this.createForm.remark
        })
        ElMessage.success('批次创建成功')
        this.createDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('创建批次失败:', err)
      } finally {
        this.createLoading = false
      }
    },
    handleDetail(row) {
      this.currentRow = row
      this.detailVisible = true
    },
    handleProcess(row) {
      ElMessageBox.confirm(
        `确定要开始处理批次「${row.batchNo}」吗？`,
        '处理确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
      ).then(async () => {
        try {
          await request.put('/finance/withdraw_batch_process', { id: row.id })
          ElMessage.success('已开始处理')
          this.fetchList()
        } catch (err) {
          console.error('处理失败:', err)
        }
      }).catch(() => {})
    },
    handleComplete(row) {
      this.completeRow = row
      this.completeForm = {
        successCount: row.totalCount,
        successAmount: row.totalAmount,
        failCount: 0,
        failAmount: 0
      }
      this.completeDialogVisible = true
    },
    async handleCompleteSubmit() {
      this.completeLoading = true
      try {
        await request.put('/finance/withdraw_batch_complete', {
          id: this.completeRow.id,
          success_count: this.completeForm.successCount,
          success_amount: this.completeForm.successAmount,
          fail_count: this.completeForm.failCount,
          fail_amount: this.completeForm.failAmount
        })
        ElMessage.success('批次已完成')
        this.completeDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('完成失败:', err)
      } finally {
        this.completeLoading = false
      }
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

.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
