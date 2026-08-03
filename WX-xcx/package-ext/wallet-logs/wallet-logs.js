/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（extApi）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const extApi = require('../../utils/ext-api');
const app = getApp();

Page({
  data: {
    tabs: ['全部', '收入', '支出'],
    tabIndex: 0,
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList();
    }
  },

  onTabTap(e) {
    const index = Number(e.currentTarget.dataset.index);
    if (index === this.data.tabIndex) return;
    this.setData({ tabIndex: index, page: 1, list: [], noMore: false });
    this.loadList();
  },

  loadList() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });
    const params = { page: this.data.page, size: this.data.pageSize };
    const ct = ['', 'income', 'expense'][this.data.tabIndex];
    if (ct) params.change_type = ct;

    extApi.listWalletLogs(params).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        typeText: item.change_type_text || '',
        tagClass: item.tag_class || '',
        amountText: item.amount_text || '',
        amountColor: item.amount_color || '',
        balanceText: item.balance_after_text || '',
        createdText: item.time_text || ''
      }));
      this.setData({
        list: this.data.list.concat(list),
        page: this.data.page + 1,
        noMore: list.length < this.data.pageSize,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  }
});
