<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">群聊监察</span>
    </div>

    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="群名称">
          <el-input
            v-model="searchForm.groupName"
            placeholder="请输入群名称"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="群类型">
          <el-select v-model="searchForm.groupType" placeholder="全部" clearable style="width: 150px">
            <el-option label="订单群" value="order" />
            <el-option label="私聊群" value="private" />
            <el-option label="公共群" value="public" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="正常" value="active" />
            <el-option label="已解散" value="disbanded" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格 -->
    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="groupName" label="群名称" min-width="160" show-overflow-tooltip />
        <el-table-column label="群类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="groupTypeTag(row.groupType)" size="small">
              {{ groupTypeLabel(row.groupType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creatorName" label="创建者" width="120" show-overflow-tooltip />
        <el-table-column prop="memberCount" label="成员数" width="80" align="center" />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '正常' : '已解散' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">详情</el-button>
            <el-button type="primary" link size="small" @click="handleViewMessages(row)">消息</el-button>
            <el-button
              v-if="row.status === 'active'"
              type="danger"
              link
              size="small"
              @click="handleDisband(row)"
            >
              解散
            </el-button>
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
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </el-card>

    <!-- 群详情抽屉 -->
    <el-drawer
      v-model="detailDrawerVisible"
      title="群聊详情"
      size="480px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading">
        <el-descriptions :column="1" border size="small" style="margin-bottom: 20px;">
          <el-descriptions-item label="群名称">{{ groupDetail.groupName }}</el-descriptions-item>
          <el-descriptions-item label="群类型">
            <el-tag :type="groupTypeTag(groupDetail.groupType)" size="small">
              {{ groupTypeLabel(groupDetail.groupType) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建者">{{ groupDetail.creatorName }}</el-descriptions-item>
          <el-descriptions-item label="成员数">{{ groupDetail.memberCount }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="groupDetail.status === 'active' ? 'success' : 'info'" size="small">
              {{ groupDetail.status === 'active' ? '正常' : '已解散' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ groupDetail.createdAt }}</el-descriptions-item>
        </el-descriptions>

        <h4 class="section-title">群公告</h4>
        <el-alert
          v-if="groupDetail.announcement"
          :title="groupDetail.announcement"
          type="info"
          :closable="false"
          show-icon
        />
        <el-empty v-else description="暂无群公告" :image-size="40" />

        <h4 class="section-title" style="margin-top: 20px;">群成员列表</h4>
        <el-table :data="groupMembers" stripe border size="small" style="width: 100%">
          <el-table-column label="头像" width="50">
            <template #default="{ row: member }">
              <el-avatar :size="28" :src="member.avatar" />
            </template>
          </el-table-column>
          <el-table-column prop="nickname" label="昵称" min-width="100" show-overflow-tooltip />
          <el-table-column label="角色" width="80" align="center">
            <template #default="{ row: member }">
              <el-tag :type="member.role === 'owner' ? 'danger' : member.role === 'admin' ? 'warning' : ''" size="small">
                {{ member.role === 'owner' ? '群主' : member.role === 'admin' ? '管理员' : '成员' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" align="center">
            <template #default="{ row: member }">
              <el-button
                v-if="member.muted"
                type="success"
                link
                size="small"
                @click="handleUnmuteMember(member)"
              >
                解除禁言
              </el-button>
              <el-button
                v-else
                type="warning"
                link
                size="small"
                @click="handleMuteMember(member)"
              >
                禁言
              </el-button>
              <el-button
                type="danger"
                link
                size="small"
                @click="handleRemoveMember(member)"
              >
                移出
              </el-button>
              <el-button
                type="danger"
                link
                size="small"
                @click="handleBanUser(member)"
              >
                封禁
              </el-button>
              <el-button
                type="danger"
                link
                size="small"
                @click="handleFreezeFunds(member)"
              >
                冻结资金
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-drawer>

    <!-- 群消息记录弹窗 -->
    <el-dialog
      v-model="messageDialogVisible"
      :title="'群消息记录 - ' + currentGroupName"
      width="780px"
      :close-on-click-modal="false"
    >
      <div class="message-list" v-loading="messageLoading">
        <div
          v-for="(msg, index) in messages"
          :key="index"
          class="message-item"
        >
          <div class="message-header">
            <el-avatar :size="28" :src="msg.avatar" style="margin-right: 8px;" />
            <span class="message-sender">{{ msg.senderName }}</span>
            <span class="message-time">{{ msg.time }}</span>
          </div>
          <div class="message-body">
            <div class="message-text" v-if="msg.textContent">{{ msg.textContent }}</div>
            <div class="message-image" v-if="msg.imageUrl">
              <el-image :src="msg.imageUrl" style="max-width: 200px; max-height: 200px;" fit="contain" />
            </div>
          </div>
        </div>
        <el-empty v-if="messages.length === 0 && !messageLoading" description="暂无消息" />
      </div>
      <div class="pagination-container" style="margin-top: 16px;">
        <el-pagination
          v-model:current-page="messagePagination.page"
          v-model:page-size="messagePagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="messagePagination.total"
          layout="total, prev, pager, next"
          small
          @size-change="fetchMessages"
          @current-change="fetchMessages"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'GroupMonitor',
  data() {
    return {
      Search,
      Refresh,
      searchForm: {
        groupName: '',
        groupType: '',
        status: ''
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      detailDrawerVisible: false,
      detailLoading: false,
      currentGroupId: null,
      groupDetail: {},
      groupMembers: [],
      messageDialogVisible: false,
      currentGroupName: '',
      messages: [],
      messageLoading: false,
      messagePagination: {
        page: 1,
        pageSize: 20,
        total: 0
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
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          groupName: this.searchForm.groupName || undefined,
          groupType: this.searchForm.groupType || undefined,
          status: this.searchForm.status || undefined
        }
        const res = await request.get('/group-monitor/groups', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取群聊列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.searchForm = {
        groupName: '',
        groupType: '',
        status: ''
      }
      this.handleSearch()
    },
    async handleViewDetail(row) {
      this.currentGroupId = row.id
      this.detailDrawerVisible = true
      this.detailLoading = true
      try {
        const res = await request.get(`/group-monitor/groups/${row.id}`)
        this.groupDetail = res.data || {}
        this.groupMembers = res.data?.members || []
      } catch (err) {
        console.error('获取群详情失败:', err)
      } finally {
        this.detailLoading = false
      }
    },
    async handleViewMessages(row) {
      this.currentGroupId = row.id
      this.currentGroupName = row.groupName
      this.messagePagination = { page: 1, pageSize: 20, total: 0 }
      this.messageDialogVisible = true
      this.fetchMessages()
    },
    async fetchMessages() {
      this.messageLoading = true
      try {
        const params = {
          page: this.messagePagination.page,
          pageSize: this.messagePagination.pageSize
        }
        const res = await request.get(`/group-monitor/groups/${this.currentGroupId}/messages`, { params })
        this.messages = res.data?.list || []
        this.messagePagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取消息记录失败:', err)
      } finally {
        this.messageLoading = false
      }
    },
    async handleDisband(row) {
      try {
        await ElMessageBox.confirm(
          `确定要解散群聊「${row.groupName}」吗？此操作不可撤销。`,
          '解散确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/group-monitor/groups/${row.id}/disband`)
        ElMessage.success('解散成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('解散失败:', err)
        }
      }
    },
    async handleMuteMember(member) {
      try {
        await ElMessageBox.prompt('请输入禁言时长（分钟）', '禁言成员', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputPattern: /^\d+$/,
          inputErrorMessage: '请输入有效的分钟数'
        })
        .then(async ({ value }) => {
          await request.post(`/group-monitor/groups/${this.currentGroupId}/mute`, {
            userId: member.userId,
            duration: Number(value)
          })
          ElMessage.success('禁言成功')
          this.handleViewDetail({ id: this.currentGroupId })
        })
      } catch (err) {
        if (err !== 'cancel') {
          console.error('禁言失败:', err)
        }
      }
    },
    async handleUnmuteMember(member) {
      try {
        await ElMessageBox.confirm(
          `确定要解除「${member.nickname}」的禁言吗？`,
          '解除禁言',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
        )
        await request.post(`/group-monitor/groups/${this.currentGroupId}/unmute`, {
          userId: member.userId
        })
        ElMessage.success('已解除禁言')
        this.handleViewDetail({ id: this.currentGroupId })
      } catch (err) {
        if (err !== 'cancel') {
          console.error('解除禁言失败:', err)
        }
      }
    },
    async handleRemoveMember(member) {
      try {
        await ElMessageBox.confirm(
          `确定要将「${member.nickname}」移出群聊吗？`,
          '移出成员',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/group-monitor/groups/${this.currentGroupId}/remove-member`, {
          userId: member.userId
        })
        ElMessage.success('移出成功')
        this.handleViewDetail({ id: this.currentGroupId })
      } catch (err) {
        if (err !== 'cancel') {
          console.error('移出成员失败:', err)
        }
      }
    },
    async handleBanUser(member) {
      try {
        await ElMessageBox.confirm(
          `确定要封禁用户「${member.nickname}」吗？`,
          '封禁用户',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/group-monitor/users/${member.userId}/ban`)
        ElMessage.success('封禁成功')
        this.handleViewDetail({ id: this.currentGroupId })
      } catch (err) {
        if (err !== 'cancel') {
          console.error('封禁用户失败:', err)
        }
      }
    },
    async handleFreezeFunds(member) {
      try {
        await ElMessageBox.confirm(
          `确定要冻结「${member.nickname}」的账户资金吗？`,
          '冻结资金',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/group-monitor/users/${member.userId}/freeze-funds`)
        ElMessage.success('冻结资金成功')
        this.handleViewDetail({ id: this.currentGroupId })
      } catch (err) {
        if (err !== 'cancel') {
          console.error('冻结资金失败:', err)
        }
      }
    },
    groupTypeTag(type) {
      const map = { order: 'primary', private: 'warning', public: 'success' }
      return map[type] || 'info'
    },
    groupTypeLabel(type) {
      const map = { order: '订单群', private: '私聊群', public: '公共群' }
      return map[type] || type
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

.section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #303133;
}

.message-list {
  max-height: 500px;
  overflow-y: auto;
}

.message-item {
  padding: 10px 12px;
  border-bottom: 1px solid #ebeef5;

  &:last-child {
    border-bottom: none;
  }
}

.message-header {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
}

.message-sender {
  font-weight: 600;
  font-size: 13px;
  color: #409eff;
}

.message-time {
  margin-left: auto;
  font-size: 12px;
  color: #909399;
}

.message-body {
  font-size: 13px;
  color: #303133;
  padding-left: 36px;
}

.message-text {
  line-height: 1.6;
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