<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">系统配置</span>
    </div>

    <div v-loading="loading">
      <!-- 业务配置组 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">业务配置</span>
        </template>
        <el-form ref="businessFormRef" :model="formData.business" :rules="businessRules" label-width="160px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="平台抽成比例(%)" prop="commissionRate">
                <el-row style="width: 100%;" :gutter="12" align="middle">
                  <el-col :span="16">
                    <el-slider
                      v-model="formData.business.commissionRate"
                      :min="0"
                      :max="50"
                      :step="0.1"
                      :marks="{ 0: '0%', 10: '10%', 20: '20%', 30: '30%', 50: '50%' }"
                    />
                  </el-col>
                  <el-col :span="8">
                    <el-input-number
                      v-model="formData.business.commissionRate"
                      :min="0"
                      :max="50"
                      :precision="1"
                      style="width: 100%"
                    >
                      <template #suffix><span style="color:#909399;">%</span></template>
                    </el-input-number>
                  </el-col>
                </el-row>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="大额订单阈值(元)" prop="largeOrderThreshold">
                <el-input-number
                  v-model="formData.business.largeOrderThreshold"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入大额订单阈值"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="未成年人单笔限额(元)" prop="minorSingleLimit">
                <el-input-number
                  v-model="formData.business.minorSingleLimit"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="请输入未成年人单笔限额"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="月累计限额(元)" prop="minorMonthlyLimit">
                <el-input-number
                  v-model="formData.business.minorMonthlyLimit"
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

      <!-- 开关配置组 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">开关配置</span>
        </template>
        <el-form ref="switchFormRef" :model="formData.switch" label-width="160px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="实名认证开关">
                <el-switch
                  v-model="formData.switch.realNameAuth"
                  active-text="开启"
                  inactive-text="关闭"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="AI风控开关">
                <el-switch
                  v-model="formData.switch.aiRiskControl"
                  active-text="开启"
                  inactive-text="关闭"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="俱乐部入驻开关">
                <el-switch
                  v-model="formData.switch.clubJoinSwitch"
                  active-text="开启"
                  inactive-text="关闭"
                />
                <div class="switch-hint">关闭后前端同步隐藏所有俱乐部入驻入口，已入驻俱乐部不受影响</div>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="个人俱乐部保证金(元)">
                <el-input-number v-model="formData.switch.clubPersonalDeposit" :min="0" :step="100" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="企业俱乐部保证金(元)">
                <el-input-number v-model="formData.switch.clubEnterpriseDeposit" :min="0" :step="100" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="服务协议URL" prop="serviceAgreementUrl">
                <el-input
                  v-model="formData.switch.serviceAgreementUrl"
                  placeholder="请输入服务协议URL"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="隐私政策URL" prop="privacyPolicyUrl">
                <el-input
                  v-model="formData.switch.privacyPolicyUrl"
                  placeholder="请输入隐私政策URL"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- 客服配置组 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">客服配置</span>
        </template>
        <el-form ref="serviceFormRef" :model="formData.service" label-width="160px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="客服微信号" prop="wechatId">
                <el-input
                  v-model="formData.service.wechatId"
                  placeholder="请输入客服微信号"
                  maxlength="50"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- API版本配置组 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-header">API版本配置</span>
        </template>
        <el-form ref="apiFormRef" :model="formData.api" label-width="160px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="当前API版本号" prop="apiVersion">
                <el-input
                  v-model="formData.api.apiVersion"
                  placeholder="请输入API版本号，如 v1.0.0"
                  maxlength="30"
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
  name: 'SystemConfig',
  data() {
    return {
      Check,
      loading: false,
      saveLoading: false,
      formData: {
        business: {
          commissionRate: 0,
          largeOrderThreshold: null,
          minorSingleLimit: null,
          minorMonthlyLimit: null
        },
        switch: {
          realNameAuth: true,
          aiRiskControl: true,
          clubJoinSwitch: true,
          clubPersonalDeposit: 0,
          clubEnterpriseDeposit: 0,
          serviceAgreementUrl: '',
          privacyPolicyUrl: ''
        },
        service: {
          wechatId: ''
        },
        api: {
          apiVersion: ''
        }
      },
      businessRules: {
        commissionRate: [{ required: true, message: '请输入平台抽成比例', trigger: 'blur' }],
        largeOrderThreshold: [{ required: true, message: '请输入大额订单阈值', trigger: 'blur' }],
        minorSingleLimit: [{ required: true, message: '请输入未成年人单笔限额', trigger: 'blur' }],
        minorMonthlyLimit: [{ required: true, message: '请输入月累计限额', trigger: 'blur' }]
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
        const res = await request.get('/system/config')
        const data = res.data || {}
        if (data.business) {
          this.formData.business = { ...this.formData.business, ...data.business }
        }
        if (data.switch) {
          this.formData.switch = { ...this.formData.switch, ...data.switch }
        }
        if (data.service) {
          this.formData.service = { ...this.formData.service, ...data.service }
        }
        if (data.api) {
          this.formData.api = { ...this.formData.api, ...data.api }
        }
      } catch (err) {
        console.error('获取系统配置失败:', err)
      } finally {
        this.loading = false
      }
    },
    async handleSave() {
      const valid = await this.validateAllForms()
      if (!valid) return
      try {
        await ElMessageBox.confirm(
          '确定要保存系统配置吗？修改后将立即生效。',
          '保存确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        this.saveLoading = true
        await request.put('/system/config', this.formData)
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
      const refs = ['businessFormRef', 'switchFormRef', 'serviceFormRef', 'apiFormRef']
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

.switch-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

@media screen and (max-width: 768px) {
  :deep(.el-col-12) {
    max-width: 100%;
    flex: 0 0 100%;
  }
}
</style>