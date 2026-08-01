const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    orderId: '',
    orderInfo: {},
    bidList: [],
    myBid: null,
    bidPrice: '',
    minBidPrice: '0.00',
    currentPrice: '0.00',
    submitting: false,
    countdown: '',
    isPlayer: false
  },

  onLoad(options) {
    const { orderId } = options;
    this.setData({ orderId });
    this.checkUserRole();
    this.loadOrderDetail();
    this.loadBidList();
    this.startCountdown();
  },

  checkUserRole() {
    request.get('/user/profile').then((res) => {
      this.setData({ isPlayer: res.is_player || false });
    }).catch(() => {});
  },

  loadOrderDetail() {
    request.get(`/orders/${this.data.orderId}`).then((res) => {
      const orderInfo = {
        ...res,
        amount_text: util.fenToYuan(res.order_amount),
        current_price_text: res.current_bid_price ? util.fenToYuan(res.current_bid_price) : util.fenToYuan(res.order_amount)
      };
      this.setData({
        orderInfo,
        currentPrice: orderInfo.current_price_text,
        minBidPrice: orderInfo.current_price_text
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    });
  },

  loadBidList() {
    request.get(`/orders/${this.data.orderId}/bids`).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        bid_price_text: util.fenToYuan(item.bid_price),
        bid_time_text: this.formatBidTime(item.bid_time)
      }));
      
      let myBid = null;
      const userId = wx.getStorageSync('userId');
      for (let i = 0; i < list.length; i++) {
        if (list[i].player_user_id == userId) {
          myBid = list[i];
          break;
        }
      }

      this.setData({ bidList: list, myBid });
    }).catch(() => {});
  },

  formatBidTime(timeStr) {
    if (!timeStr) return '';
    const date = new Date(timeStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前';
    if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前';
    return timeStr.slice(5, 16);
  },

  startCountdown() {
    this.countdownTimer = setInterval(() => {
      const endTime = new Date(this.data.orderInfo.bid_end_time || '').getTime();
      const now = Date.now();
      const diff = endTime - now;

      if (diff <= 0) {
        this.setData({ countdown: '竞价已结束' });
        clearInterval(this.countdownTimer);
        return;
      }

      const hours = Math.floor(diff / 3600000);
      const minutes = Math.floor((diff % 3600000) / 60000);
      const seconds = Math.floor((diff % 60000) / 1000);

      this.setData({
        countdown: `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
      });
    }, 1000);
  },

  onBidPriceInput(e) {
    this.setData({ bidPrice: e.detail.value });
  },

  onQuickBid(e) {
    const addAmount = e.currentTarget.dataset.amount;
    const current = parseFloat(this.data.currentPrice) || 0;
    const newPrice = (current + parseFloat(addAmount)).toFixed(2);
    this.setData({ bidPrice: newPrice });
  },

  onPlaceBid() {
    const { bidPrice, currentPrice, submitting, orderId, isPlayer } = this.data;

    if (!isPlayer) {
      wx.showToast({ title: '仅打手可参与竞价', icon: 'none' });
      return;
    }
    if (submitting) return;
    if (!bidPrice || isNaN(bidPrice)) {
      wx.showToast({ title: '请输入出价金额', icon: 'none' });
      return;
    }
    if (parseFloat(bidPrice) <= parseFloat(currentPrice)) {
      wx.showToast({ title: '出价需高于当前价格', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    request.post('/player/bid', {
      order_id: orderId,
      bid_price: bidPrice
    }).then(() => {
      wx.showToast({ title: '出价成功', icon: 'success' });
      this.setData({ bidPrice: '' });
      this.loadBidList();
      this.loadOrderDetail();
    }).catch((err) => {
      wx.showToast({ title: err.msg || '出价失败', icon: 'none' });
    }).finally(() => {
      this.setData({ submitting: false });
    });
  },

  onCancelBid() {
    wx.showModal({
      title: '取消竞价',
      content: '确定要取消此次竞价吗？',
      success: (res) => {
        if (res.confirm) {
          request.post('/player/bid/cancel', {
            order_id: this.data.orderId
          }).then(() => {
            wx.showToast({ title: '已取消', icon: 'success' });
            this.loadBidList();
          }).catch((err) => {
            wx.showToast({ title: err.msg || '取消失败', icon: 'none' });
          });
        }
      }
    });
  },

  onSelectWinner(e) {
    const bidId = e.currentTarget.dataset.bidId;
    wx.showModal({
      title: '选择中标',
      content: '确定选择该打手中标吗？',
      success: (res) => {
        if (res.confirm) {
          request.post(`/orders/${this.data.orderId}/select-winner`, {
            bid_id: bidId
          }).then(() => {
            wx.showToast({ title: '选择成功', icon: 'success' });
            this.loadOrderDetail();
            this.loadBidList();
          }).catch((err) => {
            wx.showToast({ title: err.msg || '操作失败', icon: 'none' });
          });
        }
      }
    });
  },

  onUnload() {
    if (this.countdownTimer) {
      clearInterval(this.countdownTimer);
    }
  }
});
