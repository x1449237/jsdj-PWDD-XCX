const request = require('../../utils/request');
const util = require('../../utils/util');

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
    const typeMap = {
      phone: '手机号申诉',
      order: '订单申诉',
      account: '账号申诉'
    };
    const statusMap = {
      0: '待处理',
      1: '处理中',
      2: '已完成',
      3: '已驳回'
    };
    const statusColorMap = {
      0: '#ff976a',
      1: '#0f3460',
      2: '#07c160',
      3: '#e94560'
    };

    request.get(`/appeals/${this.data.appealId}`).then((res) => {
      this.setData({
        appeal: {
          ...res,
          typeText: typeMap[res.type] || res.type,
          statusText: statusMap[res.status] || '未知',
          statusColor: statusColorMap[res.status] || '#999999',
          phone: res.phone ? util.maskPhone(res.phone) : '',
          create_time: util.formatTime(res.create_time)
        }
      });
    }).catch(() => {
      this.setData({
        appeal: {
          appeal_no: 'AP' + this.data.appealId.padStart(8, '0'),
          typeText: '手机号申诉',
          status: 1,
          statusText: '处理中',
          statusColor: '#0f3460',
          phone: '138****1234',
          description: '手机号被占用，无法绑定，请帮忙处理。运营商录屏已上传，请核实。',
          create_time: '2024-01-01 12:00:00',
          video_url: '',
          images: [],
          result: ''
        }
      });
    });
  },

  loadMessages() {
    request.get(`/appeals/${this.data.appealId}/messages`).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        create_time: util.formatRelativeTime(item.create_time)
      }));
      this.setData({ messages: list });
    }).catch(() => {
      this.setData({
        messages: [
          {
            id: 1,
            sender_name: '客服',
            content: '您好，您的申诉已收到，我们正在核实中，请耐心等待。',
            create_time: '1小时前'
          }
        ]
      });
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
      const messages = this.data.messages.concat([{
        id: res.id || Date.now(),
        sender_name: '我',
        content: text,
        create_time: '刚刚'
      }]);
      this.setData({
        messages: messages,
        replyText: '',
        replying: false
      });
    }).catch(() => {
      this.setData({ replying: false });
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