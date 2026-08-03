/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request/extApi）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const extApi = require('../../utils/ext-api');
const request = require('../../utils/request');
const app = getApp();

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
    typeOptions: [],
    typeValues: [],
    typeIndex: 0
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
    this.loadTypeOptions();
    this.loadList();
  },

  onPullDownRefresh() {
    this.loadList();
    wx.stopPullDownRefresh();
  },

  async loadTypeOptions() {
    try {
      const res = await request.get('/order-ext/template-types');
      const options = res.list || res || [];
      this.setData({
        typeOptions: options.map(o => o.text || o.label || ''),
        typeValues: options.map(o => o.value)
      });
    } catch (e) {}
  },

  loadList() {
    this.setData({ loading: true });
    extApi.listOrderTemplates().then((res) => {
      const raw = Array.isArray(res) ? res : (res && res.list) || [];
      const list = raw.map((item) => ({
        ...item,
        typeText: item.type_text || ''
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
        type: this.data.typeValues[0] || 1,
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
      'form.type': this.data.typeValues[index]
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
