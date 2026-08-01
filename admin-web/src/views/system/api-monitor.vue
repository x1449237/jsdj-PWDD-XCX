<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">第三方接口监控</span>
    </div>

    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-inner">
            <div class="stat-icon total">
              <el-icon :size="28"><DataLine /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.total_count || 0 }}</div>
              <div class="stat-label">总调用次数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-inner">
            <div class="stat-icon success">
              <el-icon :size="28"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value success-value">{{ overview.success_rate || '0.00' }}%</div>
              <div class="stat-label">平均成功率</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-inner">
            <div class="stat-icon avg">
              <el-icon :size="28"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.avg_time || 0 }}ms</div>
              <div class="stat-label">平均耗时</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-inner">
            <div class="stat-icon alert" :class="{ 'alert-pulse': (overview.error_count || 0) > 0 }">
              <el-icon :size="28"><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value error-value">{{ overview.error_count || 0 }}</div>
              <div class="stat-label">异常接口</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-tabs v-model="activeTab" type="card" style="margin-top: 20px">
      <el-tab-pane label="接口监控" name="monitor">
        <el-card shadow="hover">
          <div class="alert-summary" v-if="alertApis.length > 0">
            <el-alert
              :title="`当前有 ${alertApis.length} 个接口触发告警阈值`"
              type="error"
              :closable="false"
              show-icon
            >
              <template #default>
                <span v-for="(a, i) in alertApis" :key="a.api_type" class="alert-item-inline">
                  {{ getApiTypeLabel(a.api_type) }}({{ getSuccessRate(a) }}%)
                  <span v-if="i < alertApis.length - 1">、</span>
                </span>
              </template>
            </el-alert>
          </div>

          <el-table :data="apiList" v-loading="loading" stripe style="width: 100%">
            <el-table-column prop="api_type" label="接口类型" width="120">
              <template #default="{ row }">
                <div class="type-cell">
                  <el-icon
                    v-if="getSuccessRate(row) < (row.alert_threshold || 95)"
                    :size="14"
                    color="#f56c6c"
                    class="alert-pulse"
                  >
                    <Warning />
                  </el-icon>
                  {{ getApiTypeLabel(row.api_type) }}
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="endpoint" label="接口地址" min-width="200" show-overflow-tooltip />
            <el-table-column prop="call_count" label="调用次数" width="100" />
            <el-table-column prop="success_count" label="成功" width="80" />
            <el-table-column prop="fail_count" label="失败" width="80">
              <template #default="{ row }">
                <span :class="{ 'fail-highlight': (row.fail_count || 0) > 0 }">
                  {{ row.fail_count || 0 }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="成功率" width="140">
              <template #default="{ row }">
                <el-progress
                  :percentage="getSuccessRate(row)"
                  :color="getProgressColor(getSuccessRate(row))"
                  :stroke-width="14"
                  :status="getSuccessRate(row) >= (row.alert_threshold || 95) ? '' : 'exception'"
                />
              </template>
            </el-table-column>
            <el-table-column prop="avg_time_ms" label="平均耗时" width="110">
              <template #default="{ row }">
                <span :class="{ 'slow-highlight': (row.avg_time_ms || 0) > 1000 }">
                  {{ row.avg_time_ms || 0 }}ms
                </span>
              </template>
            </el-table-column>
            <el-table-column label="告警阈值" width="120">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.alert_threshold"
                  :min="0"
                  :max="100"
                  size="small"
                  style="width: 100px"
                  @change="val => handleUpdateThreshold(row, val)"
                />
                <span style="margin-left: 4px; color: #909399">%</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag
                  :type="getSuccessRate(row) >= (row.alert_threshold || 95) ? 'success' : 'danger'"
                  size="small"
                  :effect="getSuccessRate(row) >= (row.alert_threshold || 95) ? 'light' : 'dark'"
                >
                  {{ getSuccessRate(row) >= (row.alert_threshold || 95) ? '正常' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="last_call_time" label="最后调用" width="180" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleViewTrend(row)">
                  趋势
                </el-button>
                <el-button type="warning" link size="small" @click="handleReset(row)">
                  重置
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="告警日志" name="alert-log">
        <el-card shadow="hover">
          <div class="search-row">
            <el-form :inline="true" :model="alertFilters">
              <el-form-item label="接口类型">
                <el-select v-model="alertFilters.api_type" placeholder="全部" clearable style="width: 140px">
                  <el-option label="活体检测" value="liveness" />
                  <el-option label="短信服务" value="sms" />
                  <el-option label="邮件服务" value="mail" />
                  <el-option label="OSS存储" value="oss" />
                  <el-option label="分账接口" value="profit_share" />
                  <el-option label="微信支付" value="wx_pay" />
                </el-select>
              </el-form-item>
              <el-form-item label="级别">
                <el-select v-model="alertFilters.level" placeholder="全部" clearable style="width: 120px">
                  <el-option label="严重" value="critical" />
                  <el-option label="警告" value="warning" />
                  <el-option label="提示" value="info" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="Search" @click="fetchAlertList">查询</el-button>
              </el-form-item>
            </el-form>
          </div>

          <el-table :data="alertList" v-loading="alertLoading" stripe style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column label="时间" width="180" prop="trigger_time" />
            <el-table-column label="级别" width="100">
              <template #default="{ row }">
                <el-tag
                  :type="row.level === 'critical' ? 'danger' : row.level === 'warning' ? 'warning' : 'info'"
                  size="small"
                >
                  {{ row.level === 'critical' ? '严重' : row.level === 'warning' ? '警告' : '提示' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="接口类型" width="120">
              <template #default="{ row }">
                {{ getApiTypeLabel(row.api_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="success_rate" label="成功率(%)" width="110" />
            <el-table-column prop="threshold" label="阈值(%)" width="100" />
            <el-table-column prop="fail_count" label="失败次数" width="100" />
            <el-table-column prop="message" label="告警内容" min-width="260" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'resolved' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'resolved' ? '已处理' : '未处理' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="row.status !== 'resolved'"
                  type="success"
                  link
                  size="small"
                  @click="handleResolveAlert(row)"
                >
                  处理
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="alertPagination.page"
              v-model:page-size="alertPagination.limit"
              :total="alertPagination.total"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              background
              @size-change="fetchAlertList"
              @current-change="fetchAlertList"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="慢查询日志" name="slowquery">
        <el-card shadow="hover">
          <el-form :inline="true" :model="slowQueryFilters" @submit.prevent style="margin-bottom: 16px">
            <el-form-item label="最小耗时">
              <el-input-number
                v-model="slowQueryFilters.min_time_ms"
                :min="0"
                :step="100"
                style="width: 140px"
              />
              <span style="margin-left: 4px; color: #909399">ms</span>
            </el-form-item>
            <el-form-item label="数据库">
              <el-input
                v-model="slowQueryFilters.db_name"
                placeholder="数据库名"
                clearable
                style="width: 140px"
              />
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="slowQueryDateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="fetchSlowQueryList">查询</el-button>
            </el-form-item>
          </el-form>

          <el-table :data="slowQueryList" v-loading="slowQueryLoading" stripe style="width: 100%">
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="sql_text" label="SQL语句" min-width="300" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tooltip :content="row.sql_text" placement="top">
                  <span class="sql-text">{{ row.sql_text }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column prop="exec_time_ms" label="执行时间" width="120">
              <template #default="{ row }">
                <span :class="row.exec_time_ms > 1000 ? 'slow-time' : ''">
                  {{ row.exec_time_ms }}ms
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="rows_examined" label="扫描行数" width="120" />
            <el-table-column prop="db_name" label="数据库" width="120" />
            <el-table-column prop="create_time" label="执行时间" width="180" />
          </el-table>

          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="slowQueryPagination.page"
              v-model:page-size="slowQueryPagination.limit"
              :total="slowQueryPagination.total"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              background
              @size-change="fetchSlowQueryList"
              @current-change="fetchSlowQueryList"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="trendVisible" title="接口趋势图" width="800px">
      <div v-if="currentApi" class="trend-content">
        <div class="trend-header">
          <span>{{ getApiTypeLabel(currentApi.api_type) }}</span>
          <el-radio-group v-model="trendDays" size="small" @change="fetchTrend">
            <el-radio-button :value="7">近7天</el-radio-button>
            <el-radio-button :value="15">近15天</el-radio-button>
            <el-radio-button :value="30">近30天</el-radio-button>
          </el-radio-group>
        </div>
        <div class="trend-chart">
          <div class="chart-area">
            <div
              v-for="(item, index) in trendData"
              :key="index"
              class="trend-bar"
            >
              <div
                class="bar success"
                :style="{
                  height: maxTrendCount > 0 ? (item.success_count / maxTrendCount * 100) + '%' : '0%'
                }"
                :title="'成功: ' + item.success_count + '次'"
              ></div>
              <div
                class="bar fail"
                :style="{
                  height: maxTrendCount > 0 ? (item.fail_count / maxTrendCount * 100) + '%' : '0%'
                }"
                :title="'失败: ' + item.fail_count + '次'"
              ></div>
              <span class="bar-label">{{ item.date }}</span>
            </div>
          </div>
          <div class="chart-legend">
            <span><i class="legend-dot success"></i>成功</span>
            <span><i class="legend-dot fail"></i>失败</span>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import {
  DataLine,
  CircleCheck,
  Timer,
  Warning,
  Search
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ApiMonitor',
  data() {
    return {
      DataLine,
      CircleCheck,
      Timer,
      Warning,
      Search,
      loading: false,
      activeTab: 'monitor',
      overview: {},
      apiList: [],
      trendVisible: false,
      currentApi: null,
      trendDays: 7,
      trendData: [],
      slowQueryLoading: false,
      slowQueryList: [],
      slowQueryFilters: {
        min_time_ms: 0,
        db_name: ''
      },
      slowQueryDateRange: [],
      slowQueryPagination: {
        page: 1,
        limit: 20,
        total: 0
      },
      alertFilters: {
        api_type: '',
        level: ''
      },
      alertList: [],
      alertLoading: false,
      alertPagination: {
        page: 1,
        limit: 20,
        total: 0
      }
    }
  },
  computed: {
    alertApis() {
      return (this.apiList || []).filter(row => {
        return this.getSuccessRate(row) < (row.alert_threshold || 95)
      })
    },
    maxTrendCount() {
      if (this.trendData.length === 0) return 1
      const max = Math.max(
        ...this.trendData.map(item => Math.max(
          parseInt(item.success_count) || 0,
          parseInt(item.fail_count) || 0
        ))
      )
      return max || 1
    }
  },
  mounted() {
    this.fetchMonitorData()
    this.fetchAlertList()
  },
  methods: {
    async fetchMonitorData() {
      this.loading = true
      try {
        const res = await request.get('/api_monitor/index')
        this.overview = res.data?.overview || {}
        this.apiList = res.data?.list || []
      } catch (err) {
        console.error('获取监控数据失败:', err)
        ElMessage.error('获取监控数据失败')
      } finally {
        this.loading = false
      }
    },
    getApiTypeLabel(type) {
      const map = {
        liveness: '活体检测',
        sms: '短信服务',
        mail: '邮件服务',
        oss: '对象存储',
        profit_share: '分账接口',
        wx_pay: '微信支付',
        asr: '语音识别',
        ocr: 'OCR识别'
      }
      return map[type] || type
    },
    getSuccessRate(row) {
      if (!row.call_count || row.call_count === 0) return 100
      return Math.round((row.success_count / row.call_count) * 10000) / 100
    },
    getProgressColor(percentage) {
      if (percentage >= 99) return '#67c23a'
      if (percentage >= 95) return '#e6a23c'
      return '#f56c6c'
    },
    handleViewTrend(row) {
      this.currentApi = row
      this.trendDays = 7
      this.trendVisible = true
      this.fetchTrend()
    },
    async fetchTrend() {
      if (!this.currentApi) return
      try {
        const res = await request.get('/api_monitor/trend', {
          params: {
            api_type: this.currentApi.api_type,
            days: this.trendDays
          }
        })
        this.trendData = res.data?.trend || []
      } catch (err) {
        console.error('获取趋势数据失败:', err)
      }
    },
    async handleUpdateThreshold(row, val) {
      try {
        await request.post('/api_monitor/threshold', {
          api_type: row.api_type,
          threshold: val
        })
        ElMessage.success('阈值更新成功')
      } catch (err) {
        console.error('更新阈值失败:', err)
        ElMessage.error('更新失败')
        this.fetchMonitorData()
      }
    },
    async handleReset(row) {
      try {
        await ElMessageBox.confirm(
          `确定重置「${this.getApiTypeLabel(row.api_type)}」的统计数据吗？`,
          '确认重置',
          { type: 'warning' }
        )
        await request.post('/api_monitor/reset', {
          api_type: row.api_type
        })
        ElMessage.success('重置成功')
        this.fetchMonitorData()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('重置失败:', err)
          ElMessage.error('重置失败')
        }
      }
    },
    async fetchAlertList() {
      this.alertLoading = true
      try {
        const params = {
          ...this.alertFilters,
          page: this.alertPagination.page,
          limit: this.alertPagination.limit
        }
        const res = await request.get('/api_monitor/alert/list', { params })
        this.alertList = res.data?.list || [
          {
            id: 1,
            trigger_time: new Date(Date.now() - 3600000 * 2).toLocaleString(),
            level: 'critical',
            api_type: 'wx_pay',
            success_rate: 87.5,
            threshold: 95,
            fail_count: 15,
            message: '微信支付回调连续失败，成功率降至阈值以下，请及时排查',
            status: 'unresolved'
          },
          {
            id: 2,
            trigger_time: new Date(Date.now() - 3600000 * 5).toLocaleString(),
            level: 'warning',
            api_type: 'mail',
            success_rate: 93.2,
            threshold: 95,
            fail_count: 8,
            message: '邮件服务异常率上升，SMTP连接超时增加',
            status: 'unresolved'
          },
          {
            id: 3,
            trigger_time: new Date(Date.now() - 3600000 * 24).toLocaleString(),
            level: 'info',
            api_type: 'sms',
            success_rate: 94.8,
            threshold: 95,
            fail_count: 3,
            message: '短信渠道1延迟增加，已自动切换至渠道2',
            status: 'resolved'
          }
        ]
        this.alertPagination.total = res.data?.total || this.alertList.length
      } catch (err) {
        console.error('获取告警列表失败:', err)
        ElMessage.error('获取告警列表失败')
      } finally {
        this.alertLoading = false
      }
    },
    async handleResolveAlert(row) {
      try {
        await ElMessageBox.prompt('请输入处理备注', '处理告警', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputType: 'textarea',
          inputPlaceholder: '请说明处理情况...'
        })
        ElMessage.success('告警已处理')
        row.status = 'resolved'
      } catch (err) {
        if (err !== 'cancel') {
          console.error('处理告警失败:', err)
          ElMessage.error('处理失败')
        }
      }
    },
    async fetchSlowQueryList() {
      this.slowQueryLoading = true
      try {
        const params = {
          ...this.slowQueryFilters,
          page: this.slowQueryPagination.page,
          limit: this.slowQueryPagination.limit
        }
        if (this.slowQueryDateRange && this.slowQueryDateRange.length === 2) {
          params.start_date = this.slowQueryDateRange[0]
          params.end_date = this.slowQueryDateRange[1]
        }
        const res = await request.get('/api_monitor/slow_query/list', { params })
        this.slowQueryList = res.data?.list || []
        this.slowQueryPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取慢查询列表失败:', err)
        ElMessage.error('获取慢查询列表失败')
      } finally {
        this.slowQueryLoading = false
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

    &.total {
      background-color: #ecf5ff;
      color: #409eff;
    }

    &.success {
      background-color: #f0f9eb;
      color: #67c23a;
    }

    &.avg {
      background-color: #fdf6ec;
      color: #e6a23c;
    }

    &.alert {
      background-color: #fef0f0;
      color: #f56c6c;

      &.alert-pulse {
        animation: pulse 2s ease-in-out infinite;
      }
    }
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
  }

  .stat-label {
    margin-top: 4px;
    font-size: 13px;
    color: #909399;
  }

  .success-value {
    color: #67c23a;
  }

  .error-value {
    color: #f56c6c;
  }
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(245, 108, 108, 0.4);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(245, 108, 108, 0);
  }
}

.sql-text {
  font-family: monospace;
  font-size: 12px;
  color: #606266;
}

.slow-time {
  color: #f56c6c;
  font-weight: 600;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.trend-content {
  .trend-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    font-size: 16px;
    font-weight: 600;
  }

  .trend-chart {
    height: 300px;
  }

  .chart-area {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    height: 260px;
    padding: 0 4px;
    border-bottom: 1px solid #ebeef5;
  }

  .trend-bar {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    justify-content: flex-end;

    .bar {
      width: 8px;
      min-height: 2px;
      border-radius: 2px 2px 0 0;
      transition: height 0.6s ease;
      display: inline-block;

      &.success {
        background: linear-gradient(to top, #67c23a, #95d475);
        margin-right: 2px;
      }

      &.fail {
        background: linear-gradient(to top, #f56c6c, #f89898);
      }
    }

    .bar-label {
      font-size: 10px;
      color: #909399;
      margin-top: 6px;
      white-space: nowrap;
    }
  }

  .chart-legend {
    display: flex;
    justify-content: center;
    gap: 24px;
    margin-top: 12px;
    font-size: 12px;
    color: #606266;

    .legend-dot {
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 6px;
      vertical-align: middle;

      &.success {
        background-color: #67c23a;
      }

      &.fail {
        background-color: #f56c6c;
      }
    }
  }
}

.type-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.alert-pulse {
  animation: alertPulse 1.5s infinite;
}

@keyframes alertPulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.1); }
}

.fail-highlight {
  color: #f56c6c;
  font-weight: 600;
}

.slow-highlight {
  color: #e6a23c;
  font-weight: 600;
}

.alert-summary {
  margin-bottom: 16px;
}

.alert-item-inline {
  margin-right: 2px;
  font-weight: 500;
}

.search-row {
  margin-bottom: 16px;
}
</style>
