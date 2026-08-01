<template>
  <el-dialog
    :model-value="modelValue"
    title="批量操作确认"
    width="500px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="!confirming"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="batch-confirm-content">
      <!-- 操作摘要 -->
      <el-alert
        :title="operationSummary"
        type="warning"
        :closable="false"
        show-icon
        class="batch-alert"
      />

      <div class="batch-info">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="操作类型">
            <el-tag type="warning">{{ operationTypeLabel }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="影响数量">
            <span style="color: #f56c6c; font-weight: 600;">{{ affectedCount }} 条</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 步骤流程 -->
      <div class="confirm-steps">
        <!-- 第一步：生成二维码 -->
        <div class="step-section" v-if="!confirming">
          <el-divider content-position="left">
            <el-icon><Iphone /></el-icon>
            <span style="margin-left: 6px;">扫码确认</span>
          </el-divider>
          <p class="step-desc">
            请使用另一管理员账号扫描以下二维码进行确认，或直接点击下方按钮生成确认码。
          </p>
          <div class="qrcode-wrapper">
            <canvas ref="qrCanvasRef" class="qrcode-canvas"></canvas>
            <div v-if="qrLoading" class="qrcode-loading">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>生成中...</span>
            </div>
          </div>
          <div class="batch-id-info">
            <span class="label">批次号：</span>
            <el-tag>{{ batchId || '---' }}</el-tag>
          </div>
        </div>

        <!-- 第二步：等待确认 -->
        <div class="step-section" v-if="confirming">
          <el-divider content-position="left">
            <el-icon><Clock /></el-icon>
            <span style="margin-left: 6px;">等待确认</span>
          </el-divider>
          <div class="waiting-status">
            <el-icon class="rotating-icon" :size="48"><Loading /></el-icon>
            <p class="waiting-text">等待另一管理员扫码确认中...</p>
            <el-progress
              :percentage="pollProgress"
              :stroke-width="6"
              :show-text="false"
              style="width: 280px;"
            />
            <p class="waiting-remain">
              剩余时间：{{ countdownText }}
            </p>
          </div>
        </div>

        <!-- 第三步：结果 -->
        <div class="step-section" v-if="confirmResult">
          <el-divider content-position="left">
            <span>确认结果</span>
          </el-divider>
          <el-result
            :icon="confirmResult === 'success' ? 'success' : 'error'"
            :title="confirmResult === 'success' ? '操作已确认' : '操作失败'"
            :sub-title="confirmResultMessage"
          >
            <template #extra>
              <el-button
                v-if="confirmResult === 'success'"
                type="primary"
                @click="handleConfirmDone"
              >
                完成
              </el-button>
              <el-button
                v-else
                type="primary"
                @click="handleRetry"
              >
                重试
              </el-button>
            </template>
          </el-result>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button
        v-if="!confirming && !confirmResult"
        @click="handleCancel"
      >
        取消
      </el-button>
      <el-button
        v-if="!confirming && !confirmResult"
        type="primary"
        :loading="qrLoading"
        @click="handleGenerateQr"
      >
        生成二维码
      </el-button>
      <el-button
        v-if="confirming"
        type="warning"
        @click="handleCancelConfirm"
      >
        取消确认
      </el-button>
    </template>
  </el-dialog>
</template>

<script>
import request from '@/utils/request'
import { Iphone, Clock, Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'

export default {
  name: 'BatchConfirm',
  props: {
    modelValue: {
      type: Boolean,
      default: false
    },
    operationType: {
      type: String,
      default: ''
    },
    affectedCount: {
      type: Number,
      default: 0
    },
    orderIds: {
      type: Array,
      default: () => []
    }
  },
  emits: ['update:modelValue', 'confirmed'],
  data() {
    return {
      Iphone,
      Clock,
      Loading,
      qrLoading: false,
      batchId: '',
      confirming: false,
      pollTimer: null,
      pollCountdown: 0,
      maxPollTime: 120,
      confirmResult: '',
      confirmResultMessage: ''
    }
  },
  computed: {
    operationTypeLabel() {
      const map = {
        batch_status_change: '批量状态变更',
        batch_refund: '批量退款',
        batch_cancel: '批量取消'
      }
      return map[this.operationType] || this.operationType
    },
    operationSummary() {
      return `即将对 ${this.affectedCount} 条订单执行「${this.operationTypeLabel}」操作，此操作需要另一管理员扫码确认。`
    },
    pollProgress() {
      if (this.maxPollTime <= 0) return 0
      return Math.round(((this.maxPollTime - this.pollCountdown) / this.maxPollTime) * 100)
    },
    countdownText() {
      const m = Math.floor(this.pollCountdown / 60)
      const s = this.pollCountdown % 60
      return `${m}分${s.toString().padStart(2, '0')}秒`
    }
  },
  watch: {
    modelValue(val) {
      if (val) {
        this.resetState()
      }
    }
  },
  beforeUnmount() {
    this.clearPollTimer()
  },
  methods: {
    resetState() {
      this.batchId = ''
      this.confirming = false
      this.confirmResult = ''
      this.confirmResultMessage = ''
      this.qrLoading = false
      this.clearPollTimer()
      this.pollCountdown = 0
    },
    clearPollTimer() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer)
        this.pollTimer = null
      }
    },
    async handleGenerateQr() {
      this.qrLoading = true
      try {
        const res = await request.post('/orders/batch', {
          operationType: this.operationType,
          orderIds: this.orderIds
        })
        this.batchId = res.data?.batchId || ''
        if (!this.batchId) {
          ElMessage.error('生成批次失败')
          return
        }
        await this.$nextTick()
        const canvas = this.$refs.qrCanvasRef
        if (canvas) {
          const confirmUrl = `${window.location.origin}/#/batch-confirm?batchId=${this.batchId}&operationType=${this.operationType}&count=${this.affectedCount}`
          await QRCode.toCanvas(canvas, confirmUrl, {
            width: 200,
            margin: 1,
            color: {
              dark: '#000000',
              light: '#ffffff'
            }
          })
        }
        this.startPolling()
      } catch (err) {
        console.error('生成批量操作批次失败:', err)
        ElMessage.error('生成批次失败，请重试')
      } finally {
        this.qrLoading = false
      }
    },
    startPolling() {
      this.confirming = true
      this.pollCountdown = this.maxPollTime
      this.pollTimer = setInterval(() => {
        this.pollCountdown--
        if (this.pollCountdown <= 0) {
          this.clearPollTimer()
          this.confirming = false
          this.confirmResult = 'error'
          this.confirmResultMessage = '确认超时，请重新发起操作'
          return
        }
        this.checkConfirmStatus()
      }, 2000)
      this.checkConfirmStatus()
    },
    async checkConfirmStatus() {
      try {
        const res = await request.get(`/orders/batch/confirm/${this.batchId}`)
        const status = res.data?.status
        if (status === 'confirmed') {
          this.clearPollTimer()
          this.confirming = false
          this.confirmResult = 'success'
          this.confirmResultMessage = res.data?.message || '操作已确认，正在执行中...'
          ElMessage.success('批量操作已确认')
        } else if (status === 'rejected') {
          this.clearPollTimer()
          this.confirming = false
          this.confirmResult = 'error'
          this.confirmResultMessage = res.data?.message || '操作被拒绝'
          ElMessage.error('批量操作被拒绝')
        }
      } catch (err) {
        console.error('查询确认状态失败:', err)
      }
    },
    handleCancel() {
      this.$emit('update:modelValue', false)
    },
    handleCancelConfirm() {
      this.clearPollTimer()
      this.confirming = false
      this.confirmResult = ''
      ElMessage.info('已取消确认等待')
    },
    handleConfirmDone() {
      this.$emit('confirmed')
      this.$emit('update:modelValue', false)
    },
    handleRetry() {
      this.confirmResult = ''
      this.confirmResultMessage = ''
      this.confirming = false
      this.handleGenerateQr()
    }
  }
}
</script>

<style lang="scss" scoped>
.batch-confirm-content {
  .batch-alert {
    margin-bottom: 16px;
  }

  .batch-info {
    margin-bottom: 20px;
  }

  .confirm-steps {
    .step-section {
      margin-bottom: 16px;
    }

    .step-desc {
      font-size: 13px;
      color: #606266;
      margin: 8px 0 16px;
      line-height: 1.6;
    }
  }

  .qrcode-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px;
    background: #f5f7fa;
    border-radius: 8px;
    margin-bottom: 12px;
    min-height: 220px;
    justify-content: center;

    .qrcode-canvas {
      width: 200px;
      height: 200px;
    }

    .qrcode-loading {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 8px;
      color: #909399;
      font-size: 13px;
    }
  }

  .batch-id-info {
    text-align: center;
    font-size: 13px;
    color: #606266;

    .label {
      margin-right: 4px;
    }
  }

  .waiting-status {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 30px 0;
    gap: 16px;

    .rotating-icon {
      animation: rotating 2s linear infinite;
      color: #409eff;
    }

    .waiting-text {
      font-size: 15px;
      color: #303133;
      font-weight: 500;
    }

    .waiting-remain {
      font-size: 13px;
      color: #909399;
    }
  }
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>