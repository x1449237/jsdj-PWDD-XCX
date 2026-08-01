<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">提现管理</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="提现单号">
          <el-input
            v-model="searchForm.withdrawNo"
            placeholder="请输入提现单号"
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
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
            <el-option label="已打款" value="paid" />
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
    <el-card>
      <div class="table-toolbar">
        <el-button type="primary" :icon="Check" @click="handleBankCardVerify">
          银行卡三要素校验
        </el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="withdrawNo" label="提现单号" width="180" show-overflow-tooltip />
        <el-table-column label="用户信息" min-width="130">
          <template #default="{ row }">
            <div>{{ row.userNickname }}</div>
            <div style="color: #909399; font-size: 12px;">{{ maskPhone(row.userPhone) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="提现金额" width="110" align="center">
          <template #default="{ row }">
            <span style="color: #f56c6c; font-weight: 600;">¥{{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="fee" label="手续费" width="90" align="center">
          <template #default="{ row }">
            <span>¥{{ row.fee }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="actualAmount" label="实际到账" width="110" align="center">
          <template #default="{ row }">
            <span style="color: #67c23a; font-weight: 600;">¥{{ row.actualAmount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="银行信息" width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <div>{{ row.bankName }}</div>
            <div style="color: #909399; font-size: 12px;">{{ maskBankCard(row.bankCardNo) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="withdrawStatusTag(row.status)" size="small">
              {{ withdrawStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="申请时间" width="170" align="center" />
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button type="success" link size="small" @click="handleApprove(row)">审核通过</el-button>
              <el-button type="danger" link size="small" @click="handleReject(row)">审核拒绝</el-button>
            </template>
            <span v-else style="color: #909399;">-</span>
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

    <!-- 审核拒绝弹窗 -->
    <el-dialog
      v-model="rejectDialogVisible"
      title="审核拒绝"
      width="450px"
      :close-on-click-modal="false"
    >
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules" label-width="80px">
        <el-form-item label="提现单号">
          <span>{{ currentRejectRow?.withdrawNo }}</span>
        </el-form-item>
        <el-form-item label="拒绝原因" prop="reason">
          <el-input
            v-model="rejectForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入拒绝原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="handleRejectSubmit">确认拒绝</el-button>
      </template>
    </el-dialog>

    <!-- 银行卡三要素校验弹窗 -->
    <el-dialog
      v-model="bankVerifyDialogVisible"
      title="银行卡三要素校验"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="bankVerifyFormRef" :model="bankVerifyForm" :rules="bankVerifyRules" label-width="90px">
        <el-form-item label="真实姓名" prop="realName">
          <el-input v-model="bankVerifyForm.realName" placeholder="请输入持卡人姓名" />
        </el-form-item>
        <el-form-item label="身份证号" prop="idCard">
          <el-input v-model="bankVerifyForm.idCard" placeholder="请输入身份证号" maxlength="18" />
        </el-form-item>
        <el-form-item label="银行卡号" prop="bankCardNo">
          <el-input v-model="bankVerifyForm.bankCardNo" placeholder="请输入银行卡号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bankVerifyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="bankVerifyLoading" @click="handleBankVerifySubmit">
          开始校验
        </el-button>
      </template>
    </el-dialog>

    <!-- 校验结果 -->
    <el-dialog
      v-model="bankVerifyResultVisible"
      title="校验结果"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-result
        :icon="bankVerifyResultData?.passed ? 'success' : 'error'"
        :title="bankVerifyResultData?.passed ? '校验通过' : '校验不通过'"
        :sub-title="bankVerifyResultData?.message || ''"
      >
        <template #extra>
          <el-button type="primary" @click="bankVerifyResultVisible = false">关闭</el-button>
        </template>
      </el-result>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Check } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'FinanceWithdraw',
  data() {
    return {
      Search,
      Refresh,
      Check,
      searchForm: {
        withdrawNo: '',
        userPhone: '',
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
      rejectDialogVisible: false,
      currentRejectRow: null,
      rejectForm: {
        reason: ''
      },
      rejectRules: {
        reason: [{ required: true, message: '请输入拒绝原因', trigger: 'blur' }]
      },
      rejectLoading: false,
      bankVerifyDialogVisible: false,
      bankVerifyForm: {
        realName: '',
        idCard: '',
        bankCardNo: ''
      },
      bankVerifyRules: {
        realName: [{ required: true, message: '请输入持卡人姓名', trigger: 'blur' }],
        idCard: [{ required: true, message: '请输入身份证号', trigger: 'blur' }],
        bankCardNo: [{ required: true, message: '请输入银行卡号', trigger: 'blur' }]
      },
      bankVerifyLoading: false,
      bankVerifyResultVisible: false,
      bankVerifyResultData: null
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
          withdrawNo: this.searchForm.withdrawNo || undefined,
          userPhone: this.searchForm.userPhone || undefined,
          status: this.searchForm.status || undefined,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/finance/withdraws', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取提现列表失败:', err)
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
        withdrawNo: '',
        userPhone: '',
        status: '',
        dateRange: []
      }
      this.handleSearch()
    },
    async handleApprove(row) {
      try {
        await ElMessageBox.confirm(
          `确定要审核通过提现单「${row.withdrawNo}」吗？`,
          '审核通过确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
        )
        await request.post(`/finance/withdraws/${row.id}/approve`)
        ElMessage.success('审核通过')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('审核通过失败:', err)
        }
      }
    },
    handleReject(row) {
      this.currentRejectRow = row
      this.rejectForm.reason = ''
      this.rejectDialogVisible = true
    },
    async handleRejectSubmit() {
      const valid = await this.$refs.rejectFormRef.validate().catch(() => false)
      if (!valid) return
      this.rejectLoading = true
      try {
        await request.post(`/finance/withdraws/${this.currentRejectRow.id}/reject`, {
          reason: this.rejectForm.reason
        })
        ElMessage.success('已拒绝')
        this.rejectDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('审核拒绝失败:', err)
      } finally {
        this.rejectLoading = false
      }
    },
    handleBankCardVerify() {
      this.bankVerifyForm = {
        realName: '',
        idCard: '',
        bankCardNo: ''
      }
      this.bankVerifyDialogVisible = true
    },
    async handleBankVerifySubmit() {
      const valid = await this.$refs.bankVerifyFormRef.validate().catch(() => false)
      if (!valid) return
      this.bankVerifyLoading = true
      try {
        const res = await request.post('/finance/withdraws/verify-bank-card', {
          realName: this.bankVerifyForm.realName,
          idCard: this.bankVerifyForm.idCard,
          bankCardNo: this.bankVerifyForm.bankCardNo
        })
        this.bankVerifyResultData = res.data || {}
        this.bankVerifyDialogVisible = false
        this.bankVerifyResultVisible = true
      } catch (err) {
        console.error('银行卡校验失败:', err)
      } finally {
        this.bankVerifyLoading = false
      }
    },
    maskPhone(phone) {
      if (!phone) return '-'
      return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    },
    maskBankCard(cardNo) {
      if (!cardNo) return '-'
      return cardNo.replace(/(\d{4})\d{8,12}(\d{4})/, '$1 **** **** $2')
    },
    withdrawStatusTag(status) {
      const map = {
        pending: 'warning',
        approved: 'success',
        rejected: 'danger',
        paid: ''
      }
      return map[status] || 'info'
    },
    withdrawStatusLabel(status) {
      const map = {
        pending: '待审核',
        approved: '已通过',
        rejected: '已拒绝',
        paid: '已打款'
      }
      return map[status] || '未知'
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

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>