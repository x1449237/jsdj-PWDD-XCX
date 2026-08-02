<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">意见反馈</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" style="width: 150px">
            <el-option label="全部" :value="-1" />
            <el-option label="待处理" :value="0" />
            <el-option label="已回复" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="请输入关键词" clearable style="width: 200px" />
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
        <el-table-column prop="userNickname" label="用户昵称" min-width="130" show-overflow-tooltip />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
        <el-table-column label="截图" width="100" align="center">
          <template #default="{ row }">
            <el-image
              v-if="parseImages(row.images).length"
              :src="parseImages(row.images)[0]"
              :preview-src-list="parseImages(row.images)"
              fit="cover"
              style="width: 50px; height: 50px;"
              preview-teleported
            />
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'danger' : 'success'" size="small">
              {{ row.status === 0 ? '待处理' : '已回复' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button v-if="row.status === 0" type="primary" link size="small" @click="handleReply(row)">回复</el-button>
            <el-button v-else type="success" link size="small" @click="handleView(row)">查看回复</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <div v-if="currentRow" style="margin-bottom: 16px;">
        <div style="margin-bottom: 8px; color: #606266;">
          <strong>用户：</strong>{{ currentRow.userNickname }}
          <strong style="margin-left: 16px;">类型：</strong>{{ typeLabel(currentRow.type) }}
        </div>
        <div style="margin-bottom: 8px; color: #606266;">
          <strong>内容：</strong>{{ currentRow.content }}
        </div>
        <div v-if="currentRow.reply" style="padding: 10px; background: #f5f7fa; border-radius: 4px; color: #606266;">
          <strong>已回复：</strong>{{ currentRow.reply }}
        </div>
      </div>
      <el-form v-if="isReplyMode" :model="form" label-width="80px">
        <el-form-item label="回复内容">
          <el-input v-model="form.reply" type="textarea" :rows="4" placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button v-if="isReplyMode" type="primary" :loading="submitting" @click="handleSubmit">提交回复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { getAllFeedbacks, replyFeedback } from '@/api/extension'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'FeedbackList',
  data() {
    return {
      Search,
      Refresh,
      searchForm: {
        status: -1,
        keyword: ''
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      dialogTitle: '回复反馈',
      isReplyMode: true,
      submitting: false,
      currentRow: null,
      form: {
        reply: ''
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
          status: this.searchForm.status,
          keyword: this.searchForm.keyword || undefined
        }
        const res = await getAllFeedbacks(params)
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取反馈列表失败:', err)
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
        status: -1,
        keyword: ''
      }
      this.handleSearch()
    },
    handleReply(row) {
      this.currentRow = row
      this.isReplyMode = true
      this.dialogTitle = '回复反馈'
      this.form.reply = ''
      this.dialogVisible = true
    },
    handleView(row) {
      this.currentRow = row
      this.isReplyMode = false
      this.dialogTitle = '查看回复'
      this.dialogVisible = true
    },
    async handleSubmit() {
      if (!this.form.reply) {
        ElMessage.error('请输入回复内容')
        return
      }
      this.submitting = true
      try {
        await replyFeedback(this.currentRow.id, { reply: this.form.reply })
        ElMessage.success('回复成功')
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('回复失败:', err)
      } finally {
        this.submitting = false
      }
    },
    parseImages(images) {
      if (!images) return []
      if (Array.isArray(images)) return images
      try {
        const arr = JSON.parse(images)
        return Array.isArray(arr) ? arr : []
      } catch (e) {
        return []
      }
    },
    typeLabel(type) {
      const map = { 1: '建议', 2: '投诉', 3: '其他' }
      return map[type] || '其他'
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
