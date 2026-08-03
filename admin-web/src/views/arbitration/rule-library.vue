<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">判责规则库</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="规则类型">
          <el-select v-model="searchForm.ruleType" placeholder="全部" clearable style="width: 160px">
            <el-option label="打手迟到" value="player_late" />
            <el-option label="消极服务" value="negative_service" />
            <el-option label="玩家无故退款" value="player_unprovoked_refund" />
            <el-option label="需求变更" value="demand_change" />
            <el-option label="欺诈" value="fraud" />
          </el-select>
        </el-form-item>
        <el-form-item label="责任方">
          <el-select v-model="searchForm.faultSide" placeholder="全部" clearable style="width: 120px">
            <el-option label="打手" value="player" />
            <el-option label="买家" value="buyer" />
            <el-option label="双方" value="both" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
          <el-button type="success" :icon="Plus" @click="handleAdd">新增规则</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="规则类型" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="ruleTypeTag(row.rule_type)" size="small">
              {{ ruleTypeLabel(row.rule_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="责任方" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="faultSideTag(row.fault_side)" size="small">
              {{ faultSideLabel(row.fault_side) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处罚类型" width="130" align="center">
          <template #default="{ row }">
            {{ penaltyTypeLabel(row.penalty_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="penalty_value" label="处罚值" width="120" align="center" />
        <el-table-column prop="description" label="规则描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑规则' : '新增规则'"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="规则类型" prop="rule_type">
          <el-select v-model="form.rule_type" placeholder="请选择规则类型" style="width: 100%;">
            <el-option label="打手迟到" value="player_late" />
            <el-option label="消极服务" value="negative_service" />
            <el-option label="玩家无故退款" value="player_unprovoked_refund" />
            <el-option label="需求变更" value="demand_change" />
            <el-option label="欺诈" value="fraud" />
          </el-select>
        </el-form-item>
        <el-form-item label="责任方" prop="fault_side">
          <el-select v-model="form.fault_side" placeholder="请选择责任方" style="width: 100%;">
            <el-option label="打手" value="player" />
            <el-option label="买家" value="buyer" />
            <el-option label="双方" value="both" />
          </el-select>
        </el-form-item>
        <el-form-item label="处罚类型" prop="penalty_type">
          <el-select v-model="form.penalty_type" placeholder="请选择处罚类型" style="width: 100%;">
            <el-option label="退款比例(%)" value="refund_ratio" />
            <el-option label="扣信用分" value="deduct_credit" />
            <el-option label="扣保证金(分)" value="deduct_deposit" />
            <el-option label="封禁账号" value="ban_account" />
          </el-select>
        </el-form-item>
        <el-form-item label="处罚值" prop="penalty_value">
          <el-input v-model="form.penalty_value" placeholder="请输入处罚值" />
        </el-form-item>
        <el-form-item label="规则描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入规则描述" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ArbitrationRuleLibrary',
  components: { Search, Refresh, Plus },
  data() {
    return {
      loading: false,
      searchForm: {
        ruleType: '',
        faultSide: ''
      },
      // 全量数据(一次加载),tableData 由 computed 本地过滤,减轻后端压力
      allData: [],
      dialogVisible: false,
      isEdit: false,
      submitting: false,
      formRef: null,
      form: {
        id: 0,
        rule_type: '',
        fault_side: '',
        penalty_type: '',
        penalty_value: '',
        description: '',
        status: 1
      },
      rules: {
          rule_type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
          fault_side: [{ required: true, message: '请选择责任方', trigger: 'change' }],
          penalty_type: [{ required: true, message: '请选择处罚类型', trigger: 'change' }]
        }
    }
  },
  computed: {
    // 前端本地按规则类型 + 责任方双条件过滤,不再每次筛选都请求后端
    tableData() {
      const { ruleType, faultSide } = this.searchForm
      return this.allData.filter(r => {
        if (ruleType && r.rule_type !== ruleType) return false
        if (faultSide && r.fault_side !== faultSide) return false
        return true
      })
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    // 仅全量加载一次(及写操作后刷新),不带筛选参数
    loadData() {
      this.loading = true
      request.get('/arbitration/rule_list').then(res => {
        this.allData = res || []
      }).finally(() => {
        this.loading = false
      })
    },
    // 搜索:纯前端过滤,computed 自动响应,不调后端
    handleSearch() {
      // tableData 为 computed,searchForm 变化即自动更新
    },
    handleReset() {
      this.searchForm = {
        ruleType: '',
        faultSide: ''
      }
      // computed 自动响应,无需调后端
    },
    handleAdd() {
      this.isEdit = false
      this.form = {
        id: 0,
        rule_type: '',
        fault_side: '',
        penalty_type: '',
        penalty_value: '',
        description: '',
        status: 1
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.form = { ...row }
      this.dialogVisible = true
    },
    handleDelete(row) {
      ElMessageBox.confirm('确定删除此规则？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        request.delete('/arbitration/rule_delete', {
          data: { id: row.id }
        }).then(() => {
          ElMessage.success('删除成功')
          this.loadData()
        })
      }).catch(() => {})
    },
    handleSubmit() {
      this.$refs.formRef.validate(valid => {
        if (!valid) return
        this.submitting = true
        const requestFn = this.isEdit
          ? request.put('/arbitration/rule_update', this.form)
          : request.post('/arbitration/rule_create', this.form)
        requestFn.then(() => {
          ElMessage.success(this.isEdit ? '更新成功' : '创建成功')
          this.dialogVisible = false
          this.loadData()
        }).finally(() => {
          this.submitting = false
        })
      })
    },
    ruleTypeTag(type) {
      const map = {
        player_late: 'warning',
        negative_service: 'danger',
        player_unprovoked_refund: 'primary',
        demand_change: 'info',
        fraud: 'danger'
      }
      return map[type] || 'info'
    },
    ruleTypeLabel(type) {
      const map = {
        player_late: '打手迟到',
        negative_service: '消极服务',
        player_unprovoked_refund: '玩家无故退款',
        demand_change: '需求变更',
        fraud: '欺诈'
      }
      return map[type] || type
    },
    faultSideTag(side) {
      const map = {
        player: 'warning',
        buyer: 'primary',
        both: 'info'
      }
      return map[side] || 'info'
    },
    faultSideLabel(side) {
      const map = {
        player: '打手',
        buyer: '买家',
        both: '双方'
      }
      return map[side] || side
    },
    penaltyTypeLabel(type) {
      const map = {
        refund_ratio: '退款比例',
        deduct_credit: '扣信用分',
        deduct_deposit: '扣保证金',
        ban_account: '封禁账号'
      }
      return map[type] || type
    }
  }
}
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.page-header {
  margin-bottom: 16px;
}
.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}
.search-card,
.table-card {
  margin-bottom: 16px;
}
</style>
