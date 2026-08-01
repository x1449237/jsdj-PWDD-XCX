<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">监护绑定列表</span>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="孩子用户ID">
          <el-input
            v-model="filterForm.childUserId"
            placeholder="请输入用户ID"
            clearable
            style="width: 150px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="家长手机号">
          <el-input
            v-model="filterForm.parentPhone"
            placeholder="请输入手机号"
            clearable
            style="width: 160px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="filterForm.status"
            placeholder="全部状态"
            clearable
            style="width: 140px"
          >
            <el-option label="待确认" :value="0" />
            <el-option label="已绑定" :value="1" />
            <el-option label="已解绑" :value="2" />
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
        <el-table-column prop="id" label="绑定ID" width="90" />
        <el-table-column prop="child_user_id" label="孩子用户ID" width="110" />
        <el-table-column label="孩子昵称" width="140">
          <template #default="{ row }">
            {{ row.child?.nickname || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="parent_phone" label="家长手机号" width="140" />
        <el-table-column prop="parent_openid" label="家长OpenID" min-width="160" show-overflow-tooltip />
        <el-table-column label="月消费限额" width="130">
          <template #default="{ row }">
            ¥{{ ((row.monthly_limit || 0) / 100).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="账号状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_frozen ? 'danger' : 'success'" size="small">
              {{ row.is_frozen ? '已冻结' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="绑定状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTypeMap[row.status]" size="small">
              {{ statusTextMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="bind_time" label="绑定时间" width="180" />
        <el-table-column prop="expire_time" label="过期时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              type="danger"
              size="small"
              link
              @click="handleForceUnbind(row)"
            >
              强制解绑
            </el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
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
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'GuardianList',
  data() {
    return {
      Search,
      Refresh,
      loading: false,
      filterForm: {
        childUserId: '',
        parentPhone: '',
        status: -1
      },
      tableData: [],
      pagination: {
        page: 1,
        limit: 20,
        total: 0
      },
      statusTextMap: {
        0: '待确认',
        1: '已绑定',
        2: '已解绑'
      },
      statusTypeMap: {
        0: 'warning',
        1: 'success',
        2: 'info'
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
        if (this.filterForm.childUserId) {
          params.child_user_id = this.filterForm.childUserId
        }
        if (this.filterForm.parentPhone) {
          params.parent_phone = this.filterForm.parentPhone
        }
        if (this.filterForm.status >= 0) {
          params.status = this.filterForm.status
        }

        const res = await request.get('/minor/guardian_list', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取监护绑定列表失败:', err)
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
        childUserId: '',
        parentPhone: '',
        status: -1
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
    },
    async handleForceUnbind(row) {
      try {
        await ElMessageBox.confirm(
          `确定要强制解绑该监护关系吗？ID: ${row.id}`,
          '强制解绑确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post('/minor/force_unbind', { bind_id: row.id })
        ElMessage.success('解绑成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('强制解绑失败:', err)
        }
      }
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
