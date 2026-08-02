const request = require('../../utils/request');

// 需求 93~98: 样式完全由后端下发,前端仅渲染,禁止本地写死颜色值
const FALLBACK_STYLE = {
  role_key: 'user',
  bubble_bg: '#FFFFFF',
  bubble_radius: 12,
  bubble_shadow: '',
  text_color: '#303133',
  text_stroke_color: '',
  text_stroke_width: 0,
  text_bold_important: 0,
  voice_wave_color: '#C0C4CC'
};

// 全局缓存:样式只拉取一次,所有消息组件实例共享(修复: 原代码每条消息都发HTTP请求)
let globalStyleMap = null;
let globalStylePromise = null;

const innerAudioContext = wx.createInnerAudioContext();

Component({
  properties: {
    message: {
      type: Object,
      value: {
        id: '', type: 'text', content: '', isSelf: false, avatar: '',
        senderRoleKey: 'user',
        senderName: '',
        time: '', isRead: false, voiceDuration: 0, imageUrl: '', voiceUrl: '',
        officialVSize: 0, vBadgeType: '', officialTag: '',
        isImportant: false, verticalMark: 0,
        avatarFrameStyle: {}
      }
    },
    extraClass: { type: String, value: '' },
    myUid: { type: String, value: '' }
  },
  data: {
    playing: false,
    roleStyle: FALLBACK_STYLE,
    avfStyle: {},
    // 修复: 预计算 style 字符串,WXML 不能调用 JS 方法,只能用 data 字段
    bubbleStyleStr: '',
    textStyleStr: '',
    voiceWaveStr: '',
    avatarFrameStr: ''
  },
  lifetimes: {
    attached() {
      this.computeStyles();
      this.ensureStylesLoaded();
    },
    detached() { innerAudioContext.stop(); }
  },
  observers: {
    'message': function () {
      this.computeStyles();
    }
  },
  methods: {
    // 修复: 全局只拉取一次样式,所有组件实例共享缓存
    ensureStylesLoaded() {
      if (globalStyleMap) {
        this.applyStyles(globalStyleMap);
        return;
      }
      if (globalStylePromise) {
        globalStylePromise.then(map => this.applyStyles(map));
        return;
      }
      globalStylePromise = request.get('/platform-im/styles/all').then(res => {
        const data = res || {};
        const bubbles = data.bubbles || [];
        const map = {};
        bubbles.forEach(s => { map[s.role_key] = s; });
        globalStyleMap = map;
        this.applyStyles(map);
        return map;
      }).catch(() => {});
    },
    applyStyles(map) {
      const rk = (this.data.message && this.data.message.senderRoleKey) || 'user';
      const r = map[rk] || FALLBACK_STYLE;
      this.setData({ roleStyle: r }, () => this.computeStyles());
    },
    // 修复: 预计算所有 style 字符串,WXML 只读 data 不能调方法
    computeStyles() {
      const s = this.data.roleStyle || FALLBACK_STYLE;
      const msg = this.data.message || {};
      const isImportant = msg.isImportant || s.text_bold_important === 1;

      // 气泡样式
      const radius = parseInt(s.bubble_radius || 12);
      const bubbleParts = [
        'background:' + (s.bubble_bg || '#fff'),
        'border-radius:' + radius + 'rpx',
        s.bubble_shadow ? ('box-shadow:' + s.bubble_shadow) : ''
      ].filter(Boolean);
      if (isImportant) bubbleParts.push('font-weight:700');

      // 文字样式
      const sw = parseFloat(s.text_stroke_width || 0);
      const textParts = ['color:' + (s.text_color || '#303133')];
      if (isImportant) {
        textParts.push('font-weight:800');
        if (sw > 0) {
          textParts.push('text-shadow:0 0 ' + (sw * 2) + 'px ' + (s.text_stroke_color || '#000'));
        }
      }
      if (sw > 0) {
        textParts.push('-webkit-text-stroke:' + sw + 'px ' + (s.text_stroke_color || 'transparent'));
      }

      // 语音波形
      const voiceStr = 'background:' + (s.voice_wave_color || '#C0C4CC');

      // 头像框
      const avf = (msg.avatarFrameStyle) || {};
      let avfStr = avf.frame_style || '';

      // 长消息行间距
      if (msg.content && msg.content.length > 60) {
        textParts.push('line-height:1.7em');
      }

      this.setData({
        bubbleStyleStr: bubbleParts.join(';'),
        textStyleStr: textParts.join(';'),
        voiceWaveStr: voiceStr,
        avatarFrameStr: avfStr
      });
    },
    onPlayVoice() {
      const { message, playing } = this.data;
      if (!message.voiceUrl) { wx.showToast({ title: '语音文件不存在', icon: 'none' }); return; }
      if (playing) { innerAudioContext.stop(); this.setData({ playing: false }); return; }
      innerAudioContext.src = message.voiceUrl;
      innerAudioContext.play();
      this.setData({ playing: true });
      innerAudioContext.onEnded(() => this.setData({ playing: false }));
      innerAudioContext.onError(() => { this.setData({ playing: false }); wx.showToast({ title: '播放失败', icon: 'none' }); });
    },
    onPreviewImage() {
      const { message } = this.data;
      if (message.imageUrl) wx.previewImage({ current: message.imageUrl, urls: [message.imageUrl] });
    },
    onLongPress() {
      const { message } = this.data;
      if (!message.isSelf) return;
      wx.showActionSheet({
        itemList: ['撤回', '复制', '删除'],
        success: (res) => {
          if (res.tapIndex === 0) this.triggerEvent('recall', { message });
          else if (res.tapIndex === 1) {
            if (message.type === 'text') wx.setClipboardData({ data: message.content, success: () => wx.showToast({ title: '已复制', icon: 'none' }) });
          } else if (res.tapIndex === 2) this.triggerEvent('delete', { message });
        }
      });
    }
  }
});
