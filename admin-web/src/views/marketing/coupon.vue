<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">优惠券模板管理</span>
    </div>

    <el-card class="section-card">
      <template #header>
        <span class="card-header">新建/编辑优惠券</span>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="优惠券名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入优惠券名称" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="优惠券类型" prop="type">
              <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
                <el-option label="满减券" value="full_reduction" />
                <el-option label="新人券" value="new_user" />
                <el-option label="补偿券" value="compensation" />
                <el-option label="俱乐部专属" value="club_exclusive" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="优惠金额" prop="value">
              <el-input-number v-model="form.value" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="最低消费" prop="min_amount">
              <el-input-number v-model="form.min_amount" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="发放总量" prop="total_count">
              <el-input-number v-model="form.total_count" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="有效期(天)" prop="validity_days">
              <el-input-number v-model="form.validity_days" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="生效时间" prop="start_time">
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
            <el-form-item label="适用范围" prop="applicable_scope">
              <el-select v-model="form.applicable_scope" placeholder="请选择" style="width: 100%">
                <el-option label="全部通用" value="all" />
                <el-option label="指定游戏" value="game" />
                <el-option label="指定俱乐部" value="club" />
              </el-select>
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
            {{ form.id ? '保存修改' : '创建优惠券' }}
          </el-button>
          <el-button v-if="form.id" :icon="RefreshLeft" @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>优惠券列表</span>
        </div>
      </template>

      <div class="search-row">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="关键词">
            <el-input
              v-model="searchForm.keyword"
              placeholder="优惠券名称"
              clearable
              style="width: 200px"
              @keyup.enter="fetchList"
            />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 140px">
              <el-option label="满减券" value="full_reduction" />
              <el-option label="新人券" value="new_user" />
              <el-option label="补偿券" value="compensation" />
              <el-option label="俱乐部专属" value="club_exclusive" />
            </el-select>
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
        <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="优惠金额" width="100" align="center">
          <template #default="{ row }">¥{{ row.value }}</template>
        </el-table-column>
        <el-table-column prop="min_amount" label="最低消费" width="100" align="center">
          <template #default="{ row }">¥{{ row.min_amount }}</template>
        </el-table-column>
        <el-table-column label="发放/已用" width="120" align="center">
          <template #default="{ row }">
            {{ row.total_count === 0 ? '不限' : row.total_count }} / {{ row.used_count }}
          </template>
        </el-table-column>
        <el-table-column label="有效期" width="160" align="center">
          <template #default="{ row }">
            {{ row.validity_days > 0 ? row.validity_days + '天' : (row.start_time + ' ~ ' + row.end_time) }}
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
  name: 'MarketingCoupon',
  data() {
    return {
      Plus,
      Search,
      Refresh,
      RefreshLeft,
      form: {
        id: 0,
        name: '',
        type: 'full_reduction',
        value: 0,
        min_amount: 0,
        total_count: 0,
        validity_days: 0,
        start_time: '',
        end_time: '',
        applicable_scope: 'all',
        applicable_id: 0,
        sort: 0,
        status: 1
      },
      rules: {
        name: [{ required: true, message: '请输入优惠券名称', trigger: 'blur' }],
        type: [{ required: true, message: '请选择优惠券类型', trigger: 'change' }],
        value: [{ required: true, message: '请输入优惠金额', trigger: 'blur' }]
      },
      submitLoading: false,
      toggleLoading: {},
      searchForm: {
        keyword: '',
        type: '',
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
    typeLabel(type) {
      const map = {
        full_reduction: '满减券',
        new_user: '新人券',
        compensation: '补偿券',
        club_exclusive: '俱乐部专属'
      }
      return map[type] || type
    },
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/marketing/coupon/list', {
          params: {
            page: this.pagination.page,
            limit: this.pagination.pageSize,
            keyword: this.searchForm.keyword,
            type: this.searchForm.type,
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
      this.searchForm = { keyword: '', type: '', status: '' }
      this.pagination.page = 1
      this.fetchList()
    },
    resetForm() {
      this.form = {
        id: 0,
        name: '',
        type: 'full_reduction',
        value: 0,
        min_amount: 0,
        total_count: 0,
        validity_days: 0,
        start_time: '',
        end_time: '',
        applicable_scope: 'all',
        applicable_id: 0,
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
          ? '/marketing/coupon/update'
          : '/marketing/coupon/create'
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
        const res = await request.put('/marketing/coupon/toggle', { id: row.id })
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
      ElMessageBox.confirm(`确定删除优惠券「${row.name}」吗？`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        const res = await request.delete('/marketing/coupon/delete', { data: { id: row.id } })
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
