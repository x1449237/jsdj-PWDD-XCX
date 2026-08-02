const request = require('../../utils/request');

// 平台方管理人员 - 会话归类清单 列表页（需求1~45）
const TABS = [
  { key: 'all', name: '全部会话' },
  { key: 'risk', name: '风险预警' },
  { key: 'player_int', name: '玩家介入' },
  { key: 'evidence_expiring', name: '举证超时' },
  { key: 'new_today', name: '当日新增' },
  { key: 'club_groups', name: '俱乐部群聊' },
  { key: 'done', name: '办结售后' }
];

Page({
  data: {
    tabs: TABS,
    tab: 'all',
    timeRange: '',
    clubKeyword: '',
    keyword: '',
    gameID: 0,
    tagID: 0,
    onlyStarred: false,
    history: [],
    groups: [],
    loading: false
  },
  onLoad(opt) {
    this.loadHistory();
    this.loadList();
  },
  onShow() { this.loadList(); },
  onPullDownRefresh() {
    this.loadList();
    wx.stopPullDownRefresh();
  },
  onTabClick(e) {
    const k = e.currentTarget.dataset.key;
    this.setData({ tab: k }, () => this.loadList());
  },
  onTimeRange(e) {
    const v = e.currentTarget.dataset.v;
    this.setData({ timeRange: this.data.timeRange === v ? '' : v }, () => this.loadList());
  },
  onClubInput(e) { this.setData({ clubKeyword: e.detail.value }); },
  onClubBlur() { this.loadList(); },
  onGameChange(e) {
    this.setData({ gameID: parseInt(e.detail.value || '0') }, () => this.loadList());
  },
  onStarSwitch() {
    this.setData({ onlyStarred: !this.data.onlyStarred }, () => this.loadList());
  },
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
      // request.js resolve(data.data), res 即为搜索历史数组
      const list = Array.isArray(res) ? res : (res.list || []);
      this.setData({ history: list.map(i => i.keyword) });
    }).catch(()=>{});
  },
  // 修复: 折叠分组用 bucket_key (后端 snake_case)
  onToggleBucket(e) {
    const k = e.currentTarget.dataset.k;
    const groups = this.data.groups.map(g => g.bucket_key === k ? { ...g, isCollapsed: !g.isCollapsed } : g);
    this.setData({ groups });
  },
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
      // request.js resolve(data.data), res = {list: [...groups], total: N}
      const rawGroups = (res && res.list) || res || [];
      const groups = rawGroups.map(g => ({
        ...g,
        isCollapsed: g.is_collapsed || false,
        items: (g.items || []).map(s => ({
          ...s,
          unreadDisplay: this.formatUnread(s.unread_count),
          riskCls: this.getRiskCls(s.risk_flag),
          riskLabel: this.getRiskLabel(s.risk_flag)
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
  getRiskLabel(f) {
    if (f === 1) return '【敏感词触发·强制介入】';
    if (f === 2) return '【买家申请官方介入】';
    return '';
  },
  // 修复: 跳转到已有的 after-sale-room 或 room 页面,而非不存在的 platform-im-chat
  onOpenSession(e) {
    const id = e.currentTarget.dataset.id;
    const s = e.currentTarget.dataset.s;
    // 售后会话 → after-sale-room,群聊 → group-room
    const type = s && s.session_type;
    if (type === 'group_internal' || type === 'group_category') {
      wx.navigateTo({ url: '/chat/group-room/group-room?groupId=' + id });
    } else {
      wx.navigateTo({ url: '/chat/after-sale-room/after-sale-room?sessionId=' + id });
    }
  },
  onOpenTagFilter() {
    request.get('/platform-im/tags').then(res => {
      const tags = Array.isArray(res) ? res : (res.list || []);
      if (tags.length === 0) {
        wx.showToast({ title: '暂无标签', icon: 'none' });
        return;
      }
      wx.showActionSheet({
        itemList: tags.map(t => t.name),
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
  // 修复: 星标切换用后端 star_flag 字段
  onToggleStar(e) {
    if (e.stopPropagation) e.stopPropagation();
    const id = e.currentTarget.dataset.id;
    const starred = e.currentTarget.dataset.starred === 1 ? 0 : 1;
    request.post('/platform-im/sessions/' + id + '/tags', { star_flag: starred }).then(() => this.loadList());
  },
  onQuickTag(e) {
    if (e.stopPropagation) e.stopPropagation();
    const id = e.currentTarget.dataset.id;
    const that = this;
    request.get('/platform-im/tags').then(res => {
      const tags = Array.isArray(res) ? res : (res.list || []);
      if (tags.length === 0) return wx.showToast({ title:'暂无标签', icon:'none' });
      wx.showActionSheet({
        itemList: tags.map(t => t.name),
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
  onBatchMute() {
    const that = this;
    wx.showModal({ title:'批量免打扰', content:'将所有沉寂群聊设置为免打扰?', success(r){
      if (r.confirm) {
        request.post('/platform-im/groups/batch-mute', { group_ids: [], mute: 1 }).then(() => {
          wx.showToast({ title:'已处理' });
          that.loadList();
        });
      }
    }});
  }
});
