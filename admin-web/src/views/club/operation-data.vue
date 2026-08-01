<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">俱乐部运营数据看板</span>
    </div>

    <el-card class="stat-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="stat-card primary">
            <div class="stat-label">俱乐部总数</div>
            <div class="stat-value">{{ stats.total_clubs || 0 }}</div>
            <div class="stat-sub">新增：{{ stats.new_clubs || 0 }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card success">
            <div class="stat-label">成员总数</div>
            <div class="stat-value">{{ stats.total_members || 0 }}</div>
            <div class="stat-sub">
              创始人{{ stats.member_by_role?.founder || 0 }} /
              管理{{ stats.member_by_role?.manager || 0 }}
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card warning">
            <div class="stat-label">内部订单数</div>
            <div class="stat-value">{{ stats.total_orders || 0 }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card danger">
            <div class="stat-label">总营收</div>
            <div class="stat-value">¥{{ stats.total_revenue_yuan || '0.00' }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card>
          <div class="card-header">
            <span class="card-title">趋势数据</span>
            <div class="card-tools">
              <el-select v-model="selectedDays" style="width: 120px" @change="fetchDashboard">
                <el-option label="近7天" :value="7" />
                <el-option label="近30天" :value="30" />
                <el-option label="近90天" :value="90" />
              </el-select>
            </div>
          </div>
          <div v-if="trendData.length > 0" class="chart-wrap">
            <div class="chart-bar" v-for="(item, idx) in trendData" :key="idx">
              <div class="bar-label">{{ item.date }}</div>
              <div class="bar-group">
                <div class="bar-item">
                  <div class="bar-title">订单</div>
                  <div class="bar-bg">
                    <div class="bar-fill blue" :style="{ height: getBarHeight(item.order_count, 'order') + '%' }"></div>
                  </div>
                  <div class="bar-value">{{ item.order_count }}</div>
                </div>
                <div class="bar-item">
                  <div class="bar-title">营收</div>
                  <div class="bar-bg">
                    <div class="bar-fill green" :style="{ height: getBarHeight(item.total_revenue_yuan, 'revenue') + '%' }"></div>
                  </div>
                  <div class="bar-value">¥{{ item.total_revenue_yuan }}</div>
                </div>
                <div class="bar-item">
                  <div class="bar-title">新成员</div>
                  <div class="bar-bg">
                    <div class="bar-fill orange" :style="{ height: getBarHeight(item.new_member_count, 'member') + '%' }"></div>
                  </div>
                  <div class="bar-value">{{ item.new_member_count }}</div>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无数据" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const stats = reactive({
  total_clubs: 0,
  total_members: 0,
  total_orders: 0,
  total_revenue_yuan: '0.00',
  new_clubs: 0,
  member_by_role: { founder: 0, manager: 0, member: 0 },
})

const selectedDays = ref(30)
const trendData = ref([])

const maxValues = computed(() => {
  const maxOrder = Math.max(...trendData.value.map(d => d.order_count || 0), 1)
  const maxRev = Math.max(...trendData.value.map(d => parseFloat(d.total_revenue_yuan) || 0), 1)
  const maxMem = Math.max(...trendData.value.map(d => d.new_member_count || 0), 1)
  return { maxOrder, maxRev, maxMem }
})

const getBarHeight = (value, type) => {
  const v = type === 'revenue' ? parseFloat(value) : (value || 0)
  const max = type === 'order' ? maxValues.value.maxOrder
    : type === 'revenue' ? maxValues.value.maxRev
    : maxValues.value.maxMem
  return Math.min(100, (v / max) * 100)
}

const fetchDashboard = async () => {
  try {
    const res = await request.get('/club/operation_data', { days: selectedDays.value })
    Object.assign(stats, res)
    trendData.value = res.trend || []
  } catch (e) {
    ElMessage.error(e.message || '获取数据失败')
  }
}

onMounted(() => {
  fetchDashboard()
})
</script>

<style scoped>
.stat-cards {
  background: transparent;
  border: none;
  padding: 0;
}
.stat-cards :deep(.el-card__body) {
  padding: 0;
}
.stat-card {
  padding: 20px;
  border-radius: 8px;
  color: #fff;
}
.stat-card.primary { background: linear-gradient(135deg, #409EFF, #66b1ff); }
.stat-card.success { background: linear-gradient(135deg, #67C23A, #85ce61); }
.stat-card.warning { background: linear-gradient(135deg, #E6A23C, #ebb563); }
.stat-card.danger  { background: linear-gradient(135deg, #F56C6C, #f78989); }
.stat-label { font-size: 14px; opacity: 0.9; }
.stat-value { font-size: 28px; font-weight: bold; margin: 8px 0; }
.stat-sub   { font-size: 12px; opacity: 0.85; }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.card-title { font-size: 16px; font-weight: 500; }

.chart-wrap {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding: 10px 0;
  min-height: 300px;
}
.chart-bar {
  flex: 0 0 auto;
  width: 80px;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.bar-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}
.bar-group {
  display: flex;
  gap: 6px;
  height: 200px;
  align-items: flex-end;
}
.bar-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 20px;
}
.bar-title {
  font-size: 11px;
  color: #909399;
  margin-bottom: 4px;
}
.bar-bg {
  width: 100%;
  height: 160px;
  background: #f0f2f5;
  border-radius: 4px 4px 0 0;
  display: flex;
  align-items: flex-end;
}
.bar-fill {
  width: 100%;
  border-radius: 4px 4px 0 0;
  transition: height 0.3s;
}
.bar-fill.blue { background: #409EFF; }
.bar-fill.green { background: #67C23A; }
.bar-fill.orange { background: #E6A23C; }
.bar-value {
  font-size: 11px;
  color: #606266;
  margin-top: 4px;
}
</style>
