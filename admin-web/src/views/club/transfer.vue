<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">对公打款验证</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="验证状态">
          <el-select v-model="searchForm.verificationStatus" placeholder="全部" clearable style="width: 120px">
            <el-option label="待确认" :value="1" />
            <el-option label="已通过" :value="2" />
            <el-option label="已驳回" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="club_name" label="俱乐部" min-width="140" />
        <el-table-column label="对公账户" min-width="200">
          <template #default="{ row }">{{ row.corporate_bank }} / {{ row.corporate_account }}</template>
        </el-table-column>
        <el-table-column label="验证金额" width="120">
          <template #default="{ row }">{{ row.verification_amount }} 元</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.verification_status === 1" type="warning">待确认</el-tag>
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
        <el-table-column prop="update_time" label="提交时间" width="170" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <template v-if="row.verification_status === 1">
              <el-button type="success" size="small" @click="handleVerify(row, 'pass')">通过</el-button>
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
      searchForm: { verificationStatus: '' },
      tableData: [], loading: false, page: 1, limit: 20, total: 0
    }
  },
  mounted() { this.fetchList() },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/transfer_list', { page: this.page, limit: this.limit, ...this.searchForm })
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    previewReceipt(row) {
      window.open(row.verification_receipt, '_blank')
    },
    async handleVerify(row, action) {
      const label = action === 'pass' ? '通过' : '驳回'
      try {
        if (action === 'fail') {
          const { value: reason } = await ElMessageBox.prompt('请输入驳回原因', '驳回验证', { type: 'warning' })
          await request.put('/club/verify_transfer', { id: row.id, action, reason })
        } else {
          await ElMessageBox.confirm(`确认对公打款验证通过？俱乐部"${row.club_name}"验证金额${row.verification_amount}元。`, '确认通过', { type: 'success' })
          await request.put('/club/verify_transfer', { id: row.id, action })
        }
        ElMessage.success(`已${label}`)
        this.fetchList()
      } catch (e) { /* cancel */ }
    }
  }
}
</script>