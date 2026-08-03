// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验、权限控制、折扣计算、建议出价计算。
const request = require('../../utils/request');

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
    this.loadOrderDetail();
    this.loadBidList();
  },

  loadOrderDetail() {
    request.get(`/orders/${this.data.orderId}`).then((res) => {
      // 金额/当前价/最低出价均使用后端返回的 *_text 字段
      const currentPriceText = res.current_price_text || res.amount_text || '';
      this.setData({
        orderInfo: res,
        currentPrice: currentPriceText,
        minBidPrice: res.min_bid_price_text || currentPriceText,
        // is_player 仅作 UI 提示，竞价权限由后端在 /player/bid 接口拦截
        isPlayer: !!res.is_player
      });
      this.startCountdown();
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    });
  },

  loadBidList() {
    request.get(`/orders/${this.data.orderId}/bids`).then((res) => {
      // 直接使用后端返回的 bid_price_text / bid_time_text，my_bid 由后端识别当前用户返回
      this.setData({
        bidList: res.list || [],
        myBid: res.my_bid || null
      });
    }).catch(() => {});
  },

  // 纯 UI 计时：以后端 remain_seconds 为初始值递减用于平滑显示；
  // 初始文本使用后端 remain_time_text；到期/自动取消等业务由后端处理。
  startCountdown() {
    if (this.countdownTimer) {
      clearInterval(this.countdownTimer);
    }
    let remainSeconds = this.data.orderInfo.remain_seconds || 0;
    if (remainSeconds <= 0) {
      this.setData({ countdown: this.data.orderInfo.remain_time_text || '竞价已结束' });
      return;
    }
    this.setData({ countdown: this.data.orderInfo.remain_time_text || this.formatCountdown(remainSeconds) });
    this.countdownTimer = setInterval(() => {
      remainSeconds--;
      if (remainSeconds <= 0) {
        this.setData({ countdown: '竞价已结束' });
        clearInterval(this.countdownTimer);
        return;
      }
      this.setData({ countdown: this.formatCountdown(remainSeconds) });
    }, 1000);
  },

  // 纯 UI 格式化：将秒数格式化为 HH:MM:SS 供倒计时显示
  formatCountdown(totalSeconds) {
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
  },

  onBidPriceInput(e) {
    this.setData({ bidPrice: e.detail.value });
  },

  // 快捷加价：建议出价由后端计算返回 suggested_bid_text，前端不做加法
  onQuickBid(e) {
    const addAmount = e.currentTarget.dataset.amount;
    request.get(`/orders/${this.data.orderId}/bid/preview`, {
      add_amount: addAmount
    }).then((res) => {
      this.setData({ bidPrice: res.suggested_bid_text || '' });
    }).catch(() => {
      wx.showToast({ title: '获取建议出价失败', icon: 'none' });
    });
  },

  onPlaceBid() {
    const { bidPrice, submitting, orderId } = this.data;

    if (submitting) return;
    // 非空提示（纯 UI 校验），出价上限/低于当前价/打手身份等业务校验由后端完成
    if (!bidPrice) {
      wx.showToast({ title: '请输入出价金额', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    // 直接提交元字符串，由后端转换与校验
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
