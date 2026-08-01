<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">UP主管理</span>
    </div>

    <el-tabs v-model="activeTab" style="margin-top: 0;">
      <el-tab-pane label="UP主申请列表" name="apply">
        <el-card class="table-card">
          <div class="search-row">
            <el-form :inline="true" :model="searchForm">
              <el-form-item label="申请人UID">
                <el-input
                  v-model="searchForm.user_id"
                  placeholder="用户ID"
                  clearable
                  style="width: 160px"
                />
              </el-form-item>
              <el-form-item label="平台">
                <el-select v-model="searchForm.platform" placeholder="全部" clearable style="width: 140px">
                  <el-option label="抖音" value="douyin" />
                  <el-option label="B站" value="bilibili" />
                  <el-option label="快手" value="kuaishou" />
                  <el-option label="小红书" value="xiaohongshu" />
                  <el-option label="视频号" value="wechat_channel" />
                  <el-option label="其他" value="other" />
                </el-select>
              </el-form-item>
              <el-form-item label="状态">
                <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
                  <el-option label="待审核" value="pending" />
                  <el-option label="已通过" value="approved" />
                  <el-option label="已驳回" value="rejected" />
                  <el-option label="已吊销" value="revoked" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
                <el-button :icon="Refresh" @click="handleReset">重置</el-button>
              </el-form-item>
            </el-form>
          </div>

          <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="user_id" label="申请人UID" width="110" align="center" />
            <el-table-column label="平台" width="110" align="center">
              <template #default="{ row }">
                <el-tag :type="platformTagType(row.platform)" size="small">
                  {{ platformLabel(row.platform) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="account_name" label="账号名称" min-width="140" show-overflow-tooltip />
            <el-table-column prop="account_url" label="账号链接" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <el-link v-if="row.account_url" :href="row.account_url" target="_blank" type="primary">
                  {{ row.account_url }}
                </el-link>
                <span v-else style="color: #c0c4cc;">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="followers_count" label="粉丝数" width="120" align="center">
              <template #default="{ row }">
                {{ formatFans(row.followers_count) }}
              </template>
            </el-table-column>
            <el-table-column label="档位" width="120" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.tier_name" :type="tierTagType(row.tier_level)" size="small">
                  {{ row.tier_name }}
                </el-tag>
                <span v-else style="color: #c0c4cc;">-</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="110" align="center">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">
                  {{ statusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="apply_time" label="申请时间" width="170" align="center" />
            <el-table-column label="操作" width="260" fixed="right" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 'pending'"
                  type="success"
                  link
                  size="small"
                  @click="handleApprove(row)"
                >
                  通过
                </el-button>
                <el-button
                  v-if="row.status === 'pending'"
                  type="danger"
                  link
                  size="small"
                  @click="handleReject(row)"
                >
                  驳回
                </el-button>
                <el-button
                  v-if="row.status === 'approved'"
                  type="warning"
                  link
                  size="small"
                  @click="handleRevoke(row)"
                >
                  吊销
                </el-button>
                <el-button
                  type="primary"
                  link
                  size="small"
                  @click="handleViewDetail(row)"
                >
                  详情
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
              @size-change="fetchList"
              @current-change="fetchList"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="档位配置" name="tier">
        <el-card class="table-card">
          <div class="table-toolbar">
            <el-button type="primary" :icon="Plus" @click="handleTierAdd">新增档位</el-button>
          </div>

          <el-table :data="tierList" v-loading="tierLoading" stripe border style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="name" label="档位名称" min-width="140" />
            <el-table-column label="等级" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="tierTagType(row.level)" size="small">L{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="粉丝区间" width="200" align="center">
              <template #default="{ row }">
                {{ formatFans(row.min_followers) }} - {{ formatFans(row.max_followers) }}
              </template>
            </el-table-column>
            <el-table-column label="分成比例" width="120" align="center">
              <template #default="{ row }">
                <span style="color: #f56c6c; font-weight: 600;">{{ row.commission_rate }}%</span>
              </template>
            </el-table-column>
            <el-table-column prop="monthly_bonus" label="月度额外奖励" width="140" align="center">
              <template #default="{ row }">¥{{ row.monthly_bonus || 0 }}</template>
            </el-table-column>
            <el-table-column prop="description" label="说明" min-width="200" show-overflow-tooltip />
            <el-table-column label="操作" width="180" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleTierEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleTierDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="detailVisible" title="UP主申请详情" width="640px">
      <el-descriptions v-if="currentDetail" :column="2" border>
        <el-descriptions-item label="申请ID">{{ currentDetail.id }}</el-descriptions-item>
        <el-descriptions-item label="申请人UID">{{ currentDetail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="平台">
          <el-tag :type="platformTagType(currentDetail.platform)" size="small">
            {{ platformLabel(currentDetail.platform) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="账号名称">{{ currentDetail.account_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="粉丝数">{{ formatFans(currentDetail.followers_count) }}</el-descriptions-item>
        <el-descriptions-item label="申请档位">
          {{ currentDetail.tier_name || currentDetail.tier_level ? 'L' + currentDetail.tier_level + ' ' + (currentDetail.tier_name || '') : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="账号链接" :span="2">
          <el-link v-if="currentDetail.account_url" :href="currentDetail.account_url" target="_blank" type="primary">
            {{ currentDetail.account_url }}
          </el-link>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="真实姓名">{{ currentDetail.real_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ currentDetail.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="当前状态">
          <el-tag :type="statusTagType(currentDetail.status)" size="small">
            {{ statusLabel(currentDetail.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ currentDetail.apply_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注说明" :span="2">{{ currentDetail.remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="驳回/吊销原因" :span="2">
          {{ currentDetail.reject_reason || currentDetail.revoke_reason || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="tierDialogVisible" :title="tierForm.id ? '编辑档位' : '新增档位'" width="520px">
      <el-form ref="tierFormRef" :model="tierForm" :rules="tierRules" label-width="120px">
        <el-form-item label="档位名称" prop="name">
          <el-input v-model="tierForm.name" placeholder="请输入档位名称，如：腰部UP主" maxlength="50" />
        </el-form-item>
        <el-form-item label="等级" prop="level">
          <el-input-number v-model="tierForm.level" :min="1" :max="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最小粉丝数" prop="min_followers">
          <el-input-number v-model="tierForm.min_followers" :min="0" :step="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大粉丝数" prop="max_followers">
          <el-input-number v-model="tierForm.max_followers" :min="0" :step="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="分成比例(%)" prop="commission_rate">
          <el-input-number v-model="tierForm.commission_rate" :min="0" :max="100" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="月度奖励(元)" prop="monthly_bonus">
          <el-input-number v-model="tierForm.monthly_bonus" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="说明" prop="description">
          <el-input
            v-model="tierForm.description"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请输入档位说明"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tierDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="tierSubmitLoading" @click="handleTierSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="rejectDialogVisible" title="驳回申请" width="480px">
      <el-form :model="rejectForm" label-width="100px">
        <el-form-item label="驳回原因" prop="reason">
          <el-input
            v-model="rejectForm.reason"
            type="textarea"
            :rows="4"
            placeholder="请输入驳回原因"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="actionLoading" @click="confirmReject">确认驳回</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="revokeDialogVisible" title="吊销认证" width="480px">
      <el-form :model="revokeForm" label-width="100px">
        <el-form-item label="吊销原因" prop="reason">
          <el-input
            v-model="revokeForm.reason"
            type="textarea"
            :rows="4"
            placeholder="请输入吊销原因"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="revokeDialogVisible = false">取消</el-button>
        <el-button type="warning" :loading="actionLoading" @click="confirmRevoke">确认吊销</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'UpMasterManage',
  data() {
    return {
      Search,
      Refresh,
      Plus,
      activeTab: 'apply',
      searchForm: {
        user_id: '',
        platform: '',
        status: ''
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      detailVisible: false,
      currentDetail: null,
      actionLoading: false,
      rejectDialogVisible: false,
      rejectForm: { id: 0, reason: '' },
      revokeDialogVisible: false,
      revokeForm: { id: 0, reason: '' },
      tierList: [],
      tierLoading: false,
      tierDialogVisible: false,
      tierSubmitLoading: false,
      tierForm: {
        id: 0,
        name: '',
        level: 1,
        min_followers: 0,
        max_followers: 0,
        commission_rate: 50,
        monthly_bonus: 0,
        description: ''
      },
      tierRules: {
        name: [{ required: true, message: '请输入档位名称', trigger: 'blur' }],
        level: [{ required: true, message: '请选择等级', trigger: 'change' }],
        commission_rate: [{ required: true, message: '请输入分成比例', trigger: 'blur' }]
      }
    }
  },
  mounted() {
    this.fetchList()
    this.fetchTierList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/up_master/list', {
          params: {
            page: this.pagination.page,
            limit: this.pagination.pageSize,
            user_id: this.searchForm.user_id,
            platform: this.searchForm.platform,
            status: this.searchForm.status
          }
        })
        this.tableData = res.data?.list || res.data || []
        this.pagination.total = res.data?.total || res.total || this.tableData.length
      } finally {
        this.loading = false
      }
    },
    handleReset() {
      this.searchForm = { user_id: '', platform: '', status: '' }
      this.pagination.page = 1
      this.fetchList()
    },
    async handleApprove(row) {
      try {
        await ElMessageBox.confirm(
          `确定通过 UID「${row.user_id}」的UP主认证申请吗？`,
          '确认通过',
          { confirmButtonText: '通过', cancelButtonText: '取消', type: 'success' }
        )
      } catch (e) {
        return
      }
      this.actionLoading = true
      try {
        await request.post('/up_master/approve', { id: row.id })
        ElMessage.success('审核通过')
        this.fetchList()
      } catch (err) {
        ElMessage.error(err.message || '操作失败')
      } finally {
        this.actionLoading = false
      }
    },
    handleReject(row) {
      this.rejectForm = { id: row.id, reason: '' }
      this.rejectDialogVisible = true
    },
    async confirmReject() {
      if (!this.rejectForm.reason.trim()) {
        ElMessage.warning('请输入驳回原因')
        return
      }
      this.actionLoading = true
      try {
        await request.post('/up_master/reject', {
          id: this.rejectForm.id,
          reason: this.rejectForm.reason.trim()
        })
        ElMessage.success('已驳回')
        this.rejectDialogVisible = false
        this.fetchList()
      } catch (err) {
        ElMessage.error(err.message || '操作失败')
      } finally {
        this.actionLoading = false
      }
    },
    handleRevoke(row) {
      this.revokeForm = { id: row.id, reason: '' }
      this.revokeDialogVisible = true
    },
    async confirmRevoke() {
      if (!this.revokeForm.reason.trim()) {
        ElMessage.warning('请输入吊销原因')
        return
      }
      this.actionLoading = true
      try {
        await request.post('/up_master/revoke', {
          id: this.revokeForm.id,
          reason: this.revokeForm.reason.trim()
        })
        ElMessage.success('已吊销')
        this.revokeDialogVisible = false
        this.fetchList()
      } catch (err) {
        ElMessage.error(err.message || '操作失败')
      } finally {
        this.actionLoading = false
      }
    },
    async handleViewDetail(row) {
      this.detailVisible = true
      this.currentDetail = null
      try {
        const res = await request.get('/up_master/detail', { params: { id: row.id } })
        this.currentDetail = res.data || row
      } catch (e) {
        this.currentDetail = row
      }
    },
    async fetchTierList() {
      this.tierLoading = true
      try {
        const res = await request.get('/up_master/tier_config/list')
        this.tierList = res.data?.list || res.data || []
      } finally {
        this.tierLoading = false
      }
    },
    handleTierAdd() {
      this.tierForm = {
        id: 0,
        name: '',
        level: 1,
        min_followers: 0,
        max_followers: 0,
        commission_rate: 50,
        monthly_bonus: 0,
        description: ''
      }
      this.$refs.tierFormRef?.clearValidate()
      this.tierDialogVisible = true
    },
    handleTierEdit(row) {
      this.tierForm = { ...row }
      this.$refs.tierFormRef?.clearValidate()
      this.tierDialogVisible = true
    },
    async handleTierSubmit() {
      const valid = await this.$refs.tierFormRef?.validate().catch(() => false)
      if (!valid) return
      this.tierSubmitLoading = true
      try {
        const url = this.tierForm.id
          ? '/up_master/tier_config/update'
          : '/up_master/tier_config/create'
        const method = this.tierForm.id ? 'put' : 'post'
        const res = await request[method](url, this.tierForm)
        if (res.code === 200 || !res.code) {
          ElMessage.success(this.tierForm.id ? '修改成功' : '创建成功')
          this.tierDialogVisible = false
          this.fetchTierList()
        } else {
          ElMessage.error(res.msg || '操作失败')
        }
      } finally {
        this.tierSubmitLoading = false
      }
    },
    async handleTierDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定删除档位「${row.name}」吗？`,
          '删除确认',
          { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
        )
      } catch (e) {
        return
      }
      try {
        await request.delete('/up_master/tier_config/delete', { data: { id: row.id } })
        ElMessage.success('删除成功')
        this.fetchTierList()
      } catch (err) {
        ElMessage.error(err.message || '删除失败')
      }
    },
    platformLabel(p) {
      const map = {
        douyin: '抖音',
        bilibili: 'B站',
        kuaishou: '快手',
        xiaohongshu: '小红书',
        wechat_channel: '视频号',
        other: '其他'
      }
      return map[p] || p || '-'
    },
    platformTagType(p) {
      const map = {
        douyin: 'danger',
        bilibili: 'primary',
        kuaishou: 'warning',
        xiaohongshu: 'danger',
        wechat_channel: 'success',
        other: 'info'
      }
      return map[p] || 'info'
    },
    statusLabel(s) {
      const map = {
        pending: '待审核',
        approved: '已通过',
        rejected: '已驳回',
        revoked: '已吊销'
      }
      return map[s] || s || '未知'
    },
    statusTagType(s) {
      const map = {
        pending: 'warning',
        approved: 'success',
        rejected: 'info',
        revoked: 'danger'
      }
      return map[s] || 'info'
    },
    tierTagType(level) {
      const lv = parseInt(level) || 1
      if (lv >= 5) return 'danger'
      if (lv >= 3) return 'warning'
      if (lv >= 2) return 'primary'
      return 'info'
    },
    formatFans(count) {
      const n = parseInt(count) || 0
      if (n >= 10000) return (n / 10000).toFixed(1) + '万'
      return n.toString()
    }
  }
}
</script>

<style lang="scss" scoped>
.page-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
}

.table-card {
  .search-row {
    margin-bottom: 16px;
  }
  .table-toolbar {
    margin-bottom: 16px;
  }
  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
