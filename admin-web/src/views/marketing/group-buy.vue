<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">拼团活动配置</span>
    </div>

    <el-card class="section-card">
      <template #header>
        <span class="card-header">新建/编辑拼团活动</span>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="活动名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入活动名称" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="对应服务ID" prop="service_id">
              <el-input-number v-model="form.service_id" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="游戏ID" prop="game_id">
              <el-input-number v-model="form.game_id" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="原价" prop="original_price">
              <el-input-number v-model="form.original_price" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="拼团价" prop="group_price">
              <el-input-number v-model="form.group_price" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="折扣比例(%)" prop="discount_ratio">
              <el-input-number v-model="form.discount_ratio" :min="1" :max="100" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="最少人数" prop="min_people">
              <el-input-number v-model="form.min_people" :min="2" :max="10" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="最多人数" prop="max_people">
              <el-input-number v-model="form.max_people" :min="2" :max="20" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="最低消费(元)" prop="min_consume">
              <el-input-number v-model="form.min_consume" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="有效期开始" prop="start_time">
              <el-date-picker
                v-model="form.start_time"
                type="datetime"
                placeholder="选择开始时间"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="有效期结束" prop="end_time">
              <el-date-picker
                v-model="form.end_time"
                type="datetime"
                placeholder="选择结束时间"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="拼团时长(小时)" prop="duration_hours">
              <el-input-number v-model="form.duration_hours" :min="1" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="状态" prop="status">
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-button type="primary" :icon="Plus" :loading="submitLoading" @click="handleSubmit">
            {{ form.id ? '保存修改' : '创建活动' }}
          </el-button>
          <el-button v-if="form.id" :icon="RefreshLeft" @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-tabs v-model="activeTab" style="margin-top: 20px;">
      <el-tab-pane label="活动列表" name="list">
        <el-card class="table-card">
          <div class="search-row">
            <el-form :inline="true" :model="searchForm">
              <el-form-item label="关键词">
                <el-input
                  v-model="searchForm.keyword"
                  placeholder="活动名称"
                  clearable
                  style="width: 200px"
                  @keyup.enter="fetchList"
                />
              </el-form-item>
              <el-form-item label="状态">
                <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px">
                  <el-option label="启用" :value="1" />
                  <el-option label="禁用" :value="0" />
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
            <el-table-column prop="name" label="活动名称" min-width="160" show-overflow-tooltip />
            <el-table-column label="原价/拼团价" width="160" align="center">
              <template #default="{ row }">
                <div>¥{{ row.original_price }}</div>
                <div style="color: #f56c6c">¥{{ row.group_price }}</div>
              </template>
            </el-table-column>
            <el-table-column label="折扣" width="90" align="center">
              <template #default="{ row }">
                <el-tag type="warning" size="small">{{ row.discount_ratio || 100 }}%</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="人数要求" width="120" align="center">
              <template #default="{ row }">{{ row.min_people }}-{{ row.max_people }}人</template>
            </el-table-column>
            <el-table-column label="最低消费" width="100" align="center">
              <template #default="{ row }">¥{{ row.min_consume || 0 }}</template>
            </el-table-column>
            <el-table-column label="有效期" width="200" align="center">
              <template #default="{ row }">
                <div style="font-size: 12px;">{{ row.start_time || '-' }}</div>
                <div style="font-size: 12px; color: #909399;">至 {{ row.end_time || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90" align="center">
              <template #default="{ row }">
                <el-switch
                  :model-value="row.status === 1"
                  :loading="toggleLoading[row.id]"
                  @change="handleToggle(row)"
                />
              </template>
            </el-table-column>
            <el-table-column prop="create_time" label="创建时间" width="170" align="center" />
            <el-table-column label="操作" width="300" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleMonitor(row)">成团监控</el-button>
                <el-button type="success" link size="small" @click="handleEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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

      <el-tab-pane label="成团监控" name="monitor">
        <el-card class="table-card">
          <div class="search-row">
            <el-form :inline="true" :model="monitorSearch">
              <el-form-item label="活动">
                <el-select v-model="monitorSearch.activity_id" placeholder="全部" clearable style="width: 180px">
                  <el-option
                    v-for="act in tableData"
                    :key="act.id"
                    :label="act.name"
                    :value="act.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="团状态">
                <el-select v-model="monitorSearch.group_status" placeholder="全部" clearable style="width: 140px">
                  <el-option label="拼团中" value="pending" />
                  <el-option label="成团成功" value="success" />
                  <el-option label="拼团失败" value="failed" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="Search" @click="fetchGroupList">搜索</el-button>
              </el-form-item>
            </el-form>
          </div>

          <el-table :data="groupList" v-loading="groupLoading" stripe border style="width: 100%">
            <el-table-column prop="id" label="团ID" width="100" align="center" />
            <el-table-column prop="activity_name" label="活动名称" min-width="160" show-overflow-tooltip />
            <el-table-column label="参团进度" width="160" align="center">
              <template #default="{ row }">
                <el-progress
                  :percentage="Math.round((row.current_people / row.max_people) * 100)"
                  :stroke-width="12"
                  :text-inside="true"
                  :status="row.current_people >= row.max_people ? 'success' : ''"
                />
                <div style="margin-top: 4px; font-size: 12px;">
                  {{ row.current_people }}/{{ row.max_people }}人
                </div>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="110" align="center">
              <template #default="{ row }">
                <el-tag
                  :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'"
                  size="small"
                >
                  {{ row.status === 'success' ? '成团成功' : row.status === 'failed' ? '拼团失败' : '拼团中' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="剩余时间" width="130" align="center">
              <template #default="{ row }">
                <span v-if="row.status === 'pending'" style="color: #e6a23c;">
                  {{ row.remain_time_text || '-' }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="group_price" label="拼团价" width="100" align="center">
              <template #default="{ row }">¥{{ row.group_price }}</template>
            </el-table-column>
            <el-table-column prop="creator_name" label="开团人" width="120" show-overflow-tooltip />
            <el-table-column prop="create_time" label="开团时间" width="170" align="center" />
            <el-table-column label="操作" width="260" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleViewMembers(row)">团员明细</el-button>
                <el-button
                  v-if="row.status === 'failed'"
                  type="warning"
                  link
                  size="small"
                  @click="handleRefundFailedGroup(row)"
                >
                  退款
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="groupPagination.page"
              v-model:page-size="groupPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="groupPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchGroupList"
              @current-change="fetchGroupList"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="membersVisible" :title="'团员明细 - 团号 ' + currentGroupId" width="700px">
      <el-table :data="membersList" v-loading="membersLoading" stripe style="width: 100%">
        <el-table-column type="index" label="#" width="60" align="center" />
        <el-table-column label="角色" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_creator ? 'warning' : 'info'" size="small">
              {{ row.is_creator ? '团长' : '团员' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="user_id" label="用户ID" width="100" align="center" />
        <el-table-column prop="nickname" label="昵称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" width="130" align="center">
          <template #default="{ row }">
            {{ row.phone ? row.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2') : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="paid_amount" label="支付金额" width="100" align="center">
          <template #default="{ row }">¥{{ row.paid_amount || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.paid_status === 'paid' ? 'success' : 'warning'" size="small">
              {{ row.paid_status === 'paid' ? '已支付' : '未支付' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="join_time" label="参团时间" width="170" align="center" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus, Search, Refresh, RefreshLeft } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'MarketingGroupBuy',
  data() {
    return {
      Plus,
      Search,
      Refresh,
      RefreshLeft,
      activeTab: 'list',
      form: {
        id: 0,
        game_id: 0,
        service_id: 0,
        name: '',
        original_price: 0,
        group_price: 0,
        discount_ratio: 100,
        min_consume: 0,
        min_people: 3,
        max_people: 5,
        duration_hours: 24,
        start_time: '',
        end_time: '',
        sort: 0,
        status: 1
      },
      rules: {
        name: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
        original_price: [{ required: true, message: '请输入原价', trigger: 'blur' }],
        group_price: [{ required: true, message: '请输入拼团价', trigger: 'blur' }]
      },
      submitLoading: false,
      toggleLoading: {},
      searchForm: {
        keyword: '',
        status: ''
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      monitorSearch: {
        activity_id: '',
        group_status: ''
      },
      groupList: [],
      groupLoading: false,
      groupPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      membersVisible: false,
      currentGroupId: 0,
      membersList: [],
      membersLoading: false
    }
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/marketing/group_buy/list', {
          params: {
            page: this.pagination.page,
            limit: this.pagination.pageSize,
            keyword: this.searchForm.keyword,
            status: this.searchForm.status
          }
        })
        if (res.code === 200 || res.data) {
          this.tableData = (res.data?.list || res.data || [])
          this.pagination.total = res.data?.total || res.total || this.tableData.length
        }
      } finally {
        this.loading = false
      }
    },
    handleReset() {
      this.searchForm = { keyword: '', status: '' }
      this.pagination.page = 1
      this.fetchList()
    },
    resetForm() {
      this.form = {
        id: 0,
        game_id: 0,
        service_id: 0,
        name: '',
        original_price: 0,
        group_price: 0,
        discount_ratio: 100,
        min_consume: 0,
        min_people: 3,
        max_people: 5,
        duration_hours: 24,
        start_time: '',
        end_time: '',
        sort: 0,
        status: 1
      }
      this.$refs.formRef?.clearValidate()
    },
    async handleSubmit() {
      const valid = await this.$refs.formRef.validate().catch(() => false)
      if (!valid) return

      this.submitLoading = true
      try {
        const url = this.form.id
          ? '/marketing/group_buy/update'
          : '/marketing/group_buy/create'
        const method = this.form.id ? 'put' : 'post'
        const res = await request[method](url, this.form)
        if (res.code === 200 || !res.code) {
          ElMessage.success(this.form.id ? '修改成功' : '创建成功')
          this.resetForm()
          this.fetchList()
        } else {
          ElMessage.error(res.msg || '操作失败')
        }
      } finally {
        this.submitLoading = false
      }
    },
    handleEdit(row) {
      this.form = { ...row }
      this.$refs.formRef?.clearValidate()
    },
    async handleToggle(row) {
      this.$set(this.toggleLoading, row.id, true)
      try {
        const res = await request.put('/marketing/group_buy/toggle', { id: row.id })
        if (res.code === 200 || !res.code) {
          row.status = res.data?.status ?? (row.status === 1 ? 0 : 1)
          ElMessage.success('状态更新成功')
        } else {
          ElMessage.error(res.msg || '操作失败')
          this.fetchList()
        }
      } finally {
        this.$set(this.toggleLoading, row.id, false)
      }
    },
    handleDelete(row) {
      ElMessageBox.confirm(`确定删除活动「${row.name}」吗？`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        const res = await request.delete('/marketing/group_buy/delete', { data: { id: row.id } })
        if (res.code === 200 || !res.code) {
          ElMessage.success('删除成功')
          this.fetchList()
        } else {
          ElMessage.error(res.msg || '删除失败')
        }
      }).catch(() => {})
    },
    handleMonitor(row) {
      if (row) {
        this.monitorSearch.activity_id = row.id
      }
      this.activeTab = 'monitor'
      this.fetchGroupList()
    },
    async fetchGroupList() {
      this.groupLoading = true
      try {
        const res = await request.get('/marketing/group_buy/groups', {
          params: {
            page: this.groupPagination.page,
            limit: this.groupPagination.pageSize,
            activity_id: this.monitorSearch.activity_id,
            status: this.monitorSearch.group_status
          }
        })
        this.groupList = res.data?.list || res.data || []
        this.groupPagination.total = res.data?.total || res.total || this.groupList.length
      } finally {
        this.groupLoading = false
      }
    },
    async handleViewMembers(row) {
      this.currentGroupId = row.id
      this.membersVisible = true
      this.membersLoading = true
      try {
        const res = await request.get(`/marketing/group_buy/groups/${row.id}/members`)
        this.membersList = res.data?.list || res.data || []
      } finally {
        this.membersLoading = false
      }
    },
    async handleRefundFailedGroup(row) {
      try {
        await ElMessageBox.confirm(
          `确定对拼团失败（团号 ${row.id}）进行退款操作吗？将对所有已支付团员原路退款。`,
          '退款确认',
          { confirmButtonText: '确认退款', cancelButtonText: '取消', type: 'warning' }
        )
      } catch (e) {
        return
      }
      try {
        await request.post(`/marketing/group_buy/groups/${row.id}/refund`)
        ElMessage.success('退款申请已提交')
        this.fetchGroupList()
      } catch (err) {
        ElMessage.error(err.message || '退款失败')
      }
    }
  }
}
</script>

<style scoped lang="scss">
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

.section-card {
  margin-bottom: 20px;
}

.card-header {
  font-weight: 600;
}

.search-row {
  margin-bottom: 16px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
