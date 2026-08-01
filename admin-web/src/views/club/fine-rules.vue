<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">罚款规则备案管理</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="俱乐部">
          <el-input v-model="searchForm.keyword" placeholder="俱乐部名称/缩写" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="已下架" value="revoked" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">搜索</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="club_name" label="俱乐部" min-width="140" />
        <el-table-column prop="rule_name" label="规则名称" min-width="160" />
        <el-table-column label="罚款金额(元)" width="120">
          <template #default="{ row }">¥{{ row.amount_yuan || row.amount }}</template>
        </el-table-column>
        <el-table-column prop="scene" label="适用场景" min-width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'pending'" type="warning">待审核</el-tag>
            <el-tag v-else-if="row.status === 'approved'" type="success">已通过</el-tag>
            <el-tag v-else-if="row.status === 'rejected'" type="danger">已驳回</el-tag>
            <el-tag v-else type="info">已下架</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="提交时间" width="170" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleDetail(row)">查看详情</el-button>
            <el-button
              v-if="row.status === 'approved'"
              type="danger" size="small" @click="handleRevoke(row)">强制下架</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" v-model:page-size="limit" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="罚款规则详情" width="600px">
      <el-descriptions v-if="currentRule" :column="1" border>
        <el-descriptions-item label="规则名称">{{ currentRule.rule_name }}</el-descriptions-item>
        <el-descriptions-item label="所属俱乐部">{{ currentRule.club_name }}</el-descriptions-item>
        <el-descriptions-item label="罚款金额">¥{{ currentRule.amount_yuan || currentRule.amount }}</el-descriptions-item>
        <el-descriptions-item label="适用场景">{{ currentRule.scene }}</el-descriptions-item>
        <el-descriptions-item label="规则正文">{{ currentRule.content }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="currentRule.status === 'pending'" type="warning">待审核</el-tag>
          <el-tag v-else-if="currentRule.status === 'approved'" type="success">已通过</el-tag>
          <el-tag v-else-if="currentRule.status === 'rejected'" type="danger">已驳回</el-tag>
          <el-tag v-else type="info">已下架</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ currentRule.create_time }}</el-descriptions-item>
        <el-descriptions-item v-if="currentRule.audit_remark" label="审核备注">{{ currentRule.audit_remark }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ClubFineRules',
  data() {
    return {
      searchForm: { keyword: '', status: '' },
      tableData: [], loading: false, page: 1, limit: 20, total: 0,
      detailVisible: false,
      currentRule: null
    }
  },
  mounted() { this.fetchList() },
  methods: {
    buildParams() {
      const params = { page: this.page, limit: this.limit }
      if (this.searchForm.keyword) params.keyword = this.searchForm.keyword
      if (this.searchForm.status) params.status = this.searchForm.status
      return params
    },
    onSearch() { this.page = 1; this.fetchList() },
    onReset() {
      this.searchForm = { keyword: '', status: '' }
      this.page = 1
      this.fetchList()
    },
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/fine-rules', this.buildParams())
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    async handleDetail(row) {
      try {
        const res = await request.get('/club/fine-rules/detail', { id: row.id })
        this.currentRule = res.data
        this.detailVisible = true
      } catch (e) {
        ElMessage.error('加载详情失败')
      }
    },
    async handleRevoke(row) {
      try {
        const { value: reason } = await ElMessageBox.prompt(
          `确定强制下架俱乐部"${row.club_name}"的罚款规则"${row.rule_name}"吗？请输入下架原因。`,
          '强制下架确认',
          { type: 'error', inputPlaceholder: '请输入下架原因' }
        )
        await request.post('/club/fine-rules/revoke', { id: row.id, reason })
        ElMessage.success('已强制下架')
        this.fetchList()
      } catch (e) { /* cancel */ }
    }
  }
}
</script>
