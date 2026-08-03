// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验（宵禁等）、权限控制、折扣计算。
const request = require('../../utils/request');

Page({
  data: {
    groupList: [],
    loading: false,
    showCreateModal: false,
    createForm: {
      name: '',
      type: 'chat'
    },
    // 纯 UI 常量：创建群聊表单的类型选项
    groupTypes: [
      { value: 'chat', label: '闲聊群' },
      { value: 'welfare', label: '福利群' },
      { value: 'after_sale', label: '售后群' }
    ]
  },

  onLoad() {
    this.loadGroupList();
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
      // 直接使用后端返回的文本字段，前端不做类型映射/消息格式化/相对时间计算
      const list = (res.list || []).map(item => ({
        ...item,
        group_type_label: item.group_type_text || '',
        last_msg_preview: item.last_message_preview || '',
        last_time: item.last_time_text || ''
      }));

      this.setData({
        groupList: list,
        loading: false
      });
    }).catch(() => {
      this.setData({
        groupList: [],
        loading: false
      });
    });
  },

  onOpenGroup(e) {
    const group = e.currentTarget.dataset.group;
    // 未成年人宵禁等业务校验由后端在进入群聊/发送消息时拦截
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
    // 非空提示（纯 UI 校验）
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
