/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    pendingOrders: [],
    onlinePlayers: [],
    selectedOrder: null,
    selectedPlayer: null,
    loading: false,
    playerLoading: false,
    showConfirmModal: false,
    dispatching: false
  },

  onLoad() {
    this.loadPendingOrders();
  },

  onPullDownRefresh() {
    this.loadPendingOrders();
    if (this.data.selectedOrder) {
      this.loadOnlinePlayers();
    }
    wx.stopPullDownRefresh();
  },

  loadPendingOrders() {
    this.setData({ loading: true });

    request.get('/dispatcher/pending-orders').then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        amount: item.amount_text || '',
        create_time: item.create_time_text || item.create_time || ''
      }));

      this.setData({
        pendingOrders: list,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  loadOnlinePlayers() {
    this.setData({ playerLoading: true });

    request.get('/dispatcher/online-players').then((res) => {
      this.setData({
        onlinePlayers: res.list || [],
        playerLoading: false
      });
    }).catch(() => {
      this.setData({ playerLoading: false });
    });
  },

  onSelectOrder(e) {
    const order = e.currentTarget.dataset.order;
    this.setData({
      selectedOrder: order,
      selectedPlayer: null
    });
    this.loadOnlinePlayers();
  },

  onSelectPlayer(e) {
    const player = e.currentTarget.dataset.player;
    this.setData({ selectedPlayer: player });
  },

  onDispatch() {
    if (!this.data.selectedPlayer) {
      wx.showToast({ title: '请选择打手', icon: 'none' });
      return;
    }
    this.setData({ showConfirmModal: true });
  },

  onCloseModal() {
    this.setData({ showConfirmModal: false });
  },

  onConfirmDispatch() {
    const { selectedOrder, selectedPlayer } = this.data;

    this.setData({ dispatching: true });

    request.post('/dispatcher/dispatch', {
      order_id: selectedOrder.order_id,
      player_id: selectedPlayer.player_id
    }).then(() => {
      wx.showToast({
        title: '派单成功',
        icon: 'success',
        duration: 2000
      });

      this.setData({
        showConfirmModal: false,
        dispatching: false,
        selectedOrder: null,
        selectedPlayer: null
      });

      this.loadPendingOrders();
    }).catch((err) => {
      console.error('派单失败:', err);
      this.setData({ dispatching: false });
    });
  }
});
