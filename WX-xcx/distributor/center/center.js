/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    inviteCode: '',
    totalCommission: '0.00',
    monthCommission: '0.00',
    level1Count: 0,
    level2Count: 0
  },

  onLoad() {
    this.loadDistributorInfo();
  },

  onShow() {
    this.loadDistributorInfo();
  },

  async loadDistributorInfo() {
    try {
      const res = await request.get('/distributor/info');
      this.setData({
        inviteCode: res.inviteCode || '',
        totalCommission: res.total_commission_text || res.totalCommission || '0.00',
        monthCommission: res.month_commission_text || res.monthCommission || '0.00',
        level1Count: res.level1Count || 0,
        level2Count: res.level2Count || 0
      });
    } catch (err) {
      // 忽略错误
    }
  },

  /* ========== 复制邀请码 ========== */
  copyInviteCode() {
    if (!this.data.inviteCode) {
      wx.showToast({ title: '邀请码获取失败', icon: 'none' });
      return;
    }
    wx.setClipboardData({
      data: this.data.inviteCode,
      success: () => {
        wx.showToast({ title: '邀请码已复制', icon: 'success' });
      }
    });
  },

  /* ========== 快速入口 ========== */
  goSubordinates() {
    wx.navigateTo({
      url: '/distributor/subordinates/subordinates'
    });
  },

  goCommission() {
    wx.navigateTo({
      url: '/distributor/commission/commission'
    });
  }
});
