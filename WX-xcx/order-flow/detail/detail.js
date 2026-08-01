const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    orderId: '',
    orderInfo: {},
    timelineList: [],
    subscribeTmplIds: 'TEMPLATE_ID_PLACEHOLDER_02',
    serviceTimer: null,
    serviceDurationText: '00:00:00',
    evidenceList: [],
    isPlayer: false,
    timerInterval: null
  },

  onLoad(options) {
    const { orderId } = options;
    this.setData({ orderId });
    this.loadOrderDetail();
    this.checkUserRole();
  },

  onShow() {
    if (this.data.orderId) {
      this.loadOrderDetail();
      this.loadServiceTimer();
      this.loadEvidenceList();
    }
  },

  onUnload() {
    if (this.data.timerInterval) {
      clearInterval(this.data.timerInterval);
    }
  },

  checkUserRole() {
    request.get('/user/profile').then((res) => {
      this.setData({ isPlayer: res.is_player || false });
    }).catch(() => {});
  },

  loadServiceTimer() {
    request.get(`/orders/${this.data.orderId}/service-timer`).then((res) => {
      if (res) {
        this.setData({ serviceTimer: res });
        this.updateDurationText(res.total_seconds || 0);
        if (res.status === 1) {
          this.startTimer();
        }
      }
    }).catch(() => {});
  },

  startTimer() {
    if (this.data.timerInterval) {
      clearInterval(this.data.timerInterval);
    }
    let seconds = this.data.serviceTimer?.total_seconds || 0;
    const timer = setInterval(() => {
      seconds++;
      this.updateDurationText(seconds);
    }, 1000);
    this.setData({ timerInterval: timer });
  },

  updateDurationText(totalSeconds) {
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const text = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    this.setData({ serviceDurationText: text });
  },

  loadEvidenceList() {
    request.get(`/orders/${this.data.orderId}/evidences`).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        type_text: this.getEvidenceTypeText(item.type),
        create_time_text: util.formatTime(item.create_time, 'MM-DD HH:mm')
      }));
      this.setData({ evidenceList: list });
    }).catch(() => {});
  },

  getEvidenceTypeText(type) {
    const map = {
      'gameplay_video': '录屏',
      'rank_screenshot': '战绩截图',
      'other': '其他'
    };
    return map[type] || '其他';
  },

  onUploadEvidence() {
    wx.chooseMedia({
      count: 9,
      mediaType: ['image', 'video'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const files = res.tempFiles;
        wx.showActionSheet({
          itemList: ['录屏', '战绩截图', '其他'],
          success: (actionRes) => {
            const types = ['gameplay_video', 'rank_screenshot', 'other'];
            const type = types[actionRes.tapIndex];
            this.uploadEvidenceFiles(files, type);
          }
        });
      }
    });
  },

  uploadEvidenceFiles(files, type) {
    wx.showLoading({ title: '上传中...' });
    
    const uploadPromises = files.map(file => {
      return new Promise((resolve, reject) => {
        request.upload('/orders/evidence/upload', file.tempFilePath).then(res => {
          request.post(`/orders/${this.data.orderId}/evidence`, {
            type: type,
            file_url: res.url,
            description: ''
          }).then(resolve).catch(reject);
        }).catch(reject);
      });
    });

    Promise.all(uploadPromises).then(() => {
      wx.hideLoading();
      wx.showToast({ title: '上传成功', icon: 'success' });
      this.loadEvidenceList();
    }).catch(() => {
      wx.hideLoading();
      wx.showToast({ title: '上传失败', icon: 'none' });
    });
  },

  onPreviewImage(e) {
    const url = e.currentTarget.dataset.url;
    const urls = this.data.evidenceList
      .filter(item => item.type !== 'gameplay_video')
      .map(item => item.file_url);
    wx.previewImage({
      current: url,
      urls: urls
    });
  },

  loadOrderDetail() {
    request.get(`/orders/${this.data.orderId}`).then((res) => {
      const status = res.status;
      const orderInfo = {
        orderId: res.order_id,
        status: status,
        statusText: util.getOrderStatusText(status),
        statusColor: util.getOrderStatusColor(status),
        statusDesc: this.getStatusDesc(status),
        gameName: res.game_name || '',
        serviceName: res.service_name || '',
        rank: res.rank || '',
        amount: util.fenToYuan(res.amount),
        createTime: util.formatTime(res.create_time, 'YYYY-MM-DD HH:mm'),
        remark: res.remark || '',
        playerAvatar: res.player_avatar || '',
        playerName: res.player_name || '',
        playerRating: res.player_rating || 0,
        userAvatar: res.user_avatar || '',
        userName: res.user_name || '',
        canCancel: status === 0 || status === 1,
        canAppeal: status === 3 || status === 4
      };

      this.setData({
        orderInfo: orderInfo,
        timelineList: this.buildTimeline(res)
      });
    }).catch((err) => {
      console.error('加载订单详情失败:', err);
    });
  },

  getStatusDesc(status) {
    const descMap = {
      0: '等待打手接单',
      1: '打手已接单，准备开始服务',
      2: '服务进行中',
      3: '等待您确认完成',
      4: '订单已完成',
      5: '订单已取消',
      6: '申诉处理中'
    };
    return descMap[status] || '';
  },

  buildTimeline(res) {
    const status = res.status;
    const items = [
      { step: 1, title: '下单成功', time: util.formatTime(res.create_time, 'MM-DD HH:mm'), active: true, done: true },
      { step: 2, title: '打手接单', time: res.accept_time ? util.formatTime(res.accept_time, 'MM-DD HH:mm') : '', active: status >= 1, done: status >= 1 },
      { step: 3, title: '服务进行中', time: res.start_time ? util.formatTime(res.start_time, 'MM-DD HH:mm') : '', active: status >= 2, done: status >= 2 },
      { step: 4, title: '服务完成', time: res.finish_time ? util.formatTime(res.finish_time, 'MM-DD HH:mm') : '', active: status >= 4, done: status >= 4 }
    ];

    if (status === 5) {
      items.push({ step: 5, title: '订单已取消', time: res.cancel_time ? util.formatTime(res.cancel_time, 'MM-DD HH:mm') : '', active: true, done: true });
    } else if (status === 6) {
      items.push({ step: 5, title: '申诉中', time: res.appeal_time ? util.formatTime(res.appeal_time, 'MM-DD HH:mm') : '', active: true, done: false });
    }

    items[items.length - 1].last = true;
    return items;
  },

  onChatWithPlayer() {
    const { orderInfo } = this.data;
    wx.navigateTo({
      url: '/chat/room/room?conversationId=' + orderInfo.orderId + '&targetName=' + encodeURIComponent(orderInfo.playerName)
    });
  },

  onConfirmComplete() {
    wx.showModal({
      title: '确认完成',
      content: '确认打手已完成服务？确认后款项将结算给打手。',
      confirmText: '确认完成',
      success: (res) => {
        if (res.confirm) {
          this.confirmComplete();
        }
      }
    });
  },

  onHandleAccept() {
    wx.showModal({
      title: '确认验收',
      content: '请确认服务已完成且您已满意。验收后订单将进入T+3结算周期。',
      confirmText: '确认验收',
      success: (res) => {
        if (res.confirm) {
          this.confirmAccept();
        }
      }
    });
  },

  confirmAccept() {
    wx.showLoading({ title: '处理中...' });
    request.post(`/api/v1/orders/${this.data.orderId}/confirm`).then(() => {
      wx.hideLoading();
      wx.showToast({
        title: '验收成功，订单进入T+3结算',
        icon: 'none',
        duration: 2500
      });
      setTimeout(() => {
        wx.switchTab({
          url: '/order-flow/list/list',
          fail: () => {
            wx.navigateBack();
          }
        });
      }, 1500);
    }).catch((err) => {
      wx.hideLoading();
      console.error('验收失败:', err);
      wx.showToast({ title: err?.message || '验收失败，请稍后重试', icon: 'none' });
    });
  },

  confirmComplete() {
    request.post(`/orders/${this.data.orderId}/complete`).then(() => {
      wx.showToast({
        title: '已完成',
        icon: 'success',
        duration: 2000
      });
      this.loadOrderDetail();
    }).catch((err) => {
      console.error('确认完成失败:', err);
    });
  },

  onCancelOrder() {
    wx.showModal({
      title: '取消订单',
      content: '确认取消该订单？',
      confirmText: '确认取消',
      confirmColor: '#e94560',
      success: (res) => {
        if (res.confirm) {
          this.cancelOrder();
        }
      }
    });
  },

  cancelOrder() {
    request.post(`/orders/${this.data.orderId}/cancel`).then(() => {
      wx.showToast({
        title: '已取消',
        icon: 'success',
        duration: 2000
      });
      this.loadOrderDetail();
    }).catch((err) => {
      console.error('取消订单失败:', err);
    });
  },

  onAppeal() {
    wx.navigateTo({
      url: '/pages/appeal-submit/appeal-submit?orderId=' + this.data.orderId
    });
  },

  onGoEvaluate() {
    wx.navigateTo({
      url: '/order-flow/evaluate/evaluate?orderId=' + this.data.orderId
    });
  },

  onGoReward() {
    wx.navigateTo({
      url: '/order-flow/reward/reward?orderId=' + this.data.orderId
    });
  },

  onSubscribeResult(e) {
    console.log('订单页订阅消息授权结果:', e.detail);
  }
});