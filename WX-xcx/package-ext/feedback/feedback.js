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
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    showForm: false,
    typeOptions: [],
    typeValues: [],
    typeIndex: 0,
    content: '',
    imageList: []
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
    this.loadTypeOptions();
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList();
    }
  },

  async loadTypeOptions() {
    try {
      const res = await request.get('/user/feedback-types');
      const options = res.list || res || [];
      this.setData({
        typeOptions: options.map(o => o.text || o.label || ''),
        typeValues: options.map(o => o.value)
      });
    } catch (e) {}
  },

  loadList() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });
    extApi.listMyFeedbacks(this.data.page, this.data.pageSize).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        typeText: item.type_text || '',
        typeTagClass: item.type_tag_class || '',
        statusText: item.status_text || '',
        statusTagClass: item.status_tag_class || '',
        createdText: item.time_text || ''
      }));
      this.setData({
        list: this.data.list.concat(list),
        page: this.data.page + 1,
        noMore: list.length < this.data.pageSize,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  onOpenForm() {
    this.setData({
      showForm: true,
      typeIndex: 0,
      content: '',
      imageList: []
    });
  },

  onCloseForm() {
    this.setData({ showForm: false });
  },

  onTypeChange(e) {
    this.setData({ typeIndex: Number(e.detail.value) });
  },

  onContentInput(e) {
    this.setData({ content: e.detail.value });
  },

  onChooseImages() {
    const remain = 6 - this.data.imageList.length;
    if (remain <= 0) {
      wx.showToast({ title: '最多选择6张图片', icon: 'none' });
      return;
    }
    wx.chooseMedia({
      count: remain,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const paths = res.tempFiles.map((f) => f.tempFilePath);
        this.setData({ imageList: this.data.imageList.concat(paths) });
      }
    });
  },

  onRemoveImage(e) {
    const index = Number(e.currentTarget.dataset.index);
    const list = this.data.imageList.slice();
    list.splice(index, 1);
    this.setData({ imageList: list });
  },

  onSubmit() {
    const content = this.data.content;
    if (!content || !content.trim()) {
      wx.showToast({ title: '请填写反馈内容', icon: 'none' });
      return;
    }
    const images = this.data.imageList.join(',');
    extApi.createFeedback({
      type: this.data.typeValues[this.data.typeIndex],
      content: content.trim(),
      images
    }).then(() => {
      wx.showToast({ title: '提交成功', icon: 'success' });
      this.setData({ showForm: false, page: 1, list: [], noMore: false });
      this.loadList();
    }).catch(() => {});
  }
});
