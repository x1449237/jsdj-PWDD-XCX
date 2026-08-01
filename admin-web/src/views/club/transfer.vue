<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">对公打款验证台账</span>
      <el-button type="success" @click="handleExport" :loading="exporting">导出打款台账（带水印）</el-button>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="验证状态">
          <el-select v-model="searchForm.verificationStatus" placeholder="全部" clearable style="width: 120px">
            <el-option label="未发起" :value="0" />
            <el-option label="待确认" :value="1" />
            <el-option label="已通过" :value="2" />
            <el-option label="已驳回" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="俱乐部">
          <el-input v-model="searchForm.keyword" placeholder="名称/缩写" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="提交时间">
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
        <el-table-column prop="club_name" label="俱乐部" min-width="140" />
        <el-table-column label="对公账户" min-width="220">
          <template #default="{ row }">
            <div>{{ row.corporate_bank }}</div>
            <div class="mono">{{ row.corporate_account }}</div>
          </template>
        </el-table-column>
        <el-table-column label="验证金额" width="120">
          <template #default="{ row }">{{ row.verification_amount }} 元</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.verification_status === 0" type="info">未发起</el-tag>
            <el-tag v-else-if="row.verification_status === 1" type="warning">待确认</el-tag>
            <el-tag v-else-if="row.verification_status === 2" type="success">已通过</el-tag>
            <el-tag v-else type="danger">已驳回</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="凭证" width="80">
          <template #default="{ row }">
            <el-button v-if="row.verification_receipt" type="primary" size="small" link @click="previewReceipt(row)">查看</el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="打款发起时间" width="170" />
        <el-table-column prop="update_time" label="最近更新" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.verification_status === 0"
              type="primary" size="small" @click="handleGenerate(row)">生成打款</el-button>
            <template v-if="row.verification_status === 1">
              <el-button type="success" size="small" @click="handleVerify(row, 'pass')">确认到账</el-button>
              <el-button type="danger" size="small" @click="handleVerify(row, 'fail')">驳回</el-button>
            </template>
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
  name: 'ClubTransfer',
  data() {
    return {
      searchForm: { verificationStatus: '', keyword: '', dateRange: [] },
      tableData: [], loading: false, page: 1, limit: 20, total: 0,
      exporting: false
    }
  },
  mounted() { this.fetchList() },
  methods: {
    buildParams() {
      const params = { page: this.page, limit: this.limit }
      if (this.searchForm.verificationStatus !== '' && this.searchForm.verificationStatus !== null) {
        params.verificationStatus = this.searchForm.verificationStatus
      }
      if (this.searchForm.keyword) params.keyword = this.searchForm.keyword
      if (this.searchForm.dateRange && this.searchForm.dateRange.length === 2) {
        params.startDate = this.searchForm.dateRange[0]
        params.endDate = this.searchForm.dateRange[1]
      }
      return params
    },
    onSearch() { this.page = 1; this.fetchList() },
    onReset() {
      this.searchForm = { verificationStatus: '', keyword: '', dateRange: [] }
      this.page = 1
      this.fetchList()
    },
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/transfer-list', this.buildParams())
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    previewReceipt(row) {
      window.open(row.verification_receipt, '_blank')
    },
    async handleGenerate(row) {
      try {
        await ElMessageBox.confirm(
          `确认为俱乐部"${row.club_name}"生成对公打款验证（小额打款）吗？`,
          '生成打款确认',
          { type: 'warning', confirmButtonText: '确认生成' }
        )
        await request.post('/club/generate-transfer', { id: row.id })
        ElMessage.success('打款已生成，等待银行回执')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleVerify(row, action) {
      const label = action === 'pass' ? '确认到账' : '驳回'
      try {
        if (action === 'fail') {
          const { value: reason } = await ElMessageBox.prompt('请输入驳回原因', '驳回验证', { type: 'warning' })
          await request.post('/club/verify-transfer', { id: row.id, action, reason })
        } else {
          await ElMessageBox.confirm(
            `确认对公打款验证通过？俱乐部"${row.club_name}"验证金额${row.verification_amount}元。`,
            '确认到账',
            { type: 'success' }
          )
          await request.post('/club/verify-transfer', { id: row.id, action })
        }
        ElMessage.success(`已${label}`)
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleExport() {
      this.exporting = true
      try {
        const blob = await request.get('/club/transfer-ledger/export', { ...this.buildParams(), export: 1 })
        const url = window.URL.createObjectURL(new Blob([blob]))
        const a = document.createElement('a')
        a.href = url
        a.download = `对公打款台账_${Date.now()}.xlsx`
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
