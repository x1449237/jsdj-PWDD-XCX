const request = require('../../utils/request');
const util = require('../../utils/util');

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
    this.calculateMinDate();
    this.loadUsableCoupons();
  },

  loadOrderTypes() {
    request.get('/orders/types').then((res) => {
      this.setData({ orderTypes: res.list || [] });
    }).catch(() => {
      this.setData({
        orderTypes: [
          { type: 'instant', name: '即时单', icon: '⚡' },
          { type: 'appointment', name: '预约单', icon: '📅' },
          { type: 'team', name: '车队单', icon: '👥' },
          { type: 'teaching', name: '教学单', icon: '📚' }
        ]
      });
    });
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
      const totalAmount = util.fenToYuan(res.total_amount);
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
          orderCount: res.player_order_count || 0,
          tags: res.player_tags || []
        },
        totalAmount: totalAmount,
        payAmount: totalAmount
      });
      this.loadUsableCoupons();
    }).catch((err) => {
      console.error('加载订单信息失败:', err);
      wx.showToast({ title: '加载失败', icon: 'none' });
    });
  },

  checkUserAge() {
    request.get('/user/profile').then((res) => {
      if (res.is_minor) {
        this.setData({
          isMinor: true,
          minorLimit: res.minor_limit || 200
        });
      }
    }).catch(() => {});
  },

  loadUsableCoupons() {
    if (!this.data.totalAmount || this.data.totalAmount === '0.00') return;

    request.get('/coupons/usable', {
      amount: this.data.totalAmount
    }).then((res) => {
      this.setData({
        couponList: res.data?.list || []
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
    let discountAmount = '0.00';

    if (prevCoupon && prevCoupon.id === coupon.id) {
      newCoupon = null;
      newCouponId = 0;
    } else {
      newCoupon = coupon;
      newCouponId = coupon.id;
      discountAmount = coupon.value;
    }

    const totalAmount = parseFloat(this.data.totalAmount);
    const discount = parseFloat(discountAmount);
    const payAmount = Math.max(0, totalAmount - discount).toFixed(2);

    this.setData({
      selectedCoupon: newCoupon,
      selectedCouponId: newCouponId,
      discountAmount: discountAmount,
      payAmount: payAmount,
      showCouponPicker: false
    });
  },

  onNoCoupon() {
    const totalAmount = parseFloat(this.data.totalAmount);
    this.setData({
      selectedCoupon: null,
      selectedCouponId: 0,
      discountAmount: '0.00',
      payAmount: totalAmount.toFixed(2),
      showCouponPicker: false
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
    const { isMinor, totalAmount, minorLimit, playerId, serviceId, remark } = this.data;

    if (isMinor) {
      const currentHour = new Date().getHours();
      if (currentHour >= 22 || currentHour < 8) {
        wx.showModal({
          title: '未成年人宵禁提醒',
          content: '宵禁时间（22:00-次日08:00）未成年人无法下单，请在白天再进行操作。',
          showCancel: false,
          confirmText: '我知道了'
        });
        return;
      }
    }

    if (isMinor && parseFloat(totalAmount) > minorLimit) {
      wx.showModal({
        title: '未成年人消费限额',
        content: '您的订单金额超出未成年人消费限额，需要监护人验证后才能继续支付。',
        confirmText: '去验证',
        cancelText: '取消',
        success: (res) => {
          if (res.confirm) {
            this.onGuardianVerify();
          }
        }
      });
      return;
    }

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