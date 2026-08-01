const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    form: {
      type: 'achievement',
      title: '',
      content: '',
      images: []
    },
    submitting: false,
    typeOptions: [
      { value: 'achievement', label: '战绩' },
      { value: 'daily', label: '日常' },
      { value: 'event', label: '活动' }
    ]
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
  },

  onTypeChange(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({ 'form.type': type });
  },

  onTitleInput(e) {
    this.setData({ 'form.title': e.detail.value });
  },

  onContentInput(e) {
    this.setData({ 'form.content': e.detail.value });
  },

  chooseImage() {
    const that = this;
    wx.chooseMedia({
      count: 9 - this.data.form.images.length,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success(res) {
        const tempFiles = res.tempFiles.map(f => f.tempFilePath);
        that.setData({
          'form.images': [...that.data.form.images, ...tempFiles]
        });
      }
    });
  },

  removeImage(e) {
    const index = e.currentTarget.dataset.index;
    const images = [...this.data.form.images];
    images.splice(index, 1);
    this.setData({ 'form.images': images });
  },

  async handleSubmit() {
    if (!this.data.form.content) {
      wx.showToast({ title: '请输入内容', icon: 'none' });
      return;
    }
    if (this.data.submitting) return;

    this.setData({ submitting: true });
    try {
      await request.post('/club/dynamic/publish', {
        club_id: this.data.clubId,
        type: this.data.form.type,
        title: this.data.form.title,
        content: this.data.form.content,
        images: this.data.form.images
      });

      wx.showToast({ title: '发布成功，等待审核', icon: 'success' });
      setTimeout(() => {
        wx.navigateBack();
      }, 1500);
    } catch (e) {
      wx.showToast({ title: e.message || '发布失败', icon: 'none' });
    } finally {
      this.setData({ submitting: false });
    }
  }
});
