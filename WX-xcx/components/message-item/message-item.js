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

const innerAudioContext = wx.createInnerAudioContext();

Component({
  properties: {
    message: {
      type: Object,
      value: {
        id: '', type: 'text', content: '', isSelf: false, avatar: '',
        senderRoleKey: 'user', // platform/club_admin/user/player (后端下发)
        senderName: '',
        time: '', isRead: false, voiceDuration: 0, imageUrl: '', voiceUrl: '',
        officialVSize: 0, vBadgeType: '', officialTag: '',
        isImportant: false, verticalMark: 0, // 77
        avatarFrameStyle: {} // 后端下发头像框样式 (94)
      }
    },
    // 80~92 三级头像框样式:由后端下发,前端仅套参数渲染
    extraClass: { type: String, value: '' },
    // 平台人员账号ID (比对判断身份)
    myUid: { type: String, value: '' }
  },
  data: {
    playing: false,
    // 93-98: 从后端拉取样式,存在组件内
    styleMap: {},        // { role_key: bubbleStyle }
    roleStyle: FALLBACK_STYLE,
    avfStyle: {}         // 头像框样式 (94)
  },
  lifetimes: {
    attached() {
      this.loadRoleStyles();
    },
    detached() { innerAudioContext.stop(); }
  },
  observers: {
    'message.senderRoleKey,styleMap': function (rk, map) {
      // 98: 后端未下发则强制降级为普通样式,不允许默认展示官方样式
      const r = (map && map[rk]) || FALLBACK_STYLE;
      const avf = (this.data.message && this.data.message.avatarFrameStyle) || {};
      this.setData({ roleStyle: r, avfStyle: avf });
    }
  },
  methods: {
    // 93~96: 每次进入会话必拉取最新样式,无缓存
    loadRoleStyles() {
      request.get('/platform-im/styles/all').then(res => {
        const list = res.data || [];
        const map = {};
        list.forEach(s => { map[s.role_key] = s; });
        this.setData({ styleMap: map });
      }).catch(() => {});
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
    },
    // 64~66 动态计算:气泡样式 完全由后端下发
    bubbleStyle() {
      const s = this.data.roleStyle || FALLBACK_STYLE;
      const r = parseInt(s.bubble_radius || 12);
      return [
        'background:' + (s.bubble_bg || '#fff'),
        'border-radius:' + r + 'rpx',
        'box-shadow:' + (s.bubble_shadow || 'none'),
        (s.text_bold_important === 1 || this.data.message.isImportant) ? 'font-weight:700' : ''
      ].filter(Boolean).join(';');
    },
    // 66 文字描边 + 颜色 + 行间距
    textStyle() {
      const s = this.data.roleStyle || FALLBACK_STYLE;
      const sw = parseFloat(s.text_stroke_width || 0);
      const bold = (s.text_bold_important === 1 || this.data.message.isImportant) ? 'font-weight:800' : '';
      const doubleStroke = (s.text_bold_important === 1 || this.data.message.isImportant) && sw > 0 ?
        'text-shadow:0 0 ' + (sw*2) + 'px ' + (s.text_stroke_color || '#000') : '';
      const parts = [
        'color:' + (s.text_color || '#303133'),
        bold,
        doubleStroke
      ];
      if (sw > 0) {
        // 98: 无文字描边样式降级逻辑,严格按后端下发
        parts.push('-webkit-text-stroke:' + sw + 'px ' + (s.text_stroke_color || 'transparent'));
      }
      return parts.filter(Boolean).join(';');
    },
    voiceWaveStyle() {
      const s = this.data.roleStyle || FALLBACK_STYLE;
      return 'background:' + (s.voice_wave_color || '#C0C4CC');
    }
  }
});
