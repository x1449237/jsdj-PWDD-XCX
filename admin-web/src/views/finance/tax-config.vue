<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">个税配置</span>
    </div>

    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="role" label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="taxRate" label="税率(%)" width="150" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row.taxRate"
              :min="0"
              :max="100"
              :precision="2"
              size="small"
              style="width: 120px"
            />
          </template>
        </el-table-column>
        <el-table-column prop="threshold" label="起征点(元)" width="180" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row.threshold"
              :min="0"
              :precision="2"
              size="small"
              style="width: 150px"
            />
          </template>
        </el-table-column>
        <el-table-column prop="quickDeduction" label="速算扣除数(元)" width="180" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row.quickDeduction"
              :min="0"
              :precision="2"
              size="small"
              style="width: 150px"
            />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val) => handleToggle(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleSave(row)">保存</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card style="margin-top: 16px;">
      <template #header>
        <span>个税计算器</span>
      </template>
      <el-form :model="calculatorForm" :inline="true" label-width="80px">
        <el-form-item label="角色">
          <el-select v-model="calculatorForm.role" style="width: 140px">
            <el-option label="打手" :value="1" />
            <el-option label="俱乐部" :value="2" />
            <el-option label="分销商" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="金额(元)">
          <el-input-number
            v-model="calculatorForm.amount"
            :min="0"
            :precision="2"
            style="width: 180px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Coin" @click="handleCalculate">计算</el-button>
        </el-form-item>
      </el-form>

      <el-descriptions v-if="calculatorResult" :column="2" border style="margin-top: 16px;">
        <el-descriptions-item label="计税金额">¥{{ calculatorResult.amount }}</el-descriptions-item>
        <el-descriptions-item label="适用税率">{{ calculatorResult.taxRate }}%</el-descriptions-item>
        <el-descriptions-item label="起征点">¥{{ calculatorResult.threshold }}</el-descriptions-item>
        <el-descriptions-item label="应扣税额">
          <span style="color: #f56c6c; font-weight: 600;">¥{{ calculatorResult.taxAmount }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="税后金额" :span="2">
          <span style="color: #67c23a; font-weight: 600;">¥{{ calculatorResult.actualAmount }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card style="margin-top: 16px;">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>个税代扣记录</span>
        </div>
      </template>
      <el-form :model="taxSearchForm" :inline="true" class="search-form-inline">
        <el-form-item label="用户ID">
          <el-input
            v-model="taxSearchForm.userId"
            placeholder="请输入用户ID"
            clearable
            style="width: 140px"
          />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="taxSearchForm.role" placeholder="全部" clearable style="width: 120px">
            <el-option label="打手" :value="1" />
            <el-option label="俱乐部" :value="2" />
            <el-option label="分销商" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="月份">
          <el-date-picker
            v-model="taxSearchForm.month"
            type="month"
            value-format="YYYY-MM"
            placeholder="选择月份"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchTaxRecords">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="taxRecordList" v-loading="taxLoading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="用户" min-width="130">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '-' }}</div>
            <div style="color: #909399; font-size: 12px;">ID: {{ row.userId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="100" align="center">
          <template #default="{ row }">{{ roleLabel(row.role) }}</template>
        </el-table-column>
        <el-table-column prop="amount" label="计税金额" width="120" align="center">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column prop="taxAmount" label="代扣税额" width="120" align="center">
          <template #default="{ row }">
            <span style="color: #f56c6c;">¥{{ row.taxAmount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="taxRate" label="税率" width="90" align="center">
          <template #default="{ row }">{{ row.taxRate }}%</template>
        </el-table-column>
        <el-table-column prop="month" label="所属月份" width="110" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="taxStatusTagType(row.status)" size="small">
              {{ taxStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="certificateNo" label="完税凭证号" width="160" show-overflow-tooltip />
        <el-table-column prop="createTime" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status !== 3"
              type="primary"
              link
              size="small"
              @click="handleMarkComplete(row)"
            >标记完税</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="taxPagination.page"
          v-model:page-size="taxPagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="taxPagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchTaxRecords"
          @current-change="fetchTaxRecords"
        />
      </div>
    </el-card>

    <el-dialog v-model="completeDialogVisible" title="标记完税" width="400px">
      <el-form :model="completeForm" label-width="100px">
        <el-form-item label="凭证号">
          <el-input v-model="completeForm.certificateNo" placeholder="请输入完税凭证号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="completeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="completeLoading" @click="handleCompleteSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Coin } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'TaxConfig',
  data() {
    return {
      Search,
      Coin,
      tableData: [],
      loading: false,
      calculatorForm: {
        role: 1,
        amount: 0
      },
      calculatorResult: null,
      taxSearchForm: {
        userId: '',
        role: '',
        month: ''
      },
      taxRecordList: [],
      taxLoading: false,
      taxPagination: {
        page: 1,
        pageSize: 10,
        total: 0
      },
      completeDialogVisible: false,
      completeRow: null,
      completeForm: {
        certificateNo: ''
      },
      completeLoading: false
    }
  },
  mounted() {
    this.fetchList()
    this.fetchTaxRecords()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/tax/config_list')
        this.tableData = res.data || []
      } catch (err) {
        console.error('获取个税配置失败:', err)
      } finally {
        this.loading = false
      }
    },
    roleLabel(role) {
      const map = { 1: '打手', 2: '俱乐部', 3: '分销商' }
      return map[role] || '未知'
    },
    async handleSave(row) {
      try {
        await request.put('/tax/config_update', {
          role: row.role,
          tax_rate: row.taxRate,
          threshold: row.threshold,
          quick_deduction: row.quickDeduction
        })
        ElMessage.success('保存成功')
        this.fetchList()
      } catch (err) {
        console.error('保存失败:', err)
      }
    },
    handleToggle(row, val) {
      request.put('/tax/config_update', {
        role: row.role,
        status: val
      }).then(() => {
        ElMessage.success('操作成功')
      }).catch(() => {
        row.status = val === 1 ? 0 : 1
      })
    },
    async handleCalculate() {
      try {
        const res = await request.get('/api/v1/tax/calculator', {
          params: {
            amount: this.calculatorForm.amount,
            role: this.calculatorForm.role
          }
        })
        this.calculatorResult = res.data
      } catch (err) {
        console.error('计算失败:', err)
      }
    },
    async fetchTaxRecords() {
      this.taxLoading = true
      try {
        const params = {
          page: this.taxPagination.page,
          pageSize: this.taxPagination.pageSize,
          user_id: this.taxSearchForm.userId || undefined,
          role: this.taxSearchForm.role || undefined,
          month: this.taxSearchForm.month || undefined
        }
        const res = await request.get('/tax/record_list', { params })
        this.taxRecordList = res.data?.list || []
        this.taxPagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取个税记录失败:', err)
      } finally {
        this.taxLoading = false
      }
    },
    taxStatusLabel(status) {
      const map = { 1: '已代扣', 2: '已申报', 3: '已完税' }
      return map[status] || '未知'
    },
    taxStatusTagType(status) {
      const map = { 1: 'warning', 2: '', 3: 'success' }
      return map[status] || 'info'
    },
    handleMarkComplete(row) {
      this.completeRow = row
      this.completeForm.certificateNo = ''
      this.completeDialogVisible = true
    },
    async handleCompleteSubmit() {
      this.completeLoading = true
      try {
        await request.put('/tax/record_complete', {
          id: this.completeRow.id,
          certificate_no: this.completeForm.certificateNo
        })
        ElMessage.success('操作成功')
        this.completeDialogVisible = false
        this.fetchTaxRecords()
      } catch (err) {
        console.error('操作失败:', err)
      } finally {
        this.completeLoading = false
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.search-form-inline {
  display: flex;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
