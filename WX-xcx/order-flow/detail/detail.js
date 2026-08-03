// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验、权限控制、折扣计算、时间线构建。
const request = require('../../utils/request');

Page({
  data: {
    orderId: '',
    orderInfo: {},
    timelineList: [],
    subscribeTmplIds: 'TEMPLATE_ID_PLACEHOLDER_02',
    serviceTimer: null,
    serviceDurationText: '',
    evidenceList: [],
    isPlayer: false
  },

  onLoad(options) {
    const { orderId } = options;
    this.setData({ orderId });
    this.loadOrderDetail();
  },

  onShow() {
    if (this.data.orderId) {
      this.loadOrderDetail();
      this.loadServiceTimer();
      this.loadEvidenceList();
    }
  },

  loadServiceTimer() {
    request.get(`/orders/${this.data.orderId}/service-timer`).then((res) => {
      if (res) {
        // 服务时长文本由后端返回，前端不做秒数格式化
        this.setData({
          serviceTimer: res,
          serviceDurationText: res.service_duration_text || ''
        });
      }
    }).catch(() => {});
  },

  loadEvidenceList() {
    request.get(`/orders/${this.data.orderId}/evidences`).then((res) => {
      // 直接使用后端返回的类型/时间文本字段
      const list = (res.list || []).map(item => ({
        ...item,
        type_text: item.evidence_type_text || '',
        create_time_text: item.create_time_text || ''
      }));
      this.setData({ evidenceList: list });
    }).catch(() => {});
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
      // 状态文本/颜色/描述、金额、时间、可操作布尔位均由后端返回
      const orderInfo = {
        orderId: res.order_id,
        status: res.status,
        statusText: res.status_text || '',
        statusColor: res.status_color || '',
        statusDesc: res.status_desc || '',
        gameName: res.game_name || '',
        serviceName: res.service_name || '',
        rank: res.rank || '',
        amount: res.amount_text || '',
        createTime: res.create_time_text || '',
        remark: res.remark || '',
        playerAvatar: res.player_avatar || '',
        playerName: res.player_name || '',
        playerRating: res.player_rating || 0,
        userAvatar: res.user_avatar || '',
        userName: res.user_name || '',
        canCancel: !!res.can_cancel,
        canAppeal: !!res.can_appeal
      };

      this.setData({
        orderInfo: orderInfo,
        // 时间线由后端构建返回，前端不做状态比较/时间格式化
        timelineList: res.timeline || [],
        // is_player 仅作 UI 提示，权限由后端在接口层拦截
        isPlayer: !!res.is_player
      });
    }).catch((err) => {
      console.error('加载订单详情失败:', err);
    });
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
