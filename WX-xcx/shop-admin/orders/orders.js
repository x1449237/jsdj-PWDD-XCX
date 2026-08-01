const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    currentTab: 'all',
    currentStatus: -1,
    statusList: [
      { label: '待接单', value: 0 },
      { label: '已接单', value: 1 },
      { label: '进行中', value: 2 },
      { label: '待验收', value: 3 },
      { label: '已完成', value: 4 },
      { label: '已取消', value: 5 },
      { label: '申诉中', value: 6 }
    ],
    orderList: [],
    riskOrderList: [],
    loading: false,
    riskLoading: false,
    page: 1,
    pageSize: 20,
    hasMore: true,
    riskPage: 1,
    riskHasMore: true
  },

  onLoad() {
    this.checkAuth();
    this.loadOrders();
  },

  onShow() {
    this.checkAuth();
  },

  onPullDownRefresh() {
    if (this.data.currentTab === 'all') {
      this.setData({ page: 1, hasMore: true, orderList: [] });
      this.loadOrders();
    } else {
      this.setData({ riskPage: 1, riskHasMore: true, riskOrderList: [] });
      this.loadRiskOrders();
    }
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (this.data.currentTab === 'all' && this.data.hasMore) {
      this.loadOrders();
    } else if (this.data.currentTab === 'risk' && this.data.riskHasMore) {
      this.loadRiskOrders();
    }
  },

  checkAuth() {
    const shopAdminInfo = wx.getStorageSync('shop_admin_info');
    if (!shopAdminInfo || !shopAdminInfo.token) {
      wx.redirectTo({
        url: '/shop-admin/login/login'
      });
    }
  },

  onTabChange(e) {
    const tab = e.currentTarget.dataset.tab;
    this.setData({ currentTab: tab });
    if (tab === 'risk' && this.data.riskOrderList.length === 0) {
      this.loadRiskOrders();
    }
  },

  onStatusFilter(e) {
    const status = parseInt(e.currentTarget.dataset.status);
    this.setData({
      currentStatus: status,
      page: 1,
      hasMore: true,
      orderList: []
    });
    this.loadOrders();
  },

  loadOrders() {
    if (this.data.loading) return;
    this.setData({ loading: true });

    const params = {
      page: this.data.page,
      page_size: this.data.pageSize
    };

    if (this.data.currentStatus !== -1) {
      params.status = this.data.currentStatus;
    }

    request.get('/shop-admin/orders', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        statusText: util.getOrderStatusText(item.status),
        statusColor: util.getOrderStatusColor(item.status),
        amount: util.fenToYuan(item.amount),
        create_time: util.formatTime(item.create_time, 'YYYY-MM-DD HH:mm')
      }));

      this.setData({
        orderList: this.data.page === 1 ? list : [...this.data.orderList, ...list],
        loading: false,
        hasMore: list.length >= this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  loadRiskOrders() {
    if (this.data.riskLoading) return;
    this.setData({ riskLoading: true });

    request.get('/shop-admin/risk-orders', {
      page: this.data.riskPage,
      page_size: this.data.pageSize
    }).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        amount: util.fenToYuan(item.amount),
        create_time: util.formatTime(item.create_time, 'YYYY-MM-DD HH:mm')
      }));

      this.setData({
        riskOrderList: this.data.riskPage === 1 ? list : [...this.data.riskOrderList, ...list],
        riskLoading: false,
        riskHasMore: list.length >= this.data.pageSize,
        riskPage: this.data.riskPage + 1
      });
    }).catch(() => {
      this.setData({ riskLoading: false });
    });
  },

  onToggleOnline(e) {
    const { order, action } = e.currentTarget.dataset;
    const actionText = action === 'online' ? '上架' : '下架';

    wx.showModal({
      title: '确认操作',
      content: `确认${actionText}该订单？`,
      success: (res) => {
        if (res.confirm) {
          this.doToggleOnline(order.order_id, action);
        }
      }
    });
  },

  doToggleOnline(orderId, action) {
    request.post('/shop-admin/orders/toggle-online', {
      order_id: orderId,
      action: action
    }).then(() => {
      wx.showToast({
        title: action === 'online' ? '已上架' : '已下架',
        icon: 'success'
      });
      this.setData({ page: 1, hasMore: true, orderList: [] });
      this.loadOrders();
    }).catch((err) => {
      console.error('操作失败:', err);
    });
  },

  onViewDetail(e) {
    const orderId = e.currentTarget.dataset.orderId;
    wx.navigateTo({
      url: '/order-flow/detail/detail?orderId=' + orderId
    });
  }
});