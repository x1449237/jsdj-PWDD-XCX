const request = require('../../utils/request');

Page({
  data: {
    guardianName: '',
    guardianIdCard: '',
    guardianPhone: '',
    livenessStatus: 'idle',
    livenessLoading: false,
    livenessFailReason: '',
    disclaimerAgreed: false,
    signatureDrawn: false,
    submitting: false,
    feeAgreed: false, // 活体认证收费知情同意
    canvasCtx: null,
    startX: 0,
    startY: 0,
    drawing: false
  },

  onLoad(options) {
    this.initCanvas();
  },

  onReady() {
    this.initCanvas();
  },

  initCanvas() {
    const query = wx.createSelectorQuery();
    query.select('#signatureCanvas')
      .fields({ node: true, size: true })
      .exec((res) => {
        if (res[0]) {
          const canvas = res[0].node;
          const ctx = canvas.getContext('2d');
          const dpr = wx.getSystemInfoSync().pixelRatio;
          canvas.width = res[0].width * dpr;
          canvas.height = res[0].height * dpr;
          ctx.scale(dpr, dpr);
          ctx.lineCap = 'round';
          ctx.lineJoin = 'round';
          ctx.strokeStyle = '#333333';
          ctx.lineWidth = 3;
          this.canvas = canvas;
          this.ctx = ctx;
        } else {
          // 降级使用旧版 canvas API
          const ctx = wx.createCanvasContext('signatureCanvas');
          ctx.setLineCap('round');
          ctx.setLineJoin('round');
          ctx.setStrokeStyle('#333333');
          ctx.setLineWidth(3);
          this.ctx = ctx;
          this.canvas = null;
        }
      });
  },

  onNameInput(e) {
    this.setData({ guardianName: e.detail.value });
  },

  onIdCardInput(e) {
    this.setData({ guardianIdCard: e.detail.value });
  },

  onPhoneInput(e) {
    this.setData({ guardianPhone: e.detail.value });
  },

  // 切换活体认证收费同意状态
  toggleFeeAgree() {
    this.setData({ feeAgreed: !this.data.feeAgreed });
  },

  onStartLiveness() {
    const { guardianName, guardianIdCard, feeAgreed } = this.data;
    if (!guardianName || !guardianIdCard) {
      wx.showToast({ title: '请先填写监护人信息', icon: 'none' });
      return;
    }
    if (!feeAgreed) {
      wx.showToast({ title: '请先阅读并同意收费说明', icon: 'none' });
      return;
    }

    this.setData({
      livenessStatus: 'scanning',
      livenessLoading: true,
      livenessFailReason: ''
    });

    request.post('/auth/guardian/liveness', {
      real_name: guardianName,
      id_card: guardianIdCard
    }).then((res) => {
      this.setData({
        livenessStatus: 'success',
        livenessLoading: false
      });
    }).catch((err) => {
      this.setData({
        livenessStatus: 'fail',
        livenessLoading: false,
        livenessFailReason: err.message || '活体检测未通过，请重试'
      });
    });
  },

  onToggleDisclaimer() {
    this.setData({ disclaimerAgreed: !this.data.disclaimerAgreed });
  },

  onSignatureStart(e) {
    this.setData({
      drawing: true,
      startX: e.touches[0].x,
      startY: e.touches[0].y
    });
  },

  onSignatureMove(e) {
    if (!this.data.drawing) return;

    const { startX, startY } = this.data;
    const x = e.touches[0].x;
    const y = e.touches[0].y;

    if (this.canvas) {
      this.ctx.beginPath();
      this.ctx.moveTo(startX, startY);
      this.ctx.lineTo(x, y);
      this.ctx.stroke();
    } else {
      this.ctx.moveTo(startX, startY);
      this.ctx.lineTo(x, y);
      this.ctx.stroke();
      this.ctx.draw(true);
    }

    this.setData({
      startX: x,
      startY: y,
      signatureDrawn: true
    });
  },

  onSignatureEnd() {
    this.setData({ drawing: false });
  },

  onClearSignature() {
    if (this.canvas) {
      this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    } else {
      this.ctx.clearRect(0, 0, 750, 320);
      this.ctx.draw();
    }
    this.setData({ signatureDrawn: false });
  },

  onSubmit() {
    const {
      guardianName, guardianIdCard, guardianPhone,
      livenessStatus, disclaimerAgreed, signatureDrawn
    } = this.data;

    if (!guardianName || !guardianIdCard || !guardianPhone) {
      wx.showToast({ title: '请填写完整监护人信息', icon: 'none' });
      return;
    }

    if (livenessStatus !== 'success') {
      wx.showToast({ title: '请先完成活体检测', icon: 'none' });
      return;
    }

    if (!disclaimerAgreed) {
      wx.showToast({ title: '请先同意免责声明', icon: 'none' });
      return;
    }

    if (!signatureDrawn) {
      wx.showToast({ title: '请完成电子签名', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    // 获取签名图片
    wx.canvasToTempFilePath({
      canvasId: 'signatureCanvas',
      success: (res) => {
        // 上传签名图片
        this.uploadSignatureAndSubmit(res.tempFilePath);
      },
      fail: () => {
        this.doSubmit('');
      }
    });
  },

  uploadSignatureAndSubmit(filePath) {
    request.upload('/upload/signature', filePath, {}, { name: 'signature' }).then((data) => {
      this.doSubmit(data.url || '');
    }).catch(() => {
      this.doSubmit('');
    });
  },

  doSubmit(signatureUrl) {
    const { guardianName, guardianIdCard, guardianPhone } = this.data;

    request.post('/auth/guardian/verify', {
      guardian_name: guardianName,
      guardian_id_card: guardianIdCard,
      guardian_phone: guardianPhone,
      signature_url: signatureUrl
    }).then((res) => {
      this.setData({ submitting: false });
      wx.showToast({ title: '监护人验证成功', icon: 'success' });
      setTimeout(() => {
        wx.navigateBack();
      }, 1500);
    }).catch((err) => {
      this.setData({ submitting: false });
      wx.showToast({ title: err.message || '验证失败，请重试', icon: 'none' });
    });
  }
});