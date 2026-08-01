<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">仲裁案件列表</span>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" :inline="true" class="search-form-inline">
        <el-form-item label="案件状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 150px">
            <el-option label="待受理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已结案" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item label="纠纷类型">
          <el-select v-model="searchForm.disputeType" placeholder="全部" clearable style="width: 150px">
            <el-option label="打手迟到" value="player_late" />
            <el-option label="消极服务" value="negative_service" />
            <el-option label="玩家退款纠纷" value="player_refund" />
            <el-option label="需求变更" value="demand_change" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="案件ID/订单ID" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="案件ID" width="100" align="center" />
        <el-table-column prop="order_id" label="订单ID" width="120" align="center" />
        <el-table-column prop="applicant_id" label="申请人ID" width="120" align="center" />
        <el-table-column prop="respondent_id" label="被申请人ID" width="120" align="center" />
        <el-table-column label="纠纷类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="disputeTypeTag(row.dispute_type)" size="small">
              {{ disputeTypeLabel(row.dispute_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="handler_id" label="处理人" width="100" align="center" />
        <el-table-column prop="create_time" label="创建时间" width="170" align="center" />
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="warning"
              link
              size="small"
              @click="handleProcess(row)"
            >
              受理
            </el-button>
            <el-button
              v-if="row.status === 'processing'"
              type="success"
              link
              size="small"
              @click="handleResolve(row)"
            >
              结案
            </el-button>
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
      v-model="detailDialogVisible"
      title="案件详情"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading">
        <el-descriptions :column="2" border size="small" style="margin-bottom: 20px;">
          <el-descriptions-item label="案件ID">{{ caseDetail.id }}</el-descriptions-item>
          <el-descriptions-item label="订单ID">{{ caseDetail.order_id }}</el-descriptions-item>
          <el-descriptions-item label="申请人ID">{{ caseDetail.applicant_id }}</el-descriptions-item>
          <el-descriptions-item label="被申请人ID">{{ caseDetail.respondent_id }}</el-descriptions-item>
          <el-descriptions-item label="纠纷类型">
            {{ disputeTypeLabel(caseDetail.dispute_type) }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(caseDetail.status)" size="small">
              {{ statusLabel(caseDetail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="处理人">{{ caseDetail.handler_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ caseDetail.create_time }}</el-descriptions-item>
          <el-descriptions-item v-if="caseDetail.finish_time" label="结案时间">
            {{ caseDetail.finish_time }}
          </el-descriptions-item>
          <el-descriptions-item label="纠纷描述" :span="2">
            {{ caseDetail.description }}
          </el-descriptions-item>
          <el-descriptions-item v-if="caseDetail.result" label="处理结果" :span="2">
            {{ caseDetail.result }}
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="caseDetail.evidence_list && caseDetail.evidence_list.length > 0">
          <div class="evidence-title">举证材料</div>
          <div class="evidence-list">
            <div
              v-for="(item, index) in caseDetail.evidence_list"
              :key="item.id"
              class="evidence-item"
            >
              <div class="evidence-type">{{ evidenceTypeLabel(item.type) }}</div>
              <div class="evidence-desc">{{ item.description || '无描述' }}</div>
              <el-image
                v-if="item.type === 'image' && item.file_url"
                :src="item.file_url"
                :preview-src-list="[item.file_url]"
                fit="cover"
                class="evidence-image"
              />
              <a v-else-if="item.file_url" :href="item.file_url" target="_blank">查看文件</a>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="resolveDialogVisible"
      title="结案处理"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="resolveForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-input
            v-model="resolveForm.result"
            type="textarea"
            :rows="4"
            placeholder="请输入处理结果描述"
          />
        </el-form-item>
        <el-form-item label="处罚措施">
          <div v-for="(penalty, index) in resolveForm.penalties" :key="index" class="penalty-item">
            <el-select v-model="penalty.penalty_type" placeholder="处罚类型" style="width: 140px;">
              <el-option label="退款比例" value="refund_ratio" />
              <el-option label="扣信用分" value="deduct_credit" />
              <el-option label="扣保证金" value="deduct_deposit" />
              <el-option label="封禁账号" value="ban_account" />
            </el-select>
            <el-input v-model="penalty.penalty_value" placeholder="处罚值" style="width: 120px; margin: 0 8px;" />
            <el-select v-model="penalty.target" placeholder="对象" style="width: 100px;">
              <el-option label="打手" value="player" />
              <el-option label="买家" value="buyer" />
            </el-select>
            <el-button type="danger" link @click="removePenalty(index)">删除</el-button>
          </div>
          <el-button type="primary" link @click="addPenalty">+ 添加处罚</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resolveDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="resolving" @click="confirmResolve">确认结案</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'ArbitrationCaseList',
  components: { Search, Refresh },
  data() {
    return {
      loading: false,
      searchForm: {
        status: '',
        disputeType: '',
        keyword: ''
      },
      tableData: [],
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      detailDialogVisible: false,
      detailLoading: false,
      caseDetail: {},
      resolveDialogVisible: false,
      resolving: false,
      resolveForm: {
        caseId: 0,
        result: '',
        penalties: []
      }
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    loadData() {
      this.loading = true
      request.get('/arbitration/case_list', {
        params: {
          status: this.searchForm.status,
          dispute_type: this.searchForm.disputeType,
          keyword: this.searchForm.keyword,
          page: this.pagination.page,
          page_size: this.pagination.pageSize
        }
      }).then(res => {
        this.tableData = res.list || []
        this.pagination.total = res.total || 0
      }).finally(() => {
        this.loading = false
      })
    },
    handleSearch() {
      this.pagination.page = 1
      this.loadData()
    },
    handleReset() {
      this.searchForm = {
        status: '',
        disputeType: '',
        keyword: ''
      }
      this.pagination.page = 1
      this.loadData()
    },
    handleViewDetail(row) {
      this.detailLoading = true
      this.detailDialogVisible = true
      request.get('/arbitration/case_detail', {
        params: { case_id: row.id }
      }).then(res => {
        this.caseDetail = res
      }).finally(() => {
        this.detailLoading = false
      })
    },
    handleProcess(row) {
      ElMessageBox.confirm('确定受理此仲裁案件？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        request.post('/arbitration/process_case', {
          case_id: row.id
        }).then(() => {
          ElMessage.success('受理成功')
          this.loadData()
        })
      }).catch(() => {})
    },
    handleResolve(row) {
      this.resolveForm = {
        caseId: row.id,
        result: '',
        penalties: []
      }
      this.resolveDialogVisible = true
    },
    addPenalty() {
      this.resolveForm.penalties.push({
        penalty_type: '',
        penalty_value: '',
        target: 'player'
      })
    },
    removePenalty(index) {
      this.resolveForm.penalties.splice(index, 1)
    },
    confirmResolve() {
      if (!this.resolveForm.result.trim()) {
        ElMessage.warning('请输入处理结果')
        return
      }
      this.resolving = true
      const penalties = this.resolveForm.penalties.filter(p => p.penalty_type)
      request.post('/arbitration/resolve_case', {
        case_id: this.resolveForm.caseId,
        result: this.resolveForm.result,
        penalties: JSON.stringify(penalties)
      }).then(() => {
        ElMessage.success('结案成功')
        this.resolveDialogVisible = false
        this.loadData()
      }).finally(() => {
        this.resolving = false
      })
    },
    statusTag(status) {
      const map = {
        pending: 'warning',
        processing: 'primary',
        resolved: 'success'
      }
      return map[status] || 'info'
    },
    statusLabel(status) {
      const map = {
        pending: '待受理',
        processing: '处理中',
        resolved: '已结案'
      }
      return map[status] || status
    },
    disputeTypeTag(type) {
      const map = {
        player_late: 'warning',
        negative_service: 'danger',
        player_refund: 'primary',
        demand_change: 'info',
        other: 'success'
      }
      return map[type] || 'info'
    },
    disputeTypeLabel(type) {
      const map = {
        player_late: '打手迟到',
        negative_service: '消极服务',
        player_refund: '玩家退款纠纷',
        demand_change: '需求变更',
        other: '其他'
      }
      return map[type] || type
    },
    evidenceTypeLabel(type) {
      const map = {
        image: '图片',
        video: '视频',
        audio: '音频',
        text: '文字'
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
.pagination-container {
  margin-top: 20px;
  text-align: right;
}
.evidence-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #303133;
}
.evidence-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.evidence-item {
  width: 160px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 8px;
}
.evidence-type {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
.evidence-desc {
  font-size: 13px;
  color: #303133;
  margin-bottom: 8px;
  word-break: break-all;
}
.evidence-image {
  width: 100%;
  height: 120px;
  border-radius: 4px;
  cursor: pointer;
}
.penalty-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}
</style>
