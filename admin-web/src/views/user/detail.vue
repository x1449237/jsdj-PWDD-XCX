<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">用户详情</span>
      <el-button :icon="ArrowLeft" @click="$router.back()">返回</el-button>
    </div>

    <div v-loading="loading">
      <!-- 基本信息 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
            <div>
              <el-button
                v-if="userInfo.status === 'active'"
                type="danger"
                :icon="Lock"
                @click="handleToggleBan"
              >
                封禁用户
              </el-button>
              <el-button
                v-else
                type="success"
                :icon="Unlock"
                @click="handleToggleBan"
              >
                解封用户
              </el-button>
            </div>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户ID">{{ userInfo.id }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ userInfo.nickname }}</el-descriptions-item>
          <el-descriptions-item label="头像">
            <el-avatar :size="48" :src="userInfo.avatar" />
          </el-descriptions-item>
          <el-descriptions-item label="手机号">{{ maskPhone(userInfo.phone) }}</el-descriptions-item>
          <el-descriptions-item label="用户类型">
            <el-tag :type="userTypeTag(userInfo.userType)" size="small">
              {{ userTypeLabel(userInfo.userType) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="userInfo.status === 'active' ? 'success' : 'danger'" size="small">
              {{ userInfo.status === 'active' ? '正常' : '已封禁' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ userInfo.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="最后登录">{{ userInfo.lastLoginAt || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 实名认证信息 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">实名认证信息</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="真实姓名">{{ userInfo.realName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ maskIdCard(userInfo.idCard) }}</el-descriptions-item>
          <el-descriptions-item label="认证状态">
            <el-tag :type="realNameStatusTag(userInfo.realNameStatus)" size="small">
              {{ realNameStatusLabel(userInfo.realNameStatus) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="认证时间">{{ userInfo.realNameVerifiedAt || '-' }}</el-descriptions-item>
        </el-descriptions>
        <div v-if="livenessRecords.length > 0" class="sub-section">
          <h4>活体检测记录</h4>
          <el-table :data="livenessRecords" stripe border size="small">
            <el-table-column prop="id" label="记录ID" width="80" />
            <el-table-column prop="type" label="检测类型" width="120" />
            <el-table-column label="检测结果" width="100">
              <template #default="{ row }">
                <el-tag :type="row.passed ? 'success' : 'danger'" size="small">
                  {{ row.passed ? '通过' : '未通过' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="score" label="置信度" width="100" />
            <el-table-column prop="createdAt" label="检测时间" min-width="160" />
          </el-table>
        </div>
      </el-card>

      <!-- 订单统计 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">订单统计</span>
        </template>
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="stat-mini">
              <div class="stat-mini-value">{{ orderStats.totalOrders }}</div>
              <div class="stat-mini-label">总订单数</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-mini">
              <div class="stat-mini-value">{{ orderStats.completionRate }}%</div>
              <div class="stat-mini-label">完成率</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-mini">
              <div class="stat-mini-value">¥{{ orderStats.totalAmount }}</div>
              <div class="stat-mini-label">消费金额</div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 信用分 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header">
            <span>信用分</span>
            <el-tag type="primary" size="large">{{ userInfo.creditScore || 0 }}分</el-tag>
          </div>
        </template>
        <el-table :data="creditRecords" stripe border size="small" v-if="creditRecords.length > 0">
          <el-table-column prop="id" label="记录ID" width="80" />
          <el-table-column prop="reason" label="原因" min-width="160" />
          <el-table-column prop="change" label="变化" width="100">
            <template #default="{ row }">
              <span :style="{ color: row.change > 0 ? '#67c23a' : '#f56c6c' }">
                {{ row.change > 0 ? '+' + row.change : row.change }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="时间" width="170" />
        </el-table>
        <el-empty v-else description="暂无扣分记录" :image-size="80" />
      </el-card>

      <!-- 分销关系 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header">
            <span>分销关系</span>
            <el-button
              v-if="userInfo.inviteCode"
              type="danger"
              plain
              size="small"
              @click="handleUnbindInvite"
            >
              强制解绑邀请码
            </el-button>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="上级分销商">
            {{ userInfo.parentDistributor || '无' }}
          </el-descriptions-item>
          <el-descriptions-item label="绑定邀请码">
            {{ userInfo.inviteCode || '无' }}
          </el-descriptions-item>
          <el-descriptions-item label="绑定时间">
            {{ userInfo.inviteBoundAt || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>
    </div>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ArrowLeft, Lock, Unlock } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'UserDetail',
  data() {
    return {
      ArrowLeft,
      Lock,
      Unlock,
      loading: false,
      userId: null,
      userInfo: {},
      livenessRecords: [],
      orderStats: {
        totalOrders: 0,
        completionRate: 0,
        totalAmount: 0
      },
      creditRecords: []
    }
  },
  mounted() {
    this.userId = this.$route.params.id
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      this.loading = true
      try {
        const res = await request.get(`/users/${this.userId}`)
        const data = res.data || {}
        this.userInfo = data.userInfo || data
        this.livenessRecords = data.livenessRecords || []
        this.orderStats = data.orderStats || { totalOrders: 0, completionRate: 0, totalAmount: 0 }
        this.creditRecords = data.creditRecords || []
      } catch (err) {
        console.error('获取用户详情失败:', err)
      } finally {
        this.loading = false
      }
    },
    async handleToggleBan() {
      const isBan = this.userInfo.status === 'active'
      const action = isBan ? '封禁' : '解封'
      try {
        await ElMessageBox.confirm(
          `确定要${action}用户「${this.userInfo.nickname}」吗？`,
          `${action}确认`,
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        const url = isBan
          ? `/users/${this.userId}/ban`
          : `/users/${this.userId}/unban`
        await request.post(url)
        ElMessage.success(`${action}成功`)
        this.fetchDetail()
      } catch (err) {
        if (err !== 'cancel') {
          console.error(`${action}失败:`, err)
        }
      }
    },
    async handleUnbindInvite() {
      try {
        await ElMessageBox.confirm(
          '确定要强制解绑该用户的邀请码吗？此操作不可撤销。',
          '强制解绑确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.post(`/users/${this.userId}/unbind-invite`)
        ElMessage.success('邀请码已解绑')
        this.fetchDetail()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('解绑邀请码失败:', err)
        }
      }
    },
    maskPhone(phone) {
      if (!phone) return '-'
      return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    },
    maskIdCard(idCard) {
      if (!idCard) return '-'
      return idCard.replace(/(\d{4})\d{10}(\d{4})/, '$1**********$2')
    },
    userTypeTag(type) {
      const map = { normal: '', large_verified: 'warning', risk: 'danger' }
      return map[type] || 'info'
    },
    userTypeLabel(type) {
      const map = { normal: '普通用户', large_verified: '大额验证', risk: '风险用户' }
      return map[type] || type
    },
    realNameStatusTag(status) {
      const map = { verified: 'success', pending: 'warning', rejected: 'danger', unverified: 'info' }
      return map[status] || 'info'
    },
    realNameStatusLabel(status) {
      const map = { verified: '已认证', pending: '审核中', rejected: '已拒绝', unverified: '未认证' }
      return map[status] || '未知'
    }
  }
}
</script>

<style lang="scss" scoped>
.section-card {
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
}

.sub-section {
  margin-top: 20px;

  h4 {
    font-size: 14px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 12px;
  }
}

.stat-mini {
  text-align: center;
  padding: 16px 0;
  background: #f5f7fa;
  border-radius: 8px;

  .stat-mini-value {
    font-size: 24px;
    font-weight: 700;
    color: #409eff;
  }

  .stat-mini-label {
    margin-top: 6px;
    font-size: 13px;
    color: #909399;
  }
}

@media screen and (max-width: 768px) {
  .stat-mini-value {
    font-size: 20px;
  }
}
</style>