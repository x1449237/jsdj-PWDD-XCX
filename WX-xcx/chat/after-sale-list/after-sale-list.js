/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

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
        intervene_label: item.intervene_status_text || '',
        intervene_class: item.intervene_status_class || '',
        last_time: item.relative_time_text || '',
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
