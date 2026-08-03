/**
 * 架构规则：前端零业务逻辑。
 * 本页面仅负责：调用后端 API、setData 渲染后端返回字段、纯 UI 反馈。
 * 状态/类型文案、金额、时间、脱敏等均由后端返回：
 *   type_text / status_text / status_color / relative_time_text
 */
const request = require('../../utils/request');

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

    request.get('/appeals', {
      page: this.data.page,
      page_size: this.data.pageSize
    }).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        typeText: item.type_text || '',
        statusText: item.status_text || '',
        statusColor: item.status_color || '',
        relativeTimeText: item.relative_time_text || ''
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
