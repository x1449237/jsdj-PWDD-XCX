<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">邀请码管理</span>
    </div>

    <!-- 生成邀请码 -->
    <el-card class="section-card">
      <template #header>
        <span class="card-header">生成邀请码</span>
      </template>
      <el-form ref="generateFormRef" :model="generateForm" :rules="generateRules" label-width="120px">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="角色类型" prop="roleType">
              <el-select v-model="generateForm.roleType" placeholder="请选择角色类型" style="width: 100%">
                <el-option label="打手" value="player" />
                <el-option label="分销商" value="distributor" />
                <el-option label="派单员" value="dispatcher" />
                <el-option label="内置管理员" value="admin" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="最大使用次数" prop="maxUses">
              <el-input-number
                v-model="generateForm.maxUses"
                :min="1"
                :max="99999"
                style="width: 100%"
                placeholder="请输入最大使用次数"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="过期时间" prop="expireTime">
              <el-date-picker
                v-model="generateForm.expireTime"
                type="datetime"
                placeholder="请选择过期时间"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="生成数量" prop="count">
              <el-input-number
                v-model="generateForm.count"
                :min="1"
                :max="100"
                style="width: 100%"
                placeholder="请输入生成数量"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-button type="primary" :icon="Plus" :loading="generateLoading" @click="handleGenerate">
            批量生成
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 邀请码列表 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>邀请码列表</span>
        </div>
      </template>

      <div class="search-row">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="批次号">
            <el-input
              v-model="searchForm.batchNo"
              placeholder="请输入批次号"
              clearable
              style="width: 200px"
              @keyup.enter="handleSearch"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
            <el-button :icon="Refresh" @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="code" label="邀请码" min-width="140" show-overflow-tooltip />
        <el-table-column label="角色类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ roleTypeLabel(row.roleType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="maxUses" label="最大使用次数" width="120" align="center" />
        <el-table-column prop="usedCount" label="已使用次数" width="120" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'active' ? 'success' : row.status === 'exhausted' ? 'warning' : 'info'"
              size="small"
            >
              {{ row.status === 'active' ? '有效' : row.status === 'exhausted' ? '已用尽' : '已作废' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expireTime" label="过期时间" width="170" align="center" />
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'active'"
              type="danger"
              link
              size="small"
              @click="handleInvalidate(row)"
            >
              作废
            </el-button>
            <el-button type="primary" link size="small" @click="handleExport(row)">导出</el-button>
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
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'Invite',
  data() {
    return {
      Plus,
      Search,
      Refresh,
      generateForm: {
        roleType: '',
        maxUses: 1,
        expireTime: '',
        count: 1
      },
      generateRules: {
        roleType: [{ required: true, message: '请选择角色类型', trigger: 'change' }],
        maxUses: [{ required: true, message: '请输入最大使用次数', trigger: 'blur' }],
        expireTime: [{ required: true, message: '请选择过期时间', trigger: 'change' }],
        count: [{ required: true, message: '请输入生成数量', trigger: 'blur' }]
      },
      generateLoading: false,
      searchForm: {
        batchNo: ''
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
    roleTypeLabel(type) {
      const map = { player: '打手', distributor: '分销商', dispatcher: '派单员', admin: '内置管理员' }
      return map[type] || type
    },
    async handleGenerate() {
      const valid = await this.$refs.generateFormRef.validate().catch(() => false)
      if (!valid) return
      this.generateLoading = true
      try {
        await request.post('/invite-codes/generate', {
          roleType: this.generateForm.roleType,
          maxUses: this.generateForm.maxUses,
          expireTime: this.generateForm.expireTime,
          count: this.generateForm.count
        })
        ElMessage.success(`成功生成 ${this.generateForm.count} 个邀请码`)
        this.fetchList()
      } catch (err) {
        console.error('生成邀请码失败:', err)
      } finally {
        this.generateLoading = false
      }
    },
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          batchNo: this.searchForm.batchNo || undefined
        }
        const res = await request.get('/invite-codes', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取邀请码列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.searchForm.batchNo = ''
      this.handleSearch()
    },
    async handleInvalidate(row) {
      try {
        await ElMessageBox.confirm(
          `确定要作废邀请码「${row.code}」吗？作废后该邀请码将无法使用。`,
          '作废确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.put(`/invite-codes/${row.id}/invalidate`)
        ElMessage.success('作废成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('作废失败:', err)
        }
      }
    },
    async handleExport(row) {
      try {
        const res = await request.get(`/invite-codes/${row.id}/export`, {
          responseType: 'blob'
        })
        const url = window.URL.createObjectURL(new Blob([res]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `邀请码_${row.code}.txt`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        ElMessage.success('导出成功')
      } catch (err) {
        console.error('导出失败:', err)
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.section-card {
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
}

.search-row {
  margin-bottom: 16px;
}

.table-card {
  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}

@media screen and (max-width: 768px) {
  :deep(.el-col-sm-12) {
    max-width: 100%;
    flex: 0 0 100%;
  }
}
</style>