/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');
const websocket = require('../../utils/websocket');

Page({
  data: {
    wsConnected: false,
    showNewOrderNotice: false,
    newOrderCount: 0,
    orderList: [],
    page: 1,
    pageSize: 10,
    hasMore: true,
    loading: true,
    loadingMore: false,
    isRefreshing: false,
    showRejectModal: false,
    rejectOrderInfo: {},
    confirmTimer: null
  },

  onLoad() {
    this.checkWsStatus();
    this.initWebSocket();
    this.loadOrderList();
  },

  onShow() {
    this.checkWsStatus();
  },

  onUnload() {
    this.clearAllTimers();
    websocket.off('new_order', this.handleNewOrder);
    websocket.off('order_taken', this.handleOrderTaken);
  },

  onPullDownRefresh() {
    this.refreshList();
  },

  /* ========== WebSocket ========== */
  checkWsStatus() {
    const app = getApp();
    this.setData({ wsConnected: app.globalData.wsConnected });
  },

  initWebSocket() {
    this.handleNewOrder = (data) => {
      this.setData({
        showNewOrderNotice: true,
        newOrderCount: data.count || this.data.newOrderCount + 1
      });
      wx.vibrateShort({ type: 'medium' });
    };

    this.handleOrderTaken = (data) => {
      const { orderId } = data;
      const list = this.data.orderList.filter(item => item.orderId !== orderId);
      this.setData({ orderList: list });
    };

    websocket.on('new_order', this.handleNewOrder);
    websocket.on('order_taken', this.handleOrderTaken);
  },

  reconnectWs() {
    websocket.connect();
    setTimeout(() => {
      this.checkWsStatus();
    }, 500);
  },

  closeNewOrderNotice() {
    this.setData({ showNewOrderNotice: false });
  },

  scrollToTop() {
    this.setData({ showNewOrderNotice: false });
  },

  /* ========== 数据加载 ========== */
  async loadOrderList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/player/orders/pending', {
        page: this.data.page,
        pageSize: this.data.pageSize
      });
      const list = (res.list || []).map(item => this.formatOrderItem(item));
      this.setData({
        orderList: list,
        hasMore: res.hasMore !== false,
        loading: false
      });
    } catch (err) {
      this.setData({ loading: false });
    }
  },

  async refreshList() {
    this.setData({ isRefreshing: true, page: 1 });
    try {
      const res = await request.get('/player/orders/pending', {
        page: 1,
        pageSize: this.data.pageSize
      });
      const list = (res.list || []).map(item => this.formatOrderItem(item));
      this.setData({
        orderList: list,
        hasMore: res.hasMore !== false,
        isRefreshing: false
      });
    } catch (err) {
      this.setData({ isRefreshing: false });
    }
  },

  onRefresh() {
    this.refreshList();
  },

  async onLoadMore() {
    if (!this.data.hasMore || this.data.loadingMore) return;
    this.setData({ loadingMore: true });
    try {
      const nextPage = this.data.page + 1;
      const res = await request.get('/player/orders/pending', {
        page: nextPage,
        pageSize: this.data.pageSize
      });
      const list = (res.list || []).map(item => this.formatOrderItem(item));
      this.setData({
        orderList: [...this.data.orderList, ...list],
        page: nextPage,
        hasMore: res.hasMore !== false,
        loadingMore: false
      });
    } catch (err) {
      this.setData({ loadingMore: false });
    }
  },

  formatOrderItem(item) {
    return {
      ...item,
      gameIcon: item.gameIcon || '/assets/images/default-game.png',
      playerAvatar: item.playerAvatar || '/assets/images/default-avatar.png',
      playerName: item.player_name_masked || item.playerName || '',
      serviceTypeText: item.service_type_text || '',
      amountText: item.amount_text || '',
      publishTimeText: item.relative_time_text || '',
      isConfirming: false,
      confirmCountdown: 3,
      freeCancelCount: item.freeCancelCount || 0
    };
  },

  /* ========== 接单操作 ========== */
  acceptOrder(e) {
    const { orderId } = e.currentTarget.dataset;
    const list = this.data.orderList.map(item => {
      if (item.orderId === orderId) {
        return { ...item, isConfirming: true, confirmCountdown: 3 };
      }
      return item;
    });
    this.setData({ orderList: list });
    this.startCountdown(orderId);
  },

  startCountdown(orderId) {
    this.clearAllTimers();
    let countdown = 3;
    this.data.confirmTimer = setInterval(() => {
      countdown--;
      if (countdown <= 0) {
        this.clearAllTimers();
        this.confirmAcceptOrder(orderId);
        return;
      }
      const list = this.data.orderList.map(item => {
        if (item.orderId === orderId) {
          return { ...item, confirmCountdown: countdown };
        }
        return item;
      });
      this.setData({ orderList: list });
    }, 1000);
  },

  clearAllTimers() {
    if (this.data.confirmTimer) {
      clearInterval(this.data.confirmTimer);
      this.data.confirmTimer = null;
    }
  },

  cancelAccept(e) {
    const { orderId } = e.currentTarget.dataset;
    this.clearAllTimers();
    const list = this.data.orderList.map(item => {
      if (item.orderId === orderId) {
        return { ...item, isConfirming: false };
      }
      return item;
    });
    this.setData({ orderList: list });
  },

  confirmAccept(e) {
    const { orderId } = e.currentTarget.dataset;
    this.clearAllTimers();
    this.confirmAcceptOrder(orderId);
  },

  async confirmAcceptOrder(orderId) {
    wx.showLoading({ title: '接单中...' });
    try {
      await request.post('/player/orders/accept', { orderId });
      wx.hideLoading();
      wx.showToast({ title: '接单成功', icon: 'success' });
      const list = this.data.orderList.filter(item => item.orderId !== orderId);
      this.setData({ orderList: list });
    } catch (err) {
      wx.hideLoading();
      const list = this.data.orderList.map(item => {
        if (item.orderId === orderId) {
          return { ...item, isConfirming: false };
        }
        return item;
      });
      this.setData({ orderList: list });
    }
  },

  /* ========== 拒单操作 ========== */
  rejectOrder(e) {
    const { orderId, freeCount } = e.currentTarget.dataset;
    this.setData({
      showRejectModal: true,
      rejectOrderInfo: {
        orderId,
        freeCancelCount: parseInt(freeCount) || 0
      }
    });
  },

  closeRejectModal() {
    this.setData({ showRejectModal: false });
  },

  async confirmReject() {
    const { orderId, freeCancelCount } = this.data.rejectOrderInfo;
    this.setData({ showRejectModal: false });
    wx.showLoading({ title: '处理中...' });
    try {
      await request.post('/player/orders/reject', { orderId });
      wx.hideLoading();
      wx.showToast({ title: '已拒单', icon: 'none' });
      const list = this.data.orderList.filter(item => item.orderId !== orderId);
      this.setData({ orderList: list });
    } catch (err) {
      wx.hideLoading();
    }
  },

  noop() {}
});
