/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段（discount_label/save_amount_text/remain_time_text/progress_percent 等）
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数、折扣/进度计算。
 * 约定：request.get(url, data) 直接 resolve 出内层 data.data，res 即数据对象本身。
 */
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
      const list = (res.list || []).map(item => this.formatActivity(item));
      const newList = this.data.page === 1 ? list : this.data.activityList.concat(list);
      this.setData({
        activityList: newList,
        hasMore: list.length === this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch((err) => {
      console.error('加载活动失败:', err);
      wx.showToast({ title: '加载失败', icon: 'none' });
      if (this.data.page === 1) {
        this.setData({ activityList: [], hasMore: false });
      } else {
        this.setData({ hasMore: false });
      }
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  // 展示字段（discount_label/save_amount_text/remain_time_text/on_going_groups/progress_percent 等）均由后端返回
  formatActivity(item) {
    return { ...item };
  },

  loadMyGroups() {
    if (!this.data.isLogin) return;
    if (this.data.loading || !this.data.hasMore) return;
    this.setData({ loading: true });

    request.get('/group-buy/my', {
      page: this.data.page,
      limit: this.data.pageSize
    }).then((res) => {
      const list = (res.list || []).map(item => this.formatMyGroup(item));
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

  // 展示字段（group_progress/group_price 等）均由后端返回
  formatMyGroup(item) {
    return { ...item };
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
          url: `/pages/marketing/group-buy-detail?id=${res.id}`
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
      const groupId = res.id;
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
