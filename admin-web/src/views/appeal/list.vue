<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">申诉管理</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="申诉编号">
          <el-input
            v-model="searchForm.appealNo"
            placeholder="请输入申诉编号"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 150px">
            <el-option label="手机号申诉" value="phone" />
            <el-option label="评价申诉" value="rating" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 130px">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已办结" value="completed" />
            <el-option label="已驳回" value="rejected" />
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
    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="appealNo" label="申诉编号" width="180" show-overflow-tooltip />
        <el-table-column label="类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'phone' ? 'primary' : 'warning'" size="small">
              {{ row.type === 'phone' ? '手机号申诉' : '评价申诉' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="applicant" label="申诉人" width="120" show-overflow-tooltip />
        <el-table-column prop="summary" label="内容摘要" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'pending' ? 'warning' : row.status === 'processing' ? 'primary' : row.status === 'completed' ? 'success' : 'danger'"
              size="small"
            >
              {{ row.status === 'pending' ? '待处理' : row.status === 'processing' ? '处理中' : row.status === 'completed' ? '已办结' : '已驳回' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="提交时间" width="170" align="center" />
        <el-table-column label="催办级别" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.urgeLevel === 0 ? 'info' : row.urgeLevel === 1 ? 'warning' : row.urgeLevel === 2 ? 'danger' : 'danger'"
              size="small"
              effect="dark"
            >
              {{ row.urgeLevel === 0 ? '无' : row.urgeLevel + '级' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
            <el-button
              v-if="row.status === 'pending' || row.status === 'processing'"
              type="success"
              link
              size="small"
              @click="handleComplete(row)"
            >
              办结
            </el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="申诉详情"
      width="750px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading">
        <el-descriptions :column="2" border v-if="currentAppeal.id">
          <el-descriptions-item label="申诉编号">{{ currentAppeal.appealNo }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="currentAppeal.type === 'phone' ? 'primary' : 'warning'" size="small">
              {{ currentAppeal.type === 'phone' ? '手机号申诉' : '评价申诉' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="申诉人">{{ currentAppeal.applicant }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag
              :type="currentAppeal.status === 'pending' ? 'warning' : currentAppeal.status === 'processing' ? 'primary' : currentAppeal.status === 'completed' ? 'success' : 'danger'"
              size="small"
            >
              {{ currentAppeal.status === 'pending' ? '待处理' : currentAppeal.status === 'processing' ? '处理中' : currentAppeal.status === 'completed' ? '已办结' : '已驳回' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ currentAppeal.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="申诉内容" :span="2">{{ currentAppeal.content }}</el-descriptions-item>
        </el-descriptions>

        <!-- 催办记录 -->
        <el-card class="urge-card" v-if="urgeRecords.length > 0">
          <template #header>
            <span class="card-header">催办记录</span>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="(item, index) in urgeRecords"
              :key="index"
              :timestamp="item.time"
              placement="top"
              :type="index === 0 ? 'danger' : index === 1 ? 'warning' : 'primary'"
            >
              <el-tag
                :type="index === 0 ? 'danger' : index === 1 ? 'warning' : 'primary'"
                size="small"
                style="margin-bottom: 4px;"
              >
                {{ index + 1 }}级催办
              </el-tag>
              <div>{{ item.content }}</div>
            </el-timeline-item>
          </el-timeline>
        </el-card>

        <!-- 沟通记录 -->
        <el-card class="communication-card">
          <template #header>
            <span class="card-header">沟通记录</span>
          </template>
          <el-timeline v-if="communicationList.length > 0">
            <el-timeline-item
              v-for="(item, index) in communicationList"
              :key="index"
              :timestamp="item.time"
              placement="top"
              :type="item.role === 'admin' ? 'primary' : 'success'"
            >
              <div class="comm-sender">
                <el-tag :type="item.role === 'admin' ? 'primary' : 'success'" size="small">
                  {{ item.role === 'admin' ? '管理员' : '申诉人' }}
                </el-tag>
                <span class="comm-name">{{ item.senderName }}</span>
              </div>
              <div class="comm-content">{{ item.content }}</div>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-else description="暂无沟通记录" />

          <!-- 发送消息 -->
          <div class="send-message" v-if="currentAppeal.status !== 'completed' && currentAppeal.status !== 'rejected'">
            <el-divider />
            <div class="send-row">
              <el-input
                v-model="sendMessage"
                type="textarea"
                :rows="3"
                placeholder="请输入回复内容..."
                maxlength="500"
                show-word-limit
              />
              <el-button
                type="primary"
                :loading="sendLoading"
                style="margin-top: 10px;"
                @click="handleSendMessage"
              >
                发送
              </el-button>
            </div>
          </div>
        </el-card>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'AppealList',
  data() {
    return {
      Search,
      Refresh,
      searchForm: {
        appealNo: '',
        type: '',
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
      detailVisible: false,
      detailLoading: false,
      currentAppeal: {},
      urgeRecords: [],
      communicationList: [],
      sendMessage: '',
      sendLoading: false
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
          appealNo: this.searchForm.appealNo || undefined,
          type: this.searchForm.type || undefined,
          status: this.searchForm.status || undefined,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/appeals', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取申诉列表失败:', err)
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
        appealNo: '',
        type: '',
        status: '',
        dateRange: []
      }
      this.handleSearch()
    },
    async handleDetail(row) {
      this.detailVisible = true
      this.detailLoading = true
      this.currentAppeal = {}
      this.urgeRecords = []
      this.communicationList = []
      this.sendMessage = ''
      try {
        const res = await request.get(`/appeals/${row.id}`)
        this.currentAppeal = res.data?.appeal || {}
        this.urgeRecords = res.data?.urgeRecords || []
        this.communicationList = res.data?.communications || []
      } catch (err) {
        console.error('获取申诉详情失败:', err)
      } finally {
        this.detailLoading = false
      }
    },
    async handleComplete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要办结申诉「${row.appealNo}」吗？`,
          '办结确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.put(`/appeals/${row.id}/complete`)
        ElMessage.success('办结成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('办结失败:', err)
        }
      }
    },
    async handleSendMessage() {
      if (!this.sendMessage.trim()) {
        ElMessage.warning('请输入回复内容')
        return
      }
      this.sendLoading = true
      try {
        await request.post(`/appeals/${this.currentAppeal.id}/communications`, {
          content: this.sendMessage
        })
        ElMessage.success('发送成功')
        this.sendMessage = ''
        // 刷新沟通记录
        const res = await request.get(`/appeals/${this.currentAppeal.id}`)
        this.communicationList = res.data?.communications || []
      } catch (err) {
        console.error('发送失败:', err)
      } finally {
        this.sendLoading = false
      }
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

.table-card {
  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}

.urge-card {
  margin-top: 16px;
}

.communication-card {
  margin-top: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
}

.comm-sender {
  margin-bottom: 4px;
}

.comm-name {
  margin-left: 8px;
  font-size: 13px;
  color: #606266;
}

.comm-content {
  color: #303133;
  font-size: 14px;
  line-height: 1.6;
}

.send-message {
  .send-row {
    display: flex;
    flex-direction: column;
  }
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>