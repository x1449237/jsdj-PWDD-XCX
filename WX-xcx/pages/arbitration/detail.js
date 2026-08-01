const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    caseId: '',
    caseDetail: null,
    loading: true,
    disputeTypeLabel: '',
    statusLabel: '',
    statusTag: ''
  },

  onLoad(options) {
    const { case_id } = options;
    this.setData({ caseId: case_id || '' });
    this.loadDetail();
  },

  onPullDownRefresh() {
    this.loadDetail(() => {
      wx.stopPullDownRefresh();
    });
  },

  loadDetail(callback) {
    if (!this.data.caseId) return;

    this.setData({ loading: true });
    request.get('/arbitration/detail', {
      case_id: this.data.caseId
    }).then(res => {
      this.setData({
        caseDetail: res,
        disputeTypeLabel: this.getDisputeTypeLabel(res.dispute_type),
        statusLabel: this.getStatusLabel(res.status),
        statusTag: this.getStatusTag(res.status)
      });
    }).catch(err => {
      wx.showToast({ title: err.msg || '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
      if (callback) callback();
    });
  },

  getDisputeTypeLabel(type) {
    const map = {
      player_late: '打手迟到',
      negative_service: '消极服务',
      player_refund: '退款纠纷',
      demand_change: '需求变更',
      other: '其他'
    };
    return map[type] || type;
  },

  getStatusLabel(status) {
    const map = {
      pending: '待受理',
      processing: '处理中',
      resolved: '已结案'
    };
    return map[status] || status;
  },

  getStatusTag(status) {
    const map = {
      pending: 'warning',
      processing: 'primary',
      resolved: 'success'
    };
    return map[status] || 'info';
  },

  onPreviewImage(e) {
    const url = e.currentTarget.dataset.url;
    const urls = this.data.caseDetail.evidence_list
      ? this.data.caseDetail.evidence_list.filter(item => item.type === 'image').map(item => item.file_url)
      : [];
    wx.previewImage({
      current: url,
      urls: urls
    });
  },

  onBack() {
    wx.navigateBack();
  }
});
