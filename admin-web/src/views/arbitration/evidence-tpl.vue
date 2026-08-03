<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">举证模板管理</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="纠纷类型">
          <el-select v-model="searchForm.disputeType" placeholder="全部" clearable style="width: 160px">
            <el-option label="打手迟到" value="player_late" />
            <el-option label="消极服务" value="negative_service" />
            <el-option label="玩家退款纠纷" value="player_refund" />
            <el-option label="需求变更" value="demand_change" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
          <el-button type="success" :icon="Plus" @click="handleAdd">新增模板</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="纠纷类型" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="disputeTypeTag(row.dispute_type)" size="small">
              {{ disputeTypeLabel(row.dispute_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="模板标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="description" label="模板描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑模板' : '新增模板'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="纠纷类型" prop="dispute_type">
          <el-select v-model="form.dispute_type" placeholder="请选择纠纷类型" style="width: 100%;">
            <el-option label="打手迟到" value="player_late" />
            <el-option label="消极服务" value="negative_service" />
            <el-option label="玩家退款纠纷" value="player_refund" />
            <el-option label="需求变更" value="demand_change" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="模板标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入模板标题" />
        </el-form-item>
        <el-form-item label="模板描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入模板描述" />
        </el-form-item>
        <el-form-item label="举证项配置">
          <div v-for="(item, index) in form.required_items" :key="index" class="required-item">
            <el-input v-model="item.key" placeholder="key" style="width: 120px;" />
            <el-input v-model="item.label" placeholder="显示名称" style="width: 140px; margin: 0 8px;" />
            <el-select v-model="item.type" placeholder="类型" style="width: 100px;">
              <el-option label="图片" value="image" />
              <el-option label="视频" value="video" />
              <el-option label="音频" value="audio" />
              <el-option label="文字" value="text" />
            </el-select>
            <el-checkbox v-model="item.required" style="margin: 0 8px;">必填</el-checkbox>
            <el-button type="danger" link @click="removeRequiredItem(index)">删除</el-button>
          </div>
          <el-button type="primary" link @click="addRequiredItem">+ 添加举证项</el-button>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
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
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ArbitrationEvidenceTpl',
  components: { Search, Refresh, Plus },
  data() {
    return {
      loading: false,
      searchForm: {
        disputeType: ''
      },
      // 全量数据(一次加载),tableData 由 computed 本地过滤,减轻后端压力
      allData: [],
      dialogVisible: false,
      isEdit: false,
      submitting: false,
      formRef: null,
      form: {
        id: 0,
        dispute_type: '',
        title: '',
        description: '',
        required_items: [],
        sort: 0,
        status: 1
      },
      rules: {
          dispute_type: [{ required: true, message: '请选择纠纷类型', trigger: 'change' }],
          title: [{ required: true, message: '请输入模板标题', trigger: 'blur' }]
        }
    }
  },
  computed: {
    // 前端本地按纠纷类型过滤,不再每次筛选都请求后端
    tableData() {
      const t = this.searchForm.disputeType
      if (!t) return this.allData
      return this.allData.filter(r => r.dispute_type === t)
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    // 仅全量加载一次(及写操作后刷新),不带筛选参数
    loadData() {
      this.loading = true
      request.get('/arbitration/evidence_tpl_list').then(res => {
        this.allData = res || []
      }).finally(() => {
        this.loading = false
      })
    },
    // 搜索:纯前端过滤,computed 自动响应,不调后端
    handleSearch() {
      // tableData 为 computed,searchForm 变化即自动更新
    },
    handleReset() {
      this.searchForm = { disputeType: '' }
      // computed 自动响应,无需调后端
    },
    handleAdd() {
      this.isEdit = false
      this.form = {
        id: 0,
        dispute_type: '',
        title: '',
        description: '',
        required_items: [],
        sort: 0,
        status: 1
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.form = {
        ...row,
        required_items: row.required_items_json || []
      }
      this.dialogVisible = true
    },
    handleDelete(row) {
      ElMessageBox.confirm('确定删除此模板？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        request.delete('/arbitration/evidence_tpl_delete', {
          data: { id: row.id }
        }).then(() => {
          ElMessage.success('删除成功')
          this.loadData()
        })
      }).catch(() => {})
    },
    addRequiredItem() {
      this.form.required_items.push({
        key: '',
        label: '',
        type: 'image',
        required: true
      })
    },
    removeRequiredItem(index) {
      this.form.required_items.splice(index, 1)
    },
    handleSubmit() {
      this.$refs.formRef.validate(valid => {
        if (!valid) return
        this.submitting = true
        const data = {
          ...this.form,
          required_items_json: JSON.stringify(this.form.required_items)
        }
        const requestFn = this.isEdit
          ? request.put('/arbitration/evidence_tpl_update', data)
          : request.post('/arbitration/evidence_tpl_create', data)
        requestFn.then(() => {
          ElMessage.success(this.isEdit ? '更新成功' : '创建成功')
          this.dialogVisible = false
          this.loadData()
        }).finally(() => {
          this.submitting = false
        })
      })
    },
    disputeTypeTag(type) {
      const map = {
        player_late: 'warning',
        negative_service: 'danger',
        player_refund: 'primary',
        demand_change: 'info',
        other: 'success'
      }
      return map[type] || 'info'
    },
    disputeTypeLabel(type) {
      const map = {
        player_late: '打手迟到',
        negative_service: '消极服务',
        player_refund: '玩家退款纠纷',
        demand_change: '需求变更',
        other: '其他'
      }
      return map[type] || type
    }
  }
}
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.page-header {
  margin-bottom: 16px;
}
.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}
.search-card,
.table-card {
  margin-bottom: 16px;
}
.required-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}
</style>
