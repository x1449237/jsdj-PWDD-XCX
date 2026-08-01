const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    list: [],
    page: 1,
    limit: 20,
    total: 0,
    loading: false,
    noMore: false,
    status: '',
    statusMap: {
      1: '待接单',
      2: '已接单',
      3: '进行中',
      4: '已完成',
      5: '已取消'
    }
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadList(true);
  },

  onPullDownRefresh() {
    this.loadList(true);
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList(false);
    }
  },

  async loadList(refresh) {
    if (this.data.loading) return;
    this.setData({ loading: true });

    try {
      const page = refresh ? 1 : this.data.page;
      const res = await request.get('/club/internal-order/list', {
        club_id: this.data.clubId,
        page,
        limit: this.data.limit,
        status: this.data.status
      });

      const list = refresh ? res.list : [...this.data.list, ...res.list];
      this.setData({
        list,
        total: res.total || 0,
        page: page + 1,
        noMore: list.length >= (res.total || 0),
        loading: false
      });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }

    if (refresh) {
      wx.stopPullDownRefresh();
    }
  },

  onStatusChange(e) {
    const status = e.currentTarget.dataset.status;
    this.setData({ status, page: 1, list: [], noMore: false });
    this.loadList(true);
  },

  async handleAccept(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '确认接单',
      content: '确定要接取这个订单吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await request.post('/club/internal-order/accept', { id });
            wx.showToast({ title: '接单成功', icon: 'success' });
            this.loadList(true);
          } catch (e) {
            wx.showToast({ title: e.message || '接单失败', icon: 'none' });
          }
        }
      }
    });
  }
});
