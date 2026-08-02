<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">节日模板公告</span>
    </div>

    <el-card class="table-card">
      <div style="margin-bottom: 16px;">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增模板</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="name" label="节日名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="content" label="模板内容" min-width="260" show-overflow-tooltip />
        <el-table-column prop="scene" label="适用场景" min-width="150" show-overflow-tooltip />
        <el-table-column prop="updatedAt" label="更新时间" width="170" align="center" />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="节日名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入节日名称" />
        </el-form-item>
        <el-form-item label="模板内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="5" placeholder="请输入模板内容" />
        </el-form-item>
        <el-form-item label="适用场景" prop="scene">
          <el-input v-model="form.scene" placeholder="请输入适用场景" />
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
import { getFestivalTemplates, saveFestivalTemplate } from '@/api/extension'
import { Plus, Edit } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'FestivalTemplate',
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
        content: '',
        scene: ''
      },
      rules: {
        name: [{ required: true, message: '请输入节日名称', trigger: 'blur' }],
        content: [{ required: true, message: '请输入模板内容', trigger: 'blur' }]
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
        const res = await getFestivalTemplates()
        this.tableData = res.data || []
      } catch (err) {
        console.error('获取节日模板失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleAdd() {
      this.dialogTitle = '新增模板'
      this.form = {
        id: undefined,
        name: '',
        content: '',
        scene: ''
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogTitle = '编辑模板'
      this.form = {
        id: row.id,
        name: row.name,
        content: row.content,
        scene: row.scene || ''
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
        await saveFestivalTemplate({ ...this.form })
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
