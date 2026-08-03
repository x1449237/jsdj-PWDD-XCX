const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    orderId: '',
    sessionId: '',
    respondentId: '',
    disputeTypes: [
      { value: 'player_late', label: '打手迟到' },
      { value: 'negative_service', label: '消极服务' },
      { value: 'player_refund', label: '退款纠纷' },
      { value: 'demand_change', label: '需求变更' },
      { value: 'other', label: '其他' }
    ],
    selectedDisputeType: '',
    evidenceTpl: null,
    description: '',
    evidenceList: [],
    submitting: false
  },

  onLoad(options) {
    const { order_id, session_id, respondent_id } = options;
    this.setData({
      orderId: order_id || '',
      sessionId: session_id || '',
      respondentId: respondent_id || ''
    });
  },

  onSelectDisputeType(e) {
    const value = e.currentTarget.dataset.value;
    this.setData({ selectedDisputeType: value });
    this.loadEvidenceTpl(value);
  },

  loadEvidenceTpl(disputeType) {
    request.get('/arbitration/evidence_tpl', {
      dispute_type: disputeType
    }).then(res => {
      const tpl = Array.isArray(res) && res.length > 0 ? res[0] : null;
      this.setData({ evidenceTpl: tpl });
    }).catch(() => {});
  },

  onDescriptionInput(e) {
    this.setData({ description: e.detail.value });
  },

  onUploadImage() {
    wx.chooseMedia({
      count: 9 - this.data.evidenceList.length,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFiles = res.tempFiles;
        tempFiles.forEach(file => {
          this.uploadFile(file.tempFilePath, 'image');
        });
      }
    });
  },

  onUploadVideo() {
    wx.chooseMedia({
      count: 3 - this.data.evidenceList.filter(e => e.type === 'video').length,
      mediaType: ['video'],
      sourceType: ['album', 'camera'],
      maxDuration: 60,
      success: (res) => {
        const tempFiles = res.tempFiles;
        tempFiles.forEach(file => {
          this.uploadFile(file.tempFilePath, 'video');
        });
      }
    });
  },

  uploadFile(filePath, type) {
    wx.showLoading({ title: '上传中...', mask: true });
    const token = wx.getStorageSync('token') || '';
    wx.uploadFile({
      url: request.baseURL + '/upload/image',
      filePath: filePath,
      name: 'file',
      header: {
        'Authorization': 'Bearer ' + token
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data);
          if (data.code === 200) {
            // 仅用本地自增计数生成临时 temp_id，不使用时间戳/随机数；真实 id 由后端在提交时分配
            this._evidenceSeq = (this._evidenceSeq || 0) + 1;
            const evidence = {
              id: this._evidenceSeq,
              type: type,
              file_url: data.data.url,
              description: ''
            };
            this.setData({
              evidenceList: [...this.data.evidenceList, evidence]
            });
          } else {
            wx.showToast({ title: '上传失败', icon: 'none' });
          }
        } catch (e) {
          wx.showToast({ title: '上传失败', icon: 'none' });
        }
      },
      fail: () => {
        wx.showToast({ title: '上传失败', icon: 'none' });
      },
      complete: () => {
        wx.hideLoading();
      }
    });
  },

  onDeleteEvidence(e) {
    const index = e.currentTarget.dataset.index;
    const list = [...this.data.evidenceList];
    list.splice(index, 1);
    this.setData({ evidenceList: list });
  },

  onEvidenceDescInput(e) {
    const index = e.currentTarget.dataset.index;
    const list = [...this.data.evidenceList];
    list[index].description = e.detail.value;
    this.setData({ evidenceList: list });
  },

  onSubmit() {
    if (!this.data.selectedDisputeType) {
      wx.showToast({ title: '请选择纠纷类型', icon: 'none' });
      return;
    }
    if (!this.data.description.trim()) {
      wx.showToast({ title: '请输入纠纷描述', icon: 'none' });
      return;
    }

    const tpl = this.data.evidenceTpl;
    if (tpl && tpl.required_items_json) {
      const requiredItems = tpl.required_items_json.filter(item => item.required);
      for (const item of requiredItems) {
        const hasEvidence = this.data.evidenceList.some(e => {
          if (item.type === 'image') return e.type === 'image';
          if (item.type === 'video') return e.type === 'video';
          return true;
        });
        if (!hasEvidence) {
          wx.showToast({ title: '请上传' + item.label, icon: 'none' });
          return;
        }
      }
    }

    this.setData({ submitting: true });
    request.post('/arbitration/apply', {
      order_id: this.data.orderId,
      session_id: this.data.sessionId,
      respondent_id: this.data.respondentId,
      dispute_type: this.data.selectedDisputeType,
      description: this.data.description,
      evidence: JSON.stringify(this.data.evidenceList)
    }).then(res => {
      wx.showToast({ title: '提交成功', icon: 'success' });
      setTimeout(() => {
        wx.redirectTo({
          url: '/pages/arbitration/detail?case_id=' + res.id
        });
      }, 1500);
    }).catch(err => {
      wx.showToast({ title: err.msg || '提交失败', icon: 'none' });
    }).finally(() => {
      this.setData({ submitting: false });
    });
  },

  onBack() {
    wx.navigateBack();
  }
});
