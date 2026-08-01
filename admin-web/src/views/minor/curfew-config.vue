<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">宵禁配置</span>
    </div>

    <div v-loading="loading">
      <el-card class="section-card">
        <template #header>
          <span class="card-header">宵禁时间设置</span>
        </template>
        <el-form ref="formRef" :model="formData" :rules="rules" label-width="160px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="宵禁开关" prop="curfewEnabled">
                <el-switch
                  v-model="formData.curfewEnabled"
                  active-text="开启"
                  inactive-text="关闭"
                />
                <div class="form-hint">关闭后宵禁功能将不再生效</div>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="宵禁开始时间" prop="curfewStartHour">
                <el-select v-model="formData.curfewStartHour" style="width: 100%">
                  <el-option v-for="h in 24" :key="h - 1" :label="(h - 1) + ':00'" :value="h - 1" />
                </el-select>
                <div class="form-hint">每天此时间点开始宵禁</div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="宵禁结束时间" prop="curfewEndHour">
                <el-select v-model="formData.curfewEndHour" style="width: 100%">
                  <el-option v-for="h in 24" :key="h - 1" :label="(h - 1) + ':00'" :value="h - 1" />
                </el-select>
                <div class="form-hint">每天此时间点结束宵禁</div>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <el-card class="section-card">
        <template #header>
          <span class="card-header">消费预警设置</span>
        </template>
        <el-form ref="formRef" :model="formData" :rules="rules" label-width="160px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="默认月消费限额" prop="minorMonthlyDefaultLimit">
                <el-input-number
                  v-model="formData.minorMonthlyDefaultLimit"
                  :min="0"
                  :step="100"
                  style="width: 100%"
                />
                <div class="form-hint">单位：分，即50000 = 500元</div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="80%预警阈值" prop="warning80Threshold">
                <el-input-number
                  v-model="formData.warning80Threshold"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  :precision="2"
                  style="width: 100%"
                />
                <div class="form-hint">达到月限额此比例时发送预警</div>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="100%限额阈值" prop="warning100Threshold">
                <el-input-number
                  v-model="formData.warning100Threshold"
                  :min="0"
                  :max="2"
                  :step="0.01"
                  :precision="2"
                  style="width: 100%"
                />
                <div class="form-hint">达到此比例时需监护人二次验证</div>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <el-card class="section-card" v-if="statsData">
        <template #header>
          <span class="card-header">近7日宵禁统计</span>
        </template>
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="stat-card">
              <div class="stat-value">{{ statsData.total_blocked }}</div>
              <div class="stat-label">总拦截次数</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-card">
              <div class="stat-value">{{ statsData.unique_users }}</div>
              <div class="stat-label">涉及用户数</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-card">
              <div class="stat-value">{{ actionStatsText }}</div>
              <div class="stat-label">主要拦截类型</div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <div class="save-container">
        <el-button type="primary" :icon="Check" :loading="saveLoading" size="large" @click="handleSave">
          保存配置
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Check } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'CurfewConfig',
  data() {
    return {
      Check,
      loading: false,
      saveLoading: false,
      formData: {
        curfewEnabled: 1,
        curfewStartHour: 22,
        curfewEndHour: 8,
        minorMonthlyDefaultLimit: 50000,
        warning80Threshold: 0.8,
        warning100Threshold: 1.0
      },
      statsData: null,
      rules: {
        curfewStartHour: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
        curfewEndHour: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
        minorMonthlyDefaultLimit: [{ required: true, message: '请输入默认限额', trigger: 'blur' }]
      }
    }
  },
  computed: {
    actionStatsText() {
      if (!this.statsData || !this.statsData.action_stats || this.statsData.action_stats.length === 0) {
        return '-'
      }
      const max = this.statsData.action_stats.reduce((a, b) => a.count > b.count ? a : b)
      const typeMap = {
        order: '下单',
        pay: '支付',
        reward: '打赏',
        join_group: '进群'
      }
      return `${typeMap[max.action_type] || max.action_type} (${max.count}次)`
    }
  },
  mounted() {
    this.fetchConfig()
    this.fetchStats()
  },
  methods: {
    async fetchConfig() {
      this.loading = true
      try {
        const res = await request.get('/minor/curfew_config')
        const data = res.data || {}
        this.formData = {
          curfewEnabled: data.curfew_enabled !== undefined ? Number(data.curfew_enabled) : 1,
          curfewStartHour: data.curfew_start_hour !== undefined ? Number(data.curfew_start_hour) : 22,
          curfewEndHour: data.curfew_end_hour !== undefined ? Number(data.curfew_end_hour) : 8,
          minorMonthlyDefaultLimit: data.minor_monthly_default_limit !== undefined ? Number(data.minor_monthly_default_limit) : 50000,
          warning80Threshold: data.warning_80_threshold !== undefined ? Number(data.warning_80_threshold) : 0.8,
          warning100Threshold: data.warning_100_threshold !== undefined ? Number(data.warning_100_threshold) : 1.0
        }
      } catch (err) {
        console.error('获取宵禁配置失败:', err)
        ElMessage.error('获取配置失败')
      } finally {
        this.loading = false
      }
    },
    async fetchStats() {
      try {
        const res = await request.get('/minor/curfew_stats', { params: { days: 7 } })
        this.statsData = res.data
      } catch (err) {
        console.error('获取宵禁统计失败:', err)
      }
    },
    async handleSave() {
      try {
        await ElMessageBox.confirm(
          '确定要保存宵禁配置吗？修改后将立即生效。',
          '保存确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        this.saveLoading = true
        await request.put('/minor/curfew_config', {
          curfew_enabled: this.formData.curfewEnabled,
          curfew_start_hour: this.formData.curfewStartHour,
          curfew_end_hour: this.formData.curfewEndHour,
          minor_monthly_default_limit: this.formData.minorMonthlyDefaultLimit,
          warning_80_threshold: this.formData.warning80Threshold,
          warning_100_threshold: this.formData.warning100Threshold
        })
        ElMessage.success('配置保存成功')
        this.fetchStats()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('保存配置失败:', err)
        }
      } finally {
        this.saveLoading = false
      }
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

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.save-container {
  text-align: center;
  padding: 20px 0;
}

.stat-card {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 24px;
  text-align: center;

  .stat-value {
    font-size: 28px;
    font-weight: 600;
    color: #409eff;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    color: #606266;
  }
}
</style>
