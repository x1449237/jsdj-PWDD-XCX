<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">消费预警日志</span>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="用户ID">
          <el-input
            v-model="filterForm.userId"
            placeholder="请输入用户ID"
            clearable
            style="width: 150px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="月份">
          <el-date-picker
            v-model="filterForm.month"
            type="month"
            placeholder="选择月份"
            value-format="YYYY-MM"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="预警等级">
          <el-select
            v-model="filterForm.warningLevel"
            placeholder="全部等级"
            clearable
            style="width: 160px"
          >
            <el-option label="80%预警" :value="1" />
            <el-option label="100%限额" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" v-loading="loading">
      <el-table :data="tableData" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column label="用户昵称" width="140">
          <template #default="{ row }">
            {{ row.user?.nickname || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="month" label="月份" width="110" />
        <el-table-column label="消费金额" width="130">
          <template #default="{ row }">
            ¥{{ (row.consume_amount / 100).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="预警等级" width="120">
          <template #default="{ row }">
            <el-tag :type="row.warning_level === 2 ? 'danger' : 'warning'" size="small">
              {{ row.warning_level === 2 ? '100%限额' : '80%预警' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="guardian_openid" label="监护人OpenID" min-width="160" show-overflow-tooltip />
        <el-table-column prop="sent_at" label="发送时间" width="180" />
        <el-table-column prop="create_time" label="创建时间" width="180" />
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'

export default {
  name: 'WarningLog',
  data() {
    return {
      Search,
      Refresh,
      loading: false,
      filterForm: {
        userId: '',
        month: '',
        warningLevel: ''
      },
      tableData: [],
      pagination: {
        page: 1,
        limit: 20,
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
        const params = {
          page: this.pagination.page,
          limit: this.pagination.limit
        }
        if (this.filterForm.userId) {
          params.user_id = this.filterForm.userId
        }
        if (this.filterForm.month) {
          params.month = this.filterForm.month
        }
        if (this.filterForm.warningLevel) {
          params.warning_level = this.filterForm.warningLevel
        }

        const res = await request.get('/minor/warning_log', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取预警日志失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.filterForm = {
        userId: '',
        month: '',
        warningLevel: ''
      }
      this.pagination.page = 1
      this.fetchList()
    },
    handleSizeChange(size) {
      this.pagination.limit = size
      this.pagination.page = 1
      this.fetchList()
    },
    handlePageChange(page) {
      this.pagination.page = page
      this.fetchList()
    }
  }
}
</script>

<style lang="scss" scoped>
.filter-card {
  margin-bottom: 16px;
}

.filter-form {
  margin: 0;
}

.table-card {
  margin-bottom: 16px;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
