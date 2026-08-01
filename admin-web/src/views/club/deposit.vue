<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">保证金管理</span>
    </div>

    <!-- 个人/企业独立保证金配置 -->
    <el-card class="config-card">
      <template #header>
        <div class="card-header-row">
          <span class="card-title">保证金基础配置（个人 / 企业 双参数）</span>
          <el-button type="primary" size="small" @click="saveConfig" :loading="configSaving">保存配置</el-button>
        </div>
      </template>
      <el-form :model="configForm" inline label-width="180px">
        <el-form-item label="个人俱乐部保证金(元)">
          <el-input-number v-model="configForm.personal_deposit" :min="0" :precision="2" :step="100" style="width: 200px" />
        </el-form-item>
        <el-form-item label="企业俱乐部保证金(元)">
          <el-input-number v-model="configForm.enterprise_deposit" :min="0" :precision="2" :step="100" style="width: 200px" />
        </el-form-item>
        <el-form-item label="补缴监控阈值(元)">
          <el-input-number v-model="configForm.repay_threshold" :min="0" :precision="2" :step="100" style="width: 200px" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 16px">
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select v-model="searchForm.depositStatus" placeholder="全部" clearable style="width: 120px">
            <el-option label="未缴" :value="0" />
            <el-option label="已缴" :value="1" />
            <el-option label="已退" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="补缴状态">
          <el-select v-model="searchForm.repayStatus" placeholder="全部" clearable style="width: 140px">
            <el-option label="正常" :value="0" />
            <el-option label="待补缴" :value="1" />
            <el-option label="已逾期" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="searchForm.keyword" placeholder="名称/缩写" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">搜索</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="club_name" label="俱乐部" min-width="140" />
        <el-table-column label="缩写" width="90">
          <template #default="{ row }"><el-tag type="info" size="small">{{ row.abbreviation }}</el-tag></template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.badge_type === 'blue_v' ? 'primary' : 'success'" size="small">
              {{ row.badge_type === 'blue_v' ? '企业' : '个人' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="保证金" width="110">
          <template #default="{ row }">{{ row.deposit_amount }} 元</template>
        </el-table-column>
        <el-table-column label="缴纳状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.deposit_status === 1 ? 'success' : row.deposit_status === 2 ? 'info' : 'warning'">
              {{ row.deposit_status === 1 ? '已缴' : row.deposit_status === 2 ? '已退' : '未缴' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="补缴状态" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.repay_status === 1" type="warning">待补缴</el-tag>
            <el-tag v-else-if="row.repay_status === 2" type="danger">已逾期</el-tag>
            <el-tag v-else type="success">正常</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="交易号" width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.deposit_transaction_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="deposit_pay_time" label="缴纳时间" width="170" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.deposit_status === 0" type="success" size="small" @click="handleConfirm(row)">确认到账</el-button>
            <el-button v-if="row.deposit_status === 1" type="warning" size="small" @click="handleRefund(row)">退还</el-button>
            <el-button type="info" size="small" @click="handleViewDeduct(row)">扣除记录</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" v-model:page-size="limit" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    <!-- 扣除记录弹窗 -->
    <el-dialog v-model="deductVisible" :title="`扣除记录 - ${currentClubName}`" width="800px">
      <el-table :data="deductList" v-loading="deductLoading" size="small" border empty-text="暂无扣除记录">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="create_time" label="扣除时间" width="170" />
        <el-table-column label="扣除金额(元)" width="120">
          <template #default="{ row }">¥{{ row.amount_yuan || row.amount }}</template>
        </el-table-column>
        <el-table-column label="扣除类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'fine' ? 'danger' : 'warning'" size="small">
              {{ row.type === 'fine' ? '罚款' : '赔付' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="扣除原因" min-width="200" show-overflow-tooltip />
        <el-table-column prop="operator" label="操作人" width="120" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ClubDeposit',
  data() {
    return {
      searchForm: { depositStatus: '', repayStatus: '', keyword: '' },
      tableData: [], loading: false, page: 1, limit: 20, total: 0,
      configForm: {
        personal_deposit: 0,
        enterprise_deposit: 0,
        repay_threshold: 0
      },
      configSaving: false,
      deductVisible: false,
      deductLoading: false,
      deductList: [],
      currentClubName: ''
    }
  },
  mounted() {
    this.fetchList()
    this.fetchConfig()
  },
  methods: {
    buildParams() {
      const params = { page: this.page, limit: this.limit }
      if (this.searchForm.depositStatus !== '' && this.searchForm.depositStatus !== null) {
        params.depositStatus = this.searchForm.depositStatus
      }
      if (this.searchForm.repayStatus !== '' && this.searchForm.repayStatus !== null) {
        params.repayStatus = this.searchForm.repayStatus
      }
      if (this.searchForm.keyword) params.keyword = this.searchForm.keyword
      return params
    },
    onSearch() { this.page = 1; this.fetchList() },
    onReset() {
      this.searchForm = { depositStatus: '', repayStatus: '', keyword: '' }
      this.page = 1
      this.fetchList()
    },
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/deposit_list', this.buildParams())
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    async fetchConfig() {
      try {
        const res = await request.get('/deposit/config')
        const data = res.data || {}
        this.configForm.personal_deposit = data.personal_deposit || 0
        this.configForm.enterprise_deposit = data.enterprise_deposit || 0
        this.configForm.repay_threshold = data.repay_threshold || 0
      } catch (e) { /* ignore */ }
    },
    async saveConfig() {
      this.configSaving = true
      try {
        await request.put('/deposit/config', {
          personal_deposit: this.configForm.personal_deposit,
          enterprise_deposit: this.configForm.enterprise_deposit,
          repay_threshold: this.configForm.repay_threshold
        })
        ElMessage.success('保证金配置已更新')
      } catch (e) {
        ElMessage.error('保存失败')
      } finally {
        this.configSaving = false
      }
    },
    async handleConfirm(row) {
      try {
        await ElMessageBox.confirm(`确认俱乐部"${row.club_name}"保证金${row.deposit_amount}元已到账？`, '确认到账', { type: 'success' })
        await request.put('/club/confirm_deposit', { id: row.id })
        ElMessage.success('已确认到账，俱乐部已激活')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleRefund(row) {
      try {
        const { value: reason } = await ElMessageBox.prompt('请输入退还原因', '退还保证金', { type: 'warning' })
        await request.put('/club/refund_deposit', { id: row.id, reason })
        ElMessage.success('保证金已退还')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleViewDeduct(row) {
      this.currentClubName = row.club_name
      this.deductVisible = true
      this.deductLoading = true
      try {
        const res = await request.get('/deposit/deduct-list', { club_id: row.id })
        this.deductList = res.data?.list || []
      } catch (e) {
        ElMessage.error('加载扣除记录失败')
      } finally {
        this.deductLoading = false
      }
    }
  }
}
</script>

<style scoped>
.config-card {
  margin-bottom: 16px;
}
.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  font-size: 15px;
  font-weight: 600;
}
</style>
