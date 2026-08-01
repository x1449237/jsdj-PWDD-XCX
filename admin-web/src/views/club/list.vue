<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">俱乐部列表</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select v-model="searchForm.clubStatus" placeholder="全部" clearable style="width: 120px">
            <el-option label="审核中" value="pending" />
            <el-option label="正常运营" value="active" />
            <el-option label="冻结" value="frozen" />
            <el-option label="停业" value="closed" />
            <el-option label="注销" value="cancelled" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.badgeType" placeholder="全部" clearable style="width: 120px">
            <el-option label="企业级" value="blue_v" />
            <el-option label="个人级" value="green_v" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="searchForm.keyword" placeholder="名称/缩写/创始人" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="club_name" label="俱乐部名称" min-width="140" />
        <el-table-column label="缩写" width="100">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.abbreviation }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="V标" width="70">
          <template #default="{ row }">
            <span v-if="row.is_active && row.badge_type === 'blue_v'" class="v-badge blue-v">V</span>
            <span v-else-if="row.is_active && row.badge_type === 'green_v'" class="v-badge green-v">V</span>
            <span v-else class="v-badge off">-</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.badge_type === 'blue_v' ? 'primary' : 'success'" size="small">
              {{ row.badge_type === 'blue_v' ? '企业' : '个人' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创始人" width="100">
          <template #default="{ row }">{{ row.user?.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.club_status === 'pending'" type="warning">审核中</el-tag>
            <el-tag v-else-if="row.club_status === 'active'" type="success">正常</el-tag>
            <el-tag v-else-if="row.club_status === 'frozen'" type="info">冻结</el-tag>
            <el-tag v-else-if="row.club_status === 'closed'" type="danger">停业</el-tag>
            <el-tag v-else type="danger">注销</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="保证金" width="110">
          <template #default="{ row }">
            <span>{{ row.deposit_amount }}元</span>
            <el-tag :type="row.deposit_status === 1 ? 'success' : row.deposit_status === 2 ? 'info' : 'warning'" size="small" style="margin-left:4px">
              {{ row.deposit_status === 1 ? '已缴' : row.deposit_status === 2 ? '已退' : '未缴' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="170" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleDetail(row)">详情</el-button>
            <template v-if="row.club_status === 'active'">
              <el-button type="warning" size="small" @click="handleFreeze(row)">冻结</el-button>
              <el-button type="danger" size="small" @click="handleCancel(row, 'closed')">停业</el-button>
            </template>
            <template v-if="row.club_status === 'frozen'">
              <el-button type="success" size="small" @click="handleUnfreeze(row)">解冻</el-button>
            </template>
            <template v-if="row.club_status === 'closed'">
              <el-button type="danger" size="small" @click="handleCancel(row, 'cancelled')">注销</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" v-model:page-size="limit" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    
  </div>
</template>

<script>
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ClubList',
  data() {
    return {
      searchForm: { clubStatus: '', badgeType: '', keyword: '' },
      tableData: [], loading: false, page: 1, limit: 20, total: 0
    }
  },
  mounted() { this.fetchList() },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await request.get('/club/list', { page: this.page, limit: this.limit, ...this.searchForm })
        this.tableData = res.data?.list || []
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    handleDetail(row) { this.$router.push(`/club/detail/${row.id}`) },
    async handleFreeze(row) {
      try {
        const { value: reason } = await ElMessageBox.prompt('请输入冻结原因', '冻结俱乐部', { type: 'warning' })
        await request.put('/club/freeze', { id: row.id, reason })
        ElMessage.success('已冻结')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleUnfreeze(row) {
      try {
        await ElMessageBox.confirm(`确定解冻俱乐部"${row.club_name}"吗？`, '确认解冻', { type: 'success' })
        await request.put('/club/unfreeze', { id: row.id })
        ElMessage.success('已解冻')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleCancel(row, action) {
      const label = action === 'closed' ? '停业' : '注销'
      const msg = action === 'closed'
        ? `停业后俱乐部不可运营，V标熄灭，缩写永久封存不可复用。`
        : `注销后俱乐部永久关闭，缩写永久封存不可复用，此操作不可撤销！`
      try {
        const { value: reason } = await ElMessageBox.prompt(msg, `确认${label}`, { type: 'error', inputPlaceholder: '请输入原因' })
        await request.put('/club/cancel', { id: row.id, action, reason })
        ElMessage.success(`已${label}`)
        this.fetchList()
      } catch (e) { /* cancel */ }
    }
  }
}
</script>

<style lang="scss" scoped>
.v-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 20px; height: 20px; border-radius: 50%; font-size: 12px; font-weight: bold; color: #fff;
}
.blue-v { background: linear-gradient(135deg, #1890ff, #096dd9); }
.green-v { background: linear-gradient(135deg, #52c41a, #389e0d); }
.off { background: #ddd; color: #999; }
</style>