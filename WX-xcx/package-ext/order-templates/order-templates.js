const extApi = require('../../utils/ext-api');
const app = getApp();

const TYPE_OPTIONS = ['上分', '教学', '车队', '其他'];
const TYPE_VALUES = [1, 2, 3, 4];

Page({
  data: {
    list: [],
    loading: false,
    showForm: false,
    form: {
      name: '',
      type: 1,
      game_zone: '',
      game_id_text: '',
      requirement: ''
    },
    typeOptions: TYPE_OPTIONS,
    typeIndex: 0
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
    this.loadList();
  },

  onPullDownRefresh() {
    this.loadList();
    wx.stopPullDownRefresh();
  },

  loadList() {
    this.setData({ loading: true });
    extApi.listOrderTemplates().then((res) => {
      const raw = Array.isArray(res) ? res : (res && res.list) || [];
      const list = raw.map((item) => ({
        ...item,
        typeText: TYPE_OPTIONS[item.type - 1] || '其他'
      }));
      this.setData({ list, loading: false });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  onOpenForm() {
    this.setData({
      showForm: true,
      form: {
        name: '',
        type: 1,
        game_zone: '',
        game_id_text: '',
        requirement: ''
      },
      typeIndex: 0
    });
  },

  onCloseForm() {
    this.setData({ showForm: false });
  },

  onTypeChange(e) {
    const index = Number(e.detail.value);
    this.setData({
      typeIndex: index,
      'form.type': TYPE_VALUES[index]
    });
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ [`form.${field}`]: e.detail.value });
  },

  onSubmit() {
    const { name, type, game_zone, game_id_text, requirement } = this.data.form;
    if (!name || !name.trim()) {
      wx.showToast({ title: '请输入模板名称', icon: 'none' });
      return;
    }
    extApi.createOrderTemplate({
      name: name.trim(),
      type,
      game_id: game_id_text || '',
      game_zone: game_zone || '',
      game_id_text: game_id_text || '',
      requirement: requirement || ''
    }).then(() => {
      wx.showToast({ title: '已添加', icon: 'success' });
      this.setData({ showForm: false });
      this.loadList();
    }).catch(() => {});
  },

  onDelete(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '提示',
      content: '确定删除该模板吗？',
      success: (res) => {
        if (!res.confirm) return;
        extApi.deleteOrderTemplate(id).then(() => {
          wx.showToast({ title: '已删除', icon: 'success' });
          this.loadList();
        }).catch(() => {});
      }
    });
  }
});
