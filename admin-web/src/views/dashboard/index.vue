<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">仪表盘</span>
    </div>

    <!-- 核心指标卡 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6" v-for="item in statCards" :key="item.key">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-inner">
            <div class="stat-icon" :style="{ backgroundColor: item.bgColor }">
              <el-icon :size="28"><component :is="item.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">
                <span ref="countRefs">{{ animatedValues[item.key] }}</span>
              </div>
              <div class="stat-label">{{ item.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 30天趋势图 -->
      <el-col :xs="24" :md="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>近30天订单趋势</span>
            </div>
          </template>
          <div class="chart-container">
            <div v-if="trendData.length === 0" class="empty-tip">
              暂无趋势数据
            </div>
            <div v-else class="bar-chart">
              <div
                v-for="(item, index) in trendData"
                :key="index"
                class="bar-column"
                :title="item.date + '：' + item.count + '单'"
              >
                <div class="bar-value">{{ item.count }}</div>
                <div
                  class="bar"
                  :style="{
                    height: maxTrendCount > 0 ? (item.count / maxTrendCount * 100) + '%' : '0%'
                  }"
                ></div>
                <div class="bar-label">{{ item.date.slice(5) }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 待处理事项 -->
      <el-col :xs="24" :md="8">
        <el-card shadow="hover" class="pending-card">
          <template #header>
            <div class="card-header">
              <span>待处理事项</span>
              <el-badge :value="pendingTotal" :max="99" />
            </div>
          </template>
          <div class="pending-list">
            <div
              v-for="item in pendingItems"
              :key="item.key"
              class="pending-item"
              @click="handlePendingClick(item)"
            >
              <span class="red-dot" :class="{ active: item.count > 0 }"></span>
              <span class="pending-label">{{ item.label }}</span>
              <span class="pending-count" :class="{ urgent: item.count > 0 }">
                {{ item.count > 0 ? item.count + '条' : '无' }}
              </span>
              <el-icon class="pending-arrow"><ArrowRight /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作入口 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>快速操作</span>
            </div>
          </template>
          <div class="quick-actions">
            <el-button type="primary" :icon="Checked" @click="handleQuickAction('withdraw')">
              审核提现
            </el-button>
            <el-button type="warning" :icon="Warning" @click="handleQuickAction('appeal')">
              处理申诉
            </el-button>
            <el-button type="success" :icon="Document" @click="handleQuickAction('order')">
              查看订单
            </el-button>
            <el-button type="info" :icon="UserFilled" @click="handleQuickAction('user')">
              用户管理
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import request from '@/utils/request'
import { UserFilled, Money, ShoppingCart, Warning, ArrowRight, Checked, Document } from '@element-plus/icons-vue'

export default {
  name: 'Dashboard',
  components: {},
  data() {
    return {
      UserFilled,
      Money,
      ShoppingCart,
      Warning,
      ArrowRight,
      Checked,
      Document,
      statCards: [
        {
          key: 'totalUsers',
          label: '总用户数',
          icon: 'UserFilled',
          bgColor: '#e6f7ff',
          value: 0
        },
        {
          key: 'totalOrders',
          label: '总订单数',
          icon: 'Document',
          bgColor: '#f6ffed',
          value: 0
        },
        {
          key: 'todayRevenue',
          label: '今日营收',
          icon: 'Money',
          bgColor: '#fff7e6',
          value: 0
        },
        {
          key: 'pendingCount',
          label: '待处理事项',
          icon: 'Warning',
          bgColor: '#fff1f0',
          value: 0
        }
      ],
      animatedValues: {
        totalUsers: 0,
        totalOrders: 0,
        todayRevenue: 0,
        pendingCount: 0
      },
      targetValues: {
        totalUsers: 0,
        totalOrders: 0,
        todayRevenue: 0,
        pendingCount: 0
      },
      trendData: [],
      pendingItems: [
        { key: 'withdraw', label: '提现审核', count: 0, route: '/finance/withdraw' },
        { key: 'complaint', label: '投诉处理', count: 0, route: '/appeal/list' },
        { key: 'largeFail', label: '大额失败', count: 0, route: '/order/list' },
        { key: 'aiRisk', label: 'AI风险', count: 0, route: '/audit/player' },
        { key: 'appeal', label: '申诉处理', count: 0, route: '/appeal/list' }
      ],
      animationTimer: null
    }
  },
  computed: {
    pendingTotal() {
      return this.pendingItems.reduce((sum, item) => sum + item.count, 0)
    },
    maxTrendCount() {
      if (this.trendData.length === 0) return 1
      const max = Math.max(...this.trendData.map(item => item.count))
      return max || 1
    }
  },
  mounted() {
    this.fetchDashboardData()
  },
  beforeUnmount() {
    if (this.animationTimer) {
      clearInterval(this.animationTimer)
    }
  },
  methods: {
    async fetchDashboardData() {
      try {
        const res = await request.get('/dashboard/stats')
        const data = res.data || {}
        this.targetValues.totalUsers = data.totalUsers || 0
        this.targetValues.totalOrders = data.totalOrders || 0
        this.targetValues.todayRevenue = data.todayRevenue || 0
        this.targetValues.pendingCount = data.pendingCount || 0
        // 趋势数据由后端权威返回,失败/为空时显示空状态,前端不伪造 mock 数据
        this.trendData = data.trendData || []
        if (data.pendingItems) {
          this.pendingItems.forEach(item => {
            const updated = data.pendingItems[item.key]
            if (updated !== undefined) {
              item.count = updated
            }
          })
        }
        this.startCountAnimation()
      } catch (err) {
        console.error('获取仪表盘数据失败:', err)
        this.trendData = []
      }
    },
    startCountAnimation() {
      const duration = 1500
      const steps = 60
      const interval = duration / steps
      let step = 0
      if (this.animationTimer) {
        clearInterval(this.animationTimer)
      }
      const startValues = { ...this.animatedValues }
      this.animationTimer = setInterval(() => {
        step++
        const progress = step / steps
        const eased = 1 - Math.pow(1 - progress, 3)
        Object.keys(this.targetValues).forEach(key => {
          this.animatedValues[key] = Math.round(
            startValues[key] + (this.targetValues[key] - startValues[key]) * eased
          )
        })
        if (step >= steps) {
          Object.keys(this.targetValues).forEach(key => {
            this.animatedValues[key] = this.targetValues[key]
          })
          clearInterval(this.animationTimer)
          this.animationTimer = null
        }
      }, interval)
    },
    handlePendingClick(item) {
      if (item.route) {
        this.$router.push(item.route)
      }
    },
    handleQuickAction(action) {
      const routeMap = {
        withdraw: '/finance/withdraw',
        appeal: '/appeal/list',
        order: '/order/list',
        user: '/user/list'
      }
      const route = routeMap[action]
      if (route) {
        this.$router.push(route)
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.stat-row {
  margin-bottom: 0;
}

.stat-card {
  .stat-card-inner {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: #409eff;
  }

  .stat-info {
    flex: 1;
    min-width: 0;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;
    color: #303133;
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .stat-label {
    margin-top: 4px;
    font-size: 13px;
    color: #909399;
  }
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
}

.chart-container {
  height: 260px;
  padding: 10px 0;
}

.empty-tip {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 14px;
}

.bar-chart {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  height: 100%;
  padding: 0 4px;
  border-bottom: 1px solid #ebeef5;
}

.bar-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  justify-content: flex-end;
  cursor: pointer;

  .bar-value {
    font-size: 10px;
    color: #909399;
    margin-bottom: 2px;
    min-height: 14px;
  }

  .bar {
    width: 60%;
    max-width: 24px;
    min-height: 2px;
    background: linear-gradient(to top, #409eff, #79bbff);
    border-radius: 3px 3px 0 0;
    transition: height 0.6s ease;
  }

  .bar-label {
    font-size: 9px;
    color: #c0c4cc;
    margin-top: 6px;
    transform: rotate(-45deg);
    transform-origin: top left;
    white-space: nowrap;
  }

  &:hover .bar {
    background: linear-gradient(to top, #337ecc, #409eff);
  }
}

.pending-card {
  .pending-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .pending-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
      background-color: #f5f7fa;
    }
  }

  .red-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: #dcdfe6;
    flex-shrink: 0;
    margin-right: 10px;
    transition: background-color 0.3s;

    &.active {
      background-color: #f56c6c;
      animation: pulse 1.5s infinite;
    }
  }

  .pending-label {
    flex: 1;
    font-size: 14px;
    color: #303133;
  }

  .pending-count {
    font-size: 13px;
    color: #909399;
    margin-right: 8px;

    &.urgent {
      color: #f56c6c;
      font-weight: 600;
    }
  }

  .pending-arrow {
    color: #c0c4cc;
    font-size: 12px;
  }
}

.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(245, 108, 108, 0.4);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(245, 108, 108, 0);
  }
}

@media screen and (max-width: 768px) {
  .stat-value {
    font-size: 22px !important;
  }

  .stat-icon {
    width: 44px !important;
    height: 44px !important;
  }

  .chart-container {
    height: 200px;
  }

  .bar-column .bar-label {
    font-size: 7px;
  }
}
</style>