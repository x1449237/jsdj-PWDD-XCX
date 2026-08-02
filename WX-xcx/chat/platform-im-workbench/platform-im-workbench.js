const request = require('../../utils/request');
const util = require('../../utils/util');

// 平台方管理人员 IM 工作台（需求 46-52）
Page({
  data: {
    loading: true,
    overview: {},
    buckets: [], // [{bucketKey:'todo', bucketName:'待跟进', order:2, items:[...]}]
    bucketsOrder: ['emergency', 'todo', 'yesterday']
  },
  onLoad() {
    this.loadWorkbench();
  },
  onPullDownRefresh() {
    this.loadWorkbench();
    wx.stopPullDownRefresh();
  },
  loadWorkbench() {
    this.setData({ loading: true });
    request.get('/platform-im/workbench').then((res) => {
      const data = res.data || {};
      const bucketsOrder = data.bucketsOrder || ['emergency', 'todo', 'yesterday'];
      const buckets = bucketsOrder.map((k) => ({
        bucketKey: k,
        bucketName: this.getBucketName(k),
        order: k,
        items: (data.buckets && data.buckets[k]) || []
      })).filter(b => b.items.length > 0 || b.bucketKey === 'emergency');
      this.setData({
        overview: data.overview || {},
        bucketsOrder: bucketsOrder,
        buckets: buckets,
        loading: false
      });
    }).catch(() => this.setData({ loading: false }));
  },
  getBucketName(k) {
    const map = {
      emergency: '今日必须处理紧急任务',
      todo: '待跟进售后工单',
      yesterday: '昨日遗留未处理售后'
    };
    return map[k] || k;
  },
  // 快捷操作：立即进入
  onOpenSession(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: '/chat/platform-im-sessions/platform-im-sessions?openId=' + id
    });
  },
  // 快捷操作：添加标签
  onAddTag(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: '/chat/platform-im-sessions/platform-im-sessions?openId=' + id + '&action=tag'
    });
  },
  // 快捷操作：备注
  onAddNote(e) {
    const id = e.currentTarget.dataset.id;
    const that = this;
    wx.showModal({
      title: '填写处理备注',
      editable: true,
      placeholderText: '矛盾关键点/沟通要点（仅自己可见）',
      success(r) {
        if (r.confirm) {
          request.put('/platform-im/sessions/' + id + '/note', { note_text: r.content }).then(() => {
            wx.showToast({ title: '已保存', icon: 'success' });
            that.loadWorkbench();
          });
        }
      }
    });
  },
  // 完成一条
  onComplete(e) {
    const id = e.currentTarget.dataset.id;
    const that = this;
    wx.showModal({
      title: '标记办结',
      content: '办结后将移入已办结分组',
      success(r) {
        if (r.confirm) {
          request.post('/platform-im/sessions/' + id + '/close').then(() => {
            wx.showToast({ title: '已办结', icon: 'success' });
            that.loadWorkbench();
          });
        }
      }
    });
  },
  onGotoSessionList() {
    wx.navigateTo({ url: '/chat/platform-im-sessions/platform-im-sessions' });
  },
  // 自定义顺序 - 62
  onChangeOrder() {
    const that = this;
    wx.showActionSheet({
      itemList: ['紧急任务置顶 推荐', '待跟进工单置顶', '昨日遗留置顶'],
      success(r) {
        const order = [
          ['emergency', 'todo', 'yesterday'],
          ['todo', 'emergency', 'yesterday'],
          ['yesterday', 'emergency', 'todo']
        ][r.tapIndex];
        request.put('/platform-im/workbench/layout', { layout_json: { order: order } }).then(() => {
          that.loadWorkbench();
        });
      }
    });
  }
});
