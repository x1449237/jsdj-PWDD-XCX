const request = require('../../utils/request');
const util = require('../../utils/util');

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

    const statusMap = [null, 0, 2, 3, 4];
    const params = {
      page: this.data.page,
      page_size: this.data.pageSize
    };

    const status = statusMap[this.data.currentTab];
    if (status !== null) {
      params.status = status;
    }

    request.get('/orders', params).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        statusText: util.getOrderStatusText(item.status),
        statusColor: util.getOrderStatusColor(item.status),
        amount: util.fenToYuan(item.amount || item.total_fee)
      }));

      const orders = this.data.orders.concat(list);
      this.setData({
        orders: orders,
        page: this.data.page + 1,
        loading: false,
        noMore: list.length < this.data.pageSize
      });
    }).catch(() => {
      if (this.data.orders.length === 0) {
        this.setData({
          orders: this.getMockOrders(),
          loading: false
        });
      } else {
        this.setData({ loading: false });
      }
    });
  },

  getMockOrders() {
    return [
      {
        id: 1,
        order_no: '20240101000001',
        status: 0,
        statusText: '待接单',
        statusColor: '#ff976a',
        game_name: '王者荣耀',
        game_icon: '/assets/icons/game-wzry.png',
        service_name: '排位上分',
        player: null,
        amount: '50.00',
        total_fee: 5000
      },
      {
        id: 2,
        order_no: '20240101000002',
        status: 2,
        statusText: '进行中',
        statusColor: '#e94560',
        game_name: '英雄联盟',
        game_icon: '/assets/icons/game-lol.png',
        service_name: '娱乐陪玩',
        player: {
          nickname: '王者大神',
          avatar: '/assets/images/default-avatar.png',
          rank: '王者'
        },
        amount: '35.00',
        total_fee: 3500
      },
      {
        id: 3,
        order_no: '20240101000003',
        status: 4,
        statusText: '已完成',
        statusColor: '#999999',
        game_name: '和平精英',
        game_icon: '/assets/icons/game-hpjy.png',
        service_name: '教学指导',
        player: {
          nickname: '吃鸡少女',
          avatar: '/assets/images/default-avatar.png',
          rank: '超级王牌'
        },
        amount: '80.00',
        total_fee: 8000
      }
    ].filter((item) => {
      const statusMap = [null, 0, 2, 3, 4];
      const status = statusMap[this.data.currentTab];
      return status === null || item.status === status;
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