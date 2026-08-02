<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">违规处罚模板</span>
    </div>

    <el-card class="table-card">
      <div style="margin-bottom: 16px;">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增模板</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="name" label="模板名称" min-width="160" show-overflow-tooltip />
        <el-table-column label="处罚类型" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.type)" size="small">
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处罚时长" width="120" align="center">
          <template #default="{ row }">
            {{ row.duration ? row.duration + '分钟' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="是否启用" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status ? 'success' : 'info'" size="small">
              {{ row.status ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="处罚类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择处罚类型" style="width: 100%">
            <el-option label="警告" :value="1" />
            <el-option label="限制接单" :value="2" />
            <el-option label="清退俱乐部" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="处罚时长" prop="duration">
          <el-input-number v-model="form.duration" :min="0" controls-position="right" />
          <span style="margin-left: 8px; color: #909399;">分钟</span>
        </el-form-item>
        <el-form-item label="是否启用" prop="status">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="原因" prop="reason">
          <el-input v-model="form.reason" type="textarea" :rows="3" placeholder="请输入处罚原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { getPunishmentTemplates, savePunishmentTemplate } from '@/api/extension'
import { Plus, Edit } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'PunishmentTemplate',
  data() {
    return {
      Plus,
      Edit,
      tableData: [],
      loading: false,
      dialogVisible: false,
      dialogTitle: '新增模板',
      submitting: false,
      form: {
        id: undefined,
        name: '',
        type: 1,
        duration: 0,
        status: 1,
        reason: ''
      },
      rules: {
        name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
        type: [{ required: true, message: '请选择处罚类型', trigger: 'change' }]
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
        const res = await getPunishmentTemplates()
        this.tableData = res.data || []
      } catch (err) {
        console.error('获取处罚模板失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleAdd() {
      this.dialogTitle = '新增模板'
      this.form = {
        id: undefined,
        name: '',
        type: 1,
        duration: 0,
        status: 1,
        reason: ''
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogTitle = '编辑模板'
      this.form = {
        id: row.id,
        name: row.name,
        type: row.type,
        duration: row.duration,
        status: row.status,
        reason: row.reason || ''
      }
      this.dialogVisible = true
    },
    async handleSubmit() {
      try {
        await this.$refs.formRef.validate()
      } catch (e) {
        return
      }
      this.submitting = true
      try {
        await savePunishmentTemplate({ ...this.form })
        ElMessage.success(this.form.id ? '更新成功' : '新增成功')
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('保存失败:', err)
      } finally {
        this.submitting = false
      }
    },
    resetForm() {
      this.$refs.formRef && this.$refs.formRef.resetFields()
    },
    typeLabel(type) {
      const map = { 1: '警告', 2: '限制接单', 3: '清退俱乐部' }
      return map[type] || '-'
    },
    typeTag(type) {
      const map = { 1: 'warning', 2: 'info', 3: 'danger' }
      return map[type] || 'info'
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
