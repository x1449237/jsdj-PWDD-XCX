const extApi = require('../../utils/ext-api');
const app = getApp();

const STATUS_MAP = {
  0: { text: '已退', tagClass: 'tag-warning' },
  1: { text: '可用', tagClass: 'tag-success' },
  2: { text: '已过期', tagClass: 'tag-primary' }
};

const formatMoney = (v) => ((Number(v) || 0) / 100).toFixed(2);
const formatTime = (v) => v ? String(v).replace('T', ' ').slice(0, 16) : '-';

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
      const list = (res.list || []).map((item) => {
        const conf = STATUS_MAP[item.status] || { text: '未知', tagClass: 'tag-warning' };
        return {
          ...item,
          statusText: conf.text,
          tagClass: conf.tagClass,
          amountText: formatMoney(item.amount),
          balanceText: formatMoney(item.balance),
          expiredText: formatTime(item.expired_at),
          createdText: formatTime(item.created_at)
        };
      });
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
    const cents = Math.round(num * 100);
    extApi.createDeposit({ amount: cents }).then(() => {
      wx.showToast({ title: '预存成功', icon: 'success' });
      this.setData({ showForm: false, page: 1, list: [], noMore: false });
      this.loadList();
    }).catch(() => {});
  }
});
