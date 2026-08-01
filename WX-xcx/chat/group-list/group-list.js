const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    groupList: [],
    loading: false,
    showCreateModal: false,
    createForm: {
      name: '',
      type: 'chat'
    },
    groupTypes: [
      { value: 'chat', label: '闲聊群' },
      { value: 'welfare', label: '福利群' },
      { value: 'after_sale', label: '售后群' }
    ],
    isMinor: false
  },

  onLoad() {
    this.loadGroupList();
    this.checkUserAge();
  },

  checkUserAge() {
    request.get('/user/profile').then((res) => {
      if (res.is_minor) {
        this.setData({ isMinor: true });
      }
    }).catch(() => {});
  },

  onShow() {
    this.loadGroupList();
  },

  onPullDownRefresh() {
    this.loadGroupList();
    wx.stopPullDownRefresh();
  },

  loadGroupList() {
    if (this.data.loading) return;
    this.setData({ loading: true });

    request.get('/groups').then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        group_type_label: this.getGroupTypeLabel(item.group_type),
        last_msg_preview: this.formatLastMessage(item),
        last_time: util.formatRelativeTime(item.last_msg_time)
      }));

      this.setData({
        groupList: list,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  },

  getGroupTypeLabel(type) {
    const typeMap = {
      chat: '闲聊群',
      welfare: '福利群',
      after_sale: '售后群'
    };
    return typeMap[type] || '闲聊群';
  },

  formatLastMessage(item) {
    if (!item.last_msg_type) return '';
    const typeMap = {
      text: item.last_msg_content || '',
      image: '[图片]',
      voice: '[语音]',
      system: '[系统消息]',
      announcement: '[群公告]'
    };
    return typeMap[item.last_msg_type] || item.last_msg_content || '';
  },

  onOpenGroup(e) {
    const group = e.currentTarget.dataset.group;

    if (this.data.isMinor) {
      const currentHour = new Date().getHours();
      if (currentHour >= 22 || currentHour < 8) {
        wx.showModal({
          title: '未成年人宵禁提醒',
          content: '宵禁时间（22:00-次日08:00）未成年人无法进入群聊，请在白天再进行操作。',
          showCancel: false,
          confirmText: '我知道了'
        });
        return;
      }
    }

    wx.navigateTo({
      url: '/chat/group-room/group-room?groupId=' + group.group_id + '&groupName=' + encodeURIComponent(group.group_name || '群聊')
    });
  },

  onLongPress(e) {
    const group = e.currentTarget.dataset.group;
    wx.showActionSheet({
      itemList: ['退出群聊'],
      itemColor: '#e94560',
      success: (res) => {
        if (res.tapIndex === 0) {
          this.quitGroup(group);
        }
      }
    });
  },

  quitGroup(group) {
    wx.showModal({
      title: '退出群聊',
      content: '确定要退出「' + group.group_name + '」吗？',
      success: (res) => {
        if (res.confirm) {
          request.post('/groups/quit', {
            group_id: group.group_id
          }).then(() => {
            wx.showToast({ title: '已退出', icon: 'success' });
            this.loadGroupList();
          });
        }
      }
    });
  },

  onShowCreateModal() {
    this.setData({
      showCreateModal: true,
      createForm: { name: '', type: 'chat' }
    });
  },

  onHideCreateModal() {
    this.setData({ showCreateModal: false });
  },

  onNameInput(e) {
    this.setData({
      'createForm.name': e.detail.value
    });
  },

  onTypeSelect(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({
      'createForm.type': type
    });
  },

  onCreateGroup() {
    const { name, type } = this.data.createForm;
    if (!name.trim()) {
      wx.showToast({ title: '请输入群名称', icon: 'none' });
      return;
    }

    request.post('/groups', {
      name: name.trim(),
      type: type
    }).then((res) => {
      wx.showToast({ title: '创建成功', icon: 'success' });
      this.setData({ showCreateModal: false });
      this.loadGroupList();
      wx.navigateTo({
        url: '/chat/group-room/group-room?groupId=' + res.group_id + '&groupName=' + encodeURIComponent(name.trim())
      });
    });
  }
});