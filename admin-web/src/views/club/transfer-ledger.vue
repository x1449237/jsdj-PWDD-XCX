<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">小额打款独立台账</span>
      <el-button type="success" @click="handleExport" :loading="exporting">导出台账（带水印）</el-button>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="流水类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 140px">
            <el-option label="企业对公验证打款" value="verify" />
            <el-option label="原路退回" value="refund" />
          </el-select>
        </el-form-item>
        <el-form-item label="俱乐部">
          <el-input v-model="searchForm.keyword" placeholder="名称/缩写" clearable style="width: 180px" />
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
          <el-button type="primary" @click="onSearch">搜索</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="trade_no" label="流水号" width="200" show-overflow-tooltip />
        <el-table-column prop="club_name" label="俱乐部" min-width="140" />
        <el-table-column label="流水类型" width="160">
          <template #default="{ row }">
            <el-tag :type="row.type === 'verify' ? 'primary' : 'warning'" size="small">
              {{ row.type === 'verify' ? '企业对公验证打款' : '原路退回' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额(元)" width="120">
          <template #default="{ row }">¥{{ row.amount_yuan || row.amount }}</template>
        </el-table-column>
        <el-table-column label="收款账户" min-width="220">
          <template #default="{ row }">
            <div>{{ row.bank_name }}</div>
            <div class="mono">{{ row.account_no }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 0" type="info">处理中</el-tag>
            <el-tag v-else-if="row.status === 1" type="success">成功</el-tag>
            <el-tag v-else type="danger">失败</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="170" />
        <el-table-column prop="finish_time" label="完成时间" width="170" />
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
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
import { ElMessage } from 'element-plus'

export default {
  name: 'ClubTransferLedger',
  data() {
    return {
      searchForm: { type: '', keyword: '', dateRange: [] },
      tableData: [], loading: false, page: 1, limit: 20, total: 0,
      exporting: false
    }
  },
  mounted() { this.fetchList() },
  methods: {
    buildParams() {
      const params = { page: this.page, limit: this.limit }
      if (this.searchForm.type) params.type = this.searchForm.type
      if (this.searchForm.keyword) params.keyword = this.searchForm.keyword
      if (this.searchForm.dateRange && this.searchForm.dateRange.length === 2) {
        params.startDate = this.searchForm.dateRange[0]
        params.endDate = this.searchForm.dateRange[1]
      }
      return params
    },
    onSearch() { this.page = 1; this.fetchList() },
    onReset() {
      this.searchForm = { type: '', keyword: '', dateRange: [] }
      this.page = 1
      this.fetchList()
    },
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/transfer-ledger/list', this.buildParams())
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    async handleExport() {
      this.exporting = true
      try {
        const blob = await request.get('/club/transfer-ledger/export', { ...this.buildParams(), export: 1 })
        const url = window.URL.createObjectURL(new Blob([blob]))
        const a = document.createElement('a')
        a.href = url
        a.download = `小额打款台账_${Date.now()}.xlsx`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        window.URL.revokeObjectURL(url)
        ElMessage.success('导出成功')
      } catch (e) {
        ElMessage.error('导出失败')
      } finally {
        this.exporting = false
      }
    }
  }
}
</script>

<style scoped>
.mono {
  font-family: Menlo, Consolas, monospace;
  font-size: 12px;
  color: #606266;
}
</style>
