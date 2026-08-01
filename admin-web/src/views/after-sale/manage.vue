<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">售后介入管理</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="介入状态">
          <el-select v-model="searchForm.interventionStatus" placeholder="全部" clearable style="width: 150px">
            <el-option label="待介入" value="pending" />
            <el-option label="已介入" value="intervened" />
            <el-option label="已解决" value="resolved" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否高风险">
          <el-select v-model="searchForm.isHighRisk" placeholder="全部" clearable style="width: 120px">
            <el-option label="是" :value="true" />
            <el-option label="否" :value="false" />
          </el-select>
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
        <el-button type="primary" :icon="Download" :loading="exportLoading" @click="handleExport">
          导出介入记录
        </el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="sessionId" label="会话编号" width="180" show-overflow-tooltip />
        <el-table-column prop="orderId" label="订单ID" width="120" show-overflow-tooltip />
        <el-table-column prop="playerName" label="申诉玩家" min-width="120" show-overflow-tooltip />
        <el-table-column label="介入状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.interventionStatus)" size="small">
              {{ statusLabel(row.interventionStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="高风险" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isHighRisk ? 'danger' : 'success'" size="small">
              {{ row.isHighRisk ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">详情</el-button>
            <el-button type="primary" link size="small" @click="handleViewRecords(row)">介入记录</el-button>
            <el-button
              v-if="row.interventionStatus === 'pending'"
              type="warning"
              link
              size="small"
              @click="handleProcess(row)"
            >
              处理
            </el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="会话详情"
      width="780px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading">
        <el-descriptions :column="2" border size="small" style="margin-bottom: 20px;">
          <el-descriptions-item label="会话编号">{{ sessionDetail.sessionId }}</el-descriptions-item>
          <el-descriptions-item label="订单ID">{{ sessionDetail.orderId }}</el-descriptions-item>
          <el-descriptions-item label="申诉玩家">{{ sessionDetail.playerName }}</el-descriptions-item>
          <el-descriptions-item label="介入状态">
            <el-tag :type="statusTag(sessionDetail.interventionStatus)" size="small">
              {{ statusLabel(sessionDetail.interventionStatus) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="高风险">
            <el-tag :type="sessionDetail.isHighRisk ? 'danger' : 'success'" size="small">
              {{ sessionDetail.isHighRisk ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ sessionDetail.createdAt }}</el-descriptions-item>
        </el-descriptions>

        <h4 class="section-title">申诉原因</h4>
        <el-alert
          v-if="sessionDetail.appealReason"
          :title="sessionDetail.appealReason"
          type="warning"
          :closable="false"
          show-icon
        />
        <el-empty v-else description="暂无申诉原因" :image-size="40" />

        <h4 class="section-title" style="margin-top: 20px;">聊天消息记录</h4>
        <div class="message-list">
          <div
            v-for="(msg, index) in chatMessages"
            :key="index"
            class="message-item"
          >
            <div class="message-header">
              <span class="message-sender" :class="msg.role">{{ msg.role === 'player' ? '玩家' : '用户' }} - {{ msg.senderName }}</span>
              <span class="message-time">{{ msg.time }}</span>
            </div>
            <div class="message-body">
              <div class="message-text" v-if="msg.textContent">{{ msg.textContent }}</div>
            </div>
          </div>
          <el-empty v-if="chatMessages.length === 0 && !detailLoading" description="暂无消息" />
        </div>
      </div>
    </el-dialog>

    <!-- 介入记录弹窗 -->
    <el-dialog
      v-model="recordsDialogVisible"
      title="介入记录"
      width="700px"
      :close-on-click-modal="false"
    >
      <el-table :data="interventionRecords" v-loading="recordsLoading" stripe border style="width: 100%">
        <el-table-column label="触发方式" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.triggerType === 'auto' ? 'warning' : 'primary'" size="small">
              {{ row.triggerType === 'auto' ? '自动触发' : '手动触发' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="triggerDetail" label="触发详情" min-width="180" show-overflow-tooltip />
        <el-table-column label="处理结果" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="resultTag(row.result)" size="small">
              {{ resultLabel(row.result) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operator" label="操作人" width="100" />
        <el-table-column prop="createdAt" label="时间" width="170" align="center" />
      </el-table>
      <el-empty v-if="interventionRecords.length === 0 && !recordsLoading" description="暂无介入记录" />
    </el-dialog>

    <!-- 处理纠纷弹窗 -->
    <el-dialog
      v-model="processDialogVisible"
      title="处理纠纷"
      width="520px"
      :close-on-click-modal="false"
      @closed="handleProcessDialogClosed"
    >
      <el-form ref="processFormRef" :model="processForm" :rules="processFormRules" label-width="100px">
        <el-form-item label="处理结果" prop="result">
          <el-input
            v-model="processForm.result"
            type="textarea"
            :rows="3"
            placeholder="请输入处理结果说明"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="处理动作" prop="action">
          <el-select v-model="processForm.action" placeholder="请选择处理动作" style="width: 100%">
            <el-option label="调解" value="mediation" />
            <el-option label="退款" value="refund" />
            <el-option label="处罚" value="penalty" />
            <el-option label="驳回" value="dismiss" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="handleProcessSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'AfterSaleManage',
  data() {
    return {
      Search,
      Refresh,
      Download,
      searchForm: {
        interventionStatus: '',
        isHighRisk: ''
      },
      tableData: [],
      loading: false,
      exportLoading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      detailDialogVisible: false,
      detailLoading: false,
      currentSessionId: null,
      sessionDetail: {},
      chatMessages: [],
      recordsDialogVisible: false,
      recordsLoading: false,
      interventionRecords: [],
      processDialogVisible: false,
      processLoading: false,
      processForm: {
        result: '',
        action: 'mediation'
      },
      processFormRules: {
        result: [
          { required: true, message: '请输入处理结果', trigger: 'blur' }
        ],
        action: [
          { required: true, message: '请选择处理动作', trigger: 'change' }
        ]
      }
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
          interventionStatus: this.searchForm.interventionStatus || undefined,
          isHighRisk: this.searchForm.isHighRisk !== '' ? this.searchForm.isHighRisk : undefined
        }
        const res = await request.get('/after-sale/sessions', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取售后会话列表失败:', err)
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
        interventionStatus: '',
        isHighRisk: ''
      }
      this.handleSearch()
    },
    async handleViewDetail(row) {
      this.currentSessionId = row.sessionId
      this.detailDialogVisible = true
      this.detailLoading = true
      try {
        const res = await request.get(`/after-sale/sessions/${row.sessionId}`)
        this.sessionDetail = res.data || {}
        this.chatMessages = res.data?.messages || []
      } catch (err) {
        console.error('获取会话详情失败:', err)
      } finally {
        this.detailLoading = false
      }
    },
    async handleViewRecords(row) {
      this.currentSessionId = row.sessionId
      this.recordsDialogVisible = true
      this.recordsLoading = true
      try {
        const res = await request.get(`/after-sale/sessions/${row.sessionId}/records`)
        this.interventionRecords = res.data?.list || []
      } catch (err) {
        console.error('获取介入记录失败:', err)
      } finally {
        this.recordsLoading = false
      }
    },
    handleProcess(row) {
      this.currentSessionId = row.sessionId
      this.processForm = {
        result: '',
        action: 'mediation'
      }
      this.processDialogVisible = true
    },
    handleProcessDialogClosed() {
      this.$refs.processFormRef?.resetFields()
    },
    async handleProcessSubmit() {
      try {
        await this.$refs.processFormRef.validate()
      } catch {
        return
      }
      this.processLoading = true
      try {
        await request.post(`/after-sale/sessions/${this.currentSessionId}/process`, {
          result: this.processForm.result,
          action: this.processForm.action
        })
        ElMessage.success('处理成功')
        this.processDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('处理失败:', err)
      } finally {
        this.processLoading = false
      }
    },
    async handleExport() {
      try {
        await ElMessageBox.confirm(
          '确定要导出当前筛选条件下的介入记录吗？',
          '导出确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
        )
        this.exportLoading = true
        const params = {
          interventionStatus: this.searchForm.interventionStatus || undefined,
          isHighRisk: this.searchForm.isHighRisk !== '' ? this.searchForm.isHighRisk : undefined
        }
        const res = await request.get('/after-sale/sessions/export', {
          params,
          responseType: 'blob'
        })
        const url = window.URL.createObjectURL(new Blob([res]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `售后介入记录_${new Date().toISOString().slice(0, 10)}.xlsx`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        ElMessage.success('导出成功')
      } catch (err) {
        if (err !== 'cancel') {
          console.error('导出失败:', err)
        }
      } finally {
        this.exportLoading = false
      }
    },
    statusTag(status) {
      const map = { pending: 'warning', intervened: 'primary', resolved: 'success', closed: 'info' }
      return map[status] || 'info'
    },
    statusLabel(status) {
      const map = { pending: '待介入', intervened: '已介入', resolved: '已解决', closed: '已关闭' }
      return map[status] || status
    },
    resultTag(result) {
      const map = { mediation: 'primary', refund: 'success', penalty: 'danger', dismiss: 'info' }
      return map[result] || 'info'
    },
    resultLabel(result) {
      const map = { mediation: '调解', refund: '退款', penalty: '处罚', dismiss: '驳回' }
      return map[result] || result
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

.table-card {
  .table-toolbar {
    margin-bottom: 16px;
  }
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #303133;
}

.message-list {
  max-height: 300px;
  overflow-y: auto;
}

.message-item {
  padding: 10px 12px;
  border-bottom: 1px solid #ebeef5;

  &:last-child {
    border-bottom: none;
  }
}

.message-header {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
}

.message-sender {
  font-weight: 600;
  font-size: 13px;

  &.player {
    color: #e6a23c;
  }

  &.user {
    color: #409eff;
  }
}

.message-time {
  margin-left: auto;
  font-size: 12px;
  color: #909399;
}

.message-body {
  font-size: 13px;
  color: #303133;
}

.message-text {
  line-height: 1.6;
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>