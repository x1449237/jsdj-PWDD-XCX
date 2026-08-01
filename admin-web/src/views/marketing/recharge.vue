<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">充值活动配置</span>
    </div>

    <el-card class="section-card">
      <template #header>
        <span class="card-header">新建/编辑充值活动</span>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="活动名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入活动名称" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="充值金额" prop="recharge_amount">
              <el-input-number v-model="form.recharge_amount" :min="0" :precision="2" :step="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="赠送金额" prop="bonus_amount">
              <el-input-number v-model="form.bonus_amount" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="赠送类型" prop="bonus_type">
              <el-select v-model="form.bonus_type" placeholder="请选择" style="width: 100%">
                <el-option label="赠送余额" value="balance" />
                <el-option label="赠送优惠券" value="coupon" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="赠送优惠券ID" prop="bonus_coupon_id">
              <el-input-number v-model="form.bonus_coupon_id" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="开始时间" prop="start_time">
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
            <el-form-item label="结束时间" prop="end_time">
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

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>活动列表</span>
        </div>
      </template>

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
            <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
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
        <el-table-column prop="recharge_amount" label="充值金额" width="120" align="center">
          <template #default="{ row }">¥{{ row.recharge_amount }}</template>
        </el-table-column>
        <el-table-column prop="bonus_amount" label="赠送金额" width="120" align="center">
          <template #default="{ row }">¥{{ row.bonus_amount }}</template>
        </el-table-column>
        <el-table-column label="赠送类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.bonus_type === 'balance' ? 'success' : 'warning'">
              {{ row.bonus_type === 'balance' ? '余额' : '优惠券' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="活动时间" width="300" align="center">
          <template #default="{ row }">
            {{ row.start_time || '不限' }} ~ {{ row.end_time || '不限' }}
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
import { Plus, Search, Refresh, RefreshLeft } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'MarketingRecharge',
  data() {
    return {
      Plus,
      Search,
      Refresh,
      RefreshLeft,
      form: {
        id: 0,
        name: '',
        recharge_amount: 0,
        bonus_amount: 0,
        bonus_type: 'balance',
        bonus_coupon_id: 0,
        start_time: '',
        end_time: '',
        sort: 0,
        status: 1
      },
      rules: {
        name: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
        recharge_amount: [{ required: true, message: '请输入充值金额', trigger: 'blur' }]
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
        const res = await request.get('/marketing/recharge/list', {
          params: {
            page: this.pagination.page,
            limit: this.pagination.pageSize,
            keyword: this.searchForm.keyword,
            status: this.searchForm.status
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
    handleReset() {
      this.searchForm = { keyword: '', status: '' }
      this.pagination.page = 1
      this.fetchList()
    },
    resetForm() {
      this.form = {
        id: 0,
        name: '',
        recharge_amount: 0,
        bonus_amount: 0,
        bonus_type: 'balance',
        bonus_coupon_id: 0,
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
          ? '/marketing/recharge/update'
          : '/marketing/recharge/create'
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
        const res = await request.put('/marketing/recharge/toggle', { id: row.id })
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
      ElMessageBox.confirm(`确定删除活动「${row.name}」吗？`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        const res = await request.delete('/marketing/recharge/delete', { data: { id: row.id } })
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

.search-row {
  margin-bottom: 16px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
