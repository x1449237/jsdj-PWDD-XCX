const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    club: null,
    loading: true,
    activeTab: 'announcement',
    tabs: [
      { key: 'announcement', label: '公告' },
      { key: 'coupon', label: '优惠券' },
      { key: 'dynamic', label: '动态墙' },
      { key: 'branch', label: '分店' }
    ],
    // B10：公告已读统计 - 已上报已读的公告ID集合
    readAnnouncementIds: {}
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadDetail();
  },

  async loadDetail() {
    try {
      const res = await request.get('/club/detail', {
        id: this.data.clubId
      });
      this.setData({ club: res.data, loading: false });
      // B10：进入公告Tab时自动上报首条置顶公告已读
      this.autoReportFirstAnnouncement();
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }
  },

  onTabChange(e) {
    const key = e.currentTarget.dataset.key;
    this.setData({ activeTab: key });
    // B10：切换到公告Tab时自动上报首条未读公告
    if (key === 'announcement') {
      this.autoReportFirstAnnouncement();
    }
  },

  // B10：自动上报首条未读公告已读
  autoReportFirstAnnouncement() {
    const announcements = (this.data.club && this.data.club.announcements) || [];
    if (announcements.length === 0) return;
    const first = announcements[0];
    if (!first || !first.id) return;
    if (this.data.readAnnouncementIds[first.id]) return;
    this.reportAnnouncementRead(first.id);
  },

  // B10：上报公告已读 + 同步已读/未读统计
  async reportAnnouncementRead(announcementId) {
    if (!announcementId || this.data.readAnnouncementIds[announcementId]) return;
    try {
      const res = await request.post('/club/announcement/read', {
        club_id: this.data.clubId,
        announcement_id: announcementId
      });
      // 标记本地已读
      const readMap = Object.assign({}, this.data.readAnnouncementIds);
      readMap[announcementId] = true;
      this.setData({ readAnnouncementIds: readMap });

      // 同步更新本地公告已读统计
      const club = this.data.club;
      if (club && club.announcements) {
        const announcements = club.announcements.map(item => {
          if (item.id === announcementId) {
            return Object.assign({}, item, {
              has_read: true,
              read_count: (item.read_count || 0) + 1
            });
          }
          return item;
        });
        this.setData({ 'club.announcements': announcements });
      }

      // 后端可能返回最新统计，优先使用
      if (res && res.data && res.data.read_count !== undefined) {
        const announcements = (this.data.club.announcements || []).map(item => {
          if (item.id === announcementId) {
            return Object.assign({}, item, { read_count: res.data.read_count });
          }
          return item;
        });
        this.setData({ 'club.announcements': announcements });
      }
    } catch (e) {
      // 静默失败，不影响阅读体验
    }
  },

  // B10：点击公告项上报已读
  onAnnouncementTap(e) {
    const id = e.currentTarget.dataset.id;
    if (!id) return;
    this.reportAnnouncementRead(id);
  },

  goManage() {
    if (this.data.club && this.data.club.my_role) {
      wx.navigateTo({ url: '/pages/club/manage/index?id=' + this.data.clubId });
    }
  },

  goJoinApply() {
    wx.showModal({
      title: '加入俱乐部',
      content: '确定申请加入该俱乐部吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await request.post('/club/member/join-apply', { club_id: this.data.clubId });
            wx.showToast({ title: '申请已提交', icon: 'success' });
          } catch (e) {
            wx.showToast({ title: e.message || '申请失败', icon: 'none' });
          }
        }
      }
    });
  },

  goDynamicDetail(e) {
    // 可以跳转到动态详情，这里先保留
  },

  previewImage(e) {
    const urls = e.currentTarget.dataset.urls;
    const current = e.currentTarget.dataset.current;
    if (urls && urls.length > 0) {
      wx.previewImage({ urls, current });
    }
  },

  receiveCoupon(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '领取优惠券',
      content: '确定领取这张优惠券吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await request.post('/club/coupon/receive', { id });
            wx.showToast({ title: '领取成功', icon: 'success' });
            this.loadDetail();
          } catch (e) {
            wx.showToast({ title: e.message || '领取失败', icon: 'none' });
          }
        }
      }
    });
  }
});
