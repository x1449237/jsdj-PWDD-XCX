const request = require('../../utils/request');

// 平台方管理人员 - 会话归类清单 列表页（需求1~45）
const TABS = [
  { key: 'all', name: '全部会话', desc: '11 展示群聊+售后对话' },
  { key: 'risk', name: '风险预警会话', desc: '12 敏感词/强制介入' },
  { key: 'player_intervene', name: '玩家申请介入售后', desc: '13 手动发起介入' },
  { key: 'evidence_expiring', name: '待举证超时预警', desc: '14 2h内' },
  { key: 'new_today', name: '当日新增售后', desc: '15 今天新建' },
  { key: 'club_groups', name: '所有俱乐部群聊', desc: '16 闲聊/福利群' },
  { key: 'done', name: '办结售后', desc: '17 已处理完毕' }
];

Page({
  data: {
    tabs: TABS,
    tab: 'all',
    timeRange: '',        // yesterday/3d/7d (需求18)
    clubKeyword: '',
    keyword: '',
    gameID: 0,
    tagID: 0,
    onlyStarred: false,
    history: [],          // 需求25
    groups: [],           // 分组结果 [{bucketKey, bucketName, isCollapsed, items:[...]}]
    loading: false
  },
  onLoad(opt) {
    this.loadHistory();
    this.loadList();
    if (opt && opt.openId) {
      this._openSessionId = opt.openId;
    }
  },
  onShow() { this.loadList(); },
  onPullDownRefresh() {
    this.loadList();
    wx.stopPullDownRefresh();
  },
  // 顶部7大筛选 + 历史/俱乐部分组 需求11-20
  onTabClick(e) {
    const k = e.currentTarget.dataset.key;
    this.setData({ tab: k }, () => this.loadList());
  },
  onTimeRange(e) {
    const v = e.currentTarget.dataset.v;
    this.setData({ timeRange: this.data.timeRange === v ? '' : v }, () => this.loadList());
  },
  onClubInput(e) {
    this.setData({ clubKeyword: e.detail.value });
  },
  onClubBlur() { this.loadList(); },
  onGameChange(e) {
    this.setData({ gameID: parseInt(e.detail.value || '0') }, () => this.loadList());
  },
  onStarSwitch() {
    this.setData({ onlyStarred: !this.data.onlyStarred }, () => this.loadList());
  },
  // 搜索 需求21-26
  onSearchInput(e) { this.setData({ keyword: e.detail.value }); },
  onSearchConfirm() {
    const kw = (this.data.keyword || '').trim();
    if (!kw) return;
    this.loadList();
  },
  onHistoryClick(e) {
    const k = e.currentTarget.dataset.k;
    this.setData({ keyword: k }, () => this.loadList());
  },
  onHistoryClear() {
    const that = this;
    wx.showModal({
      title: '清空搜索历史?',
      success(r) {
        if (r.confirm) {
          request.post('/platform-im/search-history/clear').then(() => that.loadHistory());
        }
      }
    });
  },
  loadHistory() {
    request.get('/platform-im/search-history').then(res => {
      this.setData({ history: (res.data || []).map(i => i.keyword) });
    }).catch(()=>{});
  },
  // 分组折叠 需求43
  onToggleBucket(e) {
    const k = e.currentTarget.dataset.k;
    const groups = this.data.groups.map(g => g.bucketKey === k ? { ...g, isCollapsed: !g.isCollapsed } : g);
    this.setData({ groups });
  },
  // 列表加载
  loadList() {
    this.setData({ loading: true });
    const params = {
      tab: this.data.tab,
      page: 1, page_size: 100,
      time_range: this.data.timeRange,
      club_keyword: this.data.clubKeyword,
      keyword: this.data.keyword,
      game_id: this.data.gameID,
      tag_id: this.data.tagID,
      starred: this.data.onlyStarred ? 1 : 0
    };
    request.get('/platform-im/sessions', params).then(res => {
      const groups = (res.list || []).map(g => ({
        ...g,
        items: (g.items || []).map(s => ({
          ...s,
          unreadDisplay: this.formatUnread(s.unread_count),
          riskCls: this.getRiskCls(s.risk_flag),
          riskLabel: this.getRiskLabel(s.risk_flag, s.remark_tags)
        }))
      }));
      this.setData({ groups, loading: false });
    }).catch(() => this.setData({ loading: false }));
  },
  formatUnread(c) {
    if (!c) return '';
    return c > 99 ? '99+' : '' + c;
  },
  getRiskCls(f) {
    const m = { 4: 'risk-4', 3: 'risk-3', 2: 'risk-2', 1: 'risk-1' };
    return m[f] || '';
  },
  getRiskLabel(f, tags) {
    // 36/37/38
    if (f === 4) return '【敏感词触发·强制介入】';
    if (f === 3) return '【买家申请官方介入】';
    return '';
  },
  // 打开会话
  onOpenSession(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: '/chat/platform-im-chat/platform-im-chat?sessionId=' + id
    });
  },
  // 标签筛选 需求30
  onOpenTagFilter() {
    request.get('/platform-im/tags').then(res => {
      const tags = res.data || [];
      if (tags.length === 0) {
        wx.showToast({ title: '暂无标签，请先在工作台创建', icon: 'none' });
        return;
      }
      wx.showActionSheet({
        itemList: tags.map(t => t.tag_name + '  ·  (' + (t.color || '') + ')'),
        success: (r) => {
          const t = tags[r.tapIndex];
          this.setData({ tagID: t.id }, () => this.loadList());
        }
      });
    });
  },
  onClearTagFilter() {
    this.setData({ tagID: 0 }, () => this.loadList());
  },
  // 标记办结 需求31
  onMarkDone(e) {
    const id = e.currentTarget.dataset.id;
    const that = this;
    wx.showModal({ title:'确认办结该会话?', success(r){
      if (r.confirm) {
        request.post('/platform-im/sessions/' + id + '/close').then(() => {
          wx.showToast({ title: '已办结' });
          that.loadList();
        });
      }
    }});
  },
  // 加星标 需求34
  onToggleStar(e) {
    e.stopPropagation && e.stopPropagation();
    const id = e.currentTarget.dataset.id;
    const tag = (this.data.groups || []).reduce((acc,g)=>{
      const it = (g.items||[]).find(x=>x.session_id===id);
      return it || acc;
    }, null);
    const starred = tag && tag.star_flag === 1 ? 0 : 1;
    request.post('/platform-im/sessions/' + id + '/tags', {
      tag_ids: [], star_flag: starred
    }).then(() => this.loadList());
  },
  // 快捷标签操作
  onQuickTag(e) {
    e.stopPropagation && e.stopPropagation();
    const id = e.currentTarget.dataset.id;
    const that = this;
    request.get('/platform-im/tags').then(res => {
      const tags = res.data || [];
      if (tags.length === 0) return wx.showToast({ title:'暂无标签', icon:'none' });
      wx.showActionSheet({
        itemList: tags.map(t => t.tag_name),
        success(r) {
          const t = tags[r.tapIndex];
          request.post('/platform-im/sessions/' + id + '/tags', { tag_ids:[t.id] }).then(() => {
            wx.showToast({ title:'已打标' });
            that.loadList();
          });
        }
      });
    });
  },
  // 批量免打扰 需求41
  onBatchMute() {
    const that = this;
    wx.showModal({ title:'批量免打扰', content:'将所有沉寂群聊设置为免打扰?', success(r){
      if (r.confirm) {
        request.post('/platform-im/groups/batch-mute', { group_ids: [], mute: 1, apply_silent: 1 }).then(() => {
          wx.showToast({ title:'已处理' });
          that.loadList();
        });
      }
    }});
  }
});
