<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">俱乐部列表</span>
    </div>

    <el-card>
      <el-form :model="searchForm" inline>
        <el-form-item label="入驻状态">
          <el-select v-model="searchForm.clubStatus" placeholder="全部" clearable style="width: 130px">
            <el-option label="审核中" value="pending" />
            <el-option label="正常运营" value="active" />
            <el-option label="冻结" value="frozen" />
            <el-option label="停业" value="closed" />
            <el-option label="注销" value="cancelled" />
          </el-select>
        </el-form-item>
        <el-form-item label="个人/企业">
          <el-select v-model="searchForm.badgeType" placeholder="全部" clearable style="width: 120px">
            <el-option label="企业级" value="blue_v" />
            <el-option label="个人级" value="green_v" />
          </el-select>
        </el-form-item>
        <el-form-item label="提交时间">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item label="是否缴纳保证金">
          <el-select v-model="searchForm.depositStatus" placeholder="全部" clearable style="width: 140px">
            <el-option label="已缴" :value="1" />
            <el-option label="未缴" :value="0" />
            <el-option label="已退" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否被驳回">
          <el-select v-model="searchForm.rejected" placeholder="全部" clearable style="width: 120px">
            <el-option label="是" :value="1" />
            <el-option label="否" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="searchForm.keyword" placeholder="名称/缩写/创始人" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">搜索</el-button>
          <el-button @click="onReset">重置</el-button>
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
        <el-table-column label="V标" width="80">
          <template #default="{ row }">
            <span v-if="row.is_active && row.badge_type === 'blue_v'" class="v-badge blue-v">V</span>
            <span v-else-if="row.is_active && row.badge_type === 'green_v'" class="v-badge green-v">V</span>
            <span v-else class="v-badge off">-</span>
            <el-tag v-if="row.vbadge_hidden" type="danger" size="small" style="margin-left:4px">已隐藏</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="入驻类型" width="100">
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
        <el-table-column label="保证金状态" width="120">
          <template #default="{ row }">
            <span>{{ row.deposit_amount }}元</span>
            <el-tag :type="row.deposit_status === 1 ? 'success' : row.deposit_status === 2 ? 'info' : 'warning'" size="small" style="margin-left:4px">
              {{ row.deposit_status === 1 ? '已缴' : row.deposit_status === 2 ? '已退' : '未缴' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="驳回次数" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.reject_count > 0 ? 'danger' : 'info'" size="small">{{ row.reject_count || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="170" />
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleDetail(row)">详情</el-button>
            <el-button
              v-if="row.is_active && !row.vbadge_hidden"
              type="warning" size="small" @click="handleHideVBadge(row)">隐藏V标</el-button>
            <el-button
              v-if="row.vbadge_hidden"
              type="success" size="small" @click="handleRestoreVBadge(row)">恢复V标</el-button>
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
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ClubList',
  data() {
    return {
      searchForm: {
        clubStatus: '',
        badgeType: '',
        keyword: '',
        dateRange: [],
        depositStatus: '',
        rejected: ''
      },
      tableData: [], loading: false, page: 1, limit: 20, total: 0
    }
  },
  mounted() { this.fetchList() },
  methods: {
    buildParams() {
      const params = { page: this.page, limit: this.limit }
      if (this.searchForm.clubStatus) params.clubStatus = this.searchForm.clubStatus
      if (this.searchForm.badgeType) params.badgeType = this.searchForm.badgeType
      if (this.searchForm.keyword) params.keyword = this.searchForm.keyword
      if (this.searchForm.depositStatus !== '' && this.searchForm.depositStatus !== null) {
        params.depositStatus = this.searchForm.depositStatus
      }
      if (this.searchForm.rejected !== '' && this.searchForm.rejected !== null) {
        params.rejected = this.searchForm.rejected
      }
      if (this.searchForm.dateRange && this.searchForm.dateRange.length === 2) {
        params.startDate = this.searchForm.dateRange[0]
        params.endDate = this.searchForm.dateRange[1]
      }
      return params
    },
    onSearch() {
      this.page = 1
      this.fetchList()
    },
    onReset() {
      this.searchForm = {
        clubStatus: '', badgeType: '', keyword: '',
        dateRange: [], depositStatus: '', rejected: ''
      }
      this.page = 1
      this.fetchList()
    },
    async fetchList() {
      this.loading = true
      try {
        const res = await this.$api.getClubList(this.buildParams())
        this.tableData = (res.data?.list || []).map(row => ({
          ...row,
          deposit_amount: row.deposit_amount ? (row.deposit_amount / 100).toFixed(2) : '0.00'
        }))
        this.total = res.data?.total || 0
      } catch (e) { ElMessage.error('加载失败') }
      finally { this.loading = false }
    },
    handleDetail(row) { this.$router.push(`/club/detail/${row.id}`) },
    async handleFreeze(row) {
      try {
        const { value: reason } = await ElMessageBox.prompt('请输入冻结原因', '冻结俱乐部', { type: 'warning' })
        await this.$api.freezeClub({ id: row.id, reason })
        ElMessage.success('已冻结')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleUnfreeze(row) {
      try {
        await ElMessageBox.confirm(`确定解冻俱乐部"${row.club_name}"吗？`, '确认解冻', { type: 'success' })
        await this.$api.unfreezeClub({ id: row.id })
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
        await this.$api.cancelClub({ id: row.id, action, reason })
        ElMessage.success(`已${label}`)
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleHideVBadge(row) {
      try {
        await ElMessageBox.confirm(
          `确定隐藏俱乐部"${row.club_name}"的V标吗？隐藏后该俱乐部前端将不展示V标，但俱乐部仍可正常运营。`,
          '隐藏V标确认',
          { type: 'warning', confirmButtonText: '确认隐藏' }
        )
        await this.$api.hideClubVBadge(row.id)
        ElMessage.success('V标已隐藏')
        this.fetchList()
      } catch (e) { /* cancel */ }
    },
    async handleRestoreVBadge(row) {
      try {
        await ElMessageBox.confirm(
          `确定恢复俱乐部"${row.club_name}"的V标吗？`,
          '恢复V标确认',
          { type: 'success', confirmButtonText: '确认恢复' }
        )
        await this.$api.restoreClubVBadge(row.id)
        ElMessage.success('V标已恢复')
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
