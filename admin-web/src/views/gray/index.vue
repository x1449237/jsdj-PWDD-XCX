<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">灰度发布</span>
    </div>

    <el-card class="table-card">
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增灰度发布</el-button>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="version" label="版本号" width="120" align="center" />
        <el-table-column label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'pending' ? 'info' : row.status === 'gray' ? 'warning' : row.status === 'full' ? 'success' : 'danger'"
              size="small"
            >
              {{ row.status === 'pending' ? '待发布' : row.status === 'gray' ? '灰度中' : row.status === 'full' ? '全量' : '已回滚' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="白名单" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.whiteList">{{ row.whiteList }}</span>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="放量比例" width="110" align="center">
          <template #default="{ row }">
            <el-progress
              :percentage="row.ratio"
              :stroke-width="8"
              :color="row.ratio >= 100 ? '#67c23a' : '#409eff'"
            />
          </template>
        </el-table-column>
        <el-table-column label="错误率" width="100" align="center">
          <template #default="{ row }">
            <span :style="{ color: row.errorRate > 5 ? '#f56c6c' : row.errorRate > 1 ? '#e6a23c' : '#67c23a' }">
              {{ row.errorRate }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="发布时间" width="170" align="center" />
        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending'"
              type="success"
              link
              size="small"
              @click="handlePublish(row)"
            >
              发布
            </el-button>
            <el-button
              v-if="row.status === 'gray'"
              type="primary"
              link
              size="small"
              @click="handleFullRelease(row)"
            >
              全量
            </el-button>
            <el-button
              v-if="row.status === 'gray' || row.status === 'full'"
              type="danger"
              link
              size="small"
              @click="handleRollback(row)"
            >
              回滚
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
            >
              删除
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

    <!-- 新增灰度发布弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      title="新增灰度发布"
      width="560px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="版本号" prop="version">
          <el-input v-model="form.version" placeholder="请输入版本号，如 v1.2.0" maxlength="30" />
        </el-form-item>
        <el-form-item label="白名单用户ID" prop="whiteList">
          <el-input
            v-model="form.whiteList"
            type="textarea"
            :rows="3"
            placeholder="请输入用户ID，多个用逗号分隔"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="放量比例" prop="ratio">
          <el-row style="width: 100%;" :gutter="12" align="middle">
            <el-col :span="16">
              <el-slider
                v-model="form.ratio"
                :min="1"
                :max="100"
                :step="1"
                :marks="{ 1: '1%', 25: '25%', 50: '50%', 75: '75%', 100: '100%' }"
              />
            </el-col>
            <el-col :span="8">
              <el-input-number
                v-model="form.ratio"
                :min="1"
                :max="100"
                style="width: 100%"
              >
                <template #suffix><span style="color:#909399;">%</span></template>
              </el-input-number>
            </el-col>
          </el-row>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'GrayRelease',
  data() {
    return {
      Plus,
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      submitLoading: false,
      form: {
        version: '',
        whiteList: '',
        ratio: 10
      },
      formRules: {
        version: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
        whiteList: [{ required: true, message: '请输入白名单用户ID', trigger: 'blur' }],
        ratio: [{ required: true, message: '请设置放量比例', trigger: 'blur' }]
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
        const res = await request.get('/gray-release', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取灰度发布列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleAdd() {
      this.form = {
        version: '',
        whiteList: '',
        ratio: 10
      }
      this.dialogVisible = true
    },
    async handleSubmit() {
      const valid = await this.$refs.formRef.validate().catch(() => false)
      if (!valid) return
      this.submitLoading = true
      try {
        await request.post('/gray-release', {
          version: this.form.version,
          whiteList: this.form.whiteList.split(',').map(s => s.trim()).filter(Boolean),
          ratio: this.form.ratio
        })
        ElMessage.success('创建成功')
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('创建失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    async handlePublish(row) {
      try {
        await ElMessageBox.confirm(
          `确定要发布版本「${row.version}」吗？`,
          '发布确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.put(`/gray-release/${row.id}/publish`)
        ElMessage.success('发布成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('发布失败:', err)
        }
      }
    },
    async handleFullRelease(row) {
      try {
        await ElMessageBox.confirm(
          `确定要将版本「${row.version}」全量发布吗？此操作将覆盖所有用户。`,
          '全量发布确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.put(`/gray-release/${row.id}/full`)
        ElMessage.success('全量发布成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('全量发布失败:', err)
        }
      }
    },
    async handleRollback(row) {
      try {
        await ElMessageBox.confirm(
          `确定要回滚版本「${row.version}」吗？回滚后将恢复到上一个版本。`,
          '回滚确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
        )
        await request.put(`/gray-release/${row.id}/rollback`)
        ElMessage.success('回滚成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('回滚失败:', err)
        }
      }
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除版本「${row.version}」的灰度发布吗？`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete(`/gray-release/${row.id}`)
        ElMessage.success('删除成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    handleDialogClosed() {
      this.$refs.formRef?.resetFields()
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