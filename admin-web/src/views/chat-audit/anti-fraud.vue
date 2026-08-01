<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">飞单风控管理</span>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 风控规则 Tab -->
      <el-tab-pane label="风控规则" name="rules">
        <el-card class="search-card">
          <el-form :model="ruleSearchForm" :inline="true" class="search-form-inline">
            <el-form-item label="规则类型">
              <el-select v-model="ruleSearchForm.ruleType" placeholder="全部" clearable style="width: 160px">
                <el-option label="微信号" value="wechat" />
                <el-option label="QQ号" value="qq" />
                <el-option label="手机号" value="phone" />
                <el-option label="线下转账" value="offline_transfer" />
                <el-option label="银行卡" value="bank_card" />
              </el-select>
            </el-form-item>
            <el-form-item label="风险等级">
              <el-select v-model="ruleSearchForm.level" placeholder="全部" clearable style="width: 140px">
                <el-option label="警告" value="warning" />
                <el-option label="禁言" value="mute" />
                <el-option label="封禁" value="ban" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="ruleSearchForm.status" placeholder="全部" clearable style="width: 120px">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="fetchRules">搜索</el-button>
              <el-button :icon="Refresh" @click="resetRuleSearch">重置</el-button>
              <el-button type="success" :icon="Plus" @click="handleAddRule">新增规则</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="ruleList" v-loading="ruleLoading" stripe border style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="rule_name" label="规则名称" min-width="140" show-overflow-tooltip />
            <el-table-column prop="rule_type" label="规则类型" width="120" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="getRuleTypeTag(row.rule_type)">{{ getRuleTypeName(row.rule_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="pattern" label="匹配规则" min-width="220" show-overflow-tooltip />
            <el-table-column prop="level" label="风险等级" width="100" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="getLevelTag(row.level)">{{ getLevelName(row.level) }}</el-tag>
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
                <el-button type="primary" link size="small" @click="handleEditRule(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteRule(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="rulePagination.page"
              v-model:page-size="rulePagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="rulePagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchRules"
              @current-change="fetchRules"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 拦截日志 Tab -->
      <el-tab-pane label="拦截日志" name="logs">
        <el-card class="search-card">
          <el-form :model="logSearchForm" :inline="true" class="search-form-inline">
            <el-form-item label="发送者ID">
              <el-input v-model="logSearchForm.senderId" placeholder="用户ID" clearable style="width: 140px" />
            </el-form-item>
            <el-form-item label="会话ID">
              <el-input v-model="logSearchForm.sessionId" placeholder="会话ID" clearable style="width: 140px" />
            </el-form-item>
            <el-form-item label="风险等级">
              <el-select v-model="logSearchForm.level" placeholder="全部" clearable style="width: 120px">
                <el-option label="警告" value="warning" />
                <el-option label="禁言" value="mute" />
                <el-option label="封禁" value="ban" />
              </el-select>
            </el-form-item>
            <el-form-item label="处理状态">
              <el-select v-model="logSearchForm.handled" placeholder="全部" clearable style="width: 120px">
                <el-option label="未处理" :value="0" />
                <el-option label="已处理" :value="1" />
              </el-select>
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="logSearchForm.dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="fetchLogs">搜索</el-button>
              <el-button :icon="Refresh" @click="resetLogSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="logList" v-loading="logLoading" stripe border style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="session_id" label="会话ID" width="110" align="center" />
            <el-table-column prop="sender_id" label="发送者ID" width="110" align="center" />
            <el-table-column prop="message_id" label="消息ID" width="110" align="center" />
            <el-table-column prop="rule_id" label="规则ID" width="100" align="center" />
            <el-table-column prop="matched_content" label="匹配内容" min-width="160" show-overflow-tooltip />
            <el-table-column prop="level" label="风险等级" width="100" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="getLevelTag(row.level)">{{ getLevelName(row.level) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="handled" label="处理状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="row.handled === 1 ? 'success' : 'warning'">
                  {{ row.handled === 1 ? '已处理' : '未处理' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="create_time" label="创建时间" width="170" align="center" />
            <el-table-column label="操作" width="140" fixed="right" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="row.handled !== 1"
                  type="primary"
                  link
                  size="small"
                  @click="handleProcessLog(row)"
                >
                  处理
                </el-button>
                <span v-else style="color: #909399;">-</span>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="logPagination.page"
              v-model:page-size="logPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="logPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchLogs"
              @current-change="fetchLogs"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 规则编辑弹窗 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="ruleDialogMode === 'add' ? '新增风控规则' : '编辑风控规则'"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form :model="ruleForm" :rules="ruleFormRules" ref="ruleFormRef" label-width="100px">
        <el-form-item label="规则名称" prop="rule_name">
          <el-input v-model="ruleForm.rule_name" placeholder="请输入规则名称" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="规则类型" prop="rule_type">
          <el-select v-model="ruleForm.rule_type" placeholder="请选择规则类型" style="width: 100%">
            <el-option label="微信号" value="wechat" />
            <el-option label="QQ号" value="qq" />
            <el-option label="手机号" value="phone" />
            <el-option label="线下转账" value="offline_transfer" />
            <el-option label="银行卡" value="bank_card" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配规则" prop="pattern">
          <el-input
            v-model="ruleForm.pattern"
            type="textarea"
            :rows="3"
            placeholder="请输入正则表达式，如 /微信|vx/i"
            maxlength="512"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="风险等级" prop="level">
          <el-radio-group v-model="ruleForm.level">
            <el-radio value="warning">警告</el-radio>
            <el-radio value="mute">禁言</el-radio>
            <el-radio value="ban">封禁</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="ruleForm.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="ruleForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSubmitting" @click="handleRuleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 处理日志弹窗 -->
    <el-dialog
      v-model="processDialogVisible"
      title="处理拦截日志"
      width="460px"
      :close-on-click-modal="false"
    >
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="日志ID">
          <span>{{ currentLog.id }}</span>
        </el-form-item>
        <el-form-item label="匹配内容">
          <span>{{ currentLog.matched_content }}</span>
        </el-form-item>
        <el-form-item label="风险等级">
          <el-tag size="small" :type="getLevelTag(currentLog.level)">{{ getLevelName(currentLog.level) }}</el-tag>
        </el-form-item>
        <el-form-item label="处理结果" prop="handle_result">
          <el-input
            v-model="processForm.handle_result"
            type="textarea"
            :rows="3"
            placeholder="请输入处理结果"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="processSubmitting" @click="handleProcessSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'AntiFraud',
  data() {
    return {
      Search,
      Refresh,
      Plus,
      activeTab: 'rules',
      // 规则相关
      ruleSearchForm: {
        ruleType: '',
        level: '',
        status: ''
      },
      ruleList: [],
      ruleLoading: false,
      rulePagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      // 日志相关
      logSearchForm: {
        senderId: '',
        sessionId: '',
        level: '',
        handled: '',
        dateRange: []
      },
      logList: [],
      logLoading: false,
      logPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      // 规则编辑弹窗
      ruleDialogVisible: false,
      ruleDialogMode: 'add',
      ruleFormRef: null,
      ruleForm: {
        id: 0,
        rule_name: '',
        rule_type: 'wechat',
        pattern: '',
        level: 'warning',
        sort: 0,
        status: 1
      },
      ruleFormRules: {
        rule_name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
        rule_type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
        pattern: [{ required: true, message: '请输入匹配规则', trigger: 'blur' }]
      },
      ruleSubmitting: false,
      // 处理日志弹窗
      processDialogVisible: false,
      currentLog: {},
      processForm: {
        handle_result: ''
      },
      processSubmitting: false
    }
  },
  mounted() {
    this.fetchRules()
  },
  methods: {
    handleTabChange(tab) {
      if (tab === 'rules') {
        this.fetchRules()
      } else if (tab === 'logs') {
        this.fetchLogs()
      }
    },
    getRuleTypeTag(type) {
      const map = {
        wechat: 'primary',
        qq: 'success',
        phone: 'warning',
        offline_transfer: 'danger',
        bank_card: 'info'
      }
      return map[type] || 'info'
    },
    getRuleTypeName(type) {
      const map = {
        wechat: '微信号',
        qq: 'QQ号',
        phone: '手机号',
        offline_transfer: '线下转账',
        bank_card: '银行卡'
      }
      return map[type] || type
    },
    getLevelTag(level) {
      const map = {
        warning: 'warning',
        mute: 'danger',
        ban: 'danger'
      }
      return map[level] || 'info'
    },
    getLevelName(level) {
      const map = {
        warning: '警告',
        mute: '禁言',
        ban: '封禁'
      }
      return map[level] || level
    },
    async fetchRules() {
      this.ruleLoading = true
      try {
        const params = {
          page: this.rulePagination.page,
          pageSize: this.rulePagination.pageSize,
          rule_type: this.ruleSearchForm.ruleType || undefined,
          level: this.ruleSearchForm.level || undefined,
          status: this.ruleSearchForm.status !== '' ? this.ruleSearchForm.status : undefined
        }
        const res = await request.get('/chat/anti_fraud_rules', { params })
        this.ruleList = res.data?.list || []
        this.rulePagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取规则列表失败:', err)
      } finally {
        this.ruleLoading = false
      }
    },
    resetRuleSearch() {
      this.ruleSearchForm = {
        ruleType: '',
        level: '',
        status: ''
      }
      this.rulePagination.page = 1
      this.fetchRules()
    },
    handleAddRule() {
      this.ruleDialogMode = 'add'
      this.ruleForm = {
        id: 0,
        rule_name: '',
        rule_type: 'wechat',
        pattern: '',
        level: 'warning',
        sort: 0,
        status: 1
      }
      this.ruleDialogVisible = true
    },
    handleEditRule(row) {
      this.ruleDialogMode = 'edit'
      this.ruleForm = { ...row }
      this.ruleDialogVisible = true
    },
    async handleDeleteRule(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除规则「${row.rule_name}」吗？`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete('/chat/anti_fraud_rule', { data: { id: row.id } })
        ElMessage.success('删除成功')
        this.fetchRules()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    async handleRuleSubmit() {
      if (this.ruleFormRef) {
        await this.ruleFormRef.validate(async (valid) => {
          if (!valid) return
        })
      }
      try {
        this.ruleSubmitting = true
        if (this.ruleDialogMode === 'add') {
          await request.post('/chat/anti_fraud_rule', this.ruleForm)
          ElMessage.success('创建成功')
        } else {
          await request.put('/chat/anti_fraud_rule', this.ruleForm)
          ElMessage.success('更新成功')
        }
        this.ruleDialogVisible = false
        this.fetchRules()
      } catch (err) {
        console.error('提交失败:', err)
      } finally {
        this.ruleSubmitting = false
      }
    },
    async fetchLogs() {
      this.logLoading = true
      try {
        const params = {
          page: this.logPagination.page,
          pageSize: this.logPagination.pageSize,
          sender_id: this.logSearchForm.senderId || undefined,
          session_id: this.logSearchForm.sessionId || undefined,
          level: this.logSearchForm.level || undefined,
          handled: this.logSearchForm.handled !== '' ? this.logSearchForm.handled : undefined,
          start_date: this.logSearchForm.dateRange?.[0] || undefined,
          end_date: this.logSearchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/chat/anti_fraud_logs', { params })
        this.logList = res.data?.list || []
        this.logPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取日志列表失败:', err)
      } finally {
        this.logLoading = false
      }
    },
    resetLogSearch() {
      this.logSearchForm = {
        senderId: '',
        sessionId: '',
        level: '',
        handled: '',
        dateRange: []
      }
      this.logPagination.page = 1
      this.fetchLogs()
    },
    handleProcessLog(row) {
      this.currentLog = row
      this.processForm = {
        handle_result: ''
      }
      this.processDialogVisible = true
    },
    async handleProcessSubmit() {
      try {
        this.processSubmitting = true
        await request.put('/chat/anti_fraud_log_handle', {
          id: this.currentLog.id,
          handle_result: this.processForm.handle_result
        })
        ElMessage.success('处理成功')
        this.processDialogVisible = false
        this.fetchLogs()
      } catch (err) {
        console.error('处理失败:', err)
      } finally {
        this.processSubmitting = false
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
