<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">打手审核</span>
    </div>

    <!-- 搜索筛选 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="昵称/手机号"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
            <el-option label="已下架" value="delisted" />
          </el-select>
        </el-form-item>
        <el-form-item label="注册时间">
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
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column label="头像" width="70" align="center">
          <template #default="{ row }">
            <el-avatar :size="36" :src="row.avatar" />
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column label="手机号" width="130">
          <template #default="{ row }">
            {{ maskPhone(row.phone) }}
          </template>
        </el-table-column>
        <el-table-column label="实名状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.realNameStatus === 'verified' ? 'success' : 'info'" size="small">
              {{ row.realNameStatus === 'verified' ? '已认证' : '未认证' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="inviteCode" label="邀请码" width="100" align="center" />
        <el-table-column prop="createdAt" label="注册时间" width="170" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button type="success" link size="small" @click="handleApprove(row)">通过</el-button>
              <el-button type="danger" link size="small" @click="handleReject(row)">拒绝</el-button>
            </template>
            <el-button type="primary" link size="small" @click="handleDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'approved'"
              type="danger"
              link
              size="small"
              @click="handleDelist(row)"
            >
              强制下架
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </el-card>

    <!-- 拒绝原因弹窗 -->
    <el-dialog
      v-model="rejectDialogVisible"
      title="审核拒绝"
      width="450px"
      :close-on-click-modal="false"
    >
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules" label-width="80px">
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

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="打手详情"
      width="650px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="昵称">{{ detailData.nickname }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ maskPhone(detailData.phone) }}</el-descriptions-item>
          <el-descriptions-item label="真实姓名">{{ detailData.realName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ maskIdCard(detailData.idCard) }}</el-descriptions-item>
          <el-descriptions-item label="实名状态">
            <el-tag :type="detailData.realNameStatus === 'verified' ? 'success' : 'info'" size="small">
              {{ detailData.realNameStatus === 'verified' ? '已认证' : '未认证' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="邀请码">{{ detailData.inviteCode || '-' }}</el-descriptions-item>
        </el-descriptions>

        <h4 class="detail-sub-title">服务列表</h4>
        <el-table :data="detailData.services || []" stripe border size="small">
          <el-table-column prop="name" label="服务名称" min-width="120" />
          <el-table-column prop="price" label="价格" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row: sRow }">
              <el-tag :type="sRow.status === 'online' ? 'success' : 'info'" size="small">
                {{ sRow.status === 'online' ? '上架' : '下架' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <h4 class="detail-sub-title">等级收益</h4>
        <el-table :data="detailData.levelIncome || []" stripe border size="small">
          <el-table-column prop="level" label="等级" width="100" />
          <el-table-column prop="totalOrders" label="总订单" width="100" />
          <el-table-column prop="totalIncome" label="总收益" width="120" />
          <el-table-column prop="monthIncome" label="本月收益" width="120" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'AuditPlayer',
  data() {
    return {
      Search,
      Refresh,
      searchForm: {
        keyword: '',
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
      rejectForm: {
        reason: ''
      },
      rejectRules: {
        reason: [{ required: true, message: '请输入拒绝原因', trigger: 'blur' }]
      },
      rejectLoading: false,
      currentRejectId: null,
      detailDialogVisible: false,
      detailLoading: false,
      detailData: {}
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
          keyword: this.searchForm.keyword || undefined,
          status: this.searchForm.status || undefined,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/audit/players', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取打手审核列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.searchForm = { keyword: '', status: '', dateRange: [] }
      this.handleSearch()
    },
    async handleApprove(row) {
      try {
        await ElMessageBox.confirm(
          `确定要通过「${row.nickname}」的打手审核吗？`,
          '审核通过确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
        )
        await request.post(`/audit/players/${row.id}/approve`)
        ElMessage.success('审核通过')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('审核通过失败:', err)
        }
      }
    },
    handleReject(row) {
      this.currentRejectId = row.id
      this.rejectForm.reason = ''
      this.rejectDialogVisible = true
    },
    async handleRejectSubmit() {
      const valid = await this.$refs.rejectFormRef.validate().catch(() => false)
      if (!valid) return
      this.rejectLoading = true
      try {
        await request.post(`/audit/players/${this.currentRejectId}/reject`, {
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
    async handleDetail(row) {
      this.detailDialogVisible = true
      this.detailLoading = true
      try {
        const res = await request.get(`/audit/players/${row.id}`)
        this.detailData = res.data || {}
      } catch (err) {
        console.error('获取打手详情失败:', err)
      } finally {
        this.detailLoading = false
      }
    },
    async handleDelist(row) {
      try {
        await ElMessageBox.confirm(
          `确定要强制下架「${row.nickname}」吗？`,
          '强制下架确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/audit/players/${row.id}/delist`)
        ElMessage.success('已下架')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('强制下架失败:', err)
        }
      }
    },
    maskPhone(phone) {
      if (!phone) return '-'
      return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    },
    maskIdCard(idCard) {
      if (!idCard) return '-'
      return idCard.replace(/(\d{4})\d{10}(\d{4})/, '$1**********$2')
    },
    statusTag(status) {
      const map = { pending: 'warning', approved: 'success', rejected: 'danger', delisted: 'info' }
      return map[status] || 'info'
    },
    statusLabel(status) {
      const map = { pending: '待审核', approved: '已通过', rejected: '已拒绝', delisted: '已下架' }
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

.detail-sub-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin: 20px 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>