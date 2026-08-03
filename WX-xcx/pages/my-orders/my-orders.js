/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段（status_text/status_color/amount_text 等）
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数。
 */
const request = require('../../utils/request');

Page({
  data: {
    tabs: ['全部', '待接单', '进行中', '待验收', '已完成'],
    currentTab: 0,
    orders: [],
    page: 1,
    pageSize: 10,
    loading: false,
    noMore: false
  },

  onLoad() {
    this.loadOrders();
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 1
      });
    }
  },

  onPullDownRefresh() {
    this.setData({
      page: 1,
      orders: [],
      noMore: false
    });
    this.loadOrders();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadOrders();
    }
  },

  onTabChange(e) {
    const index = e.currentTarget.dataset.index;
    if (index === this.data.currentTab) return;

    this.setData({
      currentTab: index,
      page: 1,
      orders: [],
      noMore: false
    });
    this.loadOrders();
  },

  loadOrders() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });

    // tab 索引到后端 status 过滤值的映射，仅用于构造请求参数，非展示用业务逻辑
    const tabStatusFilter = [null, 0, 2, 3, 4];
    const params = {
      page: this.data.page,
      page_size: this.data.pageSize
    };

    const status = tabStatusFilter[this.data.currentTab];
    if (status !== null) {
      params.status = status;
    }

    request.get('/orders', params).then((res) => {
      const list = res.list || [];
      const orders = this.data.orders.concat(list);
      this.setData({
        orders: orders,
        page: this.data.page + 1,
        loading: false,
        noMore: list.length < this.data.pageSize
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      if (this.data.page === 1) {
        this.setData({ orders: [], loading: false, noMore: true });
      } else {
        this.setData({ loading: false });
      }
    });
  },

  onOrderTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/package-order/order-detail/order-detail?id=${id}`
    });
  },

  onCancelOrder(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '提示',
      content: '确定取消该订单吗？',
      success: (res) => {
        if (res.confirm) {
          request.post(`/orders/${id}/cancel`).then(() => {
            wx.showToast({ title: '订单已取消', icon: 'success' });
            this.setData({ page: 1, orders: [], noMore: false });
            this.loadOrders();
          });
        }
      }
    });
  },

  onPayOrder(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/package-order/order-confirm/order-confirm?id=${id}`
    });
  },

  onConfirmOrder(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '确认验收',
      content: '确认服务已完成，进行验收？',
      success: (res) => {
        if (res.confirm) {
          request.post(`/orders/${id}/confirm`).then(() => {
            wx.showToast({ title: '验收成功', icon: 'success' });
            this.setData({ page: 1, orders: [], noMore: false });
            this.loadOrders();
          });
        }
      }
    });
  },

  onContactPlayer(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/package-message/message-detail/message-detail?order_id=${id}`
    });
  },

  onReOrder(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/package-order/order-create/order-create?reorder_id=${id}`
    });
  },

  onAppeal(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/appeal-submit/appeal-submit?order_id=${id}`
    });
  },

  onGoHome() {
    wx.switchTab({
      url: '/pages/index/index'
    });
  }
});
