<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">聊天审计与AI风控</span>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 会话列表 Tab -->
      <el-tab-pane label="会话列表" name="sessions">
        <el-card class="search-card">
          <el-form :model="searchForm" :inline="true" class="search-form-inline">
            <el-form-item label="用户">
              <el-input
                v-model="searchForm.userId"
                placeholder="用户ID/昵称"
                clearable
                style="width: 180px"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item label="打手">
              <el-input
                v-model="searchForm.playerId"
                placeholder="打手ID/昵称"
                clearable
                style="width: 180px"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="searchForm.dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
              <el-button :icon="Refresh" @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="sessionList" v-loading="sessionLoading" stripe border style="width: 100%">
            <el-table-column prop="sessionId" label="会话编号" width="180" show-overflow-tooltip />
            <el-table-column label="参与用户" min-width="180">
              <template #default="{ row }">
                <div class="participants">
                  <el-tag size="small" type="primary" style="margin-right: 4px;">用户</el-tag>
                  <span>{{ row.userName }}</span>
                  <el-tag size="small" type="warning" style="margin: 0 4px;">打手</el-tag>
                  <span>{{ row.playerName }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="最后消息" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="last-message">{{ row.lastMessage }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="lastMessageTime" label="最后消息时间" width="170" align="center" />
            <el-table-column label="风险标记" width="120" align="center">
              <template #default="{ row }">
                <el-tag
                  :type="row.riskLevel === 'high' ? 'danger' : row.riskLevel === 'medium' ? 'warning' : 'success'"
                  size="small"
                >
                  {{ row.riskLevel === 'high' ? '高风险' : row.riskLevel === 'medium' ? '中风险' : '正常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleViewMessages(row)">查看消息</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="sessionPagination.page"
              v-model:page-size="sessionPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="sessionPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSearch"
              @current-change="handleSearch"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 风险用户 Tab -->
      <el-tab-pane label="风险用户" name="riskUsers">
        <el-card class="table-card">
          <el-table :data="riskUserList" v-loading="riskUserLoading" stripe border style="width: 100%">
            <el-table-column label="用户" min-width="150">
              <template #default="{ row }">
                <div class="user-info">
                  <el-avatar :size="32" :src="row.avatar" style="margin-right: 8px;" />
                  <span>{{ row.userName }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="riskType" label="风险类型" width="120" align="center">
              <template #default="{ row }">
                <el-tag type="danger" size="small">{{ row.riskType }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="风险等级" width="100" align="center">
              <template #default="{ row }">
                <el-tag
                  :type="row.riskLevel === 'high' ? 'danger' : row.riskLevel === 'medium' ? 'warning' : 'info'"
                  size="small"
                >
                  {{ row.riskLevel === 'high' ? '高' : row.riskLevel === 'medium' ? '中' : '低' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="riskScore" label="风险评分" width="100" align="center">
              <template #default="{ row }">
                <span :style="{ color: row.riskScore >= 80 ? '#f56c6c' : row.riskScore >= 50 ? '#e6a23c' : '#909399' }">
                  {{ row.riskScore }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="riskReason" label="原因" min-width="180" show-overflow-tooltip />
            <el-table-column prop="detectedAt" label="检测时间" width="170" align="center" />
            <el-table-column label="处理状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag
                  :type="row.handleStatus === 'pending' ? 'warning' : row.handleStatus === 'ignored' ? 'info' : 'danger'"
                  size="small"
                >
                  {{ row.handleStatus === 'pending' ? '待处理' : row.handleStatus === 'ignored' ? '已忽略' : row.handleStatus === 'warned' ? '已警告' : '已冻结' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="row.handleStatus === 'pending'"
                  type="primary"
                  link
                  size="small"
                  @click="handleProcessRisk(row)"
                >
                  处理
                </el-button>
                <span v-else style="color: #909399;">-</span>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="riskUserPagination.page"
              v-model:page-size="riskUserPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="riskUserPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchRiskUsers"
              @current-change="fetchRiskUsers"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 消息列表对话框 -->
    <el-dialog
      v-model="messageDialogVisible"
      :title="'会话消息 - ' + currentSessionId"
      width="780px"
      :close-on-click-modal="false"
    >
      <div class="message-list" v-loading="messageLoading">
        <div
          v-for="(msg, index) in messages"
          :key="index"
          class="message-item"
          :class="{ 'is-risk': msg.hasSensitiveWord }"
        >
          <div class="message-header">
            <span class="message-sender" :class="msg.role === 'user' ? 'user' : 'player'">
              {{ msg.role === 'user' ? '用户' : '打手' }} - {{ msg.senderName }}
            </span>
            <span class="message-time">{{ msg.time }}</span>
            <el-tag
              v-if="msg.hasSensitiveWord"
              type="danger"
              size="small"
              effect="dark"
              style="margin-left: 8px;"
            >
              敏感词命中
            </el-tag>
          </div>
          <div class="message-body">
            <div class="message-text" v-if="msg.textContent">
              <span class="msg-label">文字：</span>{{ msg.textContent }}
            </div>
            <div class="message-asr" v-if="msg.asrText">
              <span class="msg-label">ASR转文字：</span>{{ msg.asrText }}
            </div>
            <div class="message-ocr" v-if="msg.ocrText">
              <span class="msg-label">OCR识别：</span>{{ msg.ocrText }}
            </div>
            <div class="message-nlp" v-if="msg.nlpResult">
              <span class="msg-label">NLP分析：</span>
              <el-tag
                :type="msg.nlpResult === '正常' ? 'success' : 'warning'"
                size="small"
              >
                {{ msg.nlpResult }}
              </el-tag>
            </div>
            <div class="message-sensitive" v-if="msg.sensitiveWords && msg.sensitiveWords.length > 0">
              <span class="msg-label">命中敏感词：</span>
              <el-tag
                v-for="(word, wi) in msg.sensitiveWords"
                :key="wi"
                type="danger"
                size="small"
                style="margin-right: 4px;"
              >
                {{ word }}
              </el-tag>
            </div>
          </div>
        </div>
        <el-empty v-if="messages.length === 0 && !messageLoading" description="暂无消息" />
      </div>
    </el-dialog>

    <!-- 处理风险用户弹窗 -->
    <el-dialog
      v-model="processDialogVisible"
      title="处理风险用户"
      width="460px"
      :close-on-click-modal="false"
    >
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="风险用户">
          <span>{{ currentRiskUser.userName }}</span>
        </el-form-item>
        <el-form-item label="风险类型">
          <el-tag type="danger" size="small">{{ currentRiskUser.riskType }}</el-tag>
        </el-form-item>
        <el-form-item label="风险等级">
          <el-tag
            :type="currentRiskUser.riskLevel === 'high' ? 'danger' : 'warning'"
            size="small"
          >
            {{ currentRiskUser.riskLevel === 'high' ? '高' : '中' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="处理方式" prop="action">
          <el-radio-group v-model="processForm.action">
            <el-radio value="ignore">忽略</el-radio>
            <el-radio value="warn">警告</el-radio>
            <el-radio value="freeze">冻结</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理备注" prop="remark">
          <el-input
            v-model="processForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入处理备注"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="handleProcessSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'ChatAudit',
  data() {
    return {
      Search,
      Refresh,
      activeTab: 'sessions',
      searchForm: {
        userId: '',
        playerId: '',
        dateRange: []
      },
      sessionList: [],
      sessionLoading: false,
      sessionPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      riskUserList: [],
      riskUserLoading: false,
      riskUserPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      messageDialogVisible: false,
      currentSessionId: '',
      messages: [],
      messageLoading: false,
      processDialogVisible: false,
      currentRiskUser: {},
      processForm: {
        action: 'ignore',
        remark: ''
      },
      processLoading: false
    }
  },
  mounted() {
    this.fetchSessions()
  },
  methods: {
    handleTabChange(tab) {
      if (tab === 'sessions') {
        this.fetchSessions()
      } else if (tab === 'riskUsers') {
        this.fetchRiskUsers()
      }
    },
    async fetchSessions() {
      this.sessionLoading = true
      try {
        const params = {
          page: this.sessionPagination.page,
          pageSize: this.sessionPagination.pageSize,
          userId: this.searchForm.userId || undefined,
          playerId: this.searchForm.playerId || undefined,
          startDate: this.searchForm.dateRange?.[0] || undefined,
          endDate: this.searchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/chat-audit/sessions', { params })
        this.sessionList = res.data?.list || []
        this.sessionPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取会话列表失败:', err)
      } finally {
        this.sessionLoading = false
      }
    },
    handleSearch() {
      this.sessionPagination.page = 1
      this.fetchSessions()
    },
    handleReset() {
      this.searchForm = {
        userId: '',
        playerId: '',
        dateRange: []
      }
      this.handleSearch()
    },
    async handleViewMessages(row) {
      this.currentSessionId = row.sessionId
      this.messageDialogVisible = true
      this.messageLoading = true
      try {
        const res = await request.get(`/chat-audit/sessions/${row.sessionId}/messages`)
        this.messages = res.data?.list || []
      } catch (err) {
        console.error('获取消息列表失败:', err)
      } finally {
        this.messageLoading = false
      }
    },
    async fetchRiskUsers() {
      this.riskUserLoading = true
      try {
        const params = {
          page: this.riskUserPagination.page,
          pageSize: this.riskUserPagination.pageSize
        }
        const res = await request.get('/chat-audit/risk-users', { params })
        this.riskUserList = res.data?.list || []
        this.riskUserPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取风险用户列表失败:', err)
      } finally {
        this.riskUserLoading = false
      }
    },
    handleProcessRisk(row) {
      this.currentRiskUser = row
      this.processForm = {
        action: 'ignore',
        remark: ''
      }
      this.processDialogVisible = true
    },
    async handleProcessSubmit() {
      const actionMap = { ignore: '忽略', warn: '警告', freeze: '冻结' }
      try {
        await ElMessageBox.confirm(
          `确定要对用户「${this.currentRiskUser.userName}」执行「${actionMap[this.processForm.action]}」操作吗？`,
          '处理确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        this.processLoading = true
        await request.post(`/chat-audit/risk-users/${this.currentRiskUser.id}/process`, {
          action: this.processForm.action,
          remark: this.processForm.remark
        })
        ElMessage.success('处理成功')
        this.processDialogVisible = false
        this.fetchRiskUsers()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('处理失败:', err)
        }
      } finally {
        this.processLoading = false
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

.participants {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px;
}

.last-message {
  color: #909399;
  font-size: 13px;
}

.user-info {
  display: flex;
  align-items: center;
}

.message-list {
  max-height: 500px;
  overflow-y: auto;
}

.message-item {
  padding: 12px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 8px;

  &.is-risk {
    background-color: #fef0f0;
    border-left: 3px solid #f56c6c;
    border-radius: 4px;
  }
}

.message-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.message-sender {
  font-weight: 600;
  font-size: 14px;

  &.user {
    color: #409eff;
  }

  &.player {
    color: #e6a23c;
  }
}

.message-time {
  margin-left: 12px;
  font-size: 12px;
  color: #909399;
}

.message-body {
  font-size: 13px;
  color: #303133;
  line-height: 1.8;
}

.msg-label {
  color: #909399;
  font-size: 12px;
}

.message-text,
.message-asr,
.message-ocr,
.message-nlp,
.message-sensitive {
  margin-bottom: 4px;
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>