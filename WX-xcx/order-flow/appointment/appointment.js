const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    playerId: '',
    serviceId: '',
    playerInfo: {},
    serviceInfo: {},
    appointDate: '',
    appointTime: '',
    minDate: '',
    remark: '',
    totalAmount: '0.00',
    submitting: false,
    timeOptions: []
  },

  onLoad(options) {
    const { playerId, serviceId } = options;
    this.setData({ playerId, serviceId });
    this.initTimeOptions();
    this.loadPlayerInfo();
    this.calculateMinDate();
  },

  calculateMinDate() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    this.setData({ minDate: `${year}-${month}-${day}` });
  },

  initTimeOptions() {
    const options = [];
    for (let h = 0; h < 24; h++) {
      for (let m = 0; m < 60; m += 30) {
        const hour = String(h).padStart(2, '0');
        const minute = String(m).padStart(2, '0');
        options.push(`${hour}:${minute}`);
      }
    }
    this.setData({ timeOptions: options });
  },

  loadPlayerInfo() {
    request.get('/orders/preview', {
      player_id: this.data.playerId,
      service_id: this.data.serviceId
    }).then((res) => {
      this.setData({
        serviceInfo: {
          gameName: res.game_name || '',
          rank: res.rank || '',
          serviceName: res.service_name || '',
          duration: res.duration || 1,
          price: util.fenToYuan(res.price)
        },
        playerInfo: {
          avatar: res.player_avatar || '',
          nickname: res.player_nickname || '',
          rating: res.player_rating || 0,
          orderCount: res.player_order_count || 0
        },
        totalAmount: util.fenToYuan(res.total_amount)
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    });
  },

  onDateChange(e) {
    this.setData({ appointDate: e.detail.value });
  },

  onTimeChange(e) {
    const index = e.detail.value;
    this.setData({ appointTime: this.data.timeOptions[index] });
  },

  onRemarkInput(e) {
    this.setData({ remark: e.detail.value });
  },

  onSubmit() {
    const { appointDate, appointTime, remark, playerId, serviceId, submitting } = this.data;

    if (submitting) return;

    if (!appointDate) {
      wx.showToast({ title: '请选择预约日期', icon: 'none' });
      return;
    }
    if (!appointTime) {
      wx.showToast({ title: '请选择预约时间', icon: 'none' });
      return;
    }

    const appointTimeStr = `${appointDate} ${appointTime}:00`;
    const appointTimestamp = new Date(appointTimeStr).getTime();
    const now = Date.now();

    if (appointTimestamp <= now) {
      wx.showToast({ title: '预约时间需晚于当前时间', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    request.post('/orders/appointments', {
      player_id: playerId,
      service_id: serviceId,
      appoint_time: appointTimeStr,
      remark: remark.trim()
    }).then((res) => {
      const orderId = res.order_id;

      this.requestPayment(res.pay_info).then(() => {
        wx.showToast({
          title: '下单成功',
          icon: 'success',
          duration: 2000
        });

        setTimeout(() => {
          wx.redirectTo({
            url: '/order-flow/detail/detail?orderId=' + orderId
          });
        }, 2000);
      }).catch((err) => {
        if (err.errMsg && err.errMsg.indexOf('cancel') === -1) {
          wx.showToast({ title: '支付失败', icon: 'none' });
        }
        this.setData({ submitting: false });
      });
    }).catch((err) => {
      console.error('创建预约单失败:', err);
      wx.showToast({ title: err.msg || '下单失败', icon: 'none' });
      this.setData({ submitting: false });
    });
  },

  requestPayment(payInfo) {
    return new Promise((resolve, reject) => {
      wx.requestPayment({
        timeStamp: payInfo.timeStamp,
        nonceStr: payInfo.nonceStr,
        package: payInfo.package,
        signType: payInfo.signType || 'MD5',
        paySign: payInfo.paySign,
        success: resolve,
        fail: reject
      });
    });
  }
});
