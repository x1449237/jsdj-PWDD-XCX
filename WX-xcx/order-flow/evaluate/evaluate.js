const request = require('../../utils/request');

Page({
  data: {
    orderId: '',
    evaluateSuccess: false,
    starRating: {
      attitude: 0,
      skill: 0,
      communication: 0
    },
    content: '',
    selectedTags: [],
    tagList: [
      '技术过硬', '耐心指导', '态度友好',
      '效率高', '值得推荐', '性价比高',
      '声音好听', '幽默风趣', '准时守信'
    ],
    isAnonymous: false,
    submitting: false,
    inCoolingPeriod: true,
    coolingRemaining: ''
  },

  onLoad(options) {
    const { orderId } = options;
    this.setData({ orderId });
    this.checkCoolingPeriod();
  },

  checkCoolingPeriod() {
    request.get(`/orders/${this.data.orderId}/evaluate/check`).then((res) => {
      if (res.cooling_remaining > 0) {
        const hours = Math.floor(res.cooling_remaining / 3600);
        const minutes = Math.floor((res.cooling_remaining % 3600) / 60);
        this.setData({
          inCoolingPeriod: true,
          coolingRemaining: hours > 0 ? hours + '小时' + minutes + '分钟' : minutes + '分钟'
        });
      } else {
        this.setData({ inCoolingPeriod: false });
      }
    }).catch(() => {});
  },

  onStarTap(e) {
    const { type, value } = e.currentTarget.dataset;
    this.setData({
      ['starRating.' + type]: value
    });
  },

  onContentInput(e) {
    this.setData({ content: e.detail.value });
  },

  onTagTap(e) {
    const tag = e.currentTarget.dataset.tag;
    const selectedTags = [...this.data.selectedTags];
    const index = selectedTags.indexOf(tag);

    if (index > -1) {
      selectedTags.splice(index, 1);
    } else {
      if (selectedTags.length >= 5) {
        wx.showToast({ title: '最多选择5个标签', icon: 'none' });
        return;
      }
      selectedTags.push(tag);
    }

    this.setData({ selectedTags });
  },

  onToggleAnonymous(e) {
    this.setData({ isAnonymous: e.detail.value });
  },

  onSubmit() {
    const { starRating, content, selectedTags, isAnonymous, orderId } = this.data;

    if (starRating.attitude === 0) {
      wx.showToast({ title: '请为服务态度评分', icon: 'none' });
      return;
    }
    if (starRating.skill === 0) {
      wx.showToast({ title: '请为技术水平评分', icon: 'none' });
      return;
    }
    if (starRating.communication === 0) {
      wx.showToast({ title: '请为沟通效率评分', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    request.post(`/orders/${orderId}/evaluate`, {
      rating_attitude: starRating.attitude,
      rating_skill: starRating.skill,
      rating_communication: starRating.communication,
      content: content.trim(),
      tags: selectedTags,
      is_anonymous: isAnonymous
    }).then(() => {
      wx.showToast({
        title: '评价成功',
        icon: 'success',
        duration: 1200
      });
      setTimeout(() => {
        this.setData({
          submitting: false,
          evaluateSuccess: true
        });
      }, 1200);
    }).catch((err) => {
      console.error('评价提交失败:', err);
      this.setData({ submitting: false });
      this.setData({ evaluateSuccess: true });
    });
  },

  onGoLottery() {
    const { orderId } = this.data;
    wx.navigateTo({
      url: `/pages/marketing/lottery?orderId=${orderId}&source=evaluate`
    });
  },

  onBack() {
    wx.switchTab({
      url: '/order-flow/list/list',
      fail: () => {
        wx.navigateBack({ delta: 3 });
      }
    });
  }
});