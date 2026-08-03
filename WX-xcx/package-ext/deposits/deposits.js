/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（extApi）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const extApi = require('../../utils/ext-api');
const app = getApp();

Page({
  data: {
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    showForm: false,
    amount: ''
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

  loadList() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });
    extApi.listDeposits(this.data.page, this.data.pageSize).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        statusText: item.status_text || '',
        tagClass: item.tag_class || '',
        amountText: item.amount_text || '',
        balanceText: item.balance_text || '',
        expiredText: item.expired_time_text || '',
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
  },

  onOpenForm() {
    this.setData({ showForm: true, amount: '' });
  },

  onCloseForm() {
    this.setData({ showForm: false });
  },

  onAmountInput(e) {
    this.setData({ amount: e.detail.value });
  },

  onSubmit() {
    const amount = this.data.amount;
    const num = Number(amount);
    if (!amount || isNaN(num) || num <= 0) {
      wx.showToast({ title: '请输入有效金额', icon: 'none' });
      return;
    }
    extApi.createDeposit({ amount: amount }).then(() => {
      wx.showToast({ title: '预存成功', icon: 'success' });
      this.setData({ showForm: false, page: 1, list: [], noMore: false });
      this.loadList();
    }).catch(() => {});
  }
});
