const request = require('../../utils/request');
const util = require('../../utils/util');
const websocket = require('../../utils/websocket');

const recorderManager = wx.getRecorderManager();
const innerAudioContext = wx.createInnerAudioContext();

Page({
  data: {
    groupId: '',
    groupName: '',
    groupType: '',
    groupTypeLabel: '',
    myAvatar: '',
    myUserId: '',
    messageList: [],
    inputText: '',
    inputMode: 'text',
    recording: false,
    willCancel: false,
    scrollToView: '',
    showGroupInfo: false,
    showMorePanel: false,
    loadingMore: false,
    hasMoreHistory: true,
    page: 1,
    pageSize: 20,
    voiceStartY: 0,
    currentPlayingId: null,
    recallTimeLimit: 120000,
    // 群信息
    memberList: [],
    announcement: '',
    isAdmin: false,
    isCreator: false,
    showAnnouncementModal: false,
    editAnnouncementText: '',
    showMuteAction: false,
    selectedMember: null,
    // 定时公告
    showSchedulePanel: false,
    scheduleList: [],
    scheduleTitle: '',
    scheduleContent: '',
    scheduleDate: '',
    scheduleTime: '',
    // 飞单风控警告弹窗
    showAntiFraudModal: false,
    antiFraudModalContent: ''
  },

  onLoad(options) {
    const { groupId, groupName } = options;
    this.setData({
      groupId: groupId || '',
      groupName: decodeURIComponent(groupName || '群聊')
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
    this.loadGroupInfo();
    this.loadMessages();
    this.initWebSocket();
  },

  onUnload() {
    this.stopVoice();
    innerAudioContext.destroy();
    websocket.off('group_chat', this.onReceiveGroupMessage);
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
    websocket.on('group_chat', this.onReceiveGroupMessage);
  },

  onReceiveGroupMessage(data) {
    if (data.group_id === this.data.groupId) {
      const msg = this.formatMessage(data);
      this.setData({
        messageList: [...this.data.messageList, msg]
      });
      this.scrollToBottom();
    }
  },

  loadGroupInfo() {
    request.get('/groups/detail', {
      group_id: this.data.groupId
    }).then((res) => {
      const typeMap = { chat: '闲聊群', welfare: '福利群', after_sale: '售后群' };
      this.setData({
        groupName: res.group_name || this.data.groupName,
        groupType: res.group_type || '',
        groupTypeLabel: typeMap[res.group_type] || '闲聊群',
        memberList: res.members || [],
        announcement: res.announcement || '',
        isAdmin: res.is_admin || false,
        isCreator: res.is_creator || false
      });
    }).catch(() => {});
  },

  loadMessages() {
    request.get('/groups/messages', {
      group_id: this.data.groupId,
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

  formatMessage(item) {
    return {
      ...item,
      from_self: item.user_id === this.data.myUserId,
      playing: false,
      sensitive_blocked: item.sensitive_blocked || false,
      anti_fraud_risky: item.anti_fraud_risky || false,
      file_size_text: item.file_size ? this.formatFileSize(item.file_size) : '',
      show_asr: false
    };
  },

  formatFileSize(bytes) {
    if (bytes < 1024) return bytes + 'B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + 'KB';
    return (bytes / (1024 * 1024)).toFixed(1) + 'MB';
  },

  getMessageTypeLabel(type) {
    const typeMap = {
      text: '',
      image: '[图片]',
      voice: '[语音]',
      system: '[系统消息]',
      announcement: '[群公告]'
    };
    return typeMap[type] || '';
  },

  scrollToBottom() {
    if (this.data.messageList.length > 0) {
      const lastMsg = this.data.messageList[this.data.messageList.length - 1];
      this.setData({
        scrollToView: 'msg-' + (lastMsg.msg_id || lastMsg.id)
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

  onTapGroupName() {
    this.setData({ showGroupInfo: !this.data.showGroupInfo });
  },

  onHideGroupInfo() {
    this.setData({ showGroupInfo: false });
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

    request.post('/groups/send-text', {
      group_id: this.data.groupId,
      content: text
    }).then((res) => {
      if (res.anti_fraud_risky) {
        this.showAntiFraudWarning(res.anti_fraud_level || 'warning');
      }
      this.updateMessageStatus(tempMsgId, res);
    }).catch((err) => {
      if (err.code === 4001) {
        this.setMessageSensitiveBlocked(tempMsgId);
      } else if (err.code === 4002) {
        this.setMessageAntiFraudBlocked(tempMsgId);
        this.showAntiFraudWarning(err.level || 'warning');
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

    request.upload('/groups/upload', filePath).then((data) => {
      return request.post('/groups/send-image', {
        group_id: this.data.groupId,
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
      duration: Math.round(duration / 1000),
      from_self: true,
      sending: true,
      user_id: this.data.myUserId
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/groups/upload', filePath).then((data) => {
      return request.post('/groups/send-voice', {
        group_id: this.data.groupId,
        voice_url: data.url,
        duration: Math.round(duration / 1000)
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

    const now = Date.now();
    const msgTime = new Date(msg.create_time).getTime();
    if (now - msgTime > 120000) {
      wx.showToast({ title: '超过2分钟无法撤回', icon: 'none' });
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
      content: '撤回后其他成员将无法看到此消息',
      success: (res) => {
        if (res.confirm) {
          request.post('/groups/recall', {
      msg_id: msg.msg_id,
      group_id: this.data.groupId
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

  onMemberAction(e) {
    const member = e.currentTarget.dataset.member;
    if (member.user_id === this.data.myUserId) return;

    const itemList = [];
    if (this.data.isAdmin || this.data.isCreator) {
      itemList.push('禁言', '移出群聊');
    }
    if (itemList.length === 0) return;

    this.setData({
      selectedMember: member,
      showMuteAction: true
    });

    wx.showActionSheet({
      itemList: itemList,
      success: (res) => {
        if (res.tapIndex === 0) {
          this.muteMember(member);
        } else if (res.tapIndex === 1) {
          this.kickMember(member);
        }
      }
    });
  },

  muteMember(member) {
    wx.showModal({
      title: '禁言成员',
      content: '确定要禁言「' + member.nickname + '」吗？',
      success: (res) => {
        if (res.confirm) {
          request.post('/groups/mute', {
            group_id: this.data.groupId,
            user_id: member.user_id
          }).then(() => {
            wx.showToast({ title: '已禁言', icon: 'success' });
            this.loadGroupInfo();
          });
        }
      }
    });
  },

  kickMember(member) {
    wx.showModal({
      title: '移出群聊',
      content: '确定要将「' + member.nickname + '」移出群聊吗？',
      success: (res) => {
        if (res.confirm) {
          request.post('/groups/kick', {
            group_id: this.data.groupId,
            user_id: member.user_id
          }).then(() => {
            wx.showToast({ title: '已移出', icon: 'success' });
            this.loadGroupInfo();
          });
        }
      }
    });
  },

  onEditAnnouncement() {
    this.setData({
      showAnnouncementModal: true,
      editAnnouncementText: this.data.announcement
    });
  },

  onAnnouncementInput(e) {
    this.setData({ editAnnouncementText: e.detail.value });
  },

  onSaveAnnouncement() {
    const text = this.data.editAnnouncementText.trim();
    request.post('/groups/announcement', {
      group_id: this.data.groupId,
      content: text
    }).then(() => {
      wx.showToast({ title: '公告已更新', icon: 'success' });
      this.setData({
        showAnnouncementModal: false,
        announcement: text
      });
    });
  },

  onCancelAnnouncement() {
    this.setData({ showAnnouncementModal: false });
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

  onChooseFile() {
    this.setData({ showMorePanel: false });
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      maxDuration: 5 * 1024 * 1024,
      success: (res) => {
        const file = res.tempFiles[0];
        if (file.size > 5 * 1024 * 1024) {
          wx.showToast({ title: '群文件不能超过5M', icon: 'none' });
          return;
        }
        this.sendFile(file.path, file.name, file.size);
      }
    });
  },

  sendFile(filePath, fileName, fileSize) {
    const tempMsgId = util.generateId();
    const tempMsg = {
      msg_id: tempMsgId,
      type: 'file',
      file_name: fileName,
      file_size: fileSize,
      file_size_text: this.formatFileSize(fileSize),
      file_url: filePath,
      from_self: true,
      sending: true,
      user_id: this.data.myUserId
    };

    this.setData({
      messageList: [...this.data.messageList, tempMsg]
    });
    this.scrollToBottom();

    request.upload('/groups/upload-file', filePath, {
      group_id: this.data.groupId,
      file_name: fileName
    }).then((data) => {
      return request.post('/groups/send-file', {
        group_id: this.data.groupId,
        file_url: data.url,
        file_name: fileName,
        file_size: fileSize,
        file_type: data.file_type || 'document'
      }).then((res) => {
        if (res.anti_fraud_risky) {
          this.showAntiFraudWarning(res.anti_fraud_level || 'warning');
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

  onManageAnnouncement() {
    this.setData({ showMorePanel: false });
    this.loadScheduleList();
    this.setData({ showSchedulePanel: true });
  },

  onCloseSchedulePanel() {
    this.setData({ showSchedulePanel: false });
  },

  loadScheduleList() {
    request.get('/groups/announcement-schedule', {
      group_id: this.data.groupId
    }).then((res) => {
      this.setData({ scheduleList: res.list || [] });
    }).catch(() => {});
  },

  onScheduleTitleInput(e) {
    this.setData({ scheduleTitle: e.detail.value });
  },

  onScheduleContentInput(e) {
    this.setData({ scheduleContent: e.detail.value });
  },

  onScheduleDateChange(e) {
    this.setData({ scheduleDate: e.detail.value });
  },

  onScheduleTimeChange(e) {
    this.setData({ scheduleTime: e.detail.value });
  },

  onCreateSchedule() {
    const title = this.data.scheduleTitle.trim();
    const content = this.data.scheduleContent.trim();
    const date = this.data.scheduleDate;
    const time = this.data.scheduleTime;

    if (!title) {
      wx.showToast({ title: '请输入公告标题', icon: 'none' });
      return;
    }
    if (!content) {
      wx.showToast({ title: '请输入公告内容', icon: 'none' });
      return;
    }
    if (!date || !time) {
      wx.showToast({ title: '请选择发送时间', icon: 'none' });
      return;
    }

    const scheduleTime = date + ' ' + time + ':00';

    request.post('/groups/announcement-schedule', {
      group_id: this.data.groupId,
      title: title,
      content: content,
      schedule_time: scheduleTime
    }).then(() => {
      wx.showToast({ title: '创建成功', icon: 'success' });
      this.setData({
        scheduleTitle: '',
        scheduleContent: '',
        scheduleDate: '',
        scheduleTime: ''
      });
      this.loadScheduleList();
    }).catch(() => {
      wx.showToast({ title: '创建失败', icon: 'none' });
    });
  },

  onDeleteSchedule(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这条定时公告吗？',
      success: (res) => {
        if (res.confirm) {
          request.post('/groups/announcement-schedule/delete', {
            id: id,
            group_id: this.data.groupId
          }).then(() => {
            wx.showToast({ title: '已删除', icon: 'success' });
            this.loadScheduleList();
          }).catch(() => {
            wx.showToast({ title: '删除失败', icon: 'none' });
          });
        }
      }
    });
  },

  showAntiFraudWarning(level) {
    let content = '检测到您发送的内容可能存在风险，为保障您的权益，请在平台内完成交易。';
    if (level === 'mute') {
      content = '警告：您发送的内容违反平台规则，已被禁言，请遵守平台规定。';
    } else if (level === 'ban') {
      content = '严重警告：您多次发送违规内容，账号已被封禁，请联系客服处理。';
    }

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