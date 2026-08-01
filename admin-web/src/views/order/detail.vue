<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">订单详情</span>
      <div>
        <el-button :icon="ArrowLeft" @click="$router.back()">返回</el-button>
        <el-button
          v-if="orderInfo.status"
          type="warning"
          :icon="Switch"
          @click="handleForceStatus"
        >
          强制状态变更
        </el-button>
        <el-button
          v-if="canRefund"
          type="danger"
          :icon="RefreshLeft"
          @click="handleRefund"
        >
          退款
        </el-button>
      </div>
    </div>

    <div v-loading="loading">
      <!-- 订单基本信息 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header">
            <span>订单基本信息</span>
            <el-tag :type="orderStatusTag(orderInfo.status)" size="large">
              {{ orderStatusLabel(orderInfo.status) }}
            </el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="订单号">{{ orderInfo.orderNo }}</el-descriptions-item>
          <el-descriptions-item label="订单状态">
            <el-tag :type="orderStatusTag(orderInfo.status)" size="small">
              {{ orderStatusLabel(orderInfo.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="订单金额">
            <span style="color: #f56c6c; font-weight: 600;">¥{{ orderInfo.amount }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="平台抽成">
            <span>¥{{ orderInfo.platformFee }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="打手收入">
            <span style="color: #67c23a; font-weight: 600;">¥{{ orderInfo.dispatcherIncome }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="大额验证">
            <el-tag :type="orderInfo.largeVerifyFailed ? 'danger' : 'success'" size="small">
              {{ orderInfo.largeVerifyFailed ? '验证失败' : '验证通过' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ orderInfo.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ orderInfo.updatedAt || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 用户信息 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">用户信息</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="下单用户">
            {{ orderInfo.userNickname || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="用户手机号">
            {{ maskPhone(orderInfo.userPhone) }}
          </el-descriptions-item>
          <el-descriptions-item label="打手">
            {{ orderInfo.dispatcherNickname || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="打手手机号">
            {{ maskPhone(orderInfo.dispatcherPhone) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 服务信息 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">服务信息</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="游戏名称">{{ orderInfo.gameName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="服务类型">{{ orderInfo.serviceType || '-' }}</el-descriptions-item>
          <el-descriptions-item label="单价">
            <span>¥{{ orderInfo.unitPrice }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="数量">{{ orderInfo.quantity || 1 }}</el-descriptions-item>
          <el-descriptions-item label="服务描述" :span="2">
            {{ orderInfo.serviceDesc || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 状态流转时间线 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">状态流转</span>
        </template>
        <el-timeline v-if="timeline.length > 0">
          <el-timeline-item
            v-for="(item, index) in timeline"
            :key="index"
            :timestamp="item.time"
            :color="item.color"
            :type="item.type"
          >
            {{ item.label }}
            <span v-if="item.remark" style="color: #909399; font-size: 12px;">
              （{{ item.remark }}）
            </span>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无状态流转记录" :image-size="60" />
      </el-card>

      <!-- 支付信息 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">支付信息</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="支付单号">
            {{ orderInfo.paymentNo || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="支付金额">
            <span style="color: #f56c6c; font-weight: 600;">¥{{ orderInfo.paymentAmount || orderInfo.amount }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="支付方式">
            {{ orderInfo.paymentMethod || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="支付时间">
            {{ orderInfo.paymentTime || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 评价信息 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">评价信息</span>
        </template>
        <div v-if="orderInfo.review">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="评分">
              <el-rate
                :model-value="orderInfo.review.rating"
                disabled
                show-score
                text-color="#ff9900"
              />
            </el-descriptions-item>
            <el-descriptions-item label="评价时间">
              {{ orderInfo.review.createdAt || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="评价内容" :span="2">
              {{ orderInfo.review.content || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="标签" :span="2">
              <el-tag
                v-for="tag in (orderInfo.review.tags || [])"
                :key="tag"
                size="small"
                style="margin-right: 8px;"
              >
                {{ tag }}
              </el-tag>
              <span v-if="!orderInfo.review.tags || orderInfo.review.tags.length === 0">-</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>
        <el-empty v-else description="暂无评价" :image-size="60" />
      </el-card>
    </div>

    <!-- 强制扭转状态弹窗 -->
    <el-dialog
      v-model="forceStatusDialogVisible"
      title="强制扭转状态"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="forceStatusFormRef" :model="forceStatusForm" :rules="forceStatusRules" label-width="90px">
        <el-form-item label="当前状态">
          <el-tag :type="orderStatusTag(orderInfo.status)" size="default">
            {{ orderStatusLabel(orderInfo.status) }}
          </el-tag>
        </el-form-item>
        <el-form-item label="目标状态" prop="targetStatus">
          <el-select v-model="forceStatusForm.targetStatus" placeholder="请选择目标状态" style="width: 100%">
            <el-option
              v-for="s in statusOptions"
              :key="s.value"
              :label="s.label"
              :value="s.value"
              :disabled="s.value === orderInfo.status"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="变更原因" prop="reason">
          <el-input
            v-model="forceStatusForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入变更原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="forceStatusDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="forceStatusLoading" @click="handleForceStatusSubmit">
          确认变更
        </el-button>
      </template>
    </el-dialog>

    <!-- 退款弹窗 -->
    <el-dialog
      v-model="refundDialogVisible"
      title="订单退款"
      width="450px"
      :close-on-click-modal="false"
    >
      <el-form ref="refundFormRef" :model="refundForm" :rules="refundRules" label-width="90px">
        <el-form-item label="订单号">
          <span>{{ orderInfo.orderNo }}</span>
        </el-form-item>
        <el-form-item label="订单金额">
          <span style="color: #f56c6c; font-weight: 600;">¥{{ orderInfo.amount }}</span>
        </el-form-item>
        <el-form-item label="退款金额" prop="refundAmount">
          <el-input-number
            v-model="refundForm.refundAmount"
            :min="0"
            :max="orderInfo.amount"
            :precision="2"
            style="width: 100%"
            placeholder="请输入退款金额"
          />
        </el-form-item>
        <el-form-item label="退款原因" prop="reason">
          <el-input
            v-model="refundForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入退款原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="refundLoading" @click="handleRefundSubmit">
          确认退款
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ArrowLeft, Switch, RefreshLeft } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'OrderDetail',
  data() {
    return {
      ArrowLeft,
      Switch,
      RefreshLeft,
      loading: false,
      orderId: null,
      orderInfo: {},
      statusOptions: [
        { value: 'created', label: '已创建' },
        { value: 'accepted', label: '已接单' },
        { value: 'processing', label: '进行中' },
        { value: 'completed', label: '已完成' },
        { value: 'settled', label: '已结算' },
        { value: 'refunded', label: '已退款' },
        { value: 'cancelled', label: '已取消' }
      ],
      forceStatusDialogVisible: false,
      forceStatusForm: {
        targetStatus: '',
        reason: ''
      },
      forceStatusRules: {
        targetStatus: [{ required: true, message: '请选择目标状态', trigger: 'change' }],
        reason: [{ required: true, message: '请输入变更原因', trigger: 'blur' }]
      },
      forceStatusLoading: false,
      refundDialogVisible: false,
      refundForm: {
        refundAmount: null,
        reason: ''
      },
      refundRules: {
        refundAmount: [{ required: true, message: '请输入退款金额', trigger: 'blur' }],
        reason: [{ required: true, message: '请输入退款原因', trigger: 'blur' }]
      },
      refundLoading: false
    }
  },
  computed: {
    canRefund() {
      const status = this.orderInfo.status
      return status && status !== 'refunded' && status !== 'cancelled'
    },
    timeline() {
      const data = this.orderInfo.timeline || []
      if (data.length > 0) return data
      const statusFlow = [
        { key: 'created', label: '订单创建', color: '#409eff' },
        { key: 'accepted', label: '打手接单', color: '#e6a23c' },
        { key: 'processing', label: '服务进行中', color: '#e6a23c' },
        { key: 'completed', label: '服务完成', color: '#67c23a' },
        { key: 'settled', label: '订单结算', color: '#67c23a' }
      ]
      const currentStatus = this.orderInfo.status
      const statusIndex = statusFlow.findIndex(s => s.key === currentStatus)
      return statusFlow.map((item, index) => ({
        label: item.label,
        time: index <= statusIndex ? (this.orderInfo.updatedAt || this.orderInfo.createdAt) : '---',
        color: index <= statusIndex ? item.color : '#c0c4cc'
      }))
    }
  },
  mounted() {
    this.orderId = this.$route.params.id
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      this.loading = true
      try {
        const res = await request.get(`/orders/${this.orderId}`)
        this.orderInfo = res.data || {}
      } catch (err) {
        console.error('获取订单详情失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleForceStatus() {
      this.forceStatusForm = {
        targetStatus: '',
        reason: ''
      }
      this.forceStatusDialogVisible = true
    },
    async handleForceStatusSubmit() {
      const valid = await this.$refs.forceStatusFormRef.validate().catch(() => false)
      if (!valid) return
      this.forceStatusLoading = true
      try {
        await request.post(`/orders/${this.orderId}/force-status`, {
          targetStatus: this.forceStatusForm.targetStatus,
          reason: this.forceStatusForm.reason
        })
        ElMessage.success('状态变更成功')
        this.forceStatusDialogVisible = false
        this.fetchDetail()
      } catch (err) {
        console.error('状态变更失败:', err)
      } finally {
        this.forceStatusLoading = false
      }
    },
    handleRefund() {
      this.refundForm = {
        refundAmount: this.orderInfo.amount,
        reason: ''
      }
      this.refundDialogVisible = true
    },
    async handleRefundSubmit() {
      const valid = await this.$refs.refundFormRef.validate().catch(() => false)
      if (!valid) return
      this.refundLoading = true
      try {
        await request.post(`/orders/${this.orderId}/refund`, {
          refundAmount: this.refundForm.refundAmount,
          reason: this.refundForm.reason
        })
        ElMessage.success('退款成功')
        this.refundDialogVisible = false
        this.fetchDetail()
      } catch (err) {
        console.error('退款失败:', err)
      } finally {
        this.refundLoading = false
      }
    },
    maskPhone(phone) {
      if (!phone) return '-'
      return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    },
    orderStatusTag(status) {
      const map = {
        created: 'info',
        accepted: 'warning',
        processing: 'warning',
        completed: 'success',
        settled: '',
        refunded: 'danger',
        cancelled: 'info'
      }
      return map[status] || 'info'
    },
    orderStatusLabel(status) {
      const map = {
        created: '已创建',
        accepted: '已接单',
        processing: '进行中',
        completed: '已完成',
        settled: '已结算',
        refunded: '已退款',
        cancelled: '已取消'
      }
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

@media screen and (max-width: 768px) {
  :deep(.el-descriptions) {
    .el-descriptions__body {
      .el-descriptions__table {
        display: block;
      }
    }
  }
}
</style>