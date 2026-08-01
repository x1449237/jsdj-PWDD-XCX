<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">售后风控关键词管理</span>
    </div>

    <!-- 开关栏 -->
    <el-card class="switch-card">
      <div class="switch-row">
        <div class="switch-item">
          <span class="switch-label">售后关键词自动介入：</span>
          <el-switch
            v-model="globalSwitch"
            active-text="开启"
            inactive-text="关闭"
            @change="handleGlobalSwitchChange"
          />
        </div>
        <div class="switch-item">
          <span class="switch-label">测试模式：</span>
          <el-switch
            v-model="testMode"
            active-text="开启"
            inactive-text="关闭"
            @change="handleTestModeChange"
          />
        </div>
      </div>
    </el-card>

    <el-card class="table-card">
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加关键词</el-button>
        <el-button :icon="Upload" @click="handleBatchImport">批量导入</el-button>
        <el-button :icon="Search" @click="handleTestMatch">测试匹配</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="keyword" label="关键词" min-width="150" show-overflow-tooltip />
        <el-table-column label="分类" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="categoryTag(row.category)" size="small">
              {{ categoryLabel(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.matchType === 'exact' ? '' : row.matchType === 'fuzzy' ? 'warning' : 'danger'" size="small">
              {{ matchTypeLabel(row.matchType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="是否启用" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
              {{ row.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="hitCount" label="命中次数" width="100" align="center" />
        <el-table-column label="操作" width="160" fixed="right" align="center">
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

    <!-- 添加/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑关键词' : '添加关键词'"
      width="520px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="关键词" prop="keyword">
          <el-input v-model="form.keyword" placeholder="请输入关键词" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="form.category" placeholder="请选择分类" style="width: 100%">
            <el-option label="欺诈" value="fraud" />
            <el-option label="滥用" value="abuse" />
            <el-option label="退款" value="refund" />
            <el-option label="威胁" value="threat" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配类型" prop="matchType">
          <el-select v-model="form.matchType" placeholder="请选择匹配类型" style="width: 100%">
            <el-option label="精确匹配" value="exact" />
            <el-option label="模糊匹配" value="fuzzy" />
            <el-option label="正则匹配" value="regex" />
          </el-select>
        </el-form-item>
        <el-form-item label="正则表达式" prop="regexPattern" v-if="form.matchType === 'regex'">
          <el-input v-model="form.regexPattern" placeholder="请输入正则表达式，如 /pattern/flags" />
        </el-form-item>
        <el-form-item label="是否启用" prop="enabled">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="批量导入关键词"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form :model="importForm" label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="importForm.category" placeholder="请选择分类" style="width: 100%">
            <el-option label="欺诈" value="fraud" />
            <el-option label="滥用" value="abuse" />
            <el-option label="退款" value="refund" />
            <el-option label="威胁" value="threat" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配类型">
          <el-select v-model="importForm.matchType" placeholder="请选择匹配类型" style="width: 100%">
            <el-option label="精确匹配" value="exact" />
            <el-option label="模糊匹配" value="fuzzy" />
            <el-option label="正则匹配" value="regex" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".txt"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">每行一个关键词，支持 .txt 文本文件</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="importLoading" :disabled="!importFile" @click="handleImportSubmit">
          导入
        </el-button>
      </template>
    </el-dialog>

    <!-- 测试匹配弹窗 -->
    <el-dialog
      v-model="testDialogVisible"
      title="测试匹配"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="testForm" label-width="80px">
        <el-form-item label="测试文本">
          <el-input
            v-model="testForm.text"
            type="textarea"
            :rows="4"
            placeholder="请输入要测试的文本内容"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="testLoading" @click="handleTestSubmit">开始匹配</el-button>
        </el-form-item>
      </el-form>
      <div v-if="testResult.length > 0" class="test-result">
        <el-tag type="warning" style="margin-bottom: 8px;">命中 {{ testResult.length }} 个关键词：</el-tag>
        <el-table :data="testResult" stripe border size="small" style="width: 100%">
          <el-table-column prop="keyword" label="关键词" />
          <el-table-column label="分类" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="categoryTag(row.category)" size="small">{{ categoryLabel(row.category) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="匹配类型" width="100" align="center">
            <template #default="{ row }">
              <el-tag size="small">{{ matchTypeLabel(row.matchType) }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <el-empty v-else-if="testSubmitted" description="未命中任何关键词" />
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus, Upload, Search } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'AfterSaleKeywords',
  data() {
    return {
      Plus,
      Upload,
      Search,
      globalSwitch: true,
      testMode: false,
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
        keyword: '',
        category: 'fraud',
        matchType: 'exact',
        regexPattern: '',
        enabled: true
      },
      formRules: {
        keyword: [
          { required: true, message: '请输入关键词', trigger: 'blur' }
        ],
        category: [
          { required: true, message: '请选择分类', trigger: 'change' }
        ],
        matchType: [
          { required: true, message: '请选择匹配类型', trigger: 'change' }
        ],
        regexPattern: [
          { required: true, message: '请输入正则表达式', trigger: 'blur' }
        ]
      },
      importDialogVisible: false,
      importForm: {
        category: 'fraud',
        matchType: 'exact'
      },
      importFile: null,
      importLoading: false,
      testDialogVisible: false,
      testForm: {
        text: ''
      },
      testResult: [],
      testLoading: false,
      testSubmitted: false
    }
  },
  mounted() {
    this.fetchList()
    this.fetchSwitchStatus()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize
        }
        const res = await request.get('/after-sale/keywords', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取关键词列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    async fetchSwitchStatus() {
      try {
        const res = await request.get('/after-sale/keywords/switch-status')
        this.globalSwitch = res.data?.globalSwitch ?? true
        this.testMode = res.data?.testMode ?? false
      } catch (err) {
        console.error('获取开关状态失败:', err)
      }
    },
    async handleGlobalSwitchChange(val) {
      try {
        await request.post('/after-sale/keywords/global-switch', { enabled: val })
        ElMessage.success(val ? '已开启自动介入' : '已关闭自动介入')
      } catch (err) {
        this.globalSwitch = !val
        console.error('切换开关失败:', err)
      }
    },
    async handleTestModeChange(val) {
      try {
        await request.post('/after-sale/keywords/test-mode', { enabled: val })
        ElMessage.success(val ? '已开启测试模式' : '已关闭测试模式')
      } catch (err) {
        this.testMode = !val
        console.error('切换测试模式失败:', err)
      }
    },
    handleAdd() {
      this.isEdit = false
      this.editId = null
      this.form = {
        keyword: '',
        category: 'fraud',
        matchType: 'exact',
        regexPattern: '',
        enabled: true
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.editId = row.id
      this.form = {
        keyword: row.keyword,
        category: row.category,
        matchType: row.matchType,
        regexPattern: row.regexPattern || '',
        enabled: row.enabled
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
          keyword: this.form.keyword,
          category: this.form.category,
          matchType: this.form.matchType,
          enabled: this.form.enabled
        }
        if (this.form.matchType === 'regex') {
          payload.regexPattern = this.form.regexPattern
        }
        if (this.isEdit) {
          await request.put(`/after-sale/keywords/${this.editId}`, payload)
          ElMessage.success('编辑成功')
        } else {
          await request.post('/after-sale/keywords', payload)
          ElMessage.success('添加成功')
        }
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('提交失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除关键词「${row.keyword}」吗？`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete(`/after-sale/keywords/${row.id}`)
        ElMessage.success('删除成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    handleBatchImport() {
      this.importForm = {
        category: 'fraud',
        matchType: 'exact'
      }
      this.importFile = null
      this.importDialogVisible = true
    },
    handleFileChange(file) {
      this.importFile = file.raw
    },
    handleFileRemove() {
      this.importFile = null
    },
    async handleImportSubmit() {
      if (!this.importFile) {
        ElMessage.warning('请选择关键词文件')
        return
      }
      this.importLoading = true
      try {
        const formData = new FormData()
        formData.append('file', this.importFile)
        formData.append('category', this.importForm.category)
        formData.append('matchType', this.importForm.matchType)
        const res = await request.post('/after-sale/keywords/import', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        ElMessage.success(`导入成功，共导入 ${res.data?.count || 0} 个关键词`)
        this.importDialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('导入失败:', err)
      } finally {
        this.importLoading = false
      }
    },
    handleTestMatch() {
      this.testForm = { text: '' }
      this.testResult = []
      this.testSubmitted = false
      this.testDialogVisible = true
    },
    async handleTestSubmit() {
      if (!this.testForm.text.trim()) {
        ElMessage.warning('请输入测试文本')
        return
      }
      this.testLoading = true
      this.testSubmitted = true
      try {
        const res = await request.post('/after-sale/keywords/test-match', {
          text: this.testForm.text
        })
        this.testResult = res.data?.list || []
      } catch (err) {
        console.error('测试匹配失败:', err)
      } finally {
        this.testLoading = false
      }
    },
    categoryTag(category) {
      const map = { fraud: 'danger', abuse: 'warning', refund: 'primary', threat: 'danger' }
      return map[category] || 'info'
    },
    categoryLabel(category) {
      const map = { fraud: '欺诈', abuse: '滥用', refund: '退款', threat: '威胁' }
      return map[category] || category
    },
    matchTypeLabel(type) {
      const map = { exact: '精确', fuzzy: '模糊', regex: '正则' }
      return map[type] || type
    }
  }
}
</script>

<style lang="scss" scoped>
.switch-card {
  margin-bottom: 16px;
}

.switch-row {
  display: flex;
  align-items: center;
  gap: 32px;
}

.switch-item {
  display: flex;
  align-items: center;
}

.switch-label {
  margin-right: 8px;
  font-size: 14px;
  color: #606266;
}

.table-card {
  .table-toolbar {
    margin-bottom: 16px;
    display: flex;
    gap: 8px;
  }
}

.test-result {
  margin-top: 16px;
}
</style>