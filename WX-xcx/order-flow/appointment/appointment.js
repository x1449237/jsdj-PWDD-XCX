// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验（预约时间合法性等）、权限控制、折扣计算。
const request = require('../../utils/request');

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
  },

  // 纯 UI：生成时间选择器候选列(每 30 分钟一格,仅作 picker 选项骨架)
  // 实际可选时段由后端在 /orders/preview 返回 available_time_slots 时覆盖,
  // 业务规则(营业时间/玩家可用时段)在后端校验
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
      // 金额使用后端返回的 *_text 字段，前端不做分转元
      // minDate 由后端返回 min_appoint_date(权威"今天"日期),前端不调 new Date()
      this.setData({
        minDate: res.min_appoint_date || '',
        serviceInfo: {
          gameName: res.game_name || '',
          rank: res.rank || '',
          serviceName: res.service_name || '',
          duration: res.duration || 1,
          price: res.price_text || ''
        },
        playerInfo: {
          avatar: res.player_avatar || '',
          nickname: res.player_nickname || '',
          rating: res.player_rating || 0,
          orderCount: res.player_order_count || 0
        },
        totalAmount: res.total_amount_text || '0.00'
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

    // 非空提示（纯 UI 校验），预约时间合法性（晚于当前时间等）由后端校验
    if (!appointDate) {
      wx.showToast({ title: '请选择预约日期', icon: 'none' });
      return;
    }
    if (!appointTime) {
      wx.showToast({ title: '请选择预约时间', icon: 'none' });
      return;
    }

    const appointTimeStr = `${appointDate} ${appointTime}:00`;

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
