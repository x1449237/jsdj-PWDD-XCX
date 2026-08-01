<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">分账规则配置</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="规则名称">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入规则名称"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 140px">
            <el-option label="默认规则" :value="1" />
            <el-option label="按服务类型" :value="2" />
            <el-option label="按俱乐部" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" @click="handleCreate">新建规则</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="name" label="规则名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="规则类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ ruleTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="playerRatio" label="打手比例" width="100" align="center">
          <template #default="{ row }">{{ row.playerRatio }}%</template>
        </el-table-column>
        <el-table-column prop="clubRatio" label="俱乐部比例" width="110" align="center">
          <template #default="{ row }">{{ row.clubRatio }}%</template>
        </el-table-column>
        <el-table-column prop="distributorRatio" label="分销商比例" width="110" align="center">
          <template #default="{ row }">{{ row.distributorRatio }}%</template>
        </el-table-column>
        <el-table-column prop="platformRatio" label="平台比例" width="100" align="center">
          <template #default="{ row }">{{ row.platformRatio }}%</template>
        </el-table-column>
        <el-table-column label="是否默认" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
            <span v-else>-</span>
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
        <el-table-column prop="createTime" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="550px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="规则类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择规则类型" style="width: 100%">
            <el-option label="默认规则" :value="1" />
            <el-option label="按服务类型" :value="2" />
            <el-option label="按俱乐部" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="打手比例" prop="playerRatio">
          <el-input-number v-model="form.playerRatio" :min="0" :max="100" :precision="2" style="width: 100%" />
          <span style="color: #909399; font-size: 12px;">单位：%</span>
        </el-form-item>
        <el-form-item label="俱乐部比例" prop="clubRatio">
          <el-input-number v-model="form.clubRatio" :min="0" :max="100" :precision="2" style="width: 100%" />
          <span style="color: #909399; font-size: 12px;">单位：%</span>
        </el-form-item>
        <el-form-item label="分销商比例" prop="distributorRatio">
          <el-input-number v-model="form.distributorRatio" :min="0" :max="100" :precision="2" style="width: 100%" />
          <span style="color: #909399; font-size: 12px;">单位：%</span>
        </el-form-item>
        <el-form-item label="平台比例" prop="platformRatio">
          <el-input-number v-model="form.platformRatio" :min="0" :max="100" :precision="2" style="width: 100%" />
          <span style="color: #909399; font-size: 12px;">单位：%</span>
        </el-form-item>
        <el-form-item label="合计">
          <span :style="{ color: totalRatio === 100 ? '#67c23a' : '#f56c6c', fontWeight: 'bold' }">
            {{ totalRatio.toFixed(2) }}%
            <span v-if="totalRatio !== 100" style="font-weight: normal;">（必须等于100%）</span>
          </span>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="form.isDefault" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
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
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ProfitShareRule',
  data() {
    return {
      Search,
      Refresh,
      Plus,
      searchForm: {
        keyword: '',
        type: '',
        status: ''
      },
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      isEdit: false,
      form: {
        id: 0,
        name: '',
        type: 1,
        playerRatio: 0,
        clubRatio: 0,
        distributorRatio: 0,
        platformRatio: 0,
        isDefault: 0,
        status: 1
      },
      rules: {
        name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }]
      },
      submitLoading: false
    }
  },
  computed: {
    dialogTitle() {
      return this.isEdit ? '编辑分账规则' : '新建分账规则'
    },
    totalRatio() {
      return Number(this.form.playerRatio) + Number(this.form.clubRatio) + Number(this.form.distributorRatio) + Number(this.form.platformRatio)
    }
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
          keyword: this.searchForm.keyword || undefined,
          type: this.searchForm.type || undefined,
          status: this.searchForm.status !== '' ? this.searchForm.status : undefined
        }
        const res = await request.get('/profit_share/rule_list', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取分账规则列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleReset() {
      this.searchForm = {
        keyword: '',
        type: '',
        status: ''
      }
      this.handleSearch()
    },
    ruleTypeLabel(type) {
      const map = { 1: '默认规则', 2: '按服务类型', 3: '按俱乐部' }
      return map[type] || '未知'
    },
    handleCreate() {
      this.isEdit = false
      this.form = {
        id: 0,
        name: '',
        type: 1,
        playerRatio: 60,
        clubRatio: 10,
        distributorRatio: 5,
        platformRatio: 25,
        isDefault: 0,
        status: 1
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.isEdit = true
      this.form = {
        id: row.id,
        name: row.name,
        type: row.type,
        playerRatio: row.playerRatio,
        clubRatio: row.clubRatio,
        distributorRatio: row.distributorRatio,
        platformRatio: row.platformRatio,
        isDefault: row.isDefault,
        status: row.status
      }
      this.dialogVisible = true
    },
    async handleSubmit() {
      if (this.totalRatio !== 100) {
        ElMessage.warning('四个角色比例之和必须等于100%')
        return
      }
      const valid = await this.$refs.formRef.validate().catch(() => false)
      if (!valid) return

      this.submitLoading = true
      try {
        const data = {
          name: this.form.name,
          type: this.form.type,
          player_ratio: this.form.playerRatio,
          club_ratio: this.form.clubRatio,
          distributor_ratio: this.form.distributorRatio,
          platform_ratio: this.form.platformRatio,
          is_default: this.form.isDefault,
          status: this.form.status
        }

        if (this.isEdit) {
          data.id = this.form.id
          await request.put('/profit_share/rule_update', data)
          ElMessage.success('更新成功')
        } else {
          await request.post('/profit_share/rule_create', data)
          ElMessage.success('创建成功')
        }
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('提交失败:', err)
      } finally {
        this.submitLoading = false
      }
    },
    async handleToggle(row, val) {
      try {
        await request.put('/profit_share/rule_toggle', { id: row.id })
        ElMessage.success('操作成功')
      } catch (err) {
          row.status = row.status === 1 ? 0 : 1
          console.error('切换状态失败:', err)
        }
    },
    handleDelete(row) {
      ElMessageBox.confirm(
        `确定要删除规则「${row.name}」吗？`,
        '删除确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      ).then(async () => {
        try {
          await request.delete(`/profit_share/rule_delete?id=${row.id}`)
          ElMessage.success('删除成功')
          this.fetchList()
        } catch (err) {
          console.error('删除失败:', err)
        }
      }).catch(() => {})
    }
  }
}
</script>

<style lang="scss" scoped>
.search-card {
  margin-bottom: 16px;
}

.search-form-inline {
  display: flex;
  flex-wrap: wrap;
}

.table-toolbar {
  margin-bottom: 16px;
}
</style>
