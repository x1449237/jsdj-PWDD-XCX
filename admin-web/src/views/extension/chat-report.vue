<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">会话举报处理</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" style="width: 150px">
            <el-option label="全部" :value="-1" />
            <el-option label="待处理" :value="0" />
            <el-option label="已处理" :value="1" />
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
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="reporterNickname" label="举报人" min-width="130" show-overflow-tooltip />
        <el-table-column prop="reportedNickname" label="被举报人" min-width="130" show-overflow-tooltip />
        <el-table-column label="会话类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ sessionTypeLabel(row.sessionType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'warning' : 'info'" size="small">
              {{ row.status === 0 ? '待处理' : '已处理' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="primary"
              link
              size="small"
              @click="handleProcess(row)"
            >
              处理
            </el-button>
            <span v-else style="color: #909399;">-</span>
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

    <el-dialog v-model="dialogVisible" title="处理举报" width="560px">
      <div v-if="currentRow" style="margin-bottom: 16px; color: #606266;">
        <div style="margin-bottom: 8px;">
          <strong>举报人：</strong>{{ currentRow.reporterNickname }}
          <strong style="margin-left: 16px;">被举报人：</strong>{{ currentRow.reportedNickname }}
        </div>
        <div style="margin-bottom: 8px;">
          <strong>会话类型：</strong>{{ sessionTypeLabel(currentRow.sessionType) }}
        </div>
        <div>
          <strong>原因：</strong>{{ currentRow.reason }}
        </div>
      </div>
      <el-form :model="form" label-width="80px">
        <el-form-item label="处理结果">
          <el-input v-model="form.result" type="textarea" :rows="4" placeholder="请输入处理结果" />
        </el-form-item>
        <el-form-item label="处理状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="已处理" :value="1" />
            <el-option label="驳回" :value="2" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { getChatReports, handleChatReport } from '@/api/extension'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'ChatReport',
  data() {
    return {
      Search,
      Refresh,
      searchForm: {
        status: -1
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      submitting: false,
      currentRow: null,
      form: {
        status: 1,
        result: ''
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
          status: this.searchForm.status
        }
        const res = await getChatReports(params)
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取举报列表失败:', err)
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
        status: -1
      }
      this.handleSearch()
    },
    handleProcess(row) {
      this.currentRow = row
      this.form = {
        status: 1,
        result: ''
      }
      this.dialogVisible = true
    },
    async handleSubmit() {
      if (!this.form.result) {
        ElMessage.error('请输入处理结果')
        return
      }
      this.submitting = true
      try {
        await handleChatReport(this.currentRow.id, { status: this.form.status, result: this.form.result })
        ElMessage.success('处理成功')
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('处理失败:', err)
      } finally {
        this.submitting = false
      }
    },
    sessionTypeLabel(type) {
      const map = { 1: '私聊', 2: '群聊' }
      return map[type] || '-'
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
</style>
