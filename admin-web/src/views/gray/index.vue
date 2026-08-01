<template>
  <div class="page-container">
    <div class="page-header">
      <span class="page-title">灰度发布</span>
    </div>

    <el-row :gutter="20" class="top-cards">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">当前放量比例</div>
          <div class="stat-value primary">
            {{ globalConfig.release_ratio ?? 0 }}%
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">错误率阈值（回滚）</div>
          <div class="stat-value" :class="(globalConfig.error_rate_threshold ?? 5) > 5 ? 'error' : 'warning'">
            {{ globalConfig.error_rate_threshold ?? 5 }}%
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">白名单用户数</div>
          <div class="stat-value success">
            {{ (globalConfig.whitelist_uids || []).length }}
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">灰度发布中</div>
          <div class="stat-value">
            <el-tag :type="tableData.some(g => g.status === 'gray') ? 'warning' : 'info'">
              {{ tableData.some(g => g.status === 'gray') ? '进行中' : '暂无' }}
            </el-tag>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="section-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>全局灰度配置（立即生效）</span>
          <el-button type="primary" :icon="Promotion" :loading="globalSaveLoading" @click="saveGlobalConfig">
            保存配置
          </el-button>
        </div>
      </template>

      <el-form label-width="160px">
        <el-form-item label="放量比例（全局）">
          <el-row style="width: 100%;" :gutter="12" align="middle">
            <el-col :span="18">
              <el-slider
                v-model="globalConfig.release_ratio"
                :min="0"
                :max="100"
                :step="1"
                :marks="{ 0: '0%', 10: '10%', 25: '25%', 50: '50%', 75: '75%', 100: '100%' }"
                :show-tooltip="true"
              />
            </el-col>
            <el-col :span="6">
              <el-input-number
                v-model="globalConfig.release_ratio"
                :min="0"
                :max="100"
                style="width: 100%"
                controls-position="right"
              >
                <template #suffix><span style="color:#909399;">%</span></template>
              </el-input-number>
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="错误率阈值（%）">
          <el-row style="width: 100%;" :gutter="12" align="middle">
            <el-col :span="18">
              <el-slider
                v-model="globalConfig.error_rate_threshold"
                :min="0"
                :max="50"
                :step="1"
                :marks="{ 1: '1%', 5: '5%', 10: '10%', 20: '20%', 50: '50%' }"
                :show-tooltip="true"
              />
            </el-col>
            <el-col :span="6">
              <el-input-number
                v-model="globalConfig.error_rate_threshold"
                :min="0"
                :max="100"
                :precision="2"
                style="width: 100%"
                controls-position="right"
              >
                <template #suffix><span style="color:#909399;">%</span></template>
              </el-input-number>
              <div class="tip-text">超过此比例将自动触发回滚</div>
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="白名单用户UID">
          <el-input
            v-model="whitelistText"
            type="textarea"
            :rows="4"
            placeholder="批量粘贴白名单用户UID，支持逗号、空格、换行、分号分隔。例如：&#10;10001,10002,10003&#10;10004 10005"
            maxlength="10000"
            show-word-limit
          />
          <div class="whitelist-tip">
            当前解析到 <span style="color: #409eff; font-weight: 600;">{{ parsedWhitelist.length }}</span> 个有效UID
            <el-button type="primary" link size="small" @click="dedupWhitelist">去重</el-button>
            <el-button type="danger" link size="small" @click="clearWhitelist">清空</el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" style="margin-top: 20px;">
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增灰度发布</el-button>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="version" label="版本号" width="120" align="center" />
        <el-table-column label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'pending' ? 'info' : row.status === 'gray' ? 'warning' : row.status === 'full' ? 'success' : 'danger'"
              size="small"
            >
              {{ row.status === 'pending' ? '待发布' : row.status === 'gray' ? '灰度中' : row.status === 'full' ? '全量' : '已回滚' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="白名单" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.whiteList && row.whiteList.length">{{ row.whiteList.length }}个UID</span>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="放量比例" width="120" align="center">
          <template #default="{ row }">
            <el-progress
              :percentage="row.ratio"
              :stroke-width="8"
              :color="row.ratio >= 100 ? '#67c23a' : '#409eff'"
            />
          </template>
        </el-table-column>
        <el-table-column label="错误率" width="110" align="center">
          <template #default="{ row }">
            <span
              :style="{
                color: (row.errorRate ?? 0) > 5 ? '#f56c6c' : (row.errorRate ?? 0) > 1 ? '#e6a23c' : '#67c23a',
                fontWeight: '600'
              }"
            >
              {{ row.errorRate ?? 0 }}%
            </span>
            <el-tooltip
              v-if="(row.errorRate ?? 0) > (globalConfig.error_rate_threshold ?? 5)"
              content="已超过回滚阈值"
              placement="top"
            >
              <el-icon :size="14" color="#f56c6c" style="margin-left: 4px;"><WarningFilled /></el-icon>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="发布时间" width="170" align="center" />
        <el-table-column label="操作" width="260" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending'"
              type="success"
              link
              size="small"
              @click="handlePublish(row)"
            >
              发布
            </el-button>
            <el-button
              v-if="row.status === 'gray'"
              type="primary"
              link
              size="small"
              @click="handleFullRelease(row)"
            >
              全量
            </el-button>
            <el-button
              v-if="row.status === 'gray' || row.status === 'full'"
              type="danger"
              link
              size="small"
              @click="handleRollback(row)"
            >
              回滚
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
            >
              删除
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
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      title="新增灰度发布"
      width="600px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="版本号" prop="version">
          <el-input v-model="form.version" placeholder="请输入版本号，如 v1.2.0" maxlength="30" />
        </el-form-item>
        <el-form-item label="白名单用户ID" prop="whiteList">
          <el-input
            v-model="form.whiteListText"
            type="textarea"
            :rows="3"
            placeholder="批量粘贴用户ID，支持逗号、空格、换行分隔"
            maxlength="2000"
            show-word-limit
          />
          <div class="tip-text">支持多个UID批量输入</div>
        </el-form-item>
        <el-form-item label="放量比例" prop="ratio">
          <el-row style="width: 100%;" :gutter="12" align="middle">
            <el-col :span="16">
              <el-slider
                v-model="form.ratio"
                :min="0"
                :max="100"
                :step="1"
                :marks="{ 0: '0%', 25: '25%', 50: '50%', 75: '75%', 100: '100%' }"
              />
            </el-col>
            <el-col :span="8">
              <el-input-number
                v-model="form.ratio"
                :min="0"
                :max="100"
                style="width: 100%"
                controls-position="right"
              >
                <template #suffix><span style="color:#909399;">%</span></template>
              </el-input-number>
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item label="错误率阈值(%)" prop="errorThreshold">
          <el-input-number
            v-model="form.errorThreshold"
            :min="0"
            :max="100"
            :precision="2"
            style="width: 200px;"
            controls-position="right"
          />
          <span style="margin-left: 8px; color: #909399;">超过此比例自动回滚</span>
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
import { Plus, Promotion, WarningFilled } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'

export default {
  name: 'GrayRelease',
  data() {
    return {
      Plus,
      Promotion,
      WarningFilled,
      tableData: [],
      loading: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      dialogVisible: false,
      submitLoading: false,
      form: {
        version: '',
        whiteList: '',
        whiteListText: '',
        ratio: 10,
        errorThreshold: 5
      },
      formRules: {
        version: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
        ratio: [{ required: true, message: '请设置放量比例', trigger: 'blur' }]
      },
      globalConfig: {
        release_ratio: 0,
        error_rate_threshold: 5,
        whitelist_uids: []
      },
      whitelistText: '',
      globalSaveLoading: false
    }
  },
  computed: {
    parsedWhitelist() {
      return this.parseUids(this.whitelistText)
    }
  },
  mounted() {
    this.fetchList()
    this.fetchGlobalConfig()
  },
  methods: {
    parseUids(text) {
      if (!text) return []
      return text
        .split(/[,，\s;；\n\r\t]+/)
        .map(s => s.trim())
        .filter(s => s && /^\d+$/.test(s))
        .map(s => parseInt(s))
    },
    dedupWhitelist() {
      const uids = Array.from(new Set(this.parsedWhitelist))
      this.whitelistText = uids.join(',')
    },
    clearWhitelist() {
      this.whitelistText = ''
    },
    async fetchGlobalConfig() {
      try {
        const res = await request.get('/gray-release/config')
        const cfg = res.data || {}
        this.globalConfig = {
          release_ratio: cfg.release_ratio ?? cfg.ratio ?? 0,
          error_rate_threshold: cfg.error_rate_threshold ?? cfg.error_threshold ?? 5,
          whitelist_uids: cfg.whitelist_uids || cfg.whitelist || []
        }
        this.whitelistText = (this.globalConfig.whitelist_uids || []).join(',')
      } catch (err) {
        console.warn('加载全局灰度配置失败:', err)
      }
    },
    async saveGlobalConfig() {
      const whitelist = this.parsedWhitelist
      this.globalSaveLoading = true
      try {
        const payload = {
          release_ratio: this.globalConfig.release_ratio ?? 0,
          error_rate_threshold: this.globalConfig.error_rate_threshold ?? 5,
          whitelist_uids: whitelist
        }
        await request.post('/gray-release/config', payload)
        ElMessage.success('全局配置已保存并立即生效')
        this.fetchGlobalConfig()
      } catch (err) {
        ElMessage.error(err.message || '保存失败')
      } finally {
        this.globalSaveLoading = false
      }
    },
    async fetchList() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          pageSize: this.pagination.pageSize
        }
        const res = await request.get('/gray-release', { params })
        this.tableData = res.data?.list || []
        this.pagination.total = res.data?.total || 0
      } catch (err) {
        console.error('获取灰度发布列表失败:', err)
      } finally {
        this.loading = false
      }
    },
    handleAdd() {
      this.form = {
        version: '',
        whiteList: '',
        whiteListText: '',
        ratio: 10,
        errorThreshold: 5
      }
      this.dialogVisible = true
    },
    async handleSubmit() {
      const valid = await this.$refs.formRef.validate().catch(() => false)
      if (!valid) return
      this.submitLoading = true
      try {
        const uids = this.parseUids(this.form.whiteListText)
        await request.post('/gray-release', {
          version: this.form.version,
          whiteList: uids,
          ratio: this.form.ratio,
          error_threshold: this.form.errorThreshold
        })
        ElMessage.success('创建成功')
        this.dialogVisible = false
        this.fetchList()
      } catch (err) {
        console.error('创建失败:', err)
        ElMessage.error(err.message || '创建失败')
      } finally {
        this.submitLoading = false
      }
    },
    async handlePublish(row) {
      try {
        await ElMessageBox.confirm(
          `确定要发布版本「${row.version}」吗？`,
          '发布确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.put(`/gray-release/${row.id}/publish`)
        ElMessage.success('发布成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('发布失败:', err)
        }
      }
    },
    async handleFullRelease(row) {
      try {
        await ElMessageBox.confirm(
          `确定要将版本「${row.version}」全量发布吗？此操作将覆盖所有用户。`,
          '全量发布确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.put(`/gray-release/${row.id}/full`)
        ElMessage.success('全量发布成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('全量发布失败:', err)
        }
      }
    },
    async handleRollback(row) {
      try {
        await ElMessageBox.confirm(
          `确定要回滚版本「${row.version}」吗？回滚后将恢复到上一个版本。`,
          '回滚确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
        )
        await request.put(`/gray-release/${row.id}/rollback`)
        ElMessage.success('回滚成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('回滚失败:', err)
        }
      }
    },
    async handleDelete(row) {
      try {
        await ElMessageBox.confirm(
          `确定要删除版本「${row.version}」的灰度发布吗？`,
          '删除确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
        await request.delete(`/gray-release/${row.id}`)
        ElMessage.success('删除成功')
        this.fetchList()
      } catch (err) {
        if (err !== 'cancel') {
          console.error('删除失败:', err)
        }
      }
    },
    handleDialogClosed() {
      this.$refs.formRef?.resetFields()
    }
  }
}
</script>

<style lang="scss" scoped>
.page-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
}

.top-cards {
  .stat-card {
    .stat-label {
      font-size: 13px;
      color: #909399;
      margin-bottom: 8px;
    }
    .stat-value {
      font-size: 24px;
      font-weight: 700;
      color: #303133;

      &.primary { color: #409eff; }
      &.success { color: #67c23a; }
      &.warning { color: #e6a23c; }
      &.error { color: #f56c6c; }
    }
  }
}

.section-card {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
  }
  .tip-text {
    margin-top: 4px;
    font-size: 12px;
    color: #909399;
  }
  .whitelist-tip {
    margin-top: 8px;
    font-size: 13px;
    color: #606266;
  }
}

.table-card {
  .table-toolbar {
    margin-bottom: 16px;
  }
  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
