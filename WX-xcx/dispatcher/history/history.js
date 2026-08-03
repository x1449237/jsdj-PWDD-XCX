/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    recordList: [],
    currentStatus: 'all',
    statusList: [
      { label: '进行中', value: 'in_progress' },
      { label: '已完成', value: 'completed' },
      { label: '已取消', value: 'cancelled' }
    ],
    loading: false,
    page: 1,
    pageSize: 20,
    hasMore: true
  },

  onLoad() {
    this.loadRecords();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true, recordList: [] });
    this.loadRecords();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadRecords();
    }
  },

  onStatusFilter(e) {
    const status = e.currentTarget.dataset.status;
    this.setData({
      currentStatus: status,
      page: 1,
      hasMore: true,
      recordList: []
    });
    this.loadRecords();
  },

  loadRecords() {
    if (this.data.loading) return;
    this.setData({ loading: true });

    const params = {
      page: this.data.page,
      page_size: this.data.pageSize
    };

    if (this.data.currentStatus !== 'all') {
      params.status = this.data.currentStatus;
    }

    request.get('/dispatcher/dispatch-records', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        statusText: item.status_text || '',
        statusColor: item.status_color || '',
        amount: item.amount_text || '',
        dispatch_time: item.dispatch_time_text || item.dispatch_time || '',
        finish_time: item.finish_time_text || item.finish_time || ''
      }));

      this.setData({
        recordList: this.data.page === 1 ? list : [...this.data.recordList, ...list],
        loading: false,
        hasMore: list.length >= this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  }
});
