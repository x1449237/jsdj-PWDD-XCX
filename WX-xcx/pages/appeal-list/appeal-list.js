const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    appeals: [],
    page: 1,
    pageSize: 10,
    loading: false,
    noMore: false
  },

  onLoad() {
    this.loadAppeals();
  },

  onPullDownRefresh() {
    this.setData({
      page: 1,
      appeals: [],
      noMore: false
    });
    this.loadAppeals();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadAppeals();
    }
  },

  loadAppeals() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });

    const typeMap = {
      phone: '手机号申诉',
      order: '订单申诉',
      account: '账号申诉'
    };
    const statusMap = {
      0: '待处理',
      1: '处理中',
      2: '已完成',
      3: '已驳回'
    };
    const statusColorMap = {
      0: '#ff976a',
      1: '#0f3460',
      2: '#07c160',
      3: '#e94560'
    };

    request.get('/appeals', {
      page: this.data.page,
      page_size: this.data.pageSize
    }).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        typeText: typeMap[item.type] || item.type,
        statusText: statusMap[item.status] || '未知',
        statusColor: statusColorMap[item.status] || '#999999',
        create_time: util.formatRelativeTime(item.create_time)
      }));

      const appeals = this.data.appeals.concat(list);
      this.setData({
        appeals: appeals,
        page: this.data.page + 1,
        loading: false,
        noMore: list.length < this.data.pageSize
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  onAppealTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/appeal-detail/appeal-detail?id=${id}`
    });
  },

  onGoSubmit() {
    wx.navigateTo({
      url: '/pages/appeal-submit/appeal-submit'
    });
  }
});