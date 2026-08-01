<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">备份恢复</span>
    </div>

    <el-card class="table-card">
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" :loading="createLoading" @click="handleCreateBackup">
          创建备份
        </el-button>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="fileName" label="文件名" min-width="220" show-overflow-tooltip />
        <el-table-column label="大小" width="120" align="center">
          <template #default="{ row }">
            {{ formatFileSize(row.fileSize) }}
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'full' ? 'primary' : 'warning'" size="small">
              {{ row.type === 'full' ? '全量' : '增量' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'completed' ? 'success' : row.status === 'in_progress' ? 'warning' : 'danger'"
              size="small"
            >
              {{ row.status === 'completed' ? '已完成' : row.status === 'in_progress' ? '进行中' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" align="center" />
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              type="warning"
              link
              size="small"
              :disabled="row.status !== 'completed'"
              @click="handleRestore(row)"
            >
              恢复
            </el-button>
            <el-button
              type="primary"
              link
              size="small"
              :disabled="row.status !== 'completed'"
              @click="handleDownload(row)"
            >
              下载
            </el-button>
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
import { Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'Backup',
  data() {
    return {
      Plus,
      tableData: [],
      loading: false,
      createLoading: false,
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
          pageSize: this.pagination.pageSize
        }
        const res = await request.get('/backups', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取备份列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    async handleCreateBackup() {
      try {
        await ElMessageBox.confirm(
          '确定要创建新的备份吗？备份过程可能需要几分钟。',
          '创建备份',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
        )
        this.createLoading = true
        await request.post('/backups')
        ElMessage.success('备份任务已创建，请稍后刷新查看')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('创建备份失败:', err)
        }
      } finally {
        this.createLoading = false
      }
    },
    async handleRestore(row) {
      try {
        await ElMessageBox.confirm(
          `警告：将恢复到「${row.createdAt}」的备份数据，当前数据将丢失！确定要继续吗？`,
          '恢复数据确认',
          {
            confirmButtonText: '确定恢复',
            cancelButtonText: '取消',
            type: 'error',
            confirmButtonClass: 'el-button--danger'
          }
        )
        await request.post(`/backups/${row.id}/restore`)
        ElMessage.success('数据恢复成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('恢复失败:', err)
        }
      }
    },
    async handleDownload(row) {
      try {
        const res = await request.get(`/backups/${row.id}/download`, {
          responseType: 'blob'
        })
        const url = window.URL.createObjectURL(new Blob([res]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', row.fileName)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        ElMessage.success('开始下载')
      } catch (err) {
        console.error('下载失败:', err)
      }
    },
    formatFileSize(bytes) {
      if (!bytes) return '-'
      const units = ['B', 'KB', 'MB', 'GB', 'TB']
      let unitIndex = 0
      let size = bytes
      while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024
        unitIndex++
      }
      return size.toFixed(2) + ' ' + units[unitIndex]
    }
  }
}
</script>

<style lang="scss" scoped>
.table-card {
  .table-toolbar {
    margin-bottom: 16px;
  }

  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>