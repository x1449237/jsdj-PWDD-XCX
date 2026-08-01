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
            <el-form-item label="游戏ID" prop="game_id">
              <el-input-number v-model="form.game_id" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="原价" prop="original_price">
              <el-input-number v-model="form.original_price" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="拼团价" prop="group_price">
              <el-input-number v-model="form.group_price" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
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
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="拼团时长(小时)" prop="duration_hours">
              <el-input-number v-model="form.duration_hours" :min="1" :step="1" style="width: 100%" />
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
          <el-form-item>
            <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
            <el-button :icon="Refresh" @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="name" label="活动名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="game_id" label="游戏ID" width="100" align="center" />
        <el-table-column label="原价/拼团价" width="160" align="center">
          <template #default="{ row }">
            <div>¥{{ row.original_price }}</div>
            <div style="color: #f56c6c">¥{{ row.group_price }}</div>
          </template>
        </el-table-column>
        <el-table-column label="人数要求" width="120" align="center">
          <template #default="{ row }">{{ row.min_people }}-{{ row.max_people }}人</template>
        </el-table-column>
        <el-table-column label="拼团时长" width="120" align="center">
          <template #default="{ row }">{{ row.duration_hours }}小时</template>
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
  name: 'MarketingGroupBuy',
  data() {
    return {
      Plus,
      Search,
      Refresh,
      RefreshLeft,
      form: {
        id: 0,
        game_id: 0,
        name: '',
        original_price: 0,
        group_price: 0,
        min_people: 3,
        max_people: 5,
        duration_hours: 24,
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
        keyword: ''
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
        const res = await request.get('/marketing/group_buy/list', {
          params: {
            page: this.pagination.page,
            limit: this.pagination.pageSize,
            keyword: this.searchForm.keyword
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
      this.searchForm = { keyword: '' }
      this.pagination.page = 1
      this.fetchList()
    },
    resetForm() {
      this.form = {
        id: 0,
        game_id: 0,
        name: '',
        original_price: 0,
        group_price: 0,
        min_people: 3,
        max_people: 5,
        duration_hours: 24,
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
        const res = await request.put('/marketing/group_buy/toggle', { id: row.id })
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
        const res = await request.delete('/marketing/group_buy/delete', { data: { id: row.id } })
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
