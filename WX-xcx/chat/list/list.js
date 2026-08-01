const request = require('../../utils/request');
const util = require('../../utils/util');
const websocket = require('../../utils/websocket');

Page({
  data: {
    currentTab: 'chat', // chat | group | after_sale
    conversationList: [],
    searchKeyword: '',
    loading: false,
    page: 1,
    pageSize: 20,
    hasMore: true
  },

  onLoad() {
    this.loadConversations();
    this.initWebSocket();
  },

  onShow() {
    this.loadConversations();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true, conversationList: [] });
    this.loadConversations();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadConversations();
    }
  },

  onUnload() {
    websocket.off('new_message', this.onNewMessage);
  },

  initWebSocket() {
    websocket.on('new_message', this.onNewMessage);
    websocket.on('message_read', this.onMessageRead);
  },

  onNewMessage(data) {
    this.loadConversations();
  },

  onMessageRead(data) {
    this.loadConversations();
  },

  onTabChange(e) {
    const tab = e.currentTarget.dataset.tab;
    if (tab === this.data.currentTab) return;

    if (tab === 'group') {
      wx.navigateTo({
        url: '/chat/group-list/group-list'
      });
      return;
    }

    if (tab === 'after_sale') {
      wx.navigateTo({
        url: '/chat/after-sale-list/after-sale-list'
      });
      return;
    }

    this.setData({ currentTab: tab });
  },

  onSearchInput(e) {
    this.setData({ searchKeyword: e.detail.value });
  },

  onSearch() {
    this.setData({ page: 1, hasMore: true, conversationList: [] });
    this.loadConversations();
  },

  loadConversations() {
    if (this.data.loading) return;
    this.setData({ loading: true });

    const params = {
      page: this.data.page,
      page_size: this.data.pageSize
    };

    if (this.data.searchKeyword.trim()) {
      params.keyword = this.data.searchKeyword.trim();
    }

    request.get('/chat/conversations', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        last_time: this.formatChatTime(item.last_time),
        last_message: this.formatLastMessage(item)
      }));

      this.setData({
        conversationList: this.data.page === 1 ? list : [...this.data.conversationList, ...list],
        loading: false,
        hasMore: list.length >= this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  formatChatTime(timestamp) {
    if (!timestamp) return '';
    const now = new Date();
    const date = new Date(timestamp);
    const diff = now - date;

    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前';
    if (diff < 86400000) return util.formatTime(timestamp, 'HH:mm');
    if (diff < 172800000) return '昨天 ' + util.formatTime(timestamp, 'HH:mm');
    if (diff < 604800000) {
      const days = ['日', '一', '二', '三', '四', '五', '六'];
      return '周' + days[date.getDay()] + ' ' + util.formatTime(timestamp, 'HH:mm');
    }
    return util.formatTime(timestamp, 'MM-DD HH:mm');
  },

  formatLastMessage(item) {
    if (!item.last_message) return '';
    const typeMap = {
      text: item.last_message_content || '',
      image: '[图片]',
      voice: '[语音]',
      system: '[系统消息]',
      order: '[订单消息]',
      recall: '消息已撤回'
    };
    return typeMap[item.last_message_type] || item.last_message_content || '';
  },

  onOpenChat(e) {
    const conversation = e.currentTarget.dataset.conversation;
    wx.navigateTo({
      url: '/chat/room/room?conversationId=' + conversation.conversation_id + '&targetName=' + encodeURIComponent(conversation.nickname)
    });
  },

  onLongPress(e) {
    const conversation = e.currentTarget.dataset.conversation;
    wx.showActionSheet({
      itemList: ['标记已读', '删除会话'],
      success: (res) => {
        if (res.tapIndex === 0) {
          this.markAsRead(conversation.conversation_id);
        } else if (res.tapIndex === 1) {
          this.deleteConversation(conversation.conversation_id);
        }
      }
    });
  },

  markAsRead(conversationId) {
    request.post('/chat/read', {
      conversation_id: conversationId
    }).then(() => {
      this.loadConversations();
    });
  },

  deleteConversation(conversationId) {
    wx.showModal({
      title: '确认删除',
      content: '删除后将清空聊天记录',
      success: (res) => {
        if (res.confirm) {
          request.del('/chat/conversations/' + conversationId).then(() => {
            wx.showToast({ title: '已删除', icon: 'success' });
            this.setData({ page: 1, hasMore: true, conversationList: [] });
            this.loadConversations();
          });
        }
      }
    });
  }
});