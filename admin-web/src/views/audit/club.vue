<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">俱乐部入驻审核</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="审核状态">
          <el-select v-model="searchForm.auditStatus" placeholder="全部" clearable style="width: 140px">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已驳回" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="俱乐部类型">
          <el-select v-model="searchForm.clubType" placeholder="全部" clearable style="width: 140px">
            <el-option label="企业级" value="blue_v" />
            <el-option label="个人级" value="green_v" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="club_name" label="俱乐部名称" min-width="150" />
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.badge_type === 'blue_v' ? 'primary' : 'success'" size="small">
              {{ row.badge_type === 'blue_v' ? '企业级' : '个人级' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="V标" width="80">
          <template #default="{ row }">
            <span v-if="row.badge_type === 'blue_v'" class="v-badge blue-v">V</span>
            <span v-else class="v-badge green-v">V</span>
          </template>
        </el-table-column>
        <el-table-column label="创始人" width="120">
          <template #default="{ row }">
            {{ row.user?.nickname || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.audit_status === 0" type="warning" size="small">待审核</el-tag>
            <el-tag v-else-if="row.audit_status === 1" type="success" size="small">已通过</el-tag>
            <el-tag v-else type="danger" size="small">已驳回</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="V标状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
              {{ row.is_active ? '已点亮' : '未点亮' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="申请时间" width="180" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <template v-if="row.audit_status === 0">
              <el-button type="success" size="small" @click="handleApprove(row)">通过</el-button>
              <el-button type="danger" size="small" @click="handleReject(row)">驳回</el-button>
            </template>
            <template v-else-if="row.audit_status === 1 && row.is_active">
              <el-button type="warning" size="small" @click="handleForceOffline(row)">强制下架</el-button>
            </template>
            <el-button type="info" size="small" @click="handleDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="limit"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchList"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    <!-- 驳回弹窗 -->
    <el-dialog v-model="rejectVisible" title="驳回申请" width="400px">
      <el-form :model="rejectForm">
        <el-form-item label="驳回理由" required>
          <el-input
            v-model="rejectForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入驳回理由"
            maxlength="200"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="confirmReject">确认驳回</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="俱乐部详情" width="500px">
      <el-descriptions v-if="currentRow" :column="1" border>
        <el-descriptions-item label="俱乐部名称">{{ currentRow.club_name }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="currentRow.badge_type === 'blue_v' ? 'primary' : 'success'" size="small">
            {{ currentRow.badge_type === 'blue_v' ? '企业级俱乐部' : '个人级俱乐部' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创始人">{{ currentRow.user?.nickname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创始人ID">{{ currentRow.user_id }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag v-if="currentRow.audit_status === 0" type="warning">待审核</el-tag>
          <el-tag v-else-if="currentRow.audit_status === 1" type="success">已通过</el-tag>
          <el-tag v-else type="danger">已驳回</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="V标状态">
          {{ currentRow.is_active ? '已点亮' : '未点亮' }}
        </el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ currentRow.create_time }}</el-descriptions-item>
        <el-descriptions-item v-if="currentRow.audit_time" label="审核时间">{{ currentRow.audit_time }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'AuditClub',
  data() {
    return {
      searchForm: {
        auditStatus: null,
        clubType: ''
      },
      tableData: [],
      loading: false,
      page: 1,
      limit: 20,
      total: 0,
      rejectVisible: false,
      rejectLoading: false,
      rejectForm: { id: 0, remark: '' },
      detailVisible: false,
      currentRow: null
    }
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/audit/club_list', {
          page: this.page,
          limit: this.limit,
          audit_status: this.searchForm.auditStatus,
          club_type: this.searchForm.clubType
        })
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) {
        ElMessage.error('加载失败')
      } finally {
        this.loading = false
      }
    },
    async handleApprove(row) {
      try {
        await ElMessageBox.confirm(`确定通过俱乐部"${row.club_name}"的入驻申请吗？`, '确认通过', { type: 'success' })
        await request.put('/audit/approve', { id: row.id, type: 'club' })
        ElMessage.success('已通过')
        this.fetchList()
      } catch (e) {
        if (e !== 'cancel') ElMessage.error('操作失败')
      }
    },
    handleReject(row) {
      this.rejectForm = { id: row.id, remark: '' }
      this.rejectVisible = true
    },
    async confirmReject() {
      if (!this.rejectForm.remark.trim()) {
        ElMessage.warning('请填写驳回理由')
        return
      }
      this.rejectLoading = true
      try {
        await request.put('/audit/reject', {
          id: this.rejectForm.id,
          type: 'club',
          remark: this.rejectForm.remark
        })
        ElMessage.success('已驳回')
        this.rejectVisible = false
        this.fetchList()
      } catch (e) {
        ElMessage.error('操作失败')
      } finally {
        this.rejectLoading = false
      }
    },
    async handleForceOffline(row) {
      try {
        await ElMessageBox.confirm(
          `确定强制下架俱乐部"${row.club_name}"吗？下架后V标将熄灭，此操作不可撤销。`,
          '强制下架确认',
          { type: 'warning', confirmButtonText: '确认下架' }
        )
        await request.put('/audit/force_offline', { id: row.id, type: 'club' })
        ElMessage.success('已强制下架')
        this.fetchList()
      } catch (e) {
        if (e !== 'cancel') ElMessage.error('操作失败')
      }
    },
    handleDetail(row) {
      this.currentRow = row
      this.detailVisible = true
    }
  }
}
</script>

<style lang="scss" scoped>
.v-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  font-size: 12px;
  font-weight: bold;
  color: #fff;
}
.blue-v {
  background: linear-gradient(135deg, #1890ff, #096dd9);
}
.green-v {
  background: linear-gradient(135deg, #52c41a, #389e0d);
}
</style>