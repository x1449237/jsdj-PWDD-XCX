<template>
  <div class="init-container">
    <div class="init-box">
      <div class="init-header">
        <h2 class="init-title">首次登录初始化</h2>
        <p class="init-subtitle">首次登录需要完成以下步骤以保障账户安全</p>
      </div>

      <el-steps :active="activeStep" align-center class="init-steps">
        <el-step title="修改密码" />
        <el-step title="绑定邮箱" />
        <el-step title="完成" />
      </el-steps>

      <div class="init-content">
        <!-- 步骤1：修改密码 -->
        <div v-show="activeStep === 0">
          <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="0" size="large">
            <el-form-item prop="newPassword">
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="请输入新密码"
                show-password
                :prefix-icon="Lock"
              />
            </el-form-item>
            <div class="password-hint">
              <p>密码要求：</p>
              <ul>
                <li :class="{ met: passwordChecks.length }">长度至少8位</li>
                <li :class="{ met: passwordChecks.hasUpper }">包含大写字母</li>
                <li :class="{ met: passwordChecks.hasLower }">包含小写字母</li>
                <li :class="{ met: passwordChecks.hasNumber }">包含数字</li>
                <li :class="{ met: passwordChecks.hasSpecial }">包含特殊字符</li>
                <li :class="{ met: passwordChecks.typesMet >= 3 }">
                  以上至少满足三种（当前：{{ passwordChecks.typesMet }}种）
                </li>
              </ul>
            </div>
            <el-form-item prop="confirmPassword">
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                placeholder="请确认新密码"
                show-password
                :prefix-icon="Lock"
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="passwordLoading"
                class="step-btn"
                @click="handleStep1"
              >
                下一步
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <!-- 步骤2：绑定邮箱 -->
        <div v-show="activeStep === 1">
          <el-form ref="emailFormRef" :model="emailForm" :rules="emailRules" label-width="0" size="large">
            <el-form-item prop="email">
              <el-input
                v-model="emailForm.email"
                placeholder="请输入邮箱地址"
                :prefix-icon="Message"
              />
            </el-form-item>
            <el-form-item prop="verifyCode">
              <el-row style="width: 100%;" :gutter="12">
                <el-col :span="15">
                  <el-input
                    v-model="emailForm.verifyCode"
                    placeholder="请输入验证码"
                    :prefix-icon="Key"
                  />
                </el-col>
                <el-col :span="9">
                  <el-button
                    :disabled="sendCountdown > 0"
                    :loading="sendLoading"
                    class="send-btn"
                    @click="handleSendCode"
                  >
                    {{ sendCountdown > 0 ? sendCountdown + 's后重发' : '发送验证码' }}
                  </el-button>
                </el-col>
              </el-row>
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="emailLoading"
                class="step-btn"
                @click="handleStep2"
              >
                下一步
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <!-- 步骤3：完成 -->
        <div v-show="activeStep === 2" class="complete-section">
          <el-result
            icon="success"
            title="初始化完成"
            sub-title="密码修改和邮箱绑定已成功完成，现在可以正常使用系统了"
          >
            <template #extra>
              <el-button type="primary" @click="handleFinish">进入系统</el-button>
            </template>
          </el-result>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Lock, Message, Key } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

export default {
  name: 'Init',
  data() {
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.passwordForm.newPassword) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }
    const validatePassword = (rule, value, callback) => {
      if (!value) {
        callback(new Error('请输入新密码'))
        return
      }
      if (value.length < 8) {
        callback(new Error('密码长度至少8位'))
        return
      }
      const checks = this.calcPasswordChecks(value)
      if (checks.typesMet < 3) {
        callback(new Error('密码需包含大写+小写+数字+特殊字符中至少三种'))
        return
      }
      callback()
    }
    return {
      Lock,
      Message,
      Key,
      activeStep: 0,
      passwordForm: {
        newPassword: '',
        confirmPassword: ''
      },
      passwordRules: {
        newPassword: [
          { required: true, validator: validatePassword, trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请确认密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ]
      },
      passwordLoading: false,
      emailForm: {
        email: '',
        verifyCode: ''
      },
      emailRules: {
        email: [
          { required: true, message: '请输入邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
        ],
        verifyCode: [
          { required: true, message: '请输入验证码', trigger: 'blur' }
        ]
      },
      emailLoading: false,
      sendLoading: false,
      sendCountdown: 0,
      sendTimer: null
    }
  },
  computed: {
    passwordChecks() {
      return this.calcPasswordChecks(this.passwordForm.newPassword)
    }
  },
  beforeUnmount() {
    if (this.sendTimer) {
      clearInterval(this.sendTimer)
    }
  },
  methods: {
    calcPasswordChecks(password) {
      const hasUpper = /[A-Z]/.test(password)
      const hasLower = /[a-z]/.test(password)
      const hasNumber = /[0-9]/.test(password)
      const hasSpecial = /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?`~]/.test(password)
      const typesMet = [hasUpper, hasLower, hasNumber, hasSpecial].filter(Boolean).length
      return {
        length: password.length >= 8,
        hasUpper,
        hasLower,
        hasNumber,
        hasSpecial,
        typesMet
      }
    },
    async handleStep1() {
      const valid = await this.$refs.passwordFormRef.validate().catch(() => false)
      if (!valid) return
      this.passwordLoading = true
      try {
        await request.post('/init/change-password', {
          newPassword: this.passwordForm.newPassword
        })
        ElMessage.success('密码修改成功')
        this.activeStep = 1
      } catch (err) {
        console.error('修改密码失败:', err)
      } finally {
        this.passwordLoading = false
      }
    },
    async handleSendCode() {
      const valid = await this.$refs.emailFormRef.validateField('email').catch(() => false)
      if (!valid) return
      this.sendLoading = true
      try {
        await request.post('/init/send-verify-code', {
          email: this.emailForm.email
        })
        ElMessage.success('验证码已发送')
        this.sendCountdown = 60
        this.sendTimer = setInterval(() => {
          this.sendCountdown--
          if (this.sendCountdown <= 0) {
            clearInterval(this.sendTimer)
            this.sendTimer = null
          }
        }, 1000)
      } catch (err) {
        console.error('发送验证码失败:', err)
      } finally {
        this.sendLoading = false
      }
    },
    async handleStep2() {
      const valid = await this.$refs.emailFormRef.validate().catch(() => false)
      if (!valid) return
      this.emailLoading = true
      try {
        await request.post('/init/verify-email', {
          email: this.emailForm.email,
          verifyCode: this.emailForm.verifyCode
        })
        ElMessage.success('邮箱验证成功')
        this.activeStep = 2
      } catch (err) {
        console.error('邮箱验证失败:', err)
      } finally {
        this.emailLoading = false
      }
    },
    async handleFinish() {
      const authStore = useAuthStore()
      await authStore.finishInit()
      this.$router.push('/dashboard')
    }
  }
}
</script>

<style lang="scss" scoped>
.init-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  padding: 20px;
}

.init-box {
  width: 480px;
  max-width: 100%;
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.2);
}

.init-header {
  text-align: center;
  margin-bottom: 32px;
}

.init-title {
  font-size: 22px;
  font-weight: 700;
  color: #303133;
  margin: 0 0 8px 0;
}

.init-subtitle {
  font-size: 13px;
  color: #909399;
  margin: 0;
}

.init-steps {
  margin-bottom: 32px;
}

.init-content {
  min-height: 300px;
}

.password-hint {
  background: #f5f7fa;
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 18px;
  font-size: 12px;
  color: #909399;

  p {
    margin: 0 0 4px 0;
    font-weight: 600;
  }

  ul {
    margin: 0;
    padding-left: 18px;
    list-style: none;

    li {
      position: relative;
      padding: 2px 0;

      &::before {
        content: '✗';
        color: #f56c6c;
        margin-right: 6px;
        font-weight: bold;
      }

      &.met::before {
        content: '✓';
        color: #67c23a;
      }
    }
  }
}

.step-btn {
  width: 100%;
}

.send-btn {
  width: 100%;
}

.complete-section {
  padding-top: 20px;
}

@media screen and (max-width: 480px) {
  .init-box {
    padding: 24px 20px;
  }
}
</style>