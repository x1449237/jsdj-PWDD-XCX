<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">聊天审计</span>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 订单私聊审计 -->
      <el-tab-pane label="订单私聊审计" name="orderChat">
        <el-card class="search-card">
          <el-form :model="orderSearchForm" :inline="true" class="search-form-inline">
            <el-form-item label="用户">
              <el-input
                v-model="orderSearchForm.userId"
                placeholder="用户ID/昵称"
                clearable
                style="width: 180px"
                @keyup.enter="handleOrderSearch"
              />
            </el-form-item>
            <el-form-item label="打手">
              <el-input
                v-model="orderSearchForm.playerId"
                placeholder="打手ID/昵称"
                clearable
                style="width: 180px"
                @keyup.enter="handleOrderSearch"
              />
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="orderSearchForm.dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleOrderSearch">搜索</el-button>
              <el-button :icon="Refresh" @click="handleOrderReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="orderSessionList" v-loading="orderLoading" stripe border style="width: 100%">
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
                <el-button type="primary" link size="small" @click="handleViewOrderMessages(row)">查看消息</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="orderPagination.page"
              v-model:page-size="orderPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="orderPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleOrderSearch"
              @current-change="handleOrderSearch"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 群聊消息审计 -->
      <el-tab-pane label="群聊消息审计" name="groupChat">
        <el-card class="search-card">
          <el-form :model="groupSearchForm" :inline="true" class="search-form-inline">
            <el-form-item label="群名称">
              <el-input
                v-model="groupSearchForm.groupName"
                placeholder="请输入群名称"
                clearable
                style="width: 200px"
                @keyup.enter="handleGroupSearch"
              />
            </el-form-item>
            <el-form-item label="群类型">
              <el-select v-model="groupSearchForm.groupType" placeholder="全部" clearable style="width: 150px">
                <el-option label="订单群" value="order" />
                <el-option label="私聊群" value="private" />
                <el-option label="公共群" value="public" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleGroupSearch">搜索</el-button>
              <el-button :icon="Refresh" @click="handleGroupReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="groupList" v-loading="groupLoading" stripe border style="width: 100%">
            <el-table-column prop="groupName" label="群名称" min-width="160" show-overflow-tooltip />
            <el-table-column label="群类型" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.groupType === 'order' ? 'primary' : row.groupType === 'private' ? 'warning' : 'success'" size="small">
                  {{ row.groupType === 'order' ? '订单群' : row.groupType === 'private' ? '私聊群' : '公共群' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="memberCount" label="成员数" width="80" align="center" />
            <el-table-column prop="lastMessageTime" label="最后消息时间" width="170" align="center" />
            <el-table-column label="操作" width="120" fixed="right" align="center">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleViewGroupMessages(row)">查看消息</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="groupPagination.page"
              v-model:page-size="groupPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="groupPagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleGroupSearch"
              @current-change="handleGroupSearch"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 订单私聊消息弹窗 -->
    <el-dialog
      v-model="orderMessageDialogVisible"
      :title="'会话消息 - ' + currentOrderSessionId"
      width="780px"
      :close-on-click-modal="false"
    >
      <div class="message-list" v-loading="orderMessageLoading">
        <div
          v-for="(msg, index) in orderMessages"
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
        <el-empty v-if="orderMessages.length === 0 && !orderMessageLoading" description="暂无消息" />
      </div>
    </el-dialog>

    <!-- 群聊消息弹窗 -->
    <el-dialog
      v-model="groupMessageDialogVisible"
      :title="'群聊消息 - ' + currentGroupName"
      width="780px"
      :close-on-click-modal="false"
    >
      <div class="message-list" v-loading="groupMessageLoading">
        <div
          v-for="(msg, index) in groupMessages"
          :key="index"
          class="message-item"
        >
          <div class="message-header">
            <el-avatar :size="28" :src="msg.avatar" style="margin-right: 8px;" />
            <span class="message-sender user">{{ msg.senderName }}</span>
            <span class="message-time">{{ msg.time }}</span>
          </div>
          <div class="message-body">
            <div class="message-text" v-if="msg.textContent">{{ msg.textContent }}</div>
            <div class="message-image" v-if="msg.imageUrl">
              <el-image :src="msg.imageUrl" style="max-width: 200px; max-height: 200px;" fit="contain" />
            </div>
          </div>
        </div>
        <el-empty v-if="groupMessages.length === 0 && !groupMessageLoading" description="暂无消息" />
      </div>
      <div class="pagination-container" style="margin-top: 16px;">
        <el-pagination
          v-model:current-page="groupMessagePagination.page"
          v-model:page-size="groupMessagePagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="groupMessagePagination.total"
          layout="total, prev, pager, next"
          small
          @size-change="fetchGroupMessages"
          @current-change="fetchGroupMessages"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'

export default {
  name: 'ChatAudit',
  data() {
    return {
      Search,
      Refresh,
      activeTab: 'orderChat',
      // 订单私聊审计
      orderSearchForm: {
        userId: '',
        playerId: '',
        dateRange: []
      },
      orderSessionList: [],
      orderLoading: false,
      orderPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      orderMessageDialogVisible: false,
      currentOrderSessionId: '',
      orderMessages: [],
      orderMessageLoading: false,
      // 群聊消息审计
      groupSearchForm: {
        groupName: '',
        groupType: ''
      },
      groupList: [],
      groupLoading: false,
      groupPagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      groupMessageDialogVisible: false,
      currentGroupId: null,
      currentGroupName: '',
      groupMessages: [],
      groupMessageLoading: false,
      groupMessagePagination: {
        page: 1,
        pageSize: 20,
        total: 0
      }
    }
  },
  mounted() {
    this.fetchOrderSessions()
  },
  methods: {
    handleTabChange(tab) {
      if (tab === 'orderChat') {
        this.fetchOrderSessions()
      } else if (tab === 'groupChat') {
        this.fetchGroupList()
      }
    },
    // ========== 订单私聊审计 ==========
    async fetchOrderSessions() {
      this.orderLoading = true
      try {
        const params = {
          page: this.orderPagination.page,
          pageSize: this.orderPagination.pageSize,
          userId: this.orderSearchForm.userId || undefined,
          playerId: this.orderSearchForm.playerId || undefined,
          startDate: this.orderSearchForm.dateRange?.[0] || undefined,
          endDate: this.orderSearchForm.dateRange?.[1] || undefined
        }
        const res = await request.get('/chat-audit/sessions', { params })
        this.orderSessionList = res.data?.list || []
        this.orderPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取会话列表失败:', err)
      } finally {
        this.orderLoading = false
      }
    },
    handleOrderSearch() {
      this.orderPagination.page = 1
      this.fetchOrderSessions()
    },
    handleOrderReset() {
      this.orderSearchForm = {
        userId: '',
        playerId: '',
        dateRange: []
      }
      this.handleOrderSearch()
    },
    async handleViewOrderMessages(row) {
      this.currentOrderSessionId = row.sessionId
      this.orderMessageDialogVisible = true
      this.orderMessageLoading = true
      try {
        const res = await request.get(`/chat-audit/sessions/${row.sessionId}/messages`)
        this.orderMessages = res.data?.list || []
      } catch (err) {
        console.error('获取消息列表失败:', err)
      } finally {
        this.orderMessageLoading = false
      }
    },
    // ========== 群聊消息审计 ==========
    async fetchGroupList() {
      this.groupLoading = true
      try {
        const params = {
          page: this.groupPagination.page,
          pageSize: this.groupPagination.pageSize,
          groupName: this.groupSearchForm.groupName || undefined,
          groupType: this.groupSearchForm.groupType || undefined
        }
        const res = await request.get('/group-monitor/groups', { params })
        this.groupList = res.data?.list || []
        this.groupPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取群聊列表失败:', err)
      } finally {
        this.groupLoading = false
      }
    },
    handleGroupSearch() {
      this.groupPagination.page = 1
      this.fetchGroupList()
    },
    handleGroupReset() {
      this.groupSearchForm = {
        groupName: '',
        groupType: ''
      }
      this.handleGroupSearch()
    },
    async handleViewGroupMessages(row) {
      this.currentGroupId = row.id
      this.currentGroupName = row.groupName
      this.groupMessagePagination = { page: 1, pageSize: 20, total: 0 }
      this.groupMessageDialogVisible = true
      this.fetchGroupMessages()
    },
    async fetchGroupMessages() {
      this.groupMessageLoading = true
      try {
        const params = {
          page: this.groupMessagePagination.page,
          pageSize: this.groupMessagePagination.pageSize
        }
        const res = await request.get(`/group-monitor/groups/${this.currentGroupId}/messages`, { params })
        this.groupMessages = res.data?.list || []
        this.groupMessagePagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取群消息失败:', err)
      } finally {
        this.groupMessageLoading = false
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
  margin-left: auto;
  font-size: 12px;
  color: #909399;
}

.message-body {
  font-size: 13px;
  color: #303133;
  line-height: 1.8;
  padding-left: 36px;
}

.msg-label {
  color: #909399;
  font-size: 12px;
}

.message-text,
.message-asr,
.message-ocr,
.message-sensitive {
  margin-bottom: 4px;
}

.message-image {
  margin-top: 4px;
}

@media screen and (max-width: 768px) {
  .search-form-inline :deep(.el-form-item) {
    display: block;
    margin-right: 0;
  }
}
</style>