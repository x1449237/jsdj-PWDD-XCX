// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验（宵禁/限额等）、权限控制、折扣计算。
const request = require('../../utils/request');

Page({
  data: {
    orderId: '',
    playerInfo: {},
    // 预设金额/快捷金额由后端下发(避免前端硬编码金额配置)
    presetAmounts: [],
    quickAmounts: [],
    selectedAmount: 0,
    customAmount: '',
    isCustomAmount: false,
    finalAmount: '',
    message: '',
    paying: false,
    isMinor: false
  },

  onLoad(options) {
    const { orderId } = options;
    this.setData({ orderId });
    this.loadPlayerInfo();
    this.checkUserAge();
  },

  // 仅拉取未成年人标识用于展示，宵禁等业务校验由后端在打赏接口拦截
  checkUserAge() {
    request.get('/user/profile').then((res) => {
      this.setData({ isMinor: !!res.is_minor });
    }).catch(() => {});
  },

  loadPlayerInfo() {
    request.get(`/orders/${this.data.orderId}/reward/info`).then((res) => {
      this.setData({
        playerInfo: {
          avatar: res.player_avatar || '',
          nickname: res.player_nickname || ''
        },
        // 后端下发预设金额/快捷金额配置(含限额规则)
        presetAmounts: res.preset_amounts || [],
        quickAmounts: res.quick_amounts || []
      });
    }).catch(() => {});
  },

  onSelectAmount(e) {
    const amount = parseFloat(e.currentTarget.dataset.amount);
    this.setData({
      selectedAmount: amount,
      customAmount: '',
      isCustomAmount: false,
      finalAmount: amount ? String(amount) : ''
    });
  },

  onCustomAmountInput(e) {
    const value = e.detail.value;
    this.setData({
      customAmount: value,
      selectedAmount: 0,
      isCustomAmount: !!value,
      finalAmount: value
    });
  },

  onCustomFocus() {
    this.setData({
      isCustomAmount: true,
      selectedAmount: 0
    });
  },

  onMessageInput(e) {
    this.setData({ message: e.detail.value });
  },

  onReward() {
    const { selectedAmount, customAmount, isCustomAmount, message, orderId } = this.data;

    // 非空提示（纯 UI 校验），单笔上限/宵禁等业务校验由后端完成
    let amount = selectedAmount;
    if (isCustomAmount) {
      amount = parseFloat(customAmount);
      if (!amount || amount <= 0) {
        wx.showToast({ title: '请输入有效的打赏金额', icon: 'none' });
        return;
      }
    }

    if (!amount || amount <= 0) {
      wx.showToast({ title: '请选择打赏金额', icon: 'none' });
      return;
    }

    this.setData({ paying: true });

    // 直接提交元字符串，由后端 ParseYuanToFen 转换
    request.post(`/orders/${orderId}/reward`, {
      amount: String(amount),
      message: message.trim()
    }).then((res) => {
      this.requestPayment(res.pay_info).then(() => {
        wx.showToast({
          title: '打赏成功',
          icon: 'success',
          duration: 2000
        });
        setTimeout(() => {
          wx.navigateBack();
        }, 2000);
      }).catch((err) => {
        if (err.errMsg && err.errMsg.indexOf('cancel') === -1) {
          wx.showToast({ title: '支付失败', icon: 'none' });
        }
        this.setData({ paying: false });
      });
    }).catch((err) => {
      console.error('打赏失败:', err);
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
