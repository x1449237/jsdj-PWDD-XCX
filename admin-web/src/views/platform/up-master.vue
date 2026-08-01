<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">UP主认证</span>
    </div>

    <el-card class="table-card">
      <div class="table-toolbar">
        <div class="toolbar-left">
          <el-select
            v-model="filters.audit_status"
            placeholder="审核状态"
            clearable
            style="width: 140px"
            @change="handleFilterChange"
          >
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已驳回" value="rejected" />
          </el-select>
          <el-select
            v-model="filters.tier"
            placeholder="等级"
            clearable
            style="width: 140px; margin-left: 12px"
            @change="handleFilterChange"
          >
            <el-option
              v-for="item in tierOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-input
            v-model="filters.club_id"
            placeholder="俱乐部ID"
            clearable
            style="width: 140px; margin-left: 12px"
            @change="handleFilterChange"
          />
        </div>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="nickname" label="用户昵称" min-width="120" show-overflow-tooltip />
        <el-table-column prop="club_name" label="所属俱乐部" min-width="140" show-overflow-tooltip />
        <el-table-column label="申请等级" width="120" align="center">
          <template #default="{ row }">
            <el-tag
              :style="{ backgroundColor: getTierColor(row.tier), color: '#fff', border: 'none' }"
              size="small"
            >
              {{ getTierLabel(row.tier) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="fan_count" label="粉丝数" width="100" align="center" />
        <el-table-column prop="platform" label="主平台" width="100" align="center" />
        <el-table-column prop="platform_account" label="平台账号" min-width="140" show-overflow-tooltip />
        <el-table-column label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.audit_status)" size="small">
              {{ getStatusLabel(row.audit_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="170" align="center" />
        <el-table-column label="操作" width="260" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">
              查看详情
            </el-button>
            <template v-if="row.audit_status === 'pending'">
              <el-button type="success" link size="small" @click="handleApprove(row)">
                通过
              </el-button>
              <el-button type="danger" link size="small" @click="handleReject(row)">
                驳回
              </el-button>
            </template>
            <template v-if="row.audit_status === 'approved'">
              <el-button type="danger" link size="small" @click="handleRevoke(row)">
                吊销
              </el-button>
            </template>
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
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 查看详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="认证详情"
      width="600px"
      :close-on-click-modal="false"
    >
      <template v-if="currentDetail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户昵称">{{ currentDetail.nickname }}</el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ currentDetail.user_id }}</el-descriptions-item>
          <el-descriptions-item label="申请等级">
            <el-tag
              :style="{ backgroundColor: getTierColor(currentDetail.tier), color: '#fff', border: 'none' }"
              size="small"
            >
              {{ getTierLabel(currentDetail.tier) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="粉丝数">{{ currentDetail.fan_count }}</el-descriptions-item>
          <el-descriptions-item label="主平台">{{ currentDetail.platform }}</el-descriptions-item>
          <el-descriptions-item label="平台账号">{{ currentDetail.platform_account }}</el-descriptions-item>
          <el-descriptions-item label="审核状态">
            <el-tag :type="getStatusType(currentDetail.audit_status)" size="small">
              {{ getStatusLabel(currentDetail.audit_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="申请时间">{{ currentDetail.created_at }}</el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.platform_url" label="平台主页" :span="2">
            <a :href="currentDetail.platform_url" target="_blank" rel="noopener noreferrer">
              {{ currentDetail.platform_url }}
            </a>
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.fan_screenshot" label="粉丝数截图" :span="2">
            <el-image
              :src="currentDetail.fan_screenshot"
              style="max-width: 100%; max-height: 400px"
              fit="contain"
              :preview-src-list="[currentDetail.fan_screenshot]"
              preview-teleported
            />
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.remark" label="审核备注" :span="2">
            {{ currentDetail.remark }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">俱乐部信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="所属俱乐部">{{ currentDetail.club_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="俱乐部类型">{{ currentDetail.club_badge_type || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">录屏视频（从手机桌面→进入平台→个人主页）</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="视频链接">
            <a v-if="currentDetail.video_url" :href="currentDetail.video_url" target="_blank" style="color: #409EFF">
              {{ currentDetail.video_url }}
            </a>
            <span v-else style="color: #999">未上传</span>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="currentDetail.video_url" style="margin-top: 12px">
          <video
            :src="currentDetail.video_url"
            controls
            style="width: 100%; max-width: 400px; border-radius: 8px; background: #000"
            preload="metadata"
          >
            您的浏览器不支持视频播放
          </video>
        </div>
      </template>
    </el-dialog>

    <!-- 审核通过弹窗 -->
    <el-dialog
      v-model="approveDialogVisible"
      title="审核通过"
      width="480px"
      :close-on-click-modal="false"
      @closed="handleApproveDialogClosed"
    >
      <el-form ref="approveFormRef" :model="approveForm" :rules="approveFormRules" label-width="110px">
        <el-form-item label="核验粉丝数" prop="verified_fan_count">
          <el-input-number
            v-model="approveForm.verified_fan_count"
            :min="0"
            :controls="false"
            placeholder="请输入核验粉丝数"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="审核备注" prop="remark">
          <el-input
            v-model="approveForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入审核备注（选填）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitApprove">确定通过</el-button>
      </template>
    </el-dialog>

    <!-- 审核驳回弹窗 -->
    <el-dialog
      v-model="rejectDialogVisible"
      title="审核驳回"
      width="480px"
      :close-on-click-modal="false"
      @closed="handleRejectDialogClosed"
    >
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectFormRules" label-width="110px">
        <el-form-item label="驳回原因" prop="remark">
          <el-input
            v-model="rejectForm.remark"
            type="textarea"
            :rows="4"
            placeholder="请输入驳回原因（必填）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="submitLoading" @click="submitReject">确定驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessageBox, ElMessage } from 'element-plus'

const TIER_MAP = {
  1: { label: '青铜UP主', color: '#1A1A1A' },
  2: { label: '进阶UP主', color: '#B8B8B8' },
  3: { label: '高阶UP主', color: '#7B2FBE' },
  4: { label: '精英UP主', color: '#E87A2A' },
  5: { label: '巨匠UP主', color: '#C44A6C' },
  6: { label: '至尊UP主', color: '#8B1A2B' }
}

const STATUS_MAP = {
  pending: { label: '待审核', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' }
}

export default {
  name: 'PlatformUpMaster',
  data() {
    return {
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      filters: {
        audit_status: '',
        tier: '',
        club_id: ''
      },
      tierOptions: Object.entries(TIER_MAP).map(([value, item]) => ({
        value: Number(value),
        label: item.label
      })),
      detailDialogVisible: false,
      currentDetail: null,
      approveDialogVisible: false,
      approveForm: {
        id: null,
        verified_fan_count: 0,
        remark: ''
      },
      approveFormRules: {
        verified_fan_count: [
          { required: true, message: '请输入核验粉丝数', trigger: 'blur' }
        ]
      },
      approveFormRef: null,
      rejectDialogVisible: false,
      rejectForm: {
        id: null,
        remark: ''
      },
      rejectFormRules: {
        remark: [
          { required: true, message: '请输入驳回原因', trigger: 'blur' },
          { min: 1, max: 200, message: '驳回原因长度在 1 到 200 个字符', trigger: 'blur' }
        ]
      },
      rejectFormRef: null,
      submitLoading: false
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
          limit: this.pagination.pageSize
        }
        if (this.filters.audit_status) {
          params.audit_status = this.filters.audit_status
        }
        if (this.filters.tier) {
          params.tier = this.filters.tier
        }
        if (this.filters.club_id) {
          params.club_id = this.filters.club_id
        }
        const res = await request.get('/up_master/list', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取UP主认证列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleFilterChange() {
      this.pagination.page = 1
      this.fetchList()
    },
    getTierLabel(tier) {
      return TIER_MAP[tier]?.label || `等级${tier}`
    },
    getTierColor(tier) {
      return TIER_MAP[tier]?.color || '#999'
    },
    getStatusLabel(status) {
      return STATUS_MAP[status]?.label || status
    },
    getStatusType(status) {
      return STATUS_MAP[status]?.type || 'info'
    },
    handleViewDetail(row) {
      this.loadDetail(row.id)
    },
    async loadDetail(id) {
      try {
        const res = await request.get('/up_master/detail', { params: { id } })
        const data = res.data || {}
        this.currentDetail = {
          ...data,
          club_name: data.club_name || '',
          club_badge_type: data.club_badge_type || '',
          video_url: data.video_url || ''
        }
        this.detailDialogVisible = true
      } catch (err) {
        console.error('获取认证详情失败:', err)
        ElMessage.error('获取认证详情失败')
      }
    },
    handleApprove(row) {
      this.approveForm = {
        id: row.id,
        verified_fan_count: row.fan_count || 0,
        remark: ''
      }
      this.approveDialogVisible = true
    },
    handleApproveDialogClosed() {
      this.$refs.approveFormRef?.resetFields()
    },
    async submitApprove() {
      try {
        await this.$refs.approveFormRef.validate()
      } catch {
        return
      }
      this.submitLoading = true
      try {
        await request.post('/up_master/approve', {
          id: this.approveForm.id,
          verified_fan_count: this.approveForm.verified_fan_count,
          remark: this.approveForm.remark
        })
        ElMessage.success('审核通过')
        this.approveDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('审核通过失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    handleReject(row) {
      this.rejectForm = {
        id: row.id,
        remark: ''
      }
      this.rejectDialogVisible = true
    },
    handleRejectDialogClosed() {
      this.$refs.rejectFormRef?.resetFields()
    },
    async submitReject() {
      try {
        await this.$refs.rejectFormRef.validate()
      } catch {
        return
      }
      this.submitLoading = true
      try {
        await request.post('/up_master/reject', {
          id: this.rejectForm.id,
          remark: this.rejectForm.remark
        })
        ElMessage.success('已驳回')
        this.rejectDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('驳回失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    async handleRevoke(row) {
      try {
        await ElMessageBox.confirm(
          `确定要吊销「${row.nickname}」的UP主认证吗？此操作不可撤销。`,
          '吊销确认',
          { confirmButtonText: '确定吊销', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post('/up_master/revoke', { id: row.id })
        ElMessage.success('已吊销')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('吊销失败:', err)
        }
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.table-card {
  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .toolbar-left {
      display: flex;
      align-items: center;
    }
  }
}
</style>