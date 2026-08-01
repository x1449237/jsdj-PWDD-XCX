const request = require('../../../utils/request');

Page({
  data: {
    bindId: 0,
    month: '',
    reportData: null,
    orders: [],
    totalAmount: 0,
    orderCount: 0,
    loading: false
  },

  onLoad(options) {
    const bindId = options.bind_id ? parseInt(options.bind_id) : 0;
    const month = options.month || this.getCurrentMonth();
    this.setData({ bindId, month });
    this.loadReport();
  },

  getCurrentMonth() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    return `${year}-${month}`;
  },

  onMonthChange(e) {
    const month = e.detail.value;
    this.setData({ month });
    this.loadReport();
  },

  async loadReport() {
    const { bindId, month } = this.data;
    if (!bindId || !month) return;

    this.setData({ loading: true });
    try {
      const res = await request.get('/guardian/consume_report', {
        bind_id: bindId,
        month: month
      });
      const data = res.data || {};
      this.setData({
        reportData: data,
        orders: data.orders || [],
        totalAmount: data.total_amount || 0,
        orderCount: data.order_count || 0
      });
    } catch (err) {
      wx.showToast({ title: err.message || '加载失败', icon: 'none' });
    } finally {
      this.setData({ loading: false });
    }
  },

  onOrderTap(e) {
    const orderId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/order-flow/detail/detail?id=${orderId}`
    });
  },

  formatAmount(amount) {
    return (amount / 100).toFixed(2);
  }
});
