<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">超时规则引擎</span>
    </div>

    <!-- 预置规则列表 -->
    <el-card class="section-card">
      <template #header>
        <div class="card-header">
          <span>预置规则</span>
          <el-tag type="info" size="small">不可删除，仅可启用/禁用</el-tag>
        </div>
      </template>
      <el-table :data="presetRules" v-loading="presetLoading" stripe border style="width: 100%">
        <el-table-column prop="name" label="规则名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="触发状态 → 目标状态" min-width="180">
          <template #default="{ row }">
            <el-tag type="warning" size="small">{{ row.triggerStatus }}</el-tag>
            <el-icon style="margin: 0 6px; vertical-align: middle;"><ArrowRight /></el-icon>
            <el-tag type="success" size="small">{{ row.targetStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="timeoutSeconds" label="超时时间" width="120" align="center">
          <template #default="{ row }">
            {{ row.timeoutSeconds }} 秒
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="是否预置" width="90" align="center">
          <template #default>
            <el-tag type="info" size="small">预置</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              :loading="row.switching"
              @change="handleTogglePreset(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditPreset(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 自定义规则列表 -->
    <el-card class="section-card">
      <template #header>
        <div class="card-header">
          <span>自定义规则</span>
          <el-button type="primary" size="small" :icon="Plus" @click="handleAdd">新增规则</el-button>
        </div>
      </template>
      <el-table :data="customRules" v-loading="customLoading" stripe border style="width: 100%">
        <el-table-column prop="name" label="规则名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="触发状态 → 目标状态" min-width="180">
          <template #default="{ row }">
            <el-tag type="warning" size="small">{{ row.triggerStatus }}</el-tag>
            <el-icon style="margin: 0 6px; vertical-align: middle;"><ArrowRight /></el-icon>
            <el-tag type="success" size="small">{{ row.targetStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="超时时间" width="120" align="center">
          <template #default="{ row }">
            {{ row.timeoutSeconds }} 秒
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="是否预置" width="90" align="center">
          <template #default>
            <el-tag size="small">自定义</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              :loading="row.switching"
              @change="handleToggleCustom(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑规则弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="540px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="ruleFormRef" :model="ruleForm" :rules="ruleFormRules" label-width="120px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="触发状态" prop="triggerStatus">
          <el-select v-model="ruleForm.triggerStatus" placeholder="请选择触发状态" style="width: 100%">
            <el-option label="待接单" value="待接单" />
            <el-option label="已接单" value="已接单" />
            <el-option label="进行中" value="进行中" />
            <el-option label="待确认" value="待确认" />
            <el-option label="待支付" value="待支付" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标状态" prop="targetStatus">
          <el-select v-model="ruleForm.targetStatus" placeholder="请选择目标状态" style="width: 100%">
            <el-option label="已取消" value="已取消" />
            <el-option label="自动确认" value="自动确认" />
            <el-option label="自动完成" value="自动完成" />
            <el-option label="已过期" value="已过期" />
          </el-select>
        </el-form-item>
        <el-form-item label="超时秒数" prop="timeoutSeconds">
          <el-input-number
            v-model="ruleForm.timeoutSeconds"
            :min="1"
            :max="86400"
            style="width: 100%"
            placeholder="请输入超时秒数"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number
            v-model="ruleForm.sort"
            :min="0"
            :max="9999"
            style="width: 100%"
            placeholder="请输入排序值，越小越靠前"
          />
        </el-form-item>
        <el-form-item label="状态" prop="enabled">
          <el-switch v-model="ruleForm.enabled" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Plus, ArrowRight } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'TimeoutRule',
  data() {
    return {
      Plus,
      ArrowRight,
      presetRules: [],
      customRules: [],
      presetLoading: false,
      customLoading: false,
      dialogVisible: false,
      dialogTitle: '新增规则',
      submitLoading: false,
      isEdit: false,
      editId: null,
      isPreset: false,
      ruleForm: {
        name: '',
        triggerStatus: '',
        targetStatus: '',
        timeoutSeconds: null,
        sort: 0,
        enabled: true
      },
      ruleFormRules: {
        name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
        triggerStatus: [{ required: true, message: '请选择触发状态', trigger: 'change' }],
        targetStatus: [{ required: true, message: '请选择目标状态', trigger: 'change' }],
        timeoutSeconds: [{ required: true, message: '请输入超时秒数', trigger: 'blur' }]
      }
    }
  },
  mounted() {
    this.fetchPresetRules()
    this.fetchCustomRules()
  },
  methods: {
    async fetchPresetRules() {
      this.presetLoading = true
      try {
        const res = await request.get('/timeout-rules/preset')
        this.presetRules = (res.data?.list || []).map(item => ({ ...item, switching: false }))
      } catch (err) {
        console.error('获取预置规则失败:', err)
      } finally {
        this.presetLoading = false
      }
    },
    async fetchCustomRules() {
      this.customLoading = true
      try {
        const res = await request.get('/timeout-rules/custom')
        this.customRules = (res.data?.list || []).map(item => ({ ...item, switching: false }))
      } catch (err) {
        console.error('获取自定义规则失败:', err)
      } finally {
        this.customLoading = false
      }
    },
    async handleTogglePreset(row) {
      row.switching = true
      try {
        await request.put(`/timeout-rules/preset/${row.id}/toggle`, { enabled: row.enabled })
        ElMessage.success(row.enabled ? '已启用' : '已禁用')
      } catch (err) {
        row.enabled = !row.enabled
        console.error('切换状态失败:', err)
      } finally {
        row.switching = false
      }
    },
    handleEditPreset(row) {
      this.isEdit = true
      this.isPreset = true
      this.editId = row.id
      this.dialogTitle = '编辑预置规则'
      this.ruleForm = {
        name: row.name,
        triggerStatus: row.triggerStatus,
        targetStatus: row.targetStatus,
        timeoutSeconds: row.timeoutSeconds,
        sort: row.sort,
        enabled: row.enabled
      }
      this.dialogVisible = true
    },
    async handleToggleCustom(row) {
      row.switching = true
      try {
        await request.put(`/timeout-rules/custom/${row.id}/toggle`, { enabled: row.enabled })
        ElMessage.success(row.enabled ? '已启用' : '已禁用')
      } catch (err) {
        row.enabled = !row.enabled
        console.error('切换状态失败:', err)
      } finally {
        row.switching = false
      }
    },
    handleAdd() {
      this.isEdit = false
      this.isPreset = false
      this.editId = null
      this.dialogTitle = '新增规则'
      this.ruleForm = {
        name: '',
        triggerStatus: '',
        targetStatus: '',
        timeoutSeconds: null,
        sort: 0,
        enabled: true
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.isPreset = false
      this.editId = row.id
      this.dialogTitle = '编辑规则'
      this.ruleForm = {
        name: row.name,
        triggerStatus: row.triggerStatus,
        targetStatus: row.targetStatus,
        timeoutSeconds: row.timeoutSeconds,
        sort: row.sort,
        enabled: row.enabled
      }
      this.dialogVisible = true
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除规则「${row.name}」吗？删除后不可恢复。`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete(`/timeout-rules/custom/${row.id}`)
        ElMessage.success('删除成功')
        this.fetchCustomRules()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    async handleSubmit() {
      const valid = await this.$refs.ruleFormRef.validate().catch(() => false)
      if (!valid) return
      this.submitLoading = true
      try {
        if (this.isEdit) {
          const url = this.isPreset
            ? `/timeout-rules/preset/${this.editId}`
            : `/timeout-rules/custom/${this.editId}`
          await request.put(url, this.ruleForm)
          ElMessage.success('编辑成功')
        } else {
          await request.post('/timeout-rules/custom', this.ruleForm)
          ElMessage.success('新增成功')
        }
        this.dialogVisible = false
        if (this.isPreset) {
          this.fetchPresetRules()
        } else {
          this.fetchCustomRules()
        }
      } catch (err) {
        console.error('提交失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    handleDialogClosed() {
      this.$refs.ruleFormRef?.resetFields()
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
</style>