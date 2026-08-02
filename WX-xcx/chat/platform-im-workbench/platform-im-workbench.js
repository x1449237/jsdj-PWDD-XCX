const request = require('../../utils/request');

// 平台方管理人员 IM 工作台（需求 46-52）
Page({
  data: {
    loading: true,
    overview: {},
    buckets: [],
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
    // request.js resolve(data.data), res = overview 对象
    request.get('/platform-im/workbench').then(res => {
      const data = res || {};
      const overview = {
        count_emergency: data.count_emergency || 0,
        count_new_today: data.count_new_today || 0,
        count_timeout: data.count_timeout || 0
      };
      const bucketsOrder = data.buckets_order || ['emergency', 'todo', 'yesterday'];
      const rawBuckets = data.buckets || {};
      const buckets = bucketsOrder.map((k) => ({
        bucketKey: k,
        bucketName: this.getBucketName(k),
        items: rawBuckets[k] || []
      })).filter(b => b.items.length > 0 || b.bucketKey === 'emergency');
      this.setData({
        overview: overview,
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
  // 修复: 跳转到会话归类清单,而非不存在的页面
  onOpenSession(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: '/chat/platform-im-sessions/platform-im-sessions'
    });
  },
  onAddTag(e) {
    const id = e.currentTarget.dataset.id;
    const that = this;
    request.get('/platform-im/tags').then(res => {
      const tags = Array.isArray(res) ? res : (res.list || []);
      if (tags.length === 0) return wx.showToast({ title: '暂无标签', icon: 'none' });
      wx.showActionSheet({
        itemList: tags.map(t => t.name),
        success(r) {
          const t = tags[r.tapIndex];
          request.post('/platform-im/sessions/' + id + '/tags', { tag_ids: [t.id] }).then(() => {
            wx.showToast({ title: '已打标' });
            that.loadWorkbench();
          });
        }
      });
    });
  },
  // 修复: 字段名用 note_text (handler 兼容)
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
  // 修复: 用 bucket_order 字段名 (handler 兼容 layout_json.order 和 bucket_order)
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
        request.put('/platform-im/workbench/layout', { bucket_order: order }).then(() => {
          that.loadWorkbench();
        });
      }
    });
  }
});
