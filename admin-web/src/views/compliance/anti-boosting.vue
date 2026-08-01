<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">代练拦截管理</span>
    </div>

    <el-tabs v-model="activeTab" type="card">
      <el-tab-pane label="违禁词规则" name="rules">
        <el-card class="search-card">
          <el-form :model="searchForm" :inline="true" class="search-form-inline">
            <el-form-item label="级别">
              <el-select v-model="searchForm.level" placeholder="全部" clearable style="width: 140px">
                <el-option label="警告" value="warn" />
                <el-option label="拦截" value="intercept" />
                <el-option label="封禁" value="ban" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadRules">搜索</el-button>
              <el-button :icon="Refresh" @click="handleReset">重置</el-button>
              <el-button type="success" :icon="Plus" @click="handleAddRule">新增规则</el-button>
              <el-button type="warning" @click="handleExpandWords">扩充违禁词库</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="rulesData" v-loading="rulesLoading" stripe border style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="keyword" label="关键词" width="180" show-overflow-tooltip />
            <el-table-column label="级别" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="levelTag(row.level)" size="small">
                  {{ levelLabel(row.level) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="create_time" label="创建时间" width="170" align="center" />
            <el-table-column label="操作" width="180" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEditRule(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteRule(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="rulesPagination.page"
              v-model:page-size="rulesPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="rulesPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadRules"
              @current-change="loadRules"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="拦截日志" name="logs">
        <el-card class="search-card">
          <el-form :model="logSearchForm" :inline="true" class="search-form-inline">
            <el-form-item label="来源">
              <el-select v-model="logSearchForm.source" placeholder="全部" clearable style="width: 140px">
                <el-option label="订单" value="order" />
                <el-option label="私聊" value="chat" />
                <el-option label="群聊" value="group_chat" />
              </el-select>
            </el-form-item>
            <el-form-item label="级别">
              <el-select v-model="logSearchForm.level" placeholder="全部" clearable style="width: 140px">
                <el-option label="警告" value="warn" />
                <el-option label="拦截" value="intercept" />
                <el-option label="封禁" value="ban" />
              </el-select>
            </el-form-item>
            <el-form-item label="用户ID">
              <el-input v-model="logSearchForm.user_id" placeholder="请输入用户ID" clearable style="width: 140px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadLogs">搜索</el-button>
              <el-button :icon="Refresh" @click="handleResetLogs">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="logsData" v-loading="logsLoading" stripe border style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column label="来源" width="100" align="center">
              <template #default="{ row }">
                {{ sourceLabel(row.source) }}
              </template>
            </el-table-column>
            <el-table-column prop="source_id" label="来源ID" width="100" align="center" />
            <el-table-column prop="user_id" label="用户ID" width="100" align="center" />
            <el-table-column prop="matched_keyword" label="匹配关键词" width="150" show-overflow-tooltip />
            <el-table-column label="级别" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="levelTag(row.level)" size="small">
                  {{ levelLabel(row.level) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="处理状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.handled ? 'success' : 'warning'" size="small">
                  {{ row.handled ? '已处理' : '待处理' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="create_time" label="时间" width="170" align="center" />
            <el-table-column label="操作" width="120" fixed="right" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="!row.handled"
                  type="primary"
                  link
                  size="small"
                  @click="handleLog(row)"
                >
                  标记处理
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="logsPagination.page"
              v-model:page-size="logsPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="logsPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadLogs"
              @current-change="loadLogs"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog
      v-model="ruleDialogVisible"
      :title="isEditRule ? '编辑规则' : '新增规则'"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form :model="ruleForm" :rules="ruleRules" ref="ruleFormRef" label-width="80px">
        <el-form-item label="关键词" prop="keyword">
          <el-input v-model="ruleForm.keyword" placeholder="请输入关键词" />
        </el-form-item>
        <el-form-item label="级别" prop="level">
          <el-select v-model="ruleForm.level" placeholder="请选择级别" style="width: 100%;">
            <el-option label="警告" value="warn" />
            <el-option label="拦截" value="intercept" />
            <el-option label="封禁" value="ban" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="ruleForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSubmitting" @click="submitRule">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'AntiBoosting',
  components: { Search, Refresh, Plus },
  data() {
    return {
      activeTab: 'rules',
      rulesLoading: false,
      searchForm: {
        level: ''
      },
      rulesData: [],
      rulesPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      logsLoading: false,
      logSearchForm: {
        source: '',
        level: '',
        user_id: ''
      },
      logsData: [],
      logsPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      ruleDialogVisible: false,
      isEditRule: false,
      ruleSubmitting: false,
      ruleFormRef: null,
      ruleForm: {
        id: 0,
        keyword: '',
        level: 'warn',
        status: 1
      },
      ruleRules: {
        keyword: [{ required: true, message: '请输入关键词', trigger: 'blur' }],
        level: [{ required: true, message: '请选择级别', trigger: 'change' }]
      }
    }
  },
  mounted() {
    this.loadRules()
  },
  methods: {
    loadRules() {
      this.rulesLoading = true
      request.get('/compliance/anti_boosting_rule_list', {
        params: {
          level: this.searchForm.level,
          page: this.rulesPagination.page,
          page_size: this.rulesPagination.pageSize
        }
      }).then(res => {
        this.rulesData = res.list || []
        this.rulesPagination.total = res.total || 0
      }).finally(() => {
        this.rulesLoading = false
      })
    },
    handleReset() {
      this.searchForm = { level: '' }
      this.rulesPagination.page = 1
      this.loadRules()
    },
    handleAddRule() {
      this.isEditRule = false
      this.ruleForm = {
        id: 0,
        keyword: '',
        level: 'warn',
        status: 1
      }
      this.ruleDialogVisible = true
    },
    handleEditRule(row) {
      this.isEditRule = true
      this.ruleForm = { ...row }
      this.ruleDialogVisible = true
    },
    handleDeleteRule(row) {
      ElMessageBox.confirm('确定删除此规则？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        request.delete('/compliance/anti_boosting_rule_delete', {
          data: { id: row.id }
        }).then(() => {
          ElMessage.success('删除成功')
          this.loadRules()
        })
      }).catch(() => {})
    },
    submitRule() {
      this.$refs.ruleFormRef.validate(valid => {
        if (!valid) return
        this.ruleSubmitting = true
        const requestFn = this.isEditRule
          ? request.put('/compliance/anti_boosting_rule_update', this.ruleForm)
          : request.post('/compliance/anti_boosting_rule_create', this.ruleForm)
        requestFn.then(() => {
          ElMessage.success(this.isEditRule ? '更新成功' : '创建成功')
          this.ruleDialogVisible = false
          this.loadRules()
        }).finally(() => {
          this.ruleSubmitting = false
        })
      })
    },
    handleExpandWords() {
      ElMessageBox.confirm('确定扩充代练违禁词库？将添加游戏代练、外挂、上分、破解、线下交易、赌博等关键词。', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        request.post('/compliance/expand_sensitive_words').then(() => {
          ElMessage.success('扩充成功')
        })
      }).catch(() => {})
    },
    loadLogs() {
      this.logsLoading = true
      request.get('/compliance/anti_boosting_log_list', {
        params: {
          source: this.logSearchForm.source,
          level: this.logSearchForm.level,
          user_id: this.logSearchForm.user_id,
          page: this.logsPagination.page,
          page_size: this.logsPagination.pageSize
        }
      }).then(res => {
        this.logsData = res.list || []
        this.logsPagination.total = res.total || 0
      }).finally(() => {
        this.logsLoading = false
      })
    },
    handleResetLogs() {
      this.logSearchForm = {
        source: '',
        level: '',
        user_id: ''
      }
      this.logsPagination.page = 1
      this.loadLogs()
    },
    handleLog(row) {
      request.post('/compliance/anti_boosting_log_handle', {
        id: row.id
      }).then(() => {
        ElMessage.success('已标记处理')
        this.loadLogs()
      })
    },
    levelTag(level) {
      const map = {
        warn: 'warning',
        intercept: 'danger',
        ban: 'danger'
      }
      return map[level] || 'info'
    },
    levelLabel(level) {
      const map = {
        warn: '警告',
        intercept: '拦截',
        ban: '封禁'
      }
      return map[level] || level
    },
    sourceLabel(source) {
      const map = {
        order: '订单',
        chat: '私聊',
        private_chat: '私聊',
        group_chat: '群聊'
      }
      return map[source] || source
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
.pagination-container {
  margin-top: 20px;
  text-align: right;
}
</style>
