<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">处罚记录</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="处罚类型">
          <el-select v-model="searchForm.punishmentType" placeholder="全部" clearable style="width: 150px">
            <el-option label="警告" value="warning" />
            <el-option label="禁言" value="mute" />
            <el-option label="封禁" value="ban" />
            <el-option label="冻结资金" value="freeze" />
          </el-select>
        </el-form-item>
        <el-form-item label="处罚对象类型">
          <el-select v-model="searchForm.targetType" placeholder="全部" clearable style="width: 150px">
            <el-option label="用户" value="user" />
            <el-option label="打手" value="player" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格 -->
    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="targetName" label="处罚对象" min-width="150" show-overflow-tooltip />
        <el-table-column label="处罚对象类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.targetType === 'user' ? 'primary' : 'warning'" size="small">
              {{ row.targetType === 'user' ? '用户' : '打手' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处罚类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="punishmentTypeTag(row.punishmentType)" size="small">
              {{ punishmentTypeLabel(row.punishmentType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时长" width="120" align="center">
          <template #default="{ row }">
            {{ row.duration ? row.duration + '分钟' : '永久' }}
          </template>
        </el-table-column>
        <el-table-column prop="operator" label="操作人" width="120" show-overflow-tooltip />
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="时间" width="170" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'danger' : 'info'" size="small">
              {{ row.status === 'active' ? '生效中' : '已撤销' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'active'"
              type="warning"
              link
              size="small"
              @click="handleRevoke(row)"
            >
              撤销处罚
            </el-button>
            <span v-else style="color: #909399;">-</span>
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
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'PunishmentRecords',
  data() {
    return {
      Search,
      Refresh,
      searchForm: {
        punishmentType: '',
        targetType: ''
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
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          punishmentType: this.searchForm.punishmentType || undefined,
          targetType: this.searchForm.targetType || undefined
        }
        const res = await request.get('/punishment/records', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取处罚记录失败:', err)
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
        punishmentType: '',
        targetType: ''
      }
      this.handleSearch()
    },
    async handleRevoke(row) {
      try {
        await ElMessageBox.confirm(
          `确定要撤销对「${row.targetName}」的「${this.punishmentTypeLabel(row.punishmentType)}」处罚吗？`,
          '撤销确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/punishment/records/${row.id}/revoke`)
        ElMessage.success('撤销成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('撤销失败:', err)
        }
      }
    },
    punishmentTypeTag(type) {
      const map = { warning: 'warning', mute: 'info', ban: 'danger', freeze: 'danger' }
      return map[type] || 'info'
    },
    punishmentTypeLabel(type) {
      const map = { warning: '警告', mute: '禁言', ban: '封禁', freeze: '冻结资金' }
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
</style>