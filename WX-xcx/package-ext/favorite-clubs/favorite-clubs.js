const extApi = require('../../utils/ext-api');
const app = getApp();

Page({
  data: {
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList();
    }
  },

  loadList() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });
    extApi.listFavoriteClubs(this.data.page, this.data.pageSize).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        intro_text: item.intro || '暂无简介'
      }));
      this.setData({
        list: this.data.list.concat(list),
        page: this.data.page + 1,
        noMore: list.length < this.data.pageSize,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  onItemTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/club/detail/detail?id=${id}`
    });
  },

  onUnfavorite(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '提示',
      content: '确定取消收藏该俱乐部吗？',
      success: (res) => {
        if (!res.confirm) return;
        extApi.unfavoriteClub(id).then(() => {
          wx.showToast({ title: '已取消收藏', icon: 'success' });
          this.setData({
            page: 1,
            list: [],
            noMore: false
          });
          this.loadList();
        }).catch(() => {});
      }
    });
  }
});
