<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">订阅消息模板</span>
    </div>

    <el-card class="table-card">
      <div class="table-toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="handleCreate">新增模板</el-button>
          <el-select
            v-model="filterScene"
            placeholder="场景筛选"
            clearable
            style="width: 200px; margin-left: 12px"
            @change="fetchList"
          >
            <el-option
              v-for="item in sceneOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </div>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="template_id" label="模板ID" min-width="160" show-overflow-tooltip />
        <el-table-column prop="template_name" label="模板名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="scene" label="场景" width="120" align="center" />
        <el-table-column prop="scene_name" label="场景名称" width="140" show-overflow-tooltip />
        <el-table-column label="是否启用" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="row.is_enabled"
              :loading="row._toggleLoading"
              @change="(val) => handleToggleStatus(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="hit_count" label="命中次数" width="100" align="center" />
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="success" link size="small" @click="handleShowLog(row)">发送日志</el-button>
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

    <!-- 创建/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑模板' : '新增模板'"
      width="540px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="110px">
        <el-form-item label="模板ID" prop="template_id">
          <el-input v-model="form.template_id" placeholder="请输入模板ID" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="模板名称" prop="template_name">
          <el-input v-model="form.template_name" placeholder="请输入模板名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="场景" prop="scene">
          <el-select v-model="form.scene" placeholder="请选择场景" style="width: 100%">
            <el-option
              v-for="item in sceneOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="场景名称" prop="scene_name">
          <el-input v-model="form.scene_name" placeholder="请输入场景名称" maxlength="30" show-word-limit />
        </el-form-item>
        <el-form-item label="参数字段映射" prop="fields">
          <el-input
            v-model="form.fields"
            type="textarea"
            :rows="6"
            placeholder='请输入JSON格式的字段映射，例如：{"thing1": {"value": "订单号"}}'
          />
        </el-form-item>
        <el-form-item label="是否启用" prop="is_enabled">
          <el-switch v-model="form.is_enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 发送日志弹窗 -->
    <el-dialog
      v-model="logDialogVisible"
      title="发送日志"
      width="900px"
      :close-on-click-modal="false"
      @closed="handleLogDialogClosed"
    >
      <div class="log-filter">
        <el-input
          v-model="logFilter.user_id"
          placeholder="用户ID"
          clearable
          style="width: 160px; margin-right: 12px"
          @clear="fetchLogList"
          @keyup.enter="fetchLogList"
        />
        <el-select
          v-model="logFilter.scene"
          placeholder="场景"
          clearable
          style="width: 160px; margin-right: 12px"
          @change="fetchLogList"
        >
          <el-option
            v-for="item in sceneOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select
          v-model="logFilter.is_success"
          placeholder="发送状态"
          clearable
          style="width: 140px; margin-right: 12px"
          @change="fetchLogList"
        >
          <el-option label="成功" :value="true" />
          <el-option label="失败" :value="false" />
        </el-select>
        <el-button type="primary" @click="fetchLogList">查询</el-button>
      </div>

      <el-table :data="logTableData" v-loading="logLoading" stripe border style="width: 100%; margin-top: 16px">
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="user_id" label="用户ID" width="120" show-overflow-tooltip />
        <el-table-column prop="template_id" label="模板ID" min-width="150" show-overflow-tooltip />
        <el-table-column prop="scene" label="场景" width="100" align="center" />
        <el-table-column label="发送状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_success ? 'success' : 'danger'" size="small">
              {{ row.is_success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="error_msg" label="错误信息" min-width="160" show-overflow-tooltip />
        <el-table-column prop="created_at" label="发送时间" width="170" align="center" />
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="logPagination.page"
          v-model:page-size="logPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="logPagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchLogList"
          @current-change="fetchLogList"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

const SCENE_OPTIONS = [
  { label: '订单状态变更', value: 'order_status' },
  { label: '订单完成通知', value: 'order_complete' },
  { label: '退款通知', value: 'refund_notify' },
  { label: '发货通知', value: 'delivery_notify' },
  { label: '支付成功通知', value: 'pay_success' },
  { label: '审核结果通知', value: 'audit_result' },
  { label: '活动提醒', value: 'activity_remind' },
  { label: '系统通知', value: 'system_notify' }
]

export default {
  name: 'PlatformSubscribe',
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
      filterScene: '',
      sceneOptions: SCENE_OPTIONS,
      dialogVisible: false,
      isEdit: false,
      editId: null,
      submitLoading: false,
      form: {
        template_id: '',
        template_name: '',
        scene: '',
        scene_name: '',
        fields: '',
        is_enabled: true
      },
      formRules: {
        template_id: [
          { required: true, message: '请输入模板ID', trigger: 'blur' }
        ],
        template_name: [
          { required: true, message: '请输入模板名称', trigger: 'blur' },
          { min: 1, max: 50, message: '模板名称长度在 1 到 50 个字符', trigger: 'blur' }
        ],
        scene: [
          { required: true, message: '请选择场景', trigger: 'change' }
        ],
        scene_name: [
          { required: true, message: '请输入场景名称', trigger: 'blur' },
          { min: 1, max: 30, message: '场景名称长度在 1 到 30 个字符', trigger: 'blur' }
        ],
        fields: [
          { required: true, message: '请输入参数字段映射', trigger: 'blur' },
          {
            validator: (rule, value, callback) => {
              if (!value) {
                callback(new Error('请输入参数字段映射'))
                return
              }
              try {
                JSON.parse(value)
                callback()
              } catch {
                callback(new Error('请输入合法的JSON格式'))
              }
            },
            trigger: 'blur'
          }
        ]
      },
      logDialogVisible: false,
      logTableData: [],
      logLoading: false,
      logPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      logFilter: {
        user_id: '',
        scene: '',
        is_success: undefined
      },
      currentLogTemplateId: ''
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
          limit: this.pagination.pageSize
        }
        if (this.filterScene) {
          params.scene = this.filterScene
        }
        const res = await request.get('/subscribe/template/list', { params })
        this.tableData = (res.data?.list || []).map((item) => ({
          ...item,
          _toggleLoading: false
        }))
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取订阅模板列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleCreate() {
      this.isEdit = false
      this.editId = null
      this.form = {
        template_id: '',
        template_name: '',
        scene: '',
        scene_name: '',
        fields: '',
        is_enabled: true
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.editId = row.id || row.template_id
      this.form = {
        template_id: row.template_id,
        template_name: row.template_name,
        scene: row.scene,
        scene_name: row.scene_name,
        fields: typeof row.fields === 'string' ? row.fields : JSON.stringify(row.fields, null, 2),
        is_enabled: !!row.is_enabled
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
        const payload = {
          template_id: this.form.template_id,
          template_name: this.form.template_name,
          scene: this.form.scene,
          scene_name: this.form.scene_name,
          fields: JSON.parse(this.form.fields),
          is_enabled: this.form.is_enabled
        }
        if (this.isEdit) {
          await request.put(`/subscribe/template/update/${this.editId}`, payload)
          ElMessage.success('编辑成功')
        } else {
          await request.post('/subscribe/template/create', payload)
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
    async handleToggleStatus(row, val) {
      row._toggleLoading = true
      try {
        await request.put(`/subscribe/template/toggle/${row.id || row.template_id}`, {
          is_enabled: val
        })
        ElMessage.success(val ? '已启用' : '已禁用')
        row.is_enabled = val
      } catch (err) {
        console.error('切换状态失败:', err)
      } finally {
        row._toggleLoading = false
      }
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除模板「${row.template_name}」吗？此操作不可撤销。`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete(`/subscribe/template/delete/${row.id || row.template_id}`)
        ElMessage.success('删除成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    handleShowLog(row) {
      this.currentLogTemplateId = row.template_id
      this.logFilter = {
        user_id: '',
        scene: '',
        is_success: undefined
      }
      this.logPagination.page = 1
      this.logPagination.pageSize = 20
      this.logDialogVisible = true
      this.fetchLogList()
    },
    handleLogDialogClosed() {
      this.logTableData = []
      this.currentLogTemplateId = ''
    },
    async fetchLogList() {
      this.logLoading = true
      try {
        const params = {
          page: this.logPagination.page,
          limit: this.logPagination.pageSize
        }
        if (this.logFilter.user_id) {
          params.user_id = this.logFilter.user_id
        }
        if (this.logFilter.scene) {
          params.scene = this.logFilter.scene
        }
        if (this.logFilter.is_success !== undefined && this.logFilter.is_success !== '') {
          params.is_success = this.logFilter.is_success
        }
        if (this.currentLogTemplateId) {
          params.template_id = this.currentLogTemplateId
        }
        const res = await request.get('/subscribe/log/list', { params })
        this.logTableData = res.data?.list || []
        this.logPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取发送日志失败:', err)
      } finally {
        this.logLoading = false
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.table-card {
  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .toolbar-left {
      display: flex;
      align-items: center;
    }
  }
}

.log-filter {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
</style>