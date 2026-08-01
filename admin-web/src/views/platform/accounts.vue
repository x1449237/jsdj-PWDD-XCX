<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">平台官方账号管理</span>
    </div>

    <el-card class="table-card">
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" @click="handleCreate">创建账号</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="账号编号" width="100" align="center" />
        <el-table-column label="头像" width="80" align="center">
          <template #default="{ row }">
            <el-avatar :size="40" :src="row.avatar" />
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="150" show-overflow-tooltip />
        <el-table-column label="V标类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="vTypeTag(row.vType)" size="small">
              {{ vTypeLabel(row.vType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button
              :type="row.status === 'active' ? 'warning' : 'success'"
              link
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'active' ? '停用' : '启用' }}
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

    <!-- 创建/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑账号' : '创建账号'"
      width="480px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" maxlength="30" show-word-limit />
        </el-form-item>
        <el-form-item label="头像" prop="avatar">
          <el-input v-model="form.avatar" placeholder="请输入头像URL" />
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
  name: 'PlatformAccounts',
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
      isEdit: false,
      editId: null,
      submitLoading: false,
      form: {
        nickname: '',
        avatar: ''
      },
      formRules: {
        nickname: [
          { required: true, message: '请输入昵称', trigger: 'blur' },
          { min: 1, max: 30, message: '昵称长度在 1 到 30 个字符', trigger: 'blur' }
        ],
        avatar: [
          { required: true, message: '请输入头像URL', trigger: 'blur' }
        ]
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
        const res = await request.get('/platform/accounts', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取平台账号列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleCreate() {
      this.isEdit = false
      this.editId = null
      this.form = { nickname: '', avatar: '' }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.editId = row.id
      this.form = {
        nickname: row.nickname,
        avatar: row.avatar
      }
      this.dialogVisible = true
    },
    handleDialogClosed() {
      this.$refs.formRef?.resetFields()
    },
    async handleSubmit() {
      try {
        await this.$refs.formRef.validate()
      } catch {
        return
      }
      this.submitLoading = true
      try {
        if (this.isEdit) {
          await request.put(`/platform/accounts/${this.editId}`, this.form)
          ElMessage.success('编辑成功')
        } else {
          await request.post('/platform/accounts', this.form)
          ElMessage.success('创建成功')
        }
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('提交失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    async handleToggleStatus(row) {
      const isActive = row.status === 'active'
      const action = isActive ? '停用' : '启用'
      try {
        await ElMessageBox.confirm(
          `确定要${action}账号「${row.nickname}」吗？`,
          `${action}确认`,
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        const url = isActive
          ? `/platform/accounts/${row.id}/disable`
          : `/platform/accounts/${row.id}/enable`
        await request.post(url)
        ElMessage.success(`${action}成功`)
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error(`${action}失败:`, err)
        }
      }
    },
    vTypeTag(type) {
      const map = { blue: '', gold: 'warning', personal: 'info', enterprise: 'success' }
      return map[type] || 'info'
    },
    vTypeLabel(type) {
      const map = { blue: '蓝V', gold: '金V', personal: '个人认证', enterprise: '企业认证' }
      return map[type] || type
    }
  }
}
</script>

<style lang="scss" scoped>
.table-card {
  .table-toolbar {
    margin-bottom: 16px;
  }
}
</style>