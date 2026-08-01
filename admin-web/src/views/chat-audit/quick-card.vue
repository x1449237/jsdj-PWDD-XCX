<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">快捷服务卡片管理</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="卡片类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 140px">
            <el-option label="报价" value="price" />
            <el-option label="套餐" value="package" />
            <el-option label="预约" value="appointment" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
          <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          <el-button type="success" :icon="Plus" @click="handleAdd">新增卡片</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="cardList" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="type" label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTag(row.type)">{{ getTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="140" show-overflow-tooltip />
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" width="120" show-overflow-tooltip />
        <el-table-column prop="icon" label="图标" width="100" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.icon" :size="20"><component :is="row.icon" /></el-icon>
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
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

    <!-- 编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '新增快捷卡片' : '编辑快捷卡片'"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="卡片类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="报价" value="price" />
            <el-option label="套餐" value="package" />
            <el-option label="预约" value="appointment" />
          </el-select>
        </el-form-item>
        <el-form-item label="卡片标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入卡片标题" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="卡片内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="3"
            placeholder="请输入卡片内容"
            maxlength="255"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="点击动作">
          <el-input v-model="form.action" placeholder="请输入点击动作，如：navigate、copy" maxlength="32" />
        </el-form-item>
        <el-form-item label="动作参数">
          <el-input
            v-model="form.params_json"
            type="textarea"
            :rows="3"
            placeholder="请输入JSON格式的参数，如：{&quot;url&quot;: &quot;/pages/...&quot;}"
          />
        </el-form-item>
        <el-form-item label="卡片图标">
          <el-input v-model="form.icon" placeholder="请输入图标名称，如：PriceTag" maxlength="32" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
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
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'QuickCard',
  data() {
    return {
      Search,
      Refresh,
      Plus,
      searchForm: {
        type: '',
        status: ''
      },
      cardList: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      dialogMode: 'add',
      formRef: null,
      form: {
        id: 0,
        type: 'price',
        title: '',
        content: '',
        action: '',
        params_json: '',
        icon: '',
        sort: 0,
        status: 1
      },
      formRules: {
        type: [{ required: true, message: '请选择卡片类型', trigger: 'change' }],
        title: [{ required: true, message: '请输入卡片标题', trigger: 'blur' }],
        content: [{ required: true, message: '请输入卡片内容', trigger: 'blur' }]
      },
      submitting: false
    }
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    getTypeTag(type) {
      const map = {
        price: 'primary',
        package: 'success',
        appointment: 'warning'
      }
      return map[type] || 'info'
    },
    getTypeName(type) {
      const map = {
        price: '报价',
        package: '套餐',
        appointment: '预约'
      }
      return map[type] || type
    },
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          type: this.searchForm.type || undefined,
          status: this.searchForm.status !== '' ? this.searchForm.status : undefined
        }
        const res = await request.get('/chat/quick_cards', { params })
        this.cardList = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取卡片列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    resetSearch() {
      this.searchForm = {
        type: '',
        status: ''
      }
      this.pagination.page = 1
      this.fetchList()
    },
    handleAdd() {
      this.dialogMode = 'add'
      this.form = {
        id: 0,
        type: 'price',
        title: '',
        content: '',
        action: '',
        params_json: '',
        icon: '',
        sort: 0,
        status: 1
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogMode = 'edit'
      this.form = {
        ...row,
        params_json: typeof row.params_json === 'object' ? JSON.stringify(row.params_json) : row.params_json
      }
      this.dialogVisible = true
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除卡片「${row.title}」吗？`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete('/chat/quick_card', { data: { id: row.id } })
        ElMessage.success('删除成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    async handleSubmit() {
      if (this.formRef) {
        await this.formRef.validate(async (valid) => {
          if (!valid) return
        })
      }
      try {
        this.submitting = true
        const submitData = { ...this.form }
        if (submitData.params_json && typeof submitData.params_json === 'string') {
          try {
            JSON.parse(submitData.params_json)
          } catch (e) {
            ElMessage.warning('动作参数格式不正确，请输入有效JSON')
            return
          }
        }
        if (this.dialogMode === 'add') {
          await request.post('/chat/quick_card', submitData)
          ElMessage.success('创建成功')
        } else {
          await request.put('/chat/quick_card', submitData)
          ElMessage.success('更新成功')
        }
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('提交失败:', err)
      } finally {
        this.submitting = false
      }
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
  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>
