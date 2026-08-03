const request = require('../../utils/request');
const util = require('../../utils/util');
const websocket = require('../../utils/websocket');

const recorderManager = wx.getRecorderManager();
const innerAudioContext = wx.createInnerAudioContext();

Page({
  data: {
    sessionId: '',
    orderSn: '',
    myAvatar: '',
    myUserId: '',
    messageList: [],
    inputText: '',
    inputMode: 'text',
    recording: false,
    willCancel: false,
    scrollToView: '',
    showMorePanel: false,
    loadingMore: false,
    hasMoreHistory: true,
    page: 1,
    pageSize: 20,
    voiceStartY: 0,
    currentPlayingId: null,
    // 介入相关
    interveneStatus: 0,
    showInterveneBanner: false,
    interveneBannerText: '',
    autoIntervene: false,
    subscribeTmplIds: 'TEMPLATE_ID_PLACEHOLDER_04',
    // 举证上传
    showEvidenceModal: false,
    evidenceDescription: '',
    evidenceFileList: [],
    // 飞单风控警告弹窗
    showAntiFraudModal: false,
    antiFraudModalContent: ''
  },

  onLoad(options) {
    const { sessionId, orderSn } = options;
    this.setData({
      sessionId: sessionId || '',
      orderSn: decodeURIComponent(orderSn || '')
    });

    const userInfo = wx.getStorageSync('user_info');
    if (userInfo) {
      this.setData({
        myAvatar: userInfo.avatar || '',
        myUserId: userInfo.user_id || ''
      });
    }

    this.initRecorder();
    this.initAudio();
    this.loadSessionDetail();
    this.loadMessages();
    this.initWebSocket();
  },

  onUnload() {
    this.stopVoice();
    innerAudioContext.destroy();
    websocket.off('after_sale', this.onReceiveAfterSaleMessage);
    websocket.off('platform_intervene', this.onPlatformIntervene);
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
    websocket.on('after_sale', this.onReceiveAfterSaleMessage);
    websocket.on('platform_intervene', this.onPlatformIntervene);
  },

  onReceiveAfterSaleMessage(data) {
    if (data.session_id === this.data.sessionId) {
      const msg = this.formatMessage(data);
      this.setData({
        messageList: [...this.data.messageList, msg]
      });
      this.scrollToBottom();
    }
  },

  onPlatformIntervene(data) {
    if (data.session_id !== this.data.sessionId) return;

    this.setData({
      interveneStatus: 1
    });

    // 触发文案由后端按当前用户身份返回 system_message / banner_text,
    // 前端不再依据 myRole 分支判断
    if (data.trigger_type === 'keyword') {
      if (data.system_message) {
        this.addSystemMessage(data.system_message);
      }
      if (data.banner_text) {
        this.setData({
          showInterveneBanner: true,
          autoIntervene: true,
          interveneBannerText: data.banner_text
        });
      }
    }
  },

  loadSessionDetail() {
    request.get('/after-sales/detail', {
      session_id: this.data.sessionId
    }).then((res) => {
      const patch = {
        orderSn: res.order_sn || this.data.orderSn,
        interveneStatus: res.intervene_status || 0
      };
      // 介入横幅文案由后端按当前用户身份返回,前端不做角色判断
      if (res.intervene_status === 1 && res.intervene_banner_text) {
        patch.showInterveneBanner = true;
        patch.interveneBannerText = res.intervene_banner_text;
      }
      this.setData(patch);
    }).catch(() => {});
  },

  loadMessages() {
    request.get('/after-sales/messages', {
      session_id: this.data.sessionId,
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
      from_self: item.user_id === this.data.myUserId,
      playing: false,
      sensitive_blocked: item.sensitive_blocked || false,
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
        scrollToView: 'msg-' + (lastMsg.msg_id || lastMsg.id)
      });
    }
  },

  addSystemMessage(content) {
    const sysMsg = {
      msg_id: 'sys-' + util.generateId(),
      type: 'system',
      content: content,
      from_self: false
    };
    this.setData({
      messageList: [...this.data.messageList, sysMsg]
    });
    this.scrollToBottom();
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
    this.setData({ inputText: e.detail.value });
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
      sending: true,
      user_id: this.data.myUserId
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg],
      inputText: ''
    });
    this.scrollToBottom();

    request.post('/after-sales/send-text', {
      session_id: this.data.sessionId,
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
      sending: true,
      user_id: this.data.myUserId
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/after-sales/upload', filePath).then((data) => {
      return request.post('/after-sales/send-image', {
        session_id: this.data.sessionId,
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
      sending: true,
      user_id: this.data.myUserId
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/after-sales/upload', filePath).then((data) => {
      return request.post('/after-sales/send-voice', {
        session_id: this.data.sessionId,
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
          request.post('/after-sales/recall', {
            msg_id: msg.msg_id,
            session_id: this.data.sessionId
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
          request.post('/after-sales/request-intervene', {
            session_id: this.data.sessionId
          }).then((res2) => {
            const patch = { interveneStatus: 1 };
            // 申请后文案由后端按当前用户身份返回,前端不做角色判断
            if (res2.system_message) {
              this.addSystemMessage(res2.system_message);
            }
            if (res2.banner_text) {
              patch.showInterveneBanner = true;
              patch.interveneBannerText = res2.banner_text;
            }
            this.setData(patch);
            wx.showToast({ title: '申请已提交', icon: 'success' });
          }).catch(() => {
            wx.showToast({ title: '申请失败', icon: 'none' });
          });
        }
      }
    });
  },

  onApplyArbitration() {
    wx.showModal({
      title: '申请仲裁',
      content: '进入仲裁流程后，将由平台仲裁员根据举证材料和规则进行判责，确定要申请仲裁吗？',
      success: (res) => {
        if (res.confirm) {
          wx.navigateTo({
            url: '/pages/arbitration/apply?order_id=' + this.data.orderSn + '&session_id=' + this.data.sessionId
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

  onToggleAsr(e) {
    const msg = e.currentTarget.dataset.msg;
    const list = this.data.messageList.map(item => {
      if (item.msg_id === msg.msg_id) {
        return { ...item, show_asr: !item.show_asr };
      }
      return item;
    });
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
      sending: true,
      user_id: this.data.myUserId
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/after-sales/upload-file', filePath, {
      session_id: this.data.sessionId,
      file_name: fileName
    }).then((data) => {
      return request.post('/after-sales/send-file', {
        session_id: this.data.sessionId,
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

  onUploadEvidence() {
    this.setData({
      showMorePanel: false,
      showEvidenceModal: true,
      evidenceDescription: '',
      evidenceFileList: []
    });
  },

  onCloseEvidenceModal() {
    this.setData({ showEvidenceModal: false });
  },

  onEvidenceDescInput(e) {
    this.setData({ evidenceDescription: e.detail.value });
  },

  onChooseEvidenceFile() {
    wx.chooseMessageFile({
      count: 5 - this.data.evidenceFileList.length,
      type: 'all',
      success: (res) => {
        // 文件大小文本由后端在上传响应中返回 size_text,本地不格式化
        const newFiles = res.tempFiles.map(f => ({
          name: f.name,
          path: f.path,
          size: f.size,
          size_text: ''
        }));
        this.setData({
          evidenceFileList: [...this.data.evidenceFileList, ...newFiles].slice(0, 5)
        });
      }
    });
  },

  onRemoveEvidenceFile(e) {
    const index = e.currentTarget.dataset.index;
    const list = [...this.data.evidenceFileList];
    list.splice(index, 1);
    this.setData({ evidenceFileList: list });
  },

  onSubmitEvidence() {
    const fileList = this.data.evidenceFileList;
    if (fileList.length === 0) {
      wx.showToast({ title: '请至少上传一个凭证文件', icon: 'none' });
      return;
    }

    wx.showLoading({ title: '提交中...' });

    const uploadPromises = fileList.map(file => {
      return request.upload('/after-sales/upload-evidence', file.path, {
        session_id: this.data.sessionId,
        description: this.data.evidenceDescription
      });
    });

    Promise.all(uploadPromises).then(() => {
      wx.hideLoading();
      wx.showToast({ title: '举证提交成功', icon: 'success' });
      this.setData({ showEvidenceModal: false });
    }).catch(() => {
      wx.hideLoading();
      wx.showToast({ title: '提交失败，请重试', icon: 'none' });
    });
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

  removeMessage(msgId) {
    const list = this.data.messageList.filter(item => item.msg_id !== msgId);
    this.setData({ messageList: list });
  },

  onSubscribeResult(e) {
    console.log('售后页订阅消息授权结果:', e.detail);
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
  }
});