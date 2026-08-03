/**
 * 架构规则：前端零业务逻辑。
 * 本页面仅负责：调用后端 API、setData 渲染后端返回字段、纯 UI 反馈。
 * 券类型文案由后端返回 type_text，前端不做类型码到文案的映射。
 */
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
    type: ''
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
      const res = await request.get('/club/coupon/list', {
        club_id: this.data.clubId,
        page,
        limit: this.data.limit,
        type: this.data.type
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

  onTypeChange(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({ type, page: 1, list: [], noMore: false });
    this.loadList(true);
  },

  async handleReceive(e) {
    const id = e.currentTarget.dataset.id;
    try {
      await request.post('/club/coupon/receive', { id });
      wx.showToast({ title: '领取成功', icon: 'success' });
      this.loadList(true);
    } catch (e) {
      wx.showToast({ title: e.message || '领取失败', icon: 'none' });
    }
  }
});
