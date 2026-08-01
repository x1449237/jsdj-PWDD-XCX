const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    // 俱乐部选择
    clubs: [],
    selectedClubId: 0,
    selectedClubName: '',
    
    // 平台信息
    platform: '',
    platformAccountId: '',
    platformAccountUrl: '',
    fanCount: '',
    
    // 截图凭证
    screenshots: [],
    
    // 录屏视频
    videoUrl: '',
    videoUploading: false,
    
    // 等级配置
    tierConfigs: [],
    
    // 平台选项
    platformOptions: ['抖音', '快手', 'B站', '小红书', '微信视频号'],
    
    submitting: false
  },

  onLoad() {
    this.loadClubs();
    this.loadTierConfigs();
  },

  // 加载我所属的俱乐部
  async loadClubs() {
    try {
      const res = await request.get('/up_master/my_clubs');
      this.setData({ clubs: res.data || [] });
      if (this.data.clubs.length === 0) {
        wx.showToast({ title: '您尚未加入任何俱乐部', icon: 'none' });
      }
    } catch (e) {
      wx.showToast({ title: '加载俱乐部失败', icon: 'none' });
    }
  },

  // 加载等级配置
  async loadTierConfigs() {
    try {
      const res = await request.get('/up_master/tier_configs');
      this.setData({ tierConfigs: res.data || [] });
    } catch (e) {
      console.error('加载等级配置失败', e);
    }
  },

  // 选择俱乐部
  onClubChange(e) {
    const index = e.detail.value;
    const club = this.data.clubs[index];
    this.setData({
      selectedClubId: club.id,
      selectedClubName: club.club_name
    });
  },

  // 选择平台
  onPlatformChange(e) {
    this.setData({ platform: this.data.platformOptions[e.detail.value] });
  },

  // 上传截图
  onUploadScreenshot() {
    const that = this;
    if (this.data.screenshots.length >= 3) {
      wx.showToast({ title: '最多上传3张截图', icon: 'none' });
      return;
    }
    wx.chooseImage({
      count: 3 - this.data.screenshots.length,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success(res) {
        // 上传到OSS
        const files = res.tempFilePaths;
        that.uploadFiles(files, 'screenshots');
      }
    });
  },

  // 删除截图
  onDeleteScreenshot(e) {
    const index = e.currentTarget.dataset.index;
    const screenshots = this.data.screenshots;
    screenshots.splice(index, 1);
    this.setData({ screenshots });
  },

  // 上传录屏视频
  onUploadVideo() {
    const that = this;
    wx.chooseVideo({
      sourceType: ['album', 'camera'],
      maxDuration: 120,
      compressed: true,
      success(res) {
        that.setData({ videoUploading: true });
        // 上传视频
        that.uploadFiles([res.tempFilePath], 'video');
      }
    });
  },

  // 上传文件到OSS
  uploadFiles(filePaths, type) {
    const that = this;
    const promises = filePaths.map(path => {
      return new Promise((resolve, reject) => {
        wx.uploadFile({
          url: request.baseURL + '/upload/file',
          filePath: path,
          name: 'file',
          header: {
            'Authorization': 'Bearer ' + wx.getStorageSync('token')
          },
          success(res) {
            const data = JSON.parse(res.data);
            if (data.code === 200) {
              resolve(data.data?.url || '');
            } else {
              reject(new Error(data.msg || '上传失败'));
            }
          },
          fail(err) {
            reject(err);
          }
        });
      });
    });

    Promise.all(promises).then(urls => {
      if (type === 'screenshots') {
        that.setData({
          screenshots: [...that.data.screenshots, ...urls.filter(u => u)]
        });
      } else {
        that.setData({
          videoUrl: urls[0] || '',
          videoUploading: false
        });
        wx.showToast({ title: '视频上传成功', icon: 'success' });
      }
    }).catch(() => {
      that.setData({ videoUploading: false });
      wx.showToast({ title: '上传失败，请重试', icon: 'none' });
    });
  },

  // 预估等级
  getEstimatedTier() {
    const fanCount = parseInt(this.data.fanCount) || 0;
    const configs = this.data.tierConfigs;
    for (let i = configs.length - 1; i >= 0; i--) {
      if (fanCount >= configs[i].fan_threshold) {
        return configs[i];
      }
    }
    return null;
  },

  // 提交申请
  async onSubmit() {
    const { selectedClubId, platform, fanCount, screenshots, videoUrl, platformAccountId, platformAccountUrl } = this.data;

    // 校验
    if (!selectedClubId) {
      wx.showToast({ title: '请选择所属俱乐部', icon: 'none' });
      return;
    }
    if (!platform) {
      wx.showToast({ title: '请选择视频平台', icon: 'none' });
      return;
    }
    if (!fanCount || parseInt(fanCount) < 100) {
      wx.showToast({ title: '粉丝数至少100', icon: 'none' });
      return;
    }
    if (screenshots.length === 0) {
      wx.showToast({ title: '请上传个人主页截图', icon: 'none' });
      return;
    }
    if (!videoUrl) {
      wx.showToast({ title: '请上传录屏视频', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    try {
      await request.post('/up_master/submit', {
        club_id: selectedClubId,
        fan_count: parseInt(fanCount),
        platform: platform,
        platform_account_id: platformAccountId,
        platform_account_url: platformAccountUrl,
        screenshots: screenshots,
        video_url: videoUrl
      });

      wx.showToast({ title: '申请已提交', icon: 'success' });
      setTimeout(() => {
        wx.navigateBack();
      }, 1500);
    } catch (e) {
      wx.showToast({ title: e.message || '提交失败', icon: 'none' });
    } finally {
      this.setData({ submitting: false });
    }
  }
});