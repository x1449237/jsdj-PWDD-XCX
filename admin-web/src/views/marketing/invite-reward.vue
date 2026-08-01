<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">邀请奖励配置</span>
    </div>

    <el-card class="section-card">
      <template #header>
        <span class="card-header">新建/编辑奖励配置</span>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="奖励类型" prop="reward_type">
              <el-select v-model="form.reward_type" placeholder="请选择" style="width: 100%">
                <el-option label="余额奖励" value="balance" />
                <el-option label="优惠券奖励" value="coupon" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="奖励值" prop="reward_value">
              <el-input-number v-model="form.reward_value" :min="0" :precision="2" :step="1" style="width: 100%" />
              <div style="font-size: 12px; color: #999; margin-top: 4px">
                {{ form.reward_type === 'balance' ? '金额(元)' : '优惠券模板ID' }}
              </div>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="触发条件" prop="condition_type">
              <el-select v-model="form.condition_type" placeholder="请选择" style="width: 100%">
                <el-option label="首单完成" value="first_order" />
                <el-option label="实名完成" value="realname" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="条件值" prop="condition_value">
              <el-input v-model="form.condition_value" placeholder="如首单金额门槛" style="width: 100%" />
              <div style="font-size: 12px; color: #999; margin-top: 4px">
                {{ form.condition_type === 'first_order' ? '首单最低金额(元)，留空不限制' : '留空即可' }}
              </div>
            </el-form-item>
          </el-col>
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
            {{ form.id ? '保存修改' : '创建配置' }}
          </el-button>
          <el-button v-if="form.id" :icon="RefreshLeft" @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>奖励配置列表</span>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="奖励类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.reward_type === 'balance' ? 'success' : 'warning'">
              {{ row.reward_type === 'balance' ? '余额' : '优惠券' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="奖励值" width="120" align="center">
          <template #default="{ row }">
            {{ row.reward_type === 'balance' ? '¥' + row.reward_value : '优惠券#' + row.reward_value }}
          </template>
        </el-table-column>
        <el-table-column label="触发条件" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ row.condition_type === 'first_order' ? '首单完成' : '实名完成' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="condition_value" label="条件值" width="150" align="center">
          <template #default="{ row }">
            {{ row.condition_value || '无' }}
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
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
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
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
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus, RefreshLeft } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'MarketingInviteReward',
  data() {
    return {
      Plus,
      RefreshLeft,
      form: {
        id: 0,
        reward_type: 'balance',
        reward_value: 0,
        condition_type: 'first_order',
        condition_value: '',
        sort: 0,
        status: 1
      },
      rules: {
        reward_type: [{ required: true, message: '请选择奖励类型', trigger: 'change' }],
        reward_value: [{ required: true, message: '请输入奖励值', trigger: 'blur' }],
        condition_type: [{ required: true, message: '请选择触发条件', trigger: 'change' }]
      },
      submitLoading: false,
      toggleLoading: {},
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
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
        const res = await request.get('/marketing/invite_reward/list', {
          params: {
            page: this.pagination.page,
            limit: this.pagination.pageSize
          }
        })
        if (res.code === 200) {
          this.tableData = res.data.list || []
          this.pagination.total = res.data.total || 0
        }
      } finally {
        this.loading = false
      }
    },
    resetForm() {
      this.form = {
        id: 0,
        reward_type: 'balance',
        reward_value: 0,
        condition_type: 'first_order',
        condition_value: '',
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
          ? '/marketing/invite_reward/update'
          : '/marketing/invite_reward/create'
        const method = this.form.id ? 'put' : 'post'
        const res = await request[method](url, this.form)
        if (res.code === 200) {
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
        const res = await request.put('/marketing/invite_reward/toggle', { id: row.id })
        if (res.code === 200) {
          row.status = res.data.status
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
      ElMessageBox.confirm('确定删除该奖励配置吗？', '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        const res = await request.delete('/marketing/invite_reward/delete', { data: { id: row.id } })
        if (res.code === 200) {
          ElMessage.success('删除成功')
          this.fetchList()
        } else {
          ElMessage.error(res.msg || '删除失败')
        }
      }).catch(() => {})
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

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
