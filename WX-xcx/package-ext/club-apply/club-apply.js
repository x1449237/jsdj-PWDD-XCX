const extApi = require('../../utils/ext-api');
const app = getApp();

const FIELD_OPTIONS = ['头像', '技能', '价格'];

Page({
  data: {
    tabs: ['退会申请', '请假申请', '资料变更'],
    tabIndex: 0,
    fieldOptions: FIELD_OPTIONS,
    resignForm: { club_id: '', reason: '' },
    leaveForm: { club_id: '', start_date: '', end_date: '', reason: '' },
    changeForm: { club_id: '', field: FIELD_OPTIONS[0], new_value: '', reason: '' },
    submitting: false
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
    }
  },

  onTabTap(e) {
    this.setData({ tabIndex: Number(e.currentTarget.dataset.index) });
  },

  onInput(e) {
    const form = e.currentTarget.dataset.form;
    const field = e.currentTarget.dataset.field;
    this.setData({ [`${form}.${field}`]: e.detail.value });
  },

  onDateChange(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ [`leaveForm.${field}`]: e.detail.value });
  },

  onFieldChange(e) {
    this.setData({ 'changeForm.field': FIELD_OPTIONS[Number(e.detail.value)] });
  },

  submitResign() {
    const { club_id, reason } = this.data.resignForm;
    if (!club_id) {
      wx.showToast({ title: '请输入俱乐部ID', icon: 'none' });
      return;
    }
    if (!reason || !reason.trim()) {
      wx.showToast({ title: '请填写退会原因', icon: 'none' });
      return;
    }
    this.setData({ submitting: true });
    extApi.createResignation({ club_id: String(club_id), reason: reason.trim() }).then(() => {
      this.setData({ submitting: false });
      wx.showToast({ title: '提交成功', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 800);
    }).catch(() => {
      this.setData({ submitting: false });
    });
  },

  submitLeave() {
    const { club_id, start_date, end_date, reason } = this.data.leaveForm;
    if (!club_id) {
      wx.showToast({ title: '请输入俱乐部ID', icon: 'none' });
      return;
    }
    if (!start_date) {
      wx.showToast({ title: '请选择开始日期', icon: 'none' });
      return;
    }
    if (!end_date) {
      wx.showToast({ title: '请选择结束日期', icon: 'none' });
      return;
    }
    if (!reason || !reason.trim()) {
      wx.showToast({ title: '请填写请假原因', icon: 'none' });
      return;
    }
    this.setData({ submitting: true });
    extApi.createLeave({
      club_id: String(club_id),
      start_date,
      end_date,
      reason: reason.trim()
    }).then(() => {
      this.setData({ submitting: false });
      wx.showToast({ title: '提交成功', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 800);
    }).catch(() => {
      this.setData({ submitting: false });
    });
  },

  submitChange() {
    const { club_id, field, new_value, reason } = this.data.changeForm;
    if (!club_id) {
      wx.showToast({ title: '请输入俱乐部ID', icon: 'none' });
      return;
    }
    if (!new_value || !String(new_value).trim()) {
      wx.showToast({ title: '请输入新值', icon: 'none' });
      return;
    }
    if (!reason || !reason.trim()) {
      wx.showToast({ title: '请填写变更原因', icon: 'none' });
      return;
    }
    this.setData({ submitting: true });
    extApi.createChangeRequest({
      club_id: String(club_id),
      field,
      new_value: String(new_value).trim(),
      reason: reason.trim()
    }).then(() => {
      this.setData({ submitting: false });
      wx.showToast({ title: '提交成功', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 800);
    }).catch(() => {
      this.setData({ submitting: false });
    });
  }
});
