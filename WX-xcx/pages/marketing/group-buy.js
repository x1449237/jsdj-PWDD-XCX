const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    isLogin: false,
    tabIndex: 0,
    tabs: ['热门拼团', '我的拼团'],
    activityList: [],
    myGroupList: [],
    loading: false,
    page: 1,
    pageSize: 20,
    hasMore: true,
    currentShareGroup: null
  },

  onLoad() {
    this.checkLogin();
    this.loadActivities();
  },

  onShow() {
    this.checkLogin();
    if (this.data.isLogin && this.data.tabIndex === 1) {
      this.loadMyGroups();
    }
  },

  checkLogin() {
    const isLogin = !!(app.globalData && app.globalData.isLogin);
    this.setData({ isLogin });
  },

  onTabTap(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ tabIndex: index, page: 1, hasMore: true });
    if (index === 0) {
      this.setData({ activityList: [] });
      this.loadActivities();
    } else {
      if (!this.data.isLogin) {
        this.onLogin();
        return;
      }
      this.setData({ myGroupList: [] });
      this.loadMyGroups();
    }
  },

  loadActivities() {
    if (this.data.loading || !this.data.hasMore) return;
    this.setData({ loading: true });

    request.get('/group-buy/activities', {
      page: this.data.page,
      limit: this.data.pageSize
    }).then((res) => {
      const list = (res.data?.list || []).map(item => this.formatActivity(item));
      const newList = this.data.page === 1 ? list : this.data.activityList.concat(list);
      this.setData({
        activityList: newList,
        hasMore: list.length === this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch((err) => {
      console.error('加载活动失败:', err);
      const mockList = this.getMockActivities();
      const newList = this.data.page === 1 ? mockList : this.data.activityList.concat(mockList);
      this.setData({
        activityList: newList,
        hasMore: false
      });
      if (!res || !res.data) {
        wx.showToast({ title: '加载失败', icon: 'none' });
      }
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  formatActivity(item) {
    const groupPrice = Number(item.group_price || 0);
    const originalPrice = Number(item.original_price || 0);
    const discountPercent = originalPrice > 0 ? Math.round((groupPrice / originalPrice) * 100) : 100;
    const discountLabel = discountPercent >= 100 ? '' : `${Math.round(discountPercent / 10)}折`;
    const saveAmount = originalPrice > groupPrice ? (originalPrice - groupPrice).toFixed(2) : '';

    const remainSeconds = this.calcRemainSeconds(item.end_time, item.duration_hours);
    const remainTimeText = this.formatRemainTime(remainSeconds);
    const onGoingGroups = Number(item.on_going_groups || Math.floor(Math.random() * 20));
    const progressPercent = Math.min(100, Math.round((onGoingGroups / Math.max(5, item.min_people * 4)) * 100));

    return {
      ...item,
      group_price: groupPrice.toFixed(2),
      original_price: originalPrice.toFixed(2),
      discount_label: item.discount_label || discountLabel,
      save_amount: item.save_amount || saveAmount,
      remain_time_text: item.remain_time_text || remainTimeText,
      on_going_groups: onGoingGroups,
      progress_percent: item.progress_percent || progressPercent
    };
  },

  calcRemainSeconds(endTime, durationHours) {
    if (endTime) {
      const end = new Date(endTime.replace(/-/g, '/')).getTime();
      const now = Date.now();
      return Math.max(0, Math.floor((end - now) / 1000));
    }
    return (durationHours || 24) * 3600;
  },

  formatRemainTime(totalSeconds) {
    if (!totalSeconds || totalSeconds <= 0) return '活动已结束';
    const days = Math.floor(totalSeconds / 86400);
    const hours = Math.floor((totalSeconds % 86400) / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    if (days > 0) return `${days}天${hours}小时`;
    if (hours > 0) return `${hours}小时${minutes}分`;
    return `${minutes}分钟`;
  },

  getMockActivities() {
    return [
      {
        id: 1001,
        name: '王者荣耀·钻石上分2人团',
        service_name: '排位代练',
        group_price: '29.90',
        original_price: '59.80',
        min_people: 2,
        max_people: 2,
        duration_hours: 48,
        min_consume: '50',
        end_time: new Date(Date.now() + 3600 * 1000 * 24 * 3).toISOString()
      },
      {
        id: 1002,
        name: '英雄联盟·黄金-铂金3人团',
        service_name: '段位陪玩',
        group_price: '49.00',
        original_price: '98.00',
        min_people: 3,
        max_people: 5,
        duration_hours: 72,
        min_consume: '80',
        end_time: new Date(Date.now() + 3600 * 1000 * 24 * 5).toISOString()
      },
      {
        id: 1003,
        name: '和平精英·吃鸡娱乐5人团',
        service_name: '娱乐陪玩',
        group_price: '19.90',
        original_price: '49.90',
        min_people: 3,
        max_people: 5,
        duration_hours: 24,
        min_consume: '30',
        end_time: new Date(Date.now() + 3600 * 1000 * 12).toISOString()
      }
    ].map(item => this.formatActivity(item));
  },

  loadMyGroups() {
    if (!this.data.isLogin) return;
    if (this.data.loading || !this.data.hasMore) return;
    this.setData({ loading: true });

    request.get('/group-buy/my', {
      page: this.data.page,
      limit: this.data.pageSize
    }).then((res) => {
      const list = (res.data?.list || []).map(item => this.formatMyGroup(item));
      const newList = this.data.page === 1 ? list : this.data.myGroupList.concat(list);
      this.setData({
        myGroupList: newList,
        hasMore: list.length === this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  formatMyGroup(item) {
    const current = Number(item.current_people || 1);
    const max = Number(item.max_people || item.min_people || 2);
    const groupProgress = max > 0 ? Math.round((current / max) * 100) : 0;
    return {
      ...item,
      group_progress: groupProgress,
      group_price: Number(item.group_price || 0).toFixed(2)
    };
  },

  onJoinGroup(e) {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    const groupId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/marketing/group-buy-detail?id=${groupId}`
    });
  },

  onJoinExistingGroup(e) {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    const activity = e.currentTarget.dataset.activity;
    wx.showActionSheet({
      itemList: ['快速加入最近的团', '查看可加入团列表'],
      success: (res) => {
        if (res.tapIndex === 0) {
          this.quickJoinGroup(activity.id);
        } else {
          wx.navigateTo({
            url: `/pages/marketing/group-list?activityId=${activity.id}`
          });
        }
      }
    });
  },

  quickJoinGroup(activityId) {
    wx.showLoading({ title: '加入中...' });
    request.post('/group-buy/join-existing', {
      activity_id: activityId
    }).then((res) => {
      wx.hideLoading();
      wx.showToast({ title: '加入成功', icon: 'success' });
      setTimeout(() => {
        wx.navigateTo({
          url: `/pages/marketing/group-buy-detail?id=${res.data?.id || res.id}`
        });
      }, 1000);
    }).catch((err) => {
      wx.hideLoading();
      wx.showToast({
        title: err?.msg || '加入失败，请稍后重试',
        icon: 'none'
      });
    });
  },

  onOpenGroup(e) {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    const activity = e.currentTarget.dataset.item;
    wx.showModal({
      title: '确认开团',
      content: `确定要开团「${activity.name}」吗？\n拼团价¥${activity.group_price}，${activity.min_people}人即可成团。`,
      confirmText: '立即开团',
      confirmColor: '#e94560',
      success: (res) => {
        if (res.confirm) {
          this.createGroup(activity.id);
        }
      }
    });
  },

  createGroup(activityId) {
    wx.showLoading({ title: '创建中...' });
    request.post('/group-buy/join', {
      activity_id: activityId
    }).then((res) => {
      wx.hideLoading();
      const groupId = res.data?.id || res.id;
      wx.showToast({ title: '开团成功', icon: 'success' });
      setTimeout(() => {
        wx.showModal({
          title: '开团成功',
          content: '快去分享给好友，邀请他们一起拼团吧！',
          confirmText: '立即分享',
          cancelText: '稍后再说',
          confirmColor: '#e94560',
          success: (modalRes) => {
            if (modalRes.confirm) {
              wx.navigateTo({
                url: `/pages/marketing/group-buy-detail?id=${groupId}&action=share`
              });
            } else {
              wx.navigateTo({
                url: `/pages/marketing/group-buy-detail?id=${groupId}`
              });
            }
          }
        });
      }, 800);
    }).catch((err) => {
      wx.hideLoading();
      wx.showToast({
        title: err?.msg || '开团失败，请重试',
        icon: 'none'
      });
    });
  },

  onShareGroup(e) {
    const item = e.currentTarget.dataset.item;
    this.setData({ currentShareGroup: item });
  },

  onShareAppMessage(res) {
    let group = this.data.currentShareGroup;
    let title = '🔥 一起拼团更划算';
    let path = '/pages/marketing/group-buy';

    if (res.from === 'button' && res.target && res.target.dataset && res.target.dataset.id) {
      const groupId = res.target.dataset.id;
      title = group ?
        `【${group.activity_name}】还差${Math.max(0, (group.max_people || 2) - (group.current_people || 1))}人成团，拼团价¥${group.group_price}` :
        '我开了一个拼团，快来加入吧！';
      path = `/pages/marketing/group-buy-detail?id=${groupId}&invite=1`;
    }

    return {
      title: title,
      path: path,
      imageUrl: ''
    };
  },

  onReachBottom() {
    if (this.data.hasMore) {
      if (this.data.tabIndex === 0) {
        this.loadActivities();
      } else {
        this.loadMyGroups();
      }
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true });
    if (this.data.tabIndex === 0) {
      this.setData({ activityList: [] });
      this.loadActivities();
    } else {
      this.setData({ myGroupList: [] });
      this.loadMyGroups();
    }
    wx.stopPullDownRefresh();
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
