<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">财务配置</span>
    </div>

    <div v-loading="loading">
      <!-- 提现配置 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">提现配置</span>
        </template>
        <el-form ref="withdrawFormRef" :model="formData.withdraw" :rules="withdrawRules" label-width="140px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="最低提现金额(元)" prop="minAmount">
                <el-input-number
                  v-model="formData.withdraw.minAmount"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入最低提现金额"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="最高提现金额(元)" prop="maxAmount">
                <el-input-number
                  v-model="formData.withdraw.maxAmount"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入最高提现金额"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="手续费率(%)" prop="feeRate">
                <el-input-number
                  v-model="formData.withdraw.feeRate"
                  :min="0"
                  :max="100"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入手续费率"
                >
                  <template #suffix>
                    <span style="color: #909399;">%</span>
                  </template>
                </el-input-number>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="冻结天数(T+N)" prop="freezeDays">
                <el-input-number
                  v-model="formData.withdraw.freezeDays"
                  :min="0"
                  :max="365"
                  style="width: 100%"
                  placeholder="请输入冻结天数"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="提现间隔(小时)" prop="withdrawInterval">
                <el-input-number
                  v-model="formData.withdraw.withdrawInterval"
                  :min="0"
                  style="width: 100%"
                  placeholder="请输入提现间隔"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- 平台抽成配置 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">平台抽成配置</span>
        </template>
        <el-form ref="commissionFormRef" :model="formData.commission" :rules="commissionRules" label-width="140px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="抽成比例(%)" prop="commissionRate">
                <el-row style="width: 100%;" :gutter="12" align="middle">
                  <el-col :span="16">
                    <el-slider
                      v-model="formData.commission.commissionRate"
                      :min="0"
                      :max="50"
                      :step="0.1"
                      :marks="commissionMarks"
                      show-input
                    />
                  </el-col>
                  <el-col :span="8">
                    <el-input-number
                      v-model="formData.commission.commissionRate"
                      :min="0"
                      :max="50"
                      :precision="1"
                      style="width: 100%"
                      placeholder="抽成比例"
                    >
                      <template #suffix>
                        <span style="color: #909399;">%</span>
                      </template>
                    </el-input-number>
                  </el-col>
                </el-row>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- 大额订单阈值配置 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">大额订单阈值配置</span>
        </template>
        <el-form ref="largeOrderFormRef" :model="formData.largeOrder" :rules="largeOrderRules" label-width="140px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="大额订单阈值(元)" prop="largeOrderThreshold">
                <el-input-number
                  v-model="formData.largeOrder.largeOrderThreshold"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入大额订单阈值"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- 未成年人限额配置 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">未成年人限额配置</span>
        </template>
        <el-form ref="minorLimitFormRef" :model="formData.minorLimit" :rules="minorLimitRules" label-width="140px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="单笔限额(元)" prop="singleLimit">
                <el-input-number
                  v-model="formData.minorLimit.singleLimit"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入单笔限额"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="月累计限额(元)" prop="monthlyLimit">
                <el-input-number
                  v-model="formData.minorLimit.monthlyLimit"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入月累计限额"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- 保存按钮 -->
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
  name: 'FinanceConfig',
  data() {
    return {
      Check,
      loading: false,
      saveLoading: false,
      formData: {
        withdraw: {
          minAmount: null,
          maxAmount: null,
          feeRate: null,
          freezeDays: null,
          withdrawInterval: null
        },
        commission: {
          commissionRate: null
        },
        largeOrder: {
          largeOrderThreshold: null
        },
        minorLimit: {
          singleLimit: null,
          monthlyLimit: null
        }
      },
      withdrawRules: {
        minAmount: [{ required: true, message: '请输入最低提现金额', trigger: 'blur' }],
        maxAmount: [{ required: true, message: '请输入最高提现金额', trigger: 'blur' }],
        feeRate: [{ required: true, message: '请输入手续费率', trigger: 'blur' }],
        freezeDays: [{ required: true, message: '请输入冻结天数', trigger: 'blur' }],
        withdrawInterval: [{ required: true, message: '请输入提现间隔', trigger: 'blur' }]
      },
      commissionRules: {
        commissionRate: [{ required: true, message: '请输入抽成比例', trigger: 'blur' }]
      },
      largeOrderRules: {
        largeOrderThreshold: [{ required: true, message: '请输入大额订单阈值', trigger: 'blur' }]
      },
      minorLimitRules: {
        singleLimit: [{ required: true, message: '请输入单笔限额', trigger: 'blur' }],
        monthlyLimit: [{ required: true, message: '请输入月累计限额', trigger: 'blur' }]
      },
      commissionMarks: {
        0: '0%',
        10: '10%',
        20: '20%',
        30: '30%',
        50: '50%'
      }
    }
  },
  mounted() {
    this.fetchConfig()
  },
  methods: {
    async fetchConfig() {
      this.loading = true
      try {
        const res = await request.get('/finance/config')
        const data = res.data || {}
        if (data.withdraw) {
          this.formData.withdraw = { ...this.formData.withdraw, ...data.withdraw }
        }
        if (data.commission) {
          this.formData.commission = { ...this.formData.commission, ...data.commission }
        }
        if (data.largeOrder) {
          this.formData.largeOrder = { ...this.formData.largeOrder, ...data.largeOrder }
        }
        if (data.minorLimit) {
          this.formData.minorLimit = { ...this.formData.minorLimit, ...data.minorLimit }
        }
      } catch (err) {
        console.error('获取财务配置失败:', err)
      } finally {
        this.loading = false
      }
    },
    async handleSave() {
      const valid = await this.validateAllForms()
      if (!valid) return
      try {
        await ElMessageBox.confirm(
          '确定要保存财务配置吗？修改后将立即生效。',
          '保存确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        this.saveLoading = true
        await request.put('/finance/config', {
          withdraw: this.formData.withdraw,
          commission: this.formData.commission,
          largeOrder: this.formData.largeOrder,
          minorLimit: this.formData.minorLimit
        })
        ElMessage.success('配置保存成功')
      } catch (err) {
        if (err !== 'cancel') {
          console.error('保存配置失败:', err)
        }
      } finally {
        this.saveLoading = false
      }
    },
    async validateAllForms() {
      const refs = ['withdrawFormRef', 'commissionFormRef', 'largeOrderFormRef', 'minorLimitFormRef']
      for (const ref of refs) {
        if (this.$refs[ref]) {
          const valid = await this.$refs[ref].validate().catch(() => false)
          if (!valid) return false
        }
      }
      return true
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

.save-container {
  text-align: center;
  padding: 20px 0;
}

@media screen and (max-width: 768px) {
  :deep(.el-col-12) {
    max-width: 100%;
    flex: 0 0 100%;
  }
}
</style>