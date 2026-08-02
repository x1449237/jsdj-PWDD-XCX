<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">推广渠道统计</span>
    </div>

    <el-card class="table-card">
      <div style="margin-bottom: 16px;">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新建渠道</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="name" label="渠道名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="code" label="渠道码" width="140" align="center" />
        <el-table-column prop="registerCount" label="注册数" width="100" align="center" />
        <el-table-column prop="orderCount" label="订单数" width="100" align="center" />
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center" />
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

    <el-dialog v-model="dialogVisible" title="新建渠道" width="480px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="渠道名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入渠道名称" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="请输入备注" />
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
import { getPromoChannels, createPromoChannel } from '@/api/extension'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'PromoChannel',
  data() {
    return {
      Plus,
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      submitting: false,
      form: {
        name: '',
        remark: ''
      },
      rules: {
        name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }]
      }
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
          pageSize: this.pagination.pageSize
        }
        const res = await getPromoChannels(params)
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取推广渠道失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchList()
    },
    handleAdd() {
      this.form = {
        name: '',
        remark: ''
      }
      this.dialogVisible = true
    },
    async handleSubmit() {
      try {
        await this.$refs.formRef.validate()
      } catch (e) {
        return
      }
      this.submitting = true
      try {
        await createPromoChannel({ ...this.form })
        ElMessage.success('创建成功')
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('创建失败:', err)
      } finally {
        this.submitting = false
      }
    },
    resetForm() {
      this.$refs.formRef && this.$refs.formRef.resetFields()
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
</style>
