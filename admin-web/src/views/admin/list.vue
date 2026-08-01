<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">管理员管理</span>
      <el-button type="primary" :icon="Plus" @click="handleAdd">新增管理员</el-button>
    </div>

    <!-- 表格 -->
    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ row.roleName || row.role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastLoginAt" label="最后登录" width="170" align="center" />
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="260" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="handleAssignRole(row)">分配角色</el-button>
            <el-button type="info" link size="small" @click="handlePasskey(row)">通行密钥</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑管理员' : '新增管理员'"
      width="520px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="formData.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="密码" :prop="isEdit ? '' : 'password'">
          <el-input
            v-model="formData.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="formData.role" placeholder="请选择角色" style="width: 100%">
            <el-option
              v-for="r in roleList"
              :key="r.value"
              :label="r.label"
              :value="r.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="formData.phone" placeholder="请输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色弹窗 -->
    <el-dialog
      v-model="roleDialogVisible"
      title="分配角色"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form label-width="80px">
        <el-form-item label="角色">
          <el-select v-model="currentRoleValue" placeholder="请选择角色" style="width: 100%">
            <el-option
              v-for="r in roleList"
              :key="r.value"
              :label="r.label"
              :value="r.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitLoading" @click="handleRoleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 通行密钥管理弹窗 -->
    <el-dialog
      v-model="passkeyDialogVisible"
      title="通行密钥管理"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-table :data="passkeyList" stripe border size="small">
        <el-table-column prop="id" label="密钥ID" width="80" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row: pkRow }">
            <el-button type="danger" link size="small" @click="handleDeletePasskey(pkRow)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="passkeyDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'AdminList',
  data() {
    return {
      Plus,
      loading: false,
      tableData: [],
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      isEdit: false,
      currentEditId: null,
      submitLoading: false,
      formData: {
        username: '',
        nickname: '',
        password: '',
        role: '',
        email: '',
        phone: ''
      },
      formRules: {
        username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
        nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur', min: 6 }],
        role: [{ required: true, message: '请选择角色', trigger: 'change' }],
        email: [{ type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }]
      },
      roleList: [],
      roleDialogVisible: false,
      currentRoleUserId: null,
      currentRoleValue: '',
      roleSubmitLoading: false,
      passkeyDialogVisible: false,
      passkeyList: [],
      currentPasskeyUserId: null
    }
  },
  mounted() {
    this.fetchList()
    this.fetchRoles()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize
        }
        const res = await request.get('/admins', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取管理员列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    async fetchRoles() {
      try {
        const res = await request.get('/roles')
        this.roleList = (res.data || []).map(r => ({
          value: r.id || r.value,
          label: r.name || r.label
        }))
      } catch (err) {
        console.error('获取角色列表失败:', err)
      }
    },
    handleAdd() {
      this.isEdit = false
      this.currentEditId = null
      this.formData = {
        username: '',
        nickname: '',
        password: '',
        role: '',
        email: '',
        phone: ''
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.currentEditId = row.id
      this.formData = {
        username: row.username,
        nickname: row.nickname,
        password: '',
        role: row.role || row.roleId,
        email: row.email || '',
        phone: row.phone || ''
      }
      this.dialogVisible = true
    },
    handleDialogClosed() {
      this.$refs.formRef?.resetFields()
    },
    async handleSubmit() {
      const valid = await this.$refs.formRef.validate().catch(() => false)
      if (!valid) return
      this.submitLoading = true
      try {
        const payload = { ...this.formData }
        if (this.isEdit && !payload.password) {
          delete payload.password
        }
        if (this.isEdit) {
          await request.put(`/admins/${this.currentEditId}`, payload)
          ElMessage.success('管理员信息已更新')
        } else {
          await request.post('/admins', payload)
          ElMessage.success('管理员创建成功')
        }
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('保存管理员失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除管理员「${row.username}」吗？此操作不可撤销。`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete(`/admins/${row.id}`)
        ElMessage.success('删除成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除管理员失败:', err)
        }
      }
    },
    handleAssignRole(row) {
      this.currentRoleUserId = row.id
      this.currentRoleValue = row.role || row.roleId || ''
      this.roleDialogVisible = true
    },
    async handleRoleSubmit() {
      this.roleSubmitLoading = true
      try {
        await request.put(`/admins/${this.currentRoleUserId}/role`, {
          role: this.currentRoleValue
        })
        ElMessage.success('角色分配成功')
        this.roleDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('分配角色失败:', err)
      } finally {
        this.roleSubmitLoading = false
      }
    },
    async handlePasskey(row) {
      this.currentPasskeyUserId = row.id
      this.passkeyDialogVisible = true
      try {
        const res = await request.get(`/admins/${row.id}/passkeys`)
        this.passkeyList = res.data || []
      } catch (err) {
        console.error('获取通行密钥失败:', err)
        this.passkeyList = []
      }
    },
    async handleDeletePasskey(pkRow) {
      try {
        await ElMessageBox.confirm('确定要删除该通行密钥吗？', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        await request.delete(`/admins/${this.currentPasskeyUserId}/passkeys/${pkRow.id}`)
        ElMessage.success('通行密钥已删除')
        this.passkeyList = this.passkeyList.filter(p => p.id !== pkRow.id)
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除通行密钥失败:', err)
        }
      }
    }
  }
}
</script>

<style lang="scss" scoped>
@media screen and (max-width: 768px) {
  :deep(.el-table) {
    font-size: 12px;
  }
}
</style>