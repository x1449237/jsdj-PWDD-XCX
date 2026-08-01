const request = require('../../utils/request');

Page({
  data: {
    currentTab: 'agreement',
    agreements: [],
    policies: [],
    contracts: [],
    loading: true
  },

  onLoad() {
    this.fetchDocuments();
  },

  onSwitchTab(e) {
    const tab = e.currentTarget.dataset.tab;
    this.setData({ currentTab: tab });
  },

  async fetchDocuments() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/config/documents');
      const allDocs = res.data || [];

      this.setData({
        agreements: allDocs.filter(d => d.doc_type === 'agreement'),
        policies: allDocs.filter(d => d.doc_type === 'policy'),
        contracts: allDocs.filter(d => d.doc_type === 'contract'),
        loading: false
      });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }
  },

  onOpenDoc(e) {
    const doc = e.currentTarget.dataset.doc;
    if (!doc || !doc.file_url) {
      wx.showToast({ title: '文档链接无效', icon: 'none' });
      return;
    }

    wx.showLoading({ title: '下载中...' });

    const baseUrl = this.getBaseUrl();
    const downloadUrl = `${baseUrl}${doc.file_url}`;

    wx.downloadFile({
      url: downloadUrl,
      success: (res) => {
        wx.hideLoading();
        if (res.statusCode === 200) {
          wx.openDocument({
            filePath: res.tempFilePath,
            fileType: 'pdf',
            success: () => {},
            fail: () => {
              wx.showToast({ title: '无法打开PDF', icon: 'none' });
            }
          });
        } else {
          wx.showToast({ title: '下载失败', icon: 'none' });
        }
      },
      fail: () => {
        wx.hideLoading();
        wx.showToast({ title: '下载失败，请检查网络', icon: 'none' });
      }
    });
  },

  getBaseUrl() {
    // 与 request.js 中的 baseUrl 保持一致
    return 'https://your-domain.com';
  }
});