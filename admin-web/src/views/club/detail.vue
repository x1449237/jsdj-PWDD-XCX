<template>
  <div class="page-container">
    <div class="page-header">
      <el-button type="default" @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon> 返回
      </el-button>
      <span class="page-title" style="margin-left:12px">俱乐部详情</span>
    </div>

    <el-card v-loading="loading" class="detail-card">
      <template v-if="club">
        <!-- 基础信息 -->
        <el-descriptions title="基础信息" :column="2" border class="section">
          <el-descriptions-item label="俱乐部ID">{{ club.id }}</el-descriptions-item>
          <el-descriptions-item label="俱乐部名称">{{ club.club_name }}</el-descriptions-item>
          <el-descriptions-item label="当前缩写">
            <el-tag type="info" size="small">{{ club.abbreviation }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="V标类型">
            <el-tag :type="club.badge_type === 'blue_v' ? 'primary' : 'success'" size="small">
              {{ club.badge_type === 'blue_v' ? '企业级蓝V' : '个人级绿V' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="V标状态">
            <el-tag v-if="club.is_active" type="success" size="small">点亮</el-tag>
            <el-tag v-else type="info" size="small">熄灭</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="运营状态">
            <el-tag v-if="club.club_status === 'pending'" type="warning">审核中</el-tag>
            <el-tag v-else-if="club.club_status === 'active'" type="success">正常运营</el-tag>
            <el-tag v-else-if="club.club_status === 'frozen'" type="info">冻结</el-tag>
            <el-tag v-else-if="club.club_status === 'closed'" type="danger">停业</el-tag>
            <el-tag v-else type="danger">注销</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ club.create_time }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ club.update_time }}</el-descriptions-item>
        </el-descriptions>

        <!-- 创始人信息 -->
        <el-descriptions title="创始人信息" :column="2" border class="section">
          <el-descriptions-item label="创始人ID">{{ club.user_id }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ club.user?.nickname || '-' }}</el-descriptions-item>
          <el-descriptions-item label="真实姓名">{{ club.real_name }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ club.id_card }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ club.phone }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ club.email || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 企业信息（仅企业俱乐部） -->
        <el-descriptions v-if="club.is_enterprise" title="企业信息" :column="2" border class="section">
          <el-descriptions-item label="企业名称">{{ club.enterprise_name }}</el-descriptions-item>
          <el-descriptions-item label="统一社会信用代码">{{ club.credit_code }}</el-descriptions-item>
          <el-descriptions-item label="营业执照" :span="2">
            <template v-if="club.business_license">
              <el-image :src="club.business_license" style="width:200px;height:140px" fit="contain" :preview-src-list="[club.business_license]" />
            </template>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="对公账户银行">{{ club.corporate_bank }}</el-descriptions-item>
          <el-descriptions-item label="对公账户号">{{ club.corporate_account }}</el-descriptions-item>
        </el-descriptions>

        <!-- 保证金信息 -->
        <el-descriptions title="保证金信息" :column="2" border class="section">
          <el-descriptions-item label="保证金金额">{{ club.deposit_amount }} 元</el-descriptions-item>
          <el-descriptions-item label="缴纳状态">
            <el-tag :type="club.deposit_status === 1 ? 'success' : club.deposit_status === 2 ? 'info' : 'warning'" size="small">
              {{ club.deposit_status === 1 ? '已缴' : club.deposit_status === 2 ? '已退' : '未缴' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="交易号">{{ club.deposit_transaction_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="缴纳时间">{{ club.deposit_pay_time || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 对公打款验证（仅企业俱乐部） -->
        <el-descriptions v-if="club.is_enterprise" title="对公打款验证" :column="2" border class="section">
          <el-descriptions-item label="验证状态">
            <el-tag v-if="club.verification_status === 0" type="info" size="small">未发起</el-tag>
            <el-tag v-else-if="club.verification_status === 1" type="warning" size="small">待确认</el-tag>
            <el-tag v-else-if="club.verification_status === 2" type="success" size="small">已通过</el-tag>
            <el-tag v-else type="danger" size="small">已驳回</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="验证金额">{{ club.verification_amount || '-' }} 元</el-descriptions-item>
          <el-descriptions-item label="验证凭证" :span="2">
            <el-image v-if="club.verification_receipt" :src="club.verification_receipt" style="width:200px;height:140px" fit="contain" :preview-src-list="[club.verification_receipt]" />
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 审核信息 -->
        <el-descriptions title="入驻审核" :column="2" border class="section">
          <el-descriptions-item label="审核状态">
            <el-tag v-if="club.audit_status === 0" type="warning">待审核</el-tag>
            <el-tag v-else-if="club.audit_status === 1" type="success">已通过</el-tag>
            <el-tag v-else-if="club.audit_status === 2" type="danger">已驳回</el-tag>
            <el-tag v-else type="info">补充资料</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="审核人">{{ club.auditor?.username || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核时间">{{ club.audit_time || '-' }}</el-descriptions-item>
          <el-descriptions-item label="驳回原因">{{ club.reject_reason || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 缩写历史 -->
        <div class="section">
          <h3 class="section-title">缩写历史记录</h3>
          <el-table :data="club.abbr_history || []" size="small" border empty-text="暂无历史记录">
            <el-table-column prop="abbreviation" label="缩写" width="120" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.club_status === 'active'" type="success" size="small">生效中</el-tag>
                <el-tag v-else-if="row.club_status === 'frozen'" type="info" size="small">冻结</el-tag>
                <el-tag v-else type="danger" size="small">已封存</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="create_time" label="创建时间" width="180" />
            <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
          </el-table>
        </div>

        <!-- 操作按钮 -->
        <div class="section" style="text-align: center; padding-top: 20px">
          <template v-if="club.club_status === 'active'">
            <el-button type="warning" @click="handleFreeze">冻结俱乐部</el-button>
            <el-button type="danger" @click="handleCancel('closed')">停业</el-button>
          </template>
          <template v-if="club.club_status === 'frozen'">
            <el-button type="success" @click="handleUnfreeze">解冻俱乐部</el-button>
          </template>
          <template v-if="club.club_status === 'closed'">
            <el-button type="danger" @click="handleCancel('cancelled')">注销</el-button>
          </template>
          <el-button type="default" @click="$router.back()">返回</el-button>
        </div>
      </template>

      <el-empty v-else-if="!loading" description="俱乐部不存在" />
    </el-card>
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ClubDetail',
  data() {
    return {
      loading: false,
      club: null
    }
  },
  mounted() {
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      const id = this.$route.params.id
      if (!id) return
      this.loading = true
      try {
        const res = await request.get('/club/detail', { id })
        this.club = res.data
      } catch (e) {
        ElMessage.error('加载详情失败')
      } finally {
        this.loading = false
      }
    },
    async handleFreeze() {
      try {
        const { value: reason } = await ElMessageBox.prompt('请输入冻结原因', '冻结俱乐部', { type: 'warning' })
        await request.put('/club/freeze', { id: this.club.id, reason })
        ElMessage.success('已冻结')
        this.fetchDetail()
      } catch (e) { /* cancel */ }
    },
    async handleUnfreeze() {
      try {
        await ElMessageBox.confirm(`确定解冻俱乐部"${this.club.club_name}"吗？`, '确认解冻', { type: 'success' })
        await request.put('/club/unfreeze', { id: this.club.id })
        ElMessage.success('已解冻')
        this.fetchDetail()
      } catch (e) { /* cancel */ }
    },
    async handleCancel(action) {
      const label = action === 'closed' ? '停业' : '注销'
      const msg = action === 'closed'
        ? `停业后俱乐部不可运营，V标熄灭，缩写永久封存不可复用。`
        : `注销后俱乐部永久关闭，缩写永久封存不可复用，此操作不可撤销！`
      try {
        const { value: reason } = await ElMessageBox.prompt(msg, `确认${label}`, { type: 'error', inputPlaceholder: '请输入原因' })
        await request.put('/club/cancel', { id: this.club.id, action, reason })
        ElMessage.success(`已${label}`)
        this.fetchDetail()
      } catch (e) { /* cancel */ }
    }
  }
}
</script>

<style lang="scss" scoped>
.detail-card {
  max-width: 1000px;
}

.section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
  padding-left: 8px;
  border-left: 3px solid #409eff;
}
</style>