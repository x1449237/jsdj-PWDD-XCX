<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">用户管理</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="用户ID/昵称/手机号"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="用户类型">
          <el-select v-model="searchForm.userType" placeholder="全部" clearable style="width: 150px">
            <el-option label="普通用户" value="normal" />
            <el-option label="大额验证" value="large_verified" />
            <el-option label="风险用户" value="risk" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="正常" value="active" />
            <el-option label="已封禁" value="banned" />
          </el-select>
        </el-form-item>
        <el-form-item label="注册时间">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格 -->
    <el-card class="table-card">
      <div class="table-toolbar">
        <el-button type="primary" :icon="Download" :loading="exportLoading" @click="handleExport">
          导出
        </el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="用户ID" width="80" align="center" />
        <el-table-column label="头像" width="70" align="center">
          <template #default="{ row }">
            <el-avatar :size="36" :src="row.avatar" />
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120" show-overflow-tooltip />
        <el-table-column label="手机号" width="130">
          <template #default="{ row }">
            {{ maskPhone(row.phone) }}
          </template>
        </el-table-column>
        <el-table-column label="用户类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="userTypeTag(row.userType)" size="small">
              {{ userTypeLabel(row.userType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="实名状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.realNameStatus === 'verified' ? 'success' : 'info'" size="small">
              {{ row.realNameStatus === 'verified' ? '已认证' : '未认证' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creditScore" label="信用分" width="80" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '已封禁' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="注册时间" width="170" align="center" />
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">详情</el-button>
            <el-button
              :type="row.status === 'active' ? 'danger' : 'success'"
              link
              size="small"
              @click="handleToggleBan(row)"
            >
              {{ row.status === 'active' ? '封禁' : '解封' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
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
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'UserList',
  data() {
    return {
      Search,
      Refresh,
      Download,
      searchForm: {
        keyword: '',
        userType: '',
        status: '',
        dateRange: []
      },
      tableData: [],
      loading: false,
      exportLoading: false,
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
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          keyword: this.searchForm.keyword || undefined,
          userType: this.searchForm.userType || undefined,
          status: this.searchForm.status || undefined,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/users', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取用户列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.searchForm = {
        keyword: '',
        userType: '',
        status: '',
        dateRange: []
      }
      this.handleSearch()
    },
    handleDetail(row) {
      this.$router.push(`/user/detail/${row.id}`)
    },
    async handleToggleBan(row) {
      const isBan = row.status === 'active'
      const action = isBan ? '封禁' : '解封'
      try {
        await ElMessageBox.confirm(
          `确定要${action}用户「${row.nickname}」吗？`,
          `${action}确认`,
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        const url = isBan ? `/users/${row.id}/ban` : `/users/${row.id}/unban`
        await request.post(url)
        ElMessage.success(`${action}成功`)
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error(`${action}失败:`, err)
        }
      }
    },
    async handleExport() {
      try {
        await ElMessageBox.confirm(
          '确定要导出当前筛选条件下的用户数据吗？',
          '导出确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
        )
        this.exportLoading = true
        const params = {
          keyword: this.searchForm.keyword || undefined,
          userType: this.searchForm.userType || undefined,
          status: this.searchForm.status || undefined,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/users/export', { params, responseType: 'blob' })
        const url = window.URL.createObjectURL(new Blob([res]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `用户数据_${new Date().toISOString().slice(0, 10)}.xlsx`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        ElMessage.success('导出成功')
      } catch (err) {
        if (err !== 'cancel') {
          console.error('导出失败:', err)
        }
      } finally {
        this.exportLoading = false
      }
    },
    maskPhone(phone) {
      if (!phone) return '-'
      return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    },
    userTypeTag(type) {
      const map = { normal: '', large_verified: 'warning', risk: 'danger' }
      return map[type] || 'info'
    },
    userTypeLabel(type) {
      const map = { normal: '普通用户', large_verified: '大额验证', risk: '风险用户' }
      return map[type] || type
    }
  }
}
</script>

<style lang="scss" scoped>
.search-card {
  margin-bottom: 16px;
}

.search-form-inline {
  display: flex;
  flex-wrap: wrap;
}

.table-card {
  .table-toolbar {
    margin-bottom: 16px;
  }
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>