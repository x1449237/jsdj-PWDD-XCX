<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">平台文档管理</span>
      <span class="page-desc">协议 / 政策 / 合同（仅限PDF格式）</span>
    </div>

    <el-card>
      <!-- 筛选 -->
      <el-form :model="searchForm" inline>
        <el-form-item label="文档类型">
          <el-select v-model="searchForm.docType" placeholder="全部" clearable style="width: 140px">
            <el-option label="协议" value="agreement" />
            <el-option label="政策" value="policy" />
            <el-option label="合同" value="contract" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-form-item>
        <el-form-item>
          <el-button type="success" @click="showUploadDialog">上传文档</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="文档标题" min-width="160" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.doc_type === 'agreement' ? 'primary' : row.doc_type === 'policy' ? 'warning' : 'danger'"
              size="small"
            >
              {{ docTypeMap[row.doc_type] || row.doc_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column label="文件大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="80" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">v{{ row.version }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
              {{ row.is_active ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作人" width="120">
          <template #default="{ row }">
            {{ row.admin?.nickname || row.admin?.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="上传时间" width="180" />
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handlePreview(row)">预览</el-button>
            <el-button type="info" size="small" @click="showVersions(row)">历史版本</el-button>
            <el-button type="warning" size="small" @click="showReplaceDialog(row)">替换</el-button>
            <el-button
              :type="row.is_active ? 'info' : 'success'"
              size="small"
              @click="handleToggle(row)"
            >
              {{ row.is_active ? '禁用' : '启用' }}
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog v-model="uploadVisible" title="上传文档" width="500px" @closed="resetUpload">
      <el-form :model="uploadForm" ref="uploadFormRef" :rules="uploadRules" label-width="80px">
        <el-form-item label="文档类型" prop="docType">
          <el-select v-model="uploadForm.docType" placeholder="请选择文档类型" style="width: 100%">
            <el-option label="协议" value="agreement" />
            <el-option label="政策" value="policy" />
            <el-option label="合同" value="contract" />
          </el-select>
        </el-form-item>
        <el-form-item label="文档标题" prop="title">
          <el-input v-model="uploadForm.title" placeholder="请输入文档标题" maxlength="128" />
        </el-form-item>
        <el-form-item label="PDF文件" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".pdf"
            :on-change="onFileChange"
            :on-remove="onFileRemove"
          >
            <el-button type="primary">选择PDF文件</el-button>
            <template #tip>
              <div class="el-upload__tip">仅支持PDF格式，文件大小不超过20MB</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="confirmUpload">确认上传</el-button>
      </template>
    </el-dialog>

    <!-- 替换弹窗 -->
    <el-dialog v-model="replaceVisible" title="替换文档" width="500px" @closed="resetReplace">
      <el-form :model="replaceForm" ref="replaceFormRef" label-width="80px">
        <el-form-item label="当前文档">
          <span class="current-doc-info">{{ replaceForm.currentTitle }}</span>
          <el-tag size="small" style="margin-left: 8px">v{{ replaceForm.currentVersion }}</el-tag>
        </el-form-item>
        <el-form-item label="新PDF文件" prop="file">
          <el-upload
            ref="replaceUploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".pdf"
            :on-change="onReplaceFileChange"
            :on-remove="onReplaceFileRemove"
          >
            <el-button type="warning">选择新PDF文件</el-button>
            <template #tip>
              <div class="el-upload__tip">替换后将自动递增版本号</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replaceVisible = false">取消</el-button>
        <el-button type="warning" :loading="replaceLoading" @click="confirmReplace">确认替换</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog v-model="previewVisible" title="文档预览" width="800px" top="40px">
      <iframe
        v-if="previewUrl"
        :src="previewUrl"
        style="width: 100%; height: 600px; border: none"
      ></iframe>
    </el-dialog>

    <!-- 历史版本弹窗 -->
    <el-dialog v-model="versionsVisible" :title="'历史版本 - ' + versionsTitle" width="800px">
      <el-table :data="versionsData" v-loading="versionsLoading" stripe>
        <el-table-column label="版本" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_current ? 'success' : 'info'" size="small">
              {{ row.is_current ? 'v' + row.version + '（当前）' : 'v' + row.version }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column label="大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openVersion(row)">打开</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'DocumentManage',
  data() {
    return {
      searchForm: { docType: '' },
      // 全量数据(一次加载),tableData 由 computed 本地过滤,减轻后端压力
      allData: [],
      loading: false,
      docTypeMap: {
        agreement: '协议',
        policy: '政策',
        contract: '合同'
      },

      // 上传
      uploadVisible: false,
      uploadLoading: false,
      uploadForm: { docType: '', title: '', file: null },
      uploadRules: {
        docType: [{ required: true, message: '请选择文档类型', trigger: 'change' }],
        title: [{ required: true, message: '请输入文档标题', trigger: 'blur' }]
      },

      // 替换
      replaceVisible: false,
      replaceLoading: false,
      replaceForm: { id: 0, currentTitle: '', currentVersion: 0, file: null },

      // 预览
      previewVisible: false,
      previewUrl: '',

      // 历史版本
      versionsVisible: false,
      versionsTitle: '',
      versionsData: [],
      versionsLoading: false,
      versionsDocumentId: 0
    }
  },
  mounted() {
    this.fetchList()
  },
  computed: {
    // 前端本地按文档类型过滤,不再每次筛选都请求后端
    tableData() {
      const t = this.searchForm.docType
      if (!t) return this.allData
      return this.allData.filter(r => r.doc_type === t)
    }
  },
  methods: {
    // 仅全量加载一次(及写操作后刷新),不带筛选参数
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/document/list')
        this.allData = res.data || []
      } catch (e) {
        ElMessage.error('加载失败')
      } finally {
        this.loading = false
      }
    },
    // 搜索:纯前端过滤,computed 自动响应,不调后端
    handleSearch() {
      // tableData 为 computed,searchForm 变化即自动更新
    },

    formatFileSize(bytes) {
      if (!bytes) return '0 B'
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    },

    // ===== 上传 =====
    showUploadDialog() {
      this.uploadForm = { docType: '', title: '', file: null }
      this.uploadVisible = true
    },
    onFileChange(file) {
      const ext = file.name.split('.').pop().toLowerCase()
      if (ext !== 'pdf') {
        ElMessage.error('仅支持PDF格式文件')
        return false
      }
      this.uploadForm.file = file.raw
    },
    onFileRemove() {
      this.uploadForm.file = null
    },
    async confirmUpload() {
      if (!this.uploadForm.docType) {
        ElMessage.warning('请选择文档类型')
        return
      }
      if (!this.uploadForm.title.trim()) {
        ElMessage.warning('请输入文档标题')
        return
      }
      if (!this.uploadForm.file) {
        ElMessage.warning('请选择PDF文件')
        return
      }

      this.uploadLoading = true
      try {
        const formData = new FormData()
        formData.append('file', this.uploadForm.file)
        formData.append('doc_type', this.uploadForm.docType)
        formData.append('title', this.uploadForm.title.trim())

        await request.post('/document/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        ElMessage.success('上传成功')
        this.uploadVisible = false
        this.fetchList()
      } catch (e) {
        ElMessage.error(e.message || '上传失败')
      } finally {
        this.uploadLoading = false
      }
    },
    resetUpload() {
      this.uploadForm = { docType: '', title: '', file: null }
    },

    // ===== 替换 =====
    showReplaceDialog(row) {
      this.replaceForm = {
        id: row.id,
        currentTitle: row.title,
        currentVersion: row.version,
        file: null
      }
      this.replaceVisible = true
    },
    onReplaceFileChange(file) {
      const ext = file.name.split('.').pop().toLowerCase()
      if (ext !== 'pdf') {
        ElMessage.error('仅支持PDF格式文件')
        return false
      }
      this.replaceForm.file = file.raw
    },
    onReplaceFileRemove() {
      this.replaceForm.file = null
    },
    async confirmReplace() {
      if (!this.replaceForm.file) {
        ElMessage.warning('请选择新的PDF文件')
        return
      }

      try {
        await ElMessageBox.confirm(
          `确定替换文档"${this.replaceForm.currentTitle}"吗？版本号将更新为 v${this.replaceForm.currentVersion + 1}`,
          '确认替换',
          { type: 'warning' }
        )
      } catch (e) {
        return
      }

      this.replaceLoading = true
      try {
        const formData = new FormData()
        formData.append('id', this.replaceForm.id)
        formData.append('file', this.replaceForm.file)

        await request.put('/document/replace', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        ElMessage.success('替换成功')
        this.replaceVisible = false
        this.fetchList()
      } catch (e) {
        ElMessage.error(e.message || '替换失败')
      } finally {
        this.replaceLoading = false
      }
    },
    resetReplace() {
      this.replaceForm = { id: 0, currentTitle: '', currentVersion: 0, file: null }
    },

    // ===== 预览 =====
    handlePreview(row) {
      this.previewUrl = row.file_url
      this.previewVisible = true
    },

    // ===== 历史版本 =====
    async showVersions(row) {
      this.versionsTitle = row.title
      this.versionsDocumentId = row.id
      this.versionsVisible = true
      this.versionsLoading = true
      try {
        const res = await request.get('/document/versions', {
          document_id: row.id
        })
        this.versionsData = res.data || []
      } catch (e) {
        ElMessage.error('加载版本历史失败')
      } finally {
        this.versionsLoading = false
      }
    },
    openVersion(row) {
      this.previewUrl = row.file_url
      this.previewVisible = true
    },

    // ===== 启用/禁用 =====
    async handleToggle(row) {
      const action = row.is_active ? '禁用' : '启用'
      try {
        await ElMessageBox.confirm(`确定${action}文档"${row.title}"吗？`, `确认${action}`)
        await request.put('/document/toggle', { id: row.id })
        ElMessage.success(`已${action}`)
        this.fetchList()
      } catch (e) {
        // 取消
      }
    },

    // ===== 删除 =====
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定删除文档"${row.title}"吗？此操作为逻辑删除，不会删除服务器上的PDF文件。`,
          '确认删除',
          { type: 'warning', confirmButtonText: '确认删除' }
        )
        await request.delete('/document/delete', { data: { id: row.id } })
        ElMessage.success('已删除')
        this.fetchList()
      } catch (e) {
        // 取消
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.page-header {
  margin-bottom: 16px;

  .page-title {
    font-size: 20px;
    font-weight: 600;
  }
  .page-desc {
    font-size: 13px;
    color: #909399;
    margin-left: 12px;
  }
}

.current-doc-info {
  font-weight: 600;
  color: #303133;
}
</style>