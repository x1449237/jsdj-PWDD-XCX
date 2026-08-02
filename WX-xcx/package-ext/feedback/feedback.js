const extApi = require('../../utils/ext-api');
const app = getApp();

const TYPE_OPTIONS = ['建议', '投诉', '其他'];
const TYPE_VALUES = [1, 2, 3];
const TYPE_TAG_CLASS = { 1: 'tag-success', 2: 'tag-primary', 3: 'tag-warning' };

const STATUS_MAP = {
  0: { text: '待处理', tagClass: 'tag-warning' },
  1: { text: '已回复', tagClass: 'tag-success' }
};

const formatTime = (v) => v ? String(v).replace('T', ' ').slice(0, 16) : '-';

Page({
  data: {
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    showForm: false,
    typeOptions: TYPE_OPTIONS,
    typeIndex: 0,
    content: '',
    imageList: []
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
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

  loadList() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });
    extApi.listMyFeedbacks(this.data.page, this.data.pageSize).then((res) => {
      const list = (res.list || []).map((item) => {
        const conf = STATUS_MAP[item.status] || { text: '未知', tagClass: 'tag-warning' };
        return {
          ...item,
          typeText: TYPE_OPTIONS[TYPE_VALUES.indexOf(item.type)] || '其他',
          typeTagClass: TYPE_TAG_CLASS[item.type] || 'tag-warning',
          statusText: conf.text,
          statusTagClass: conf.tagClass,
          createdText: formatTime(item.created_at)
        };
      });
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
      type: TYPE_VALUES[this.data.typeIndex],
      content: content.trim(),
      images
    }).then(() => {
      wx.showToast({ title: '提交成功', icon: 'success' });
      this.setData({ showForm: false, page: 1, list: [], noMore: false });
      this.loadList();
    }).catch(() => {});
  }
});
