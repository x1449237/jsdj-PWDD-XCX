// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验（宵禁/限额等）、权限控制、折扣计算。
const request = require('../../utils/request');

Page({
  data: {
    serviceInfo: {},
    playerInfo: {},
    totalAmount: '0.00',
    discountAmount: '0.00',
    payAmount: '0.00',
    isMinor: false,
    minorLimit: 200,
    remark: '',
    paying: false,
    playerId: '',
    serviceId: '',
    orderTypes: [],
    currentType: 'instant',
    appointDate: '',
    appointTime: '',
    minDate: '',
    timeOptions: [],
    selectedPackage: null,
    couponList: [],
    selectedCouponId: 0,
    selectedCoupon: null,
    showCouponPicker: false,
    groupBuyId: 0
  },

  onLoad(options) {
    const { playerId, serviceId, group_buy_id } = options;
    this.setData({ playerId, serviceId, groupBuyId: group_buy_id || 0 });
    this.loadOrderInfo(playerId, serviceId);
    this.checkUserAge();
    this.loadOrderTypes();
    this.initTimeOptions();
    this.loadUsableCoupons();
  },

  loadOrderTypes() {
    request.get('/orders/types').then((res) => {
      this.setData({ orderTypes: res.list || [] });
    }).catch(() => {
      // 移除 mock 回退，仅提示错误并置空
      wx.showToast({ title: '订单类型加载失败', icon: 'none' });
      this.setData({ orderTypes: [] });
    });
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

  onTypeChange(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({ currentType: type });

    if (type === 'appointment') {
      wx.navigateTo({
        url: `/order-flow/appointment/appointment?playerId=${this.data.playerId}&serviceId=${this.data.serviceId}`
      });
    }
  },

  onDateChange(e) {
    this.setData({ appointDate: e.detail.value });
  },

  onTimeChange(e) {
    const index = e.detail.value;
    this.setData({ appointTime: this.data.timeOptions[index] });
  },

  onOpenPackageList() {
    wx.navigateTo({
      url: `/order-flow/package-list/package-list`
    });
  },

  selectPackage(pkg) {
    this.setData({ selectedPackage: pkg });
  },

  loadOrderInfo(playerId, serviceId) {
    request.get('/orders/preview', {
      player_id: playerId,
      service_id: serviceId
    }).then((res) => {
      // 金额、单价均使用后端返回的 *_text 字段，前端不做分转元
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
          orderCount: res.player_order_count || 0,
          tags: res.player_tags || []
        },
        totalAmount: res.total_amount_text || '0.00',
        discountAmount: res.discount_amount_text || '0.00',
        payAmount: res.pay_amount_text || res.total_amount_text || '0.00'
      });
      this.loadUsableCoupons();
    }).catch((err) => {
      console.error('加载订单信息失败:', err);
      wx.showToast({ title: '加载失败', icon: 'none' });
    });
  },

  // 仅拉取未成年人标识用于展示提示，宵禁/消费限额等业务校验由后端在下单时拦截
  checkUserAge() {
    request.get('/user/profile').then((res) => {
      this.setData({
        isMinor: !!res.is_minor,
        minorLimit: res.minor_limit || 200
      });
    }).catch(() => {});
  },

  loadUsableCoupons() {
    if (!this.data.totalAmount || this.data.totalAmount === '0.00') return;

    request.get('/coupons/usable', {
      amount: this.data.totalAmount
    }).then((res) => {
      this.setData({
        couponList: res.list || []
      });
    }).catch(() => {});
  },

  onOpenCouponPicker() {
    if (this.data.couponList.length === 0) {
      wx.showToast({ title: '暂无可用优惠券', icon: 'none' });
      return;
    }
    this.setData({ showCouponPicker: true });
  },

  onCloseCouponPicker() {
    this.setData({ showCouponPicker: false });
  },

  onSelectCoupon(e) {
    const coupon = e.currentTarget.dataset.item;
    const prevCoupon = this.data.selectedCoupon;
    let newCoupon = null;
    let newCouponId = 0;

    if (prevCoupon && prevCoupon.id === coupon.id) {
      newCoupon = null;
      newCouponId = 0;
    } else {
      newCoupon = coupon;
      newCouponId = coupon.id;
    }

    // 折扣/应付金额由后端 preview 接口根据 coupon_id 计算返回，前端不做减法
    this.refreshPayAmount(newCouponId, newCoupon);
  },

  onNoCoupon() {
    this.refreshPayAmount(0, null);
  },

  refreshPayAmount(couponId, coupon) {
    request.get('/orders/preview', {
      player_id: this.data.playerId,
      service_id: this.data.serviceId,
      coupon_id: couponId
    }).then((res) => {
      this.setData({
        selectedCoupon: coupon,
        selectedCouponId: couponId,
        totalAmount: res.total_amount_text || this.data.totalAmount,
        discountAmount: res.discount_amount_text || '0.00',
        payAmount: res.pay_amount_text || res.total_amount_text || '0.00',
        showCouponPicker: false
      });
    }).catch(() => {
      this.setData({ showCouponPicker: false });
      wx.showToast({ title: '金额计算失败', icon: 'none' });
    });
  },

  onRemarkInput(e) {
    this.setData({ remark: e.detail.value });
  },

  onGuardianVerify() {
    wx.navigateTo({
      url: '/pages/guardian-verify/guardian-verify?amount=' + this.data.totalAmount
    });
  },

  onOpenESign() {
    wx.navigateTo({
      url: '/pages/e-sign/e-sign?playerId=' + this.data.playerId + '&serviceId=' + this.data.serviceId
    });
  },

  onPay() {
    // 未成年人宵禁/消费限额等业务校验由后端在下单接口拦截并返回错误
    this.checkAntiBoosting().then(allowed => {
      if (!allowed) return;
      this.doPay();
    });
  },

  checkAntiBoosting() {
    return new Promise((resolve) => {
      const { remark, serviceInfo, playerInfo } = this.data;
      const checkContent = remark + ' ' + (serviceInfo.title || '') + ' ' + (playerInfo.nickname || '');

      request.post('/compliance/check-anti-boosting', {
        content: checkContent,
        source: 'order'
      }).then(res => {
        if (res.blocked) {
          wx.showModal({
            title: '内容违规提醒',
            content: '检测到您的订单内容包含代练相关违规词汇，根据平台规定，禁止发布代练、外挂、上分等违规内容。请修改后重新下单。',
            showCancel: false,
            confirmText: '我知道了',
            confirmColor: '#e94560'
          });
          resolve(false);
        } else {
          resolve(true);
        }
      }).catch(() => {
        resolve(true);
      });
    });
  },

  doPay() {
    const { playerId, serviceId, remark, selectedCouponId, groupBuyId } = this.data;

    this.setData({ paying: true });

    request.post('/orders', {
      player_id: playerId,
      service_id: serviceId,
      remark: remark.trim(),
      coupon_id: selectedCouponId,
      group_buy_id: groupBuyId
    }).then((res) => {
      const orderId = res.order_id;

      this.requestPayment(res.pay_info).then(() => {
        wx.showToast({
          title: '支付成功',
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
        this.setData({ paying: false });
      });
    }).catch((err) => {
      console.error('创建订单失败:', err);
      this.setData({ paying: false });
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
