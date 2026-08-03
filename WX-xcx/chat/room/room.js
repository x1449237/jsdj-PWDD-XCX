const request = require('../../utils/request');
const util = require('../../utils/util');
const websocket = require('../../utils/websocket');

const recorderManager = wx.getRecorderManager();
const innerAudioContext = wx.createInnerAudioContext();

Page({
  data: {
    conversationId: '',
    targetName: '',
    targetAvatar: '',
    targetOnline: false,
    isPlatformOfficial: false,
    myAvatar: '',
    myUserId: '',
    messageList: [],
    inputText: '',
    inputMode: 'text',
    recording: false,
    willCancel: false,
    scrollToView: '',
    showMorePanel: false,
    showQuickCardPanel: false,
    quickCardList: [],
    loadingMore: false,
    hasMoreHistory: true,
    page: 1,
    pageSize: 20,
    voiceStartY: 0,
    currentPlayingId: null,
    // 介入状态
    interveneStatus: 0,
    showInterveneBanner: false,
    interveneBannerText: '',
    // 飞单风控警告弹窗
    showAntiFraudModal: false,
    antiFraudModalContent: '',
    // ---- 99~582 聊天窗口交互增强 ----
    targetTyping: false,              // 对方正在输入 (需求340, 341 仅文字触发,语音不触发)
    autoContVoice: false,             // 自动连读语音开关 (需求117)
    showShortPhrase: false,           // 快捷短语面板开关 (需求121)
    shortPhraseList: []               // 最多10条
  },

  onLoad(options) {
    const { conversationId, targetName } = options;
    this.setData({
      conversationId: conversationId || '',
      targetName: decodeURIComponent(targetName || '聊天')
    });

    const userInfo = wx.getStorageSync('user_info');
    if (userInfo) {
      this.setData({
        myAvatar: userInfo.avatar || '',
        myUserId: userInfo.user_id || ''
      });
    }

    this.loadConversationInfo();
    this.initRecorder();
    this.initAudio();
    this.loadMessages();
    this.initWebSocket();
    this.markAsRead();
    this.loadQuickCards();
    // 99~582 增强
    this.loadIMPreference();
    this.startTypingPolling();
  },

  onUnload() {
    this.stopVoice();
    innerAudioContext.destroy();
    websocket.off('chat_message', this.onReceiveMessage);
    websocket.off('platform_intervene', this.onPlatformIntervene);
    websocket.off('message_recall', this.onMessageRecalled);
    if (this._typingTimer) { clearInterval(this._typingTimer); this._typingTimer = null; }
  },

  loadConversationInfo() {
    request.get('/chat/conversation/detail', {
      conversation_id: this.data.conversationId
    }).then((res) => {
      this.setData({
        isPlatformOfficial: res.is_platform_official || false,
        targetAvatar: res.target_avatar || '',
        targetOnline: res.target_online || false,
        interveneStatus: res.intervene_status || 0
      });

      if (res.intervene_status === 1) {
        this.setData({
          showInterveneBanner: true,
          interveneBannerText: '平台已介入本次会话'
        });
      }
    }).catch(() => {});
  },

  initRecorder() {
    recorderManager.onStart(() => {
      console.log('录音开始');
    });

    recorderManager.onStop((res) => {
      if (this.data.willCancel) {
        this.setData({ recording: false, willCancel: false });
        return;
      }
      this.uploadVoice(res.tempFilePath, res.duration);
      this.setData({ recording: false, willCancel: false });
    });

    recorderManager.onError((err) => {
      console.error('录音错误:', err);
      wx.showToast({ title: '录音失败', icon: 'none' });
      this.setData({ recording: false, willCancel: false });
    });
  },

  initAudio() {
    innerAudioContext.onEnded(() => {
      this.setPlayingStatus(null);
    });

    innerAudioContext.onError((err) => {
      console.error('音频播放错误:', err);
      this.setPlayingStatus(null);
    });
  },

  initWebSocket() {
    websocket.on('chat_message', this.onReceiveMessage);
    websocket.on('platform_intervene', this.onPlatformIntervene);
    websocket.on('message_recall', this.onMessageRecalled);
  },

  onReceiveMessage(data) {
    if (data.conversation_id === this.data.conversationId) {
      const msg = this.formatMessage(data);
      this.setData({
        messageList: [...this.data.messageList, msg]
      });
      this.scrollToBottom();
    }
  },

  onMessageRecalled(data) {
    if (data.conversation_id === this.data.conversationId && data.msg_id) {
      this.setMessageRecalled(data.msg_id);
    }
  },

  onPlatformIntervene(data) {
    if (data.conversation_id !== this.data.conversationId) return;
    this.setData({
      interveneStatus: 1,
      showInterveneBanner: true,
      interveneBannerText: '平台已介入本次会话'
    });
  },

  markAsRead() {
    request.post('/chat/read', {
      conversation_id: this.data.conversationId
    }).catch(() => {});
  },

  loadMessages() {
    request.get('/chat/messages', {
      conversation_id: this.data.conversationId,
      page: this.data.page,
      page_size: this.data.pageSize
    }).then((res) => {
      const list = (res.list || []).map(item => this.formatMessage(item)).reverse();

      this.setData({
        messageList: this.data.page === 1 ? list : [...list, ...this.data.messageList],
        hasMoreHistory: list.length >= this.data.pageSize,
        page: this.data.page + 1
      });

      if (this.data.page === 1) {
        this.scrollToBottom();
      }
    }).catch(() => {
      this.setData({ loadingMore: false });
    });
  },

  // 纯 UI 状态包装:所有文本/大小/时间字段均由后端返回,前端不计算
  formatMessage(item) {
    return {
      ...item,
      from_self: item.from_self || false,
      playing: false,
      sensitive_blocked: item.sensitive_blocked || false,
      is_platform_official: item.is_platform_official || false,
      anti_fraud_risky: item.anti_fraud_risky || false,
      file_size_text: item.file_size_text || '',
      asr_text: item.asr_text || '',
      show_asr: false
    };
  },

  scrollToBottom() {
    if (this.data.messageList.length > 0) {
      const lastMsg = this.data.messageList[this.data.messageList.length - 1];
      this.setData({
        scrollToView: 'msg-' + lastMsg.msg_id
      });
    }
  },

  onLoadMore() {
    if (this.data.hasMoreHistory && !this.data.loadingMore) {
      this.setData({ loadingMore: true });
      this.loadMessages();
      setTimeout(() => {
        this.setData({ loadingMore: false });
      }, 1000);
    }
  },

  onBack() {
    wx.navigateBack();
  },

  onMore() {
    this.setData({ showMorePanel: !this.data.showMorePanel });
  },

  onTextInput(e) {
    const val = e.detail.value || '';
    this.setData({ inputText: val });
    // 99~582: 文字输入时上报正在输入状态 (语音录制不上报 需求340/341)
    this._throttledTypingReport = this._throttledTypingReport || throttle(() => {
      request.post('/im/typing', {
        session_id: parseInt(this.data.conversationId || '0', 10),
        typing_type: 'text'
      }).catch(() => {});
    }, 3000);
    if (val) this._throttledTypingReport();
  },

  onSendText() {
    const text = this.data.inputText.trim();
    if (!text) {
      wx.showToast({ title: '请输入消息内容', icon: 'none' });
      return;
    }

    const tempMsgId = util.generateId();
    const tempMsg = {
      msg_id: tempMsgId,
      type: 'text',
      content: text,
      from_self: true,
      sending: true
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg],
      inputText: ''
    });
    this.scrollToBottom();

    request.post('/chat/send', {
      conversation_id: this.data.conversationId,
      type: 'text',
      content: text
    }).then((res) => {
      if (res.anti_fraud_risky) {
        this.showAntiFraudWarning(res.warning_content || '');
      }
      this.updateMessageStatus(tempMsgId, res);
    }).catch((err) => {
      if (err.code === 4001) {
        this.setMessageSensitiveBlocked(tempMsgId);
      } else if (err.code === 4002) {
        this.setMessageAntiFraudBlocked(tempMsgId);
        this.showAntiFraudWarning(err.warning_content || '');
      } else {
        this.removeMessage(tempMsgId);
        wx.showToast({ title: '发送失败', icon: 'none' });
      }
    });
  },

  onSwitchMode() {
    this.setData({
      inputMode: this.data.inputMode === 'text' ? 'voice' : 'text',
      showMorePanel: false
    });
  },

  onChooseImage() {
    this.setData({ showMorePanel: false });
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['album'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        this.sendImage(tempFilePath);
      }
    });
  },

  onTakePhoto() {
    this.setData({ showMorePanel: false });
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        this.sendImage(tempFilePath);
      }
    });
  },

  sendImage(filePath) {
    const tempMsgId = util.generateId();
    const tempMsg = {
      msg_id: tempMsgId,
      type: 'image',
      image_url: filePath,
      from_self: true,
      sending: true
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/chat/upload', filePath).then((data) => {
      return request.post('/chat/send', {
        conversation_id: this.data.conversationId,
        type: 'image',
        image_url: data.url
      }).then((res) => {
        this.updateMessageStatus(tempMsgId, res);
      });
    }).catch(() => {
      this.removeMessage(tempMsgId);
      wx.showToast({ title: '发送失败', icon: 'none' });
    });
  },

  onVoiceStart(e) {
    this.setData({
      recording: true,
      willCancel: false,
      voiceStartY: e.touches[0].clientY
    });

    recorderManager.start({
      duration: 60000,
      sampleRate: 16000,
      numberOfChannels: 1,
      encodeBitRate: 48000,
      format: 'mp3'
    });
  },

  onVoiceMove(e) {
    const moveY = e.touches[0].clientY;
    const diff = this.data.voiceStartY - moveY;
    this.setData({
      willCancel: diff > 50
    });
  },

  onVoiceEnd() {
    recorderManager.stop();
  },

  stopVoice() {
    try {
      recorderManager.stop();
    } catch (e) {}
    this.setData({ recording: false, willCancel: false });
  },

  uploadVoice(filePath, duration) {
    const tempMsgId = util.generateId();
    const tempMsg = {
      msg_id: tempMsgId,
      type: 'voice',
      from_self: true,
      sending: true
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/chat/upload', filePath).then((data) => {
      return request.post('/chat/send', {
        conversation_id: this.data.conversationId,
        type: 'voice',
        voice_url: data.url,
        duration: duration
      }).then((res) => {
        this.updateMessageStatus(tempMsgId, res);
      });
    }).catch(() => {
      this.removeMessage(tempMsgId);
      wx.showToast({ title: '发送失败', icon: 'none' });
    });
  },

  onPlayVoice(e) {
    const msg = e.currentTarget.dataset.msg;
    if (!msg.voice_url) return;

    if (this.data.currentPlayingId === msg.msg_id) {
      innerAudioContext.stop();
      this.setPlayingStatus(null);
      return;
    }

    if (this.data.currentPlayingId) {
      innerAudioContext.stop();
    }

    innerAudioContext.src = msg.voice_url;
    innerAudioContext.play();
    this.setPlayingStatus(msg.msg_id);
  },

  setPlayingStatus(msgId) {
    const list = this.data.messageList.map(item => ({
      ...item,
      playing: item.msg_id === msgId && msgId !== null
    }));
    this.setData({
      messageList: list,
      currentPlayingId: msgId
    });
  },

  onPreviewImage(e) {
    const url = e.currentTarget.dataset.url;
    const urls = this.data.messageList
      .filter(item => item.type === 'image' && item.image_url)
      .map(item => item.image_url);
    wx.previewImage({
      current: url,
      urls: urls
    });
  },

  onImageLoad(e) {
    // 图片加载完成后，自动滚动到底部
  },

  onLongPressMessage(e) {
    const msg = e.currentTarget.dataset.msg;
    if (!msg.from_self) return;

    // 是否可撤回由后端 can_recall 字段决定，前端不解析消息时间
    if (msg.can_recall === false) {
      wx.showToast({ title: '超过可撤回时间', icon: 'none' });
      return;
    }

    wx.showActionSheet({
      itemList: ['撤回', '删除'],
      success: (res) => {
        if (res.tapIndex === 0) {
          this.recallMessage(msg);
        } else if (res.tapIndex === 1) {
          this.deleteMessage(msg);
        }
      }
    });
  },

  recallMessage(msg) {
    wx.showModal({
      title: '确认撤回',
      content: '撤回后对方将无法看到此消息',
      success: (res) => {
        if (res.confirm) {
          request.post('/chat/recall', {
            msg_id: msg.msg_id,
            conversation_id: this.data.conversationId
          }).then(() => {
            this.setMessageRecalled(msg.msg_id);
          }).catch(() => {
            wx.showToast({ title: '撤回失败', icon: 'none' });
          });
        }
      }
    });
  },

  deleteMessage(msg) {
    wx.showModal({
      title: '确认删除',
      content: '删除后该消息将不再显示',
      success: (res) => {
        if (res.confirm) {
          this.removeMessage(msg.msg_id);
        }
      }
    });
  },

  onRequestIntervene() {
    if (this.data.interveneStatus === 1) {
      wx.showToast({ title: '平台已介入', icon: 'none' });
      return;
    }

    wx.showModal({
      title: '申请平台介入',
      content: '申请后平台方将在48小时内强行介入，确定要申请吗？',
      success: (res) => {
        if (res.confirm) {
          request.post('/chat/request-intervene', {
            conversation_id: this.data.conversationId
          }).then(() => {
            this.setData({
              interveneStatus: 1,
              showInterveneBanner: true,
              interveneBannerText: '您的申请已提交，平台方将在48小时内强行介入，请耐心等待'
            });
            wx.showToast({ title: '申请已提交', icon: 'success' });
          }).catch(() => {
            wx.showToast({ title: '申请失败', icon: 'none' });
          });
        }
      }
    });
  },

  setMessageRecalled(msgId) {
    const list = this.data.messageList.map(item => {
      if (item.msg_id === msgId) {
        return { ...item, type: 'recall', content: '', image_url: '', voice_url: '' };
      }
      return item;
    });
    this.setData({ messageList: list });
  },

  setMessageSensitiveBlocked(msgId) {
    const list = this.data.messageList.map(item => {
      if (item.msg_id === msgId) {
        return { ...item, sensitive_blocked: true };
      }
      return item;
    });
    this.setData({ messageList: list });
  },

  setMessageAntiFraudBlocked(msgId) {
    const list = this.data.messageList.map(item => {
      if (item.msg_id === msgId) {
        return { ...item, anti_fraud_risky: true };
      }
      return item;
    });
    this.setData({ messageList: list });
  },

  removeMessage(msgId) {
    const list = this.data.messageList.filter(item => item.msg_id !== msgId);
    this.setData({ messageList: list });
  },

  onChooseFile() {
    this.setData({ showMorePanel: false });
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      maxDuration: 10 * 1024 * 1024,
      success: (res) => {
        const file = res.tempFiles[0];
        if (file.size > 10 * 1024 * 1024) {
          wx.showToast({ title: '文件不能超过10M', icon: 'none' });
          return;
        }
        this.sendFile(file.path, file.name, file.size);
      }
    });
  },

  sendFile(filePath, fileName, fileSize) {
    const tempMsgId = util.generateId();
    // 临时消息 file_size_text 留空,服务端响应后由后端 file_size_text 覆盖
    const tempMsg = {
      msg_id: tempMsgId,
      type: 'file',
      file_name: fileName,
      file_size: fileSize,
      file_size_text: '',
      file_url: filePath,
      from_self: true,
      sending: true
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/chat/upload-file', filePath, {
      conversation_id: this.data.conversationId,
      file_name: fileName
    }).then((data) => {
      return request.post('/chat/send-file', {
        conversation_id: this.data.conversationId,
        file_url: data.url,
        file_name: fileName,
        file_size: fileSize,
        file_type: data.file_type || 'document'
      }).then((res) => {
        if (res.anti_fraud_risky) {
          this.showAntiFraudWarning(res.warning_content || '');
        }
        this.updateMessageStatus(tempMsgId, res);
      });
    }).catch(() => {
      this.removeMessage(tempMsgId);
      wx.showToast({ title: '发送失败', icon: 'none' });
    });
  },

  onOpenFile(e) {
    const msg = e.currentTarget.dataset.msg;
    if (!msg.file_url) return;

    wx.showLoading({ title: '加载中...' });
    wx.downloadFile({
      url: msg.file_url,
      success: (res) => {
        wx.hideLoading();
        if (res.statusCode === 200) {
          wx.openDocument({
            filePath: res.tempFilePath,
            showMenu: true,
            fail: () => {
              wx.showToast({ title: '打开失败', icon: 'none' });
            }
          });
        }
      },
      fail: () => {
        wx.hideLoading();
        wx.showToast({ title: '下载失败', icon: 'none' });
      }
    });
  },

  loadQuickCards() {
    request.get('/chat/quick-cards', {
      type: 'all'
    }).then((res) => {
      this.setData({ quickCardList: res.list || [] });
    }).catch(() => {});
  },

  onShowQuickCards() {
    this.setData({ showMorePanel: false });
    if (this.data.quickCardList.length === 0) {
      wx.showToast({ title: '暂无快捷卡片', icon: 'none' });
      return;
    }
    this.setData({ showQuickCardPanel: true });
  },

  onCloseQuickCards() {
    this.setData({ showQuickCardPanel: false });
  },

  onSendQuickCard(e) {
    const card = e.currentTarget.dataset.card;
    if (!card) return;

    this.setData({ showQuickCardPanel: false });

    const tempMsgId = util.generateId();
    const tempMsg = {
      msg_id: tempMsgId,
      type: 'card',
      card_title: card.title,
      card_content: card.content,
      card_icon: card.icon,
      card_action: card.action,
      card_params: card.params_json || {},
      from_self: true,
      sending: true
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.post('/chat/send-quick-card', {
      conversation_id: this.data.conversationId,
      card_id: card.id
    }).then((res) => {
      this.updateMessageStatus(tempMsgId, res);
    }).catch(() => {
      this.removeMessage(tempMsgId);
      wx.showToast({ title: '发送失败', icon: 'none' });
    });
  },

  onCardAction(e) {
    const msg = e.currentTarget.dataset.msg;
    if (!msg || !msg.card_action) return;

    const action = msg.card_action;
    const params = msg.card_params || {};

    switch (action) {
      case 'navigate':
        if (params.url) {
          wx.navigateTo({ url: params.url });
        }
        break;
      case 'view_price':
        wx.showToast({ title: '查看报价', icon: 'none' });
        break;
      case 'view_package':
        wx.showToast({ title: '查看套餐', icon: 'none' });
        break;
      case 'make_appointment':
        wx.showToast({ title: '预约服务', icon: 'none' });
        break;
      default:
        break;
    }
  },

  // 飞单风控警告:文案由后端 warning_content 字段直接返回,前端不做级别→文案映射
  showAntiFraudWarning(content) {
    if (!content) return;
    this.setData({
      showAntiFraudModal: true,
      antiFraudModalContent: content
    });
  },

  onCloseAntiFraudModal() {
    this.setData({ showAntiFraudModal: false });
  },

  updateMessageStatus(tempMsgId, serverMsg) {
    const list = this.data.messageList.map(item => {
      if (item.msg_id === tempMsgId) {
        return {
          ...item,
          msg_id: serverMsg.msg_id || tempMsgId,
          sending: false,
          ...serverMsg
        };
      }
      return item;
    });
    this.setData({ messageList: list });
  },

  // ===== 99~582: 个性化偏好 + 正在输入轮询 + 快捷短语 =====
  async loadIMPreference() {
    try {
      const pref = await request.get('/im/preference');
      const list = pref.short_phrase || (pref && pref.shortPhrase) || [];
      this.setData({
        autoContVoice: !!(pref.auto_cont_voice || (pref && pref.autoContVoice)),
        shortPhraseList: Array.isArray(list) ? list.slice(0, 10) : []
      });
    } catch (e) {}
  },

  // 正在输入轮询(10秒一次, 切后台停掉 340~358)
  startTypingPolling() {
    this.refreshTyping();
    this._typingTimer = setInterval(() => this.refreshTyping(), 4000);
  },

  async refreshTyping() {
    try {
      const r = await request.get('/im/' + this.data.conversationId + '/typing');
      const uids = (r && r.uids) || [];
      this.setData({ targetTyping: uids.length > 0 });
    } catch (e) {}
  },

  onToggleShortPhrase() {
    this.setData({ showShortPhrase: !this.data.showShortPhrase, showMorePanel: false });
  },

  onPasteShortPhrase(e) {
    const text = e.currentTarget.dataset.text;
    if (text == null) return;
    this.setData({ inputText: String(text), showShortPhrase: false });
  },

  onToggleAutoContVoice() {
    this.setData({ autoContVoice: !this.data.autoContVoice });
    request.post('/im/preference', { auto_cont_voice: this.data.autoContVoice ? 1 : 0 }).catch(() => {});
  }
});

// ====== 通用工具:节流 ======
function throttle(fn, wait) {
  let last = 0;
  return function() {
    const now = Date.now();
    if (now - last >= wait) {
      last = now;
      fn.apply(this, arguments);
    }
  };
}