<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">抽奖活动配置</span>
    </div>

    <el-card class="section-card">
      <template #header>
        <span class="card-header">新建/编辑抽奖活动</span>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="活动名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入活动名称" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="活动类型" prop="type">
              <el-select v-model="form.type" placeholder="请选择" style="width: 100%">
                <el-option label="幸运转盘" value="wheel" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="消耗类型" prop="cost_type">
              <el-select v-model="form.cost_type" placeholder="请选择" style="width: 100%">
                <el-option label="免费" value="free" />
                <el-option label="余额" value="balance" />
                <el-option label="积分" value="points" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="消耗值" prop="cost_value">
              <el-input-number v-model="form.cost_value" :min="0" :precision="2" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="每日限制" prop="daily_limit">
              <el-input-number v-model="form.daily_limit" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="总限制" prop="total_limit">
              <el-input-number v-model="form.total_limit" :min="0" :step="1" style="width: 100%" />
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
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
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

    <el-card v-if="form.id" class="section-card">
      <template #header>
        <div class="card-header">
          <span>奖品配置</span>
          <el-button type="primary" size="small" :icon="Plus" @click="addPrize">添加奖品</el-button>
        </div>
      </template>

      <el-table :data="prizeList" border style="width: 100%">
        <el-table-column prop="sort" label="位置" width="80" align="center">
          <template #default="{ row, $index }">
            <el-input-number v-model="row.sort" :min="0" :step="1" size="small" style="width: 70px" />
          </template>
        </el-table-column>
        <el-table-column label="奖品名称" min-width="150">
          <template #default="{ row }">
            <el-input v-model="row.name" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="类型" width="130" align="center">
          <template #default="{ row }">
            <el-select v-model="row.type" size="small" style="width: 100%">
              <el-option label="优惠券" value="coupon" />
              <el-option label="免费时长" value="free_time" />
              <el-option label="余额" value="balance" />
              <el-option label="谢谢参与" value="thank" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="奖品值" width="120" align="center">
          <template #default="{ row }">
            <el-input-number v-model="row.value" :min="0" :precision="2" :step="1" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="中奖概率" width="120" align="center">
          <template #default="{ row }">
            <el-input-number v-model="row.probability" :min="0" :max="1" :precision="4" :step="0.0001" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="库存" width="100" align="center">
          <template #default="{ row }">
            <el-input-number v-model="row.stock" :min="0" :step="1" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.status" :active-value="1" :inactive-value="0" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right" align="center">
          <template #default="{ row, $index }">
            <el-button type="danger" link size="small" @click="removePrize($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px">
        <el-button type="success" :icon="Check" :loading="prizeSaving" @click="savePrizes">保存奖品</el-button>
      </div>
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
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type === 'wheel' ? '转盘' : row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="消耗" width="120" align="center">
          <template #default="{ row }">
            {{ row.cost_type === 'free' ? '免费' : row.cost_value }}
          </template>
        </el-table-column>
        <el-table-column label="活动时间" width="300" align="center">
          <template #default="{ row }">
            {{ row.start_time || '不限' }} ~ {{ row.end_time || '不限' }}
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
import { Plus, Search, Refresh, RefreshLeft, Check } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'MarketingLottery',
  data() {
    return {
      Plus,
      Search,
      Refresh,
      RefreshLeft,
      Check,
      form: {
        id: 0,
        name: '',
        type: 'wheel',
        cost_type: 'free',
        cost_value: 0,
        daily_limit: 0,
        total_limit: 0,
        start_time: '',
        end_time: '',
        sort: 0,
        status: 1
      },
      rules: {
        name: [{ required: true, message: '请输入活动名称', trigger: 'blur' }]
      },
      submitLoading: false,
      prizeSaving: false,
      toggleLoading: {},
      prizeList: [],
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
        const res = await request.get('/marketing/lottery/list', {
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
        name: '',
        type: 'wheel',
        cost_type: 'free',
        cost_value: 0,
        daily_limit: 0,
        total_limit: 0,
        start_time: '',
        end_time: '',
        sort: 0,
        status: 1
      }
      this.prizeList = []
      this.$refs.formRef?.clearValidate()
    },
    async handleSubmit() {
      const valid = await this.$refs.formRef.validate().catch(() => false)
      if (!valid) return

      this.submitLoading = true
      try {
        const url = this.form.id
          ? '/marketing/lottery/update'
          : '/marketing/lottery/create'
        const method = this.form.id ? 'put' : 'post'
        const res = await request[method](url, this.form)
        if (res.code === 200) {
          ElMessage.success(this.form.id ? '修改成功' : '创建成功')
          if (!this.form.id) {
            this.form.id = res.data.id
          }
          this.fetchList()
          this.fetchPrizes()
        } else {
          ElMessage.error(res.msg || '操作失败')
        }
      } finally {
        this.submitLoading = false
      }
    },
    async handleEdit(row) {
      this.form = { ...row }
      this.$refs.formRef?.clearValidate()
      await this.fetchPrizes()
    },
    async fetchPrizes() {
      if (!this.form.id) return
      const res = await request.get('/marketing/lottery/detail', {
        params: { id: this.form.id }
      })
      if (res.code === 200) {
        this.prizeList = res.data.prizes || []
      }
    },
    addPrize() {
      this.prizeList.push({
        id: 0,
        activity_id: this.form.id,
        name: '',
        type: 'thank',
        value: 0,
        probability: 0,
        sort: this.prizeList.length,
        image: '',
        stock: 0,
        status: 1
      })
    },
    removePrize(index) {
      this.prizeList.splice(index, 1)
    },
    async savePrizes() {
      this.prizeSaving = true
      try {
        const res = await request.post('/marketing/lottery/save_prizes', {
          activity_id: this.form.id,
          prizes: this.prizeList
        })
        if (res.code === 200) {
          ElMessage.success('保存成功')
          this.fetchPrizes()
        } else {
          ElMessage.error(res.msg || '保存失败')
        }
      } finally {
        this.prizeSaving = false
      }
    },
    async handleToggle(row) {
      this.$set(this.toggleLoading, row.id, true)
      try {
        const res = await request.put('/marketing/lottery/toggle', { id: row.id })
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
        const res = await request.delete('/marketing/lottery/delete', { data: { id: row.id } })
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
  display: flex;
  justify-content: space-between;
  align-items: center;
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
