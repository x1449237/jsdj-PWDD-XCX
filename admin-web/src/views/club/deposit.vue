<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">保证金管理</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select v-model="searchForm.depositStatus" placeholder="全部" clearable style="width: 120px">
            <el-option label="未缴" :value="0" />
            <el-option label="已缴" :value="1" />
            <el-option label="已退" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="searchForm.keyword" placeholder="名称/缩写" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
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
        <el-table-column label="交易号" width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.deposit_transaction_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="deposit_pay_time" label="缴纳时间" width="170" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.deposit_status === 0" type="success" size="small" @click="handleConfirm(row)">确认到账</el-button>
            <el-button v-if="row.deposit_status === 1" type="warning" size="small" @click="handleRefund(row)">退还</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" v-model:page-size="limit" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ClubDeposit',
  data() {
    return {
      searchForm: { depositStatus: '', keyword: '' },
      tableData: [], loading: false, page: 1, limit: 20, total: 0
    }
  },
  mounted() { this.fetchList() },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/deposit_list', { page: this.page, limit: this.limit, ...this.searchForm })
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
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
    }
  }
}
</script>