const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    sessionList: [],
    loading: false,
    showCreateModal: false,
    orderList: [],
    appealForm: {
      order_id: '',
      reason: '',
      images: []
    }
  },

  onLoad() {
    this.loadSessionList();
  },

  onShow() {
    this.loadSessionList();
  },

  onPullDownRefresh() {
    this.loadSessionList();
    wx.stopPullDownRefresh();
  },

  loadSessionList() {
    if (this.data.loading) return;
    this.setData({ loading: true });

    request.get('/after-sales').then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        intervene_label: this.getInterveneLabel(item.intervene_status),
        intervene_class: this.getInterveneClass(item.intervene_status),
        last_time: util.formatRelativeTime(item.last_msg_time),
        order_sn_preview: item.order_sn ? '订单: ' + item.order_sn : ''
      }));

      this.setData({
        sessionList: list,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  getInterveneLabel(status) {
    const statusMap = {
      0: '未介入',
      1: '介入中',
      2: '已解除'
    };
    return statusMap[status] || '未介入';
  },

  getInterveneClass(status) {
    const classMap = {
      0: 'intervene-none',
      1: 'intervene-ing',
      2: 'intervene-done'
    };
    return classMap[status] || 'intervene-none';
  },

  onOpenSession(e) {
    const session = e.currentTarget.dataset.session;
    wx.navigateTo({
      url: '/chat/after-sale-room/after-sale-room?sessionId=' + session.session_id + '&orderSn=' + encodeURIComponent(session.order_sn || '')
    });
  },

  onShowCreateModal() {
    this.loadCompletedOrders();
    this.setData({
      showCreateModal: true,
      appealForm: { order_id: '', reason: '', images: [] }
    });
  },

  onHideCreateModal() {
    this.setData({ showCreateModal: false });
  },

  loadCompletedOrders() {
    request.get('/orders/completed', {
      page: 1,
      page_size: 50
    }).then((res) => {
      this.setData({
        orderList: res.list || []
      });
    }).catch(() => {});
  },

  onOrderSelect(e) {
    const orderId = e.currentTarget.dataset.id;
    this.setData({
      'appealForm.order_id': orderId
    });
  },

  onReasonInput(e) {
    this.setData({
      'appealForm.reason': e.detail.value
    });
  },

  onChooseImage() {
    const currentCount = this.data.appealForm.images.length;
    const remainCount = 9 - currentCount;
    if (remainCount <= 0) {
      wx.showToast({ title: '最多上传9张图片', icon: 'none' });
      return;
    }

    wx.chooseMedia({
      count: remainCount,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['album'],
      success: (res) => {
        const newImages = res.tempFiles.map(f => f.tempFilePath);
        this.setData({
          'appealForm.images': [...this.data.appealForm.images, ...newImages]
        });
      }
    });
  },

  onRemoveImage(e) {
    const index = e.currentTarget.dataset.index;
    const images = this.data.appealForm.images;
    images.splice(index, 1);
    this.setData({
      'appealForm.images': images
    });
  },

  onCreateAppeal() {
    const { order_id, reason, images } = this.data.appealForm;
    if (!order_id) {
      wx.showToast({ title: '请选择订单', icon: 'none' });
      return;
    }
    if (!reason.trim()) {
      wx.showToast({ title: '请输入申诉原因', icon: 'none' });
      return;
    }

    this.uploadImages(images).then((imageUrls) => {
      return request.post('/after-sales', {
        order_id: order_id,
        reason: reason.trim(),
        images: imageUrls
      });
    }).then((res) => {
      wx.showToast({ title: '申诉提交成功', icon: 'success' });
      this.setData({ showCreateModal: false });
      this.loadSessionList();
      wx.navigateTo({
        url: '/chat/after-sale-room/after-sale-room?sessionId=' + res.session_id + '&orderSn=' + encodeURIComponent(order_id)
      });
    }).catch(() => {
      wx.showToast({ title: '提交失败', icon: 'none' });
    });
  },

  uploadImages(images) {
    if (images.length === 0) return Promise.resolve([]);

    const uploadPromises = images.map((filePath) => {
      return request.upload('/after-sales/upload', filePath).then((data) => {
        return data.url;
      });
    });

    return Promise.all(uploadPromises);
  }
});