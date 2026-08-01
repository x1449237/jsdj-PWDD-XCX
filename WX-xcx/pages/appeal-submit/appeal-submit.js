const request = require('../../utils/request');
const validator = require('../../utils/validator');

Page({
  data: {
    appealType: 'phone',
    orderList: [],
    selectedOrder: null,
    selectedOrderIndex: -1,
    phone: '',
    description: '',
    videoPath: '',
    images: [],
    submitting: false,
    canSubmit: false
  },

  onLoad(options) {
    if (options.order_id) {
      this.setData({ appealType: 'order' });
      this.loadOrderList();
    }
    this.checkCanSubmit();
  },

  loadOrderList() {
    request.get('/orders', { page: 1, page_size: 50 }).then((res) => {
      this.setData({ orderList: res.list || [] });
    });
  },

  onSelectType(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({
      appealType: type,
      selectedOrder: null,
      selectedOrderIndex: -1
    });
    if (type === 'order') {
      this.loadOrderList();
    }
    this.checkCanSubmit();
  },

  onOrderChange(e) {
    const index = e.detail.value;
    const order = this.data.orderList[index];
    this.setData({
      selectedOrder: order,
      selectedOrderIndex: index
    });
    this.checkCanSubmit();
  },

  onPhoneInput(e) {
    this.setData({ phone: e.detail.value });
    this.checkCanSubmit();
  },

  onDescriptionInput(e) {
    this.setData({ description: e.detail.value });
    this.checkCanSubmit();
  },

  onChooseVideo() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['video'],
      sourceType: ['album', 'camera'],
      maxDuration: 15,
      camera: 'back',
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        const duration = res.tempFiles[0].duration;
        if (duration > 15) {
          wx.showToast({
            title: '视频不能超过15秒',
            icon: 'none'
          });
          return;
        }
        this.setData({ videoPath: tempFilePath });
        this.checkCanSubmit();
      }
    });
  },

  onPreviewVideo() {
    if (!this.data.videoPath) return;
    wx.previewMedia({
      sources: [{
        url: this.data.videoPath,
        type: 'video'
      }]
    });
  },

  onDeleteVideo() {
    this.setData({ videoPath: '' });
    this.checkCanSubmit();
  },

  onChooseImage() {
    const remain = 3 - this.data.images.length;
    wx.chooseMedia({
      count: remain,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const newImages = res.tempFiles.map((file) => file.tempFilePath);
        this.setData({
          images: this.data.images.concat(newImages)
        });
      }
    });
  },

  onPreviewImage(e) {
    const url = e.currentTarget.dataset.url;
    wx.previewImage({
      current: url,
      urls: this.data.images
    });
  },

  onDeleteImage(e) {
    const index = e.currentTarget.dataset.index;
    const images = this.data.images;
    images.splice(index, 1);
    this.setData({ images });
  },

  checkCanSubmit() {
    const { appealType, phone, description, videoPath, selectedOrder } = this.data;
    let canSubmit = false;

    if (appealType === 'phone') {
      canSubmit = validator.validatePhone(phone).valid && description.trim().length > 0 && !!videoPath;
    } else if (appealType === 'order') {
      canSubmit = !!selectedOrder && description.trim().length > 0;
    } else {
      canSubmit = validator.validatePhone(phone).valid && description.trim().length > 0;
    }

    this.setData({ canSubmit });
  },

  onSubmit() {
    if (!this.data.canSubmit) {
      wx.showToast({
        title: '请完善申诉信息',
        icon: 'none'
      });
      return;
    }

    this.setData({ submitting: true });

    const uploadFiles = [];
    if (this.data.videoPath) {
      uploadFiles.push(this.uploadFile(this.data.videoPath, 'video'));
    }
    for (const image of this.data.images) {
      uploadFiles.push(this.uploadFile(image, 'image'));
    }

    Promise.all(uploadFiles).then((uploadResults) => {
      const videoUrl = uploadResults[0] || '';
      const imageUrls = uploadResults.slice(1);

      return request.post('/appeals', {
        type: this.data.appealType,
        order_id: this.data.selectedOrder ? this.data.selectedOrder.id : undefined,
        phone: this.data.phone,
        description: this.data.description,
        video_url: videoUrl,
        images: imageUrls
      });
    }).then(() => {
      wx.showToast({
        title: '申诉提交成功',
        icon: 'success',
        duration: 2000
      });
      setTimeout(() => {
        wx.navigateBack();
      }, 2000);
    }).catch((err) => {
      this.setData({ submitting: false });
    });
  },

  uploadFile(filePath, type) {
    return new Promise((resolve, reject) => {
      wx.uploadFile({
        url: request.baseURL + '/upload',
        filePath: filePath,
        name: 'file',
        header: {
          'Authorization': 'Bearer ' + wx.getStorageSync('token')
        },
        formData: {
          type: type
        },
        success(res) {
          const data = JSON.parse(res.data);
          if (data.code === 200) {
            resolve(data.data.url);
          } else {
            reject(new Error(data.msg));
          }
        },
        fail(err) {
          reject(err);
        }
      });
    });
  }
});