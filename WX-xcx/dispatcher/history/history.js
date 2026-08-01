const request = require('../../utils/request');
const util = require('../../utils/util');

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

    const statusTextMap = {
      in_progress: '进行中',
      completed: '已完成',
      cancelled: '已取消'
    };

    const statusColorMap = {
      in_progress: '#e94560',
      completed: '#07c160',
      cancelled: '#cccccc'
    };

    request.get('/dispatcher/dispatch-records', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        statusText: statusTextMap[item.status] || '未知',
        statusColor: statusColorMap[item.status] || '#999999',
        amount: util.fenToYuan(item.amount),
        dispatch_time: util.formatTime(item.dispatch_time, 'YYYY-MM-DD HH:mm'),
        finish_time: item.finish_time ? util.formatTime(item.finish_time, 'YYYY-MM-DD HH:mm') : ''
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