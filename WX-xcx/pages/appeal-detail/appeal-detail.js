/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段（type_text/status_text/status_color/phone_masked/time_text/relative_time_text 等）
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数。
 * 约定：request.get(url, data) 直接 resolve 出内层 data.data，res 即数据对象本身。
 */
const request = require('../../utils/request');

Page({
  data: {
    appealId: '',
    appeal: {},
    messages: [],
    replyText: '',
    replying: false
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ appealId: options.id });
      this.loadAppealDetail();
      this.loadMessages();
    }
  },

  loadAppealDetail() {
    request.get(`/appeals/${this.data.appealId}`).then((res) => {
      this.setData({ appeal: res || {} });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ appeal: {} });
    });
  },

  loadMessages() {
    request.get(`/appeals/${this.data.appealId}/messages`).then((res) => {
      this.setData({ messages: res.list || [] });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ messages: [] });
    });
  },

  onReplyInput(e) {
    this.setData({ replyText: e.detail.value });
  },

  onSendReply() {
    const text = this.data.replyText.trim();
    if (!text) return;

    this.setData({ replying: true });

    request.post(`/appeals/${this.data.appealId}/messages`, {
      content: text
    }).then((res) => {
      const messages = this.data.messages.concat([res]);
      this.setData({
        messages: messages,
        replyText: '',
        replying: false
      });
    }).catch(() => {
      this.setData({ replying: false });
      wx.showToast({ title: '发送失败', icon: 'none' });
    });
  },

  onPreviewVideo() {
    if (!this.data.appeal.video_url) return;
    wx.previewMedia({
      sources: [{
        url: this.data.appeal.video_url,
        type: 'video'
      }]
    });
  },

  onPreviewImage(e) {
    const url = e.currentTarget.dataset.url;
    wx.previewImage({
      current: url,
      urls: this.data.appeal.images || []
    });
  }
});
