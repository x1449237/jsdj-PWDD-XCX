<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">分角色协议版本管理</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="全部" clearable style="width: 140px">
            <el-option label="玩家(买家)" value="buyer" />
            <el-option label="打手" value="player" />
            <el-option label="分销商" value="distributor" />
            <el-option label="俱乐部" value="club" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议类型">
          <el-select v-model="searchForm.agreement_type" placeholder="全部" clearable style="width: 160px">
            <el-option label="用户服务协议" value="user_service" />
            <el-option label="隐私政策" value="privacy" />
            <el-option label="打手入驻协议" value="player_entry" />
            <el-option label="俱乐部入驻协议" value="club_entry" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
          <el-button type="success" :icon="Plus" @click="handleCreate">创建新版本</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="roleTag(row.role)" size="small">
              {{ roleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="协议类型" width="140" align="center">
          <template #default="{ row }">
            {{ agreementTypeLabel(row.agreement_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本号" width="100" align="center" />
        <el-table-column label="内容摘要" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.content ? row.content.substring(0, 100) + '...' : '' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
              {{ row.is_active ? '当前生效' : '历史版本' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="publish_time" label="发布时间" width="170" align="center" />
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">查看</el-button>
            <el-button
              v-if="!row.is_active"
              type="success"
              link
              size="small"
              @click="handlePublish(row)"
            >
              发布
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isView ? '查看协议' : (isEdit ? '编辑协议' : '创建新版本')"
      width="700px"
      :close-on-click-modal="false"
    >
      <el-form v-if="!isView" :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" placeholder="请选择角色" style="width: 100%;">
            <el-option label="玩家(买家)" value="buyer" />
            <el-option label="打手" value="player" />
            <el-option label="分销商" value="distributor" />
            <el-option label="俱乐部" value="club" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议类型" prop="agreement_type">
          <el-select v-model="form.agreement_type" placeholder="请选择协议类型" style="width: 100%;">
            <el-option label="用户服务协议" value="user_service" />
            <el-option label="隐私政策" value="privacy" />
            <el-option label="打手入驻协议" value="player_entry" />
            <el-option label="俱乐部入驻协议" value="club_entry" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="12"
            placeholder="请输入协议内容"
          />
        </el-form-item>
      </el-form>
      <div v-else class="view-content">
        <div class="view-info">
          <span class="info-item">角色：{{ roleLabel(form.role) }}</span>
          <span class="info-item">类型：{{ agreementTypeLabel(form.agreement_type) }}</span>
          <span class="info-item">版本：v{{ form.version }}</span>
        </div>
        <div class="content-text">{{ form.content }}</div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ isView ? '关闭' : '取消' }}</el-button>
        <el-button v-if="!isView" type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'AgreementVersion',
  components: { Search, Refresh, Plus },
  data() {
    return {
      loading: false,
      searchForm: {
        role: '',
        agreement_type: ''
      },
      // 全量数据(一次加载),tableData 由 computed 本地过滤,减轻后端压力
      allData: [],
      dialogVisible: false,
      isView: false,
      isEdit: false,
      submitting: false,
      formRef: null,
      form: {
        id: 0,
        role: '',
        agreement_type: '',
        version: 1,
        content: '',
        is_active: 0
      },
      rules: {
        role: [{ required: true, message: '请选择角色', trigger: 'change' }],
        agreement_type: [{ required: true, message: '请选择协议类型', trigger: 'change' }],
        content: [{ required: true, message: '请输入协议内容', trigger: 'blur' }]
      }
    }
  },
  computed: {
    // 前端本地按角色 + 协议类型过滤,不再每次筛选都请求后端
    tableData() {
      const { role, agreement_type } = this.searchForm
      return this.allData.filter(r => {
        if (role && r.role !== role) return false
        if (agreement_type && r.agreement_type !== agreement_type) return false
        return true
      })
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    // 仅全量加载一次(及写操作后刷新),不带筛选参数
    loadData() {
      this.loading = true
      request.get('/compliance/agreement_version_list').then(res => {
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
      this.searchForm = {
        role: '',
        agreement_type: ''
      }
      // computed 自动响应,无需调后端
    },
    handleCreate() {
      this.isView = false
      this.isEdit = false
      this.form = {
        id: 0,
        role: '',
        agreement_type: '',
        version: 1,
        content: '',
        is_active: 0
      }
      this.dialogVisible = true
    },
    handleView(row) {
      this.isView = true
      this.isEdit = false
      this.form = { ...row }
      this.dialogVisible = true
    },
    handlePublish(row) {
      ElMessageBox.confirm(
        `确定发布此版本？发布后将成为当前生效版本，用户需重新签署。`,
        '提示',
        {
          confirmButtonText: '确定发布',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(() => {
        request.post('/compliance/agreement_version_publish', {
          id: row.id
        }).then(() => {
          ElMessage.success('发布成功')
          this.loadData()
        })
      }).catch(() => {})
    },
    handleSubmit() {
      this.$refs.formRef.validate(valid => {
        if (!valid) return
        this.submitting = true
        request.post('/compliance/agreement_version_create', {
          role: this.form.role,
          agreement_type: this.form.agreement_type,
          content: this.form.content
        }).then(() => {
          ElMessage.success('创建成功')
          this.dialogVisible = false
          this.loadData()
        }).finally(() => {
          this.submitting = false
        })
      })
    },
    roleTag(role) {
      const map = {
        buyer: 'primary',
        player: 'warning',
        distributor: 'success',
        club: 'danger'
      }
      return map[role] || 'info'
    },
    roleLabel(role) {
      const map = {
        buyer: '玩家(买家)',
        player: '打手',
        distributor: '分销商',
        club: '俱乐部'
      }
      return map[role] || role
    },
    agreementTypeLabel(type) {
      const map = {
        user_service: '用户服务协议',
        privacy: '隐私政策',
        player_entry: '打手入驻协议',
        club_entry: '俱乐部入驻协议'
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
.view-content {
  max-height: 500px;
  overflow-y: auto;
}
.view-info {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}
.info-item {
  margin-right: 24px;
  font-size: 14px;
  color: #606266;
}
.content-text {
  white-space: pre-wrap;
  line-height: 1.8;
  color: #303133;
  font-size: 14px;
}
</style>
