<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">服务保证金管理</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="正常" value="active" />
            <el-option label="冻结" value="frozen" />
            <el-option label="已退出" value="withdrawn" />
          </el-select>
        </el-form-item>
        <el-form-item label="打手ID">
          <el-input v-model="searchForm.keyword" placeholder="请输入打手ID" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="loadData">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="player_user_id" label="打手ID" width="120" align="center" />
        <el-table-column label="保证金总额" width="140" align="center">
          <template #default="{ row }">
            <span class="amount">¥{{ (row.amount / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="冻结金额" width="140" align="center">
          <template #default="{ row }">
            <span class="freeze-amount">¥{{ (row.freeze_amount / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="可用余额" width="140" align="center">
          <template #default="{ row }">
            <span class="available-amount">¥{{ ((row.amount - row.freeze_amount) / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="320" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewLog(row)">流水</el-button>
            <el-button type="success" link size="small" @click="handleDeposit(row)">充值</el-button>
            <el-button type="warning" link size="small" @click="handleDeduct(row)">扣除</el-button>
            <el-button type="info" link size="small" @click="handleFreeze(row)">冻结</el-button>
            <el-button type="success" link size="small" @click="handleUnfreeze(row)">解冻</el-button>
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
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <el-dialog v-model="logDialogVisible" title="保证金流水" width="800px">
      <el-table :data="logList" v-loading="logLoading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="logTypeTag(row.type)" size="small">
              {{ logTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变动金额" width="120" align="center">
          <template #default="{ row }">
            <span :class="row.amount > 0 ? 'amount-up' : 'amount-down'">
              {{ row.amount > 0 ? '+' : '' }}{{ (row.amount / 100).toFixed(2) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="余额" width="120" align="center">
          <template #default="{ row }">
            ¥{{ (row.balance / 100).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column prop="related_id" label="关联ID" width="100" align="center" />
        <el-table-column prop="create_time" label="时间" width="170" align="center" />
      </el-table>
    </el-dialog>

    <el-dialog v-model="actionDialogVisible" :title="actionTitle" width="400px">
      <el-form :model="actionForm" label-width="80px">
        <el-form-item label="金额(元)">
          <el-input-number v-model="actionForm.amount" :min="0.01" :precision="2" :step="10" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="actionForm.description" type="textarea" :rows="2" placeholder="请输入说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="actionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="actionLoading" @click="confirmAction">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'ServiceDeposit',
  components: { Search, Refresh },
  data() {
    return {
      loading: false,
      searchForm: {
        status: '',
        keyword: ''
      },
      tableData: [],
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      logDialogVisible: false,
      logLoading: false,
      logList: [],
      actionDialogVisible: false,
      actionLoading: false,
      actionType: '',
      currentRow: null,
      actionForm: {
        amount: 0,
        description: ''
      }
    }
  },
  computed: {
    actionTitle() {
      const map = {
        deposit: '保证金充值',
        deduct: '保证金扣除',
        refund: '保证金退还',
        freeze: '保证金冻结',
        unfreeze: '保证金解冻'
      }
      return map[this.actionType] || '操作'
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    loadData() {
      this.loading = true
      request.get('/service_deposit/list', {
        params: {
          status: this.searchForm.status,
          keyword: this.searchForm.keyword,
          page: this.pagination.page,
          page_size: this.pagination.pageSize
        }
      }).then(res => {
        this.tableData = res.list || []
        this.pagination.total = res.total || 0
      }).finally(() => {
        this.loading = false
      })
    },
    handleReset() {
      this.searchForm = { status: '', keyword: '' }
      this.pagination.page = 1
      this.loadData()
    },
    handleViewLog(row) {
      this.logDialogVisible = true
      this.logLoading = true
      request.get('/service_deposit/log_list', {
        params: { player_user_id: row.player_user_id }
      }).then(res => {
        this.logList = res.list || []
      }).finally(() => {
        this.logLoading = false
      })
    },
    handleDeposit(row) {
      this.actionType = 'deposit'
      this.currentRow = row
      this.actionForm = { amount: 0, description: '' }
      this.actionDialogVisible = true
    },
    handleDeduct(row) {
      this.actionType = 'deduct'
      this.currentRow = row
      this.actionForm = { amount: 0, description: '' }
      this.actionDialogVisible = true
    },
    handleFreeze(row) {
      this.actionType = 'freeze'
      this.currentRow = row
      this.actionForm = { amount: 0, description: '' }
      this.actionDialogVisible = true
    },
    handleUnfreeze(row) {
      this.actionType = 'unfreeze'
      this.currentRow = row
      this.actionForm = { amount: 0, description: '' }
      this.actionDialogVisible = true
    },
    confirmAction() {
      if (this.actionForm.amount <= 0) {
        ElMessage.warning('请输入金额')
        return
      }
      const amountFen = Math.round(this.actionForm.amount * 100)
      const apiMap = {
        deposit: '/service_deposit/manual_deposit',
        deduct: '/service_deposit/manual_deduct',
        refund: '/service_deposit/manual_refund',
        freeze: '/service_deposit/freeze',
        unfreeze: '/service_deposit/unfreeze'
      }
      this.actionLoading = true
      request.post(apiMap[this.actionType], {
        player_user_id: this.currentRow.player_user_id,
        amount: amountFen,
        description: this.actionForm.description
      }).then(() => {
        ElMessage.success('操作成功')
        this.actionDialogVisible = false
        this.loadData()
      }).finally(() => {
        this.actionLoading = false
      })
    },
    statusTag(status) {
      const map = {
        active: 'success',
        frozen: 'warning',
        withdrawn: 'info'
      }
      return map[status] || 'info'
    },
    statusLabel(status) {
      const map = {
        active: '正常',
        frozen: '冻结',
        withdrawn: '已退出'
      }
      return map[status] || status
    },
    logTypeTag(type) {
      const map = {
        deposit: 'success',
        deduct: 'danger',
        refund: 'warning',
        freeze: 'info',
        unfreeze: 'primary'
      }
      return map[type] || 'info'
    },
    logTypeLabel(type) {
      const map = {
        deposit: '充值',
        deduct: '扣除',
        refund: '退还',
        freeze: '冻结',
        unfreeze: '解冻'
      }
      return map[type] || type
    }
  }
}
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.page-header {
  margin-bottom: 16px;
}
.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}
.search-card,
.table-card {
  margin-bottom: 16px;
}
.amount {
  color: #67c23a;
  font-weight: 600;
}
.freeze-amount {
  color: #e6a23c;
  font-weight: 600;
}
.available-amount {
  color: #409eff;
  font-weight: 600;
}
.amount-up {
  color: #67c23a;
}
.amount-down {
  color: #f56c6c;
}
.pagination-container {
  margin-top: 20px;
  text-align: right;
}
</style>
