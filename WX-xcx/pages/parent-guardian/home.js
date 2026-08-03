/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段（monthly_limit_text 等）
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数。
 * 约定：request.get(url, data) 直接 resolve 出内层 data.data，res 即数据对象本身。
 */
const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    bindList: [],
    bindListIndex: 0,
    selectedBindId: 0,
    currentChild: null,
    setting: {
      allow_order: 1,
      allow_reward: 1,
      is_frozen: 0,
      monthly_limit: 0
    },
    monthConsume: 0,
    monthlyLimit: 0,
    loading: false
  },

  onShow() {
    this.loadBindList();
  },

  onPullDownRefresh() {
    this.loadBindList().finally(() => {
      wx.stopPullDownRefresh();
    });
  },

  async loadBindList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/guardian/bind_list');
      const list = res || [];
      this.setData({ bindList: list });

      if (list.length > 0) {
        const activeIndex = list.findIndex(item => item.status === 1);
        const defaultIndex = activeIndex >= 0 ? activeIndex : 0;
        const firstBind = list[defaultIndex];
        this.setData({
          bindListIndex: defaultIndex,
          selectedBindId: firstBind.id
        });
        await this.loadChildInfo(firstBind.id);
      }
    } catch (err) {
      console.error('加载绑定列表失败:', err);
    } finally {
      this.setData({ loading: false });
    }
  },

  async loadChildInfo(bindId) {
    try {
      const [childRes, settingRes] = await Promise.all([
        request.get('/guardian/child_info', { bind_id: bindId }),
        request.get('/guardian/setting', { bind_id: bindId })
      ]);
      const childInfo = childRes || {};
      const settingInfo = settingRes || { allow_order: 1, allow_reward: 1, is_frozen: 0, monthly_limit: 0 };
      this.setData({
        currentChild: childInfo,
        setting: settingInfo,
        monthConsume: childInfo.month_consume || 0,
        monthlyLimit: settingInfo.monthly_limit || 0
      });
    } catch (err) {
      console.error('加载孩子信息失败:', err);
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({
        currentChild: null,
        setting: { allow_order: 1, allow_reward: 1, is_frozen: 0, monthly_limit: 0 },
        monthConsume: 0,
        monthlyLimit: 0
      });
    }
  },

  onBindChange(e) {
    const index = parseInt(e.detail.value || 0);
    const bind = this.data.bindList[index];
    if (bind) {
      this.setData({
        bindListIndex: index,
        selectedBindId: bind.id
      });
      this.loadChildInfo(bind.id);
    }
  },

  onGoConsumeReport() {
    const { selectedBindId } = this.data;
    if (!selectedBindId) return;
    wx.navigateTo({
      url: `/pages/parent-guardian/consume-report?bind_id=${selectedBindId}`
    });
  },

  onToggleOrder(e) {
    const allow = e.detail.value ? 1 : 0;
    this.updateSetting('allow_order', allow, allow ? '已开启下单权限' : '已关闭下单权限');
  },

  onToggleReward(e) {
    const allow = e.detail.value ? 1 : 0;
    this.updateSetting('allow_reward', allow, allow ? '已开启打赏权限' : '已关闭打赏权限');
  },

  onToggleFreeze(e) {
    const isFrozen = e.detail.value ? 1 : 0;
    const content = isFrozen ? '确定要冻结孩子账号吗？冻结后孩子将无法下单和打赏。' : '确定要解冻孩子账号吗？';
    wx.showModal({
      title: '确认操作',
      content: content,
      success: (res) => {
        if (res.confirm) {
          this.updateSetting('is_frozen', isFrozen, isFrozen ? '账号已冻结' : '账号已解冻');
        } else {
          this.setData({ 'setting.is_frozen': !isFrozen });
        }
      }
    });
  },

  async updateSetting(key, value, successMsg) {
    const { selectedBindId } = this.data;
    if (!selectedBindId) return false;

    try {
      let url = '';
      let data = {};
      if (key === 'monthly_limit') {
        url = '/guardian/monthly_limit';
        data = { bind_id: selectedBindId, monthly_limit: value };
      } else if (key === 'allow_order') {
        url = '/guardian/toggle_order';
        data = { bind_id: selectedBindId, allow: value };
      } else if (key === 'allow_reward') {
        url = '/guardian/toggle_reward';
        data = { bind_id: selectedBindId, allow: value };
      } else if (key === 'is_frozen') {
        url = '/guardian/toggle_freeze';
        data = { bind_id: selectedBindId, is_frozen: value };
      }

      await request.put(url, data);
      wx.showToast({ title: successMsg, icon: 'success' });

      const settingKey = `setting.${key}`;
      this.setData({ [settingKey]: value });
      return true;
    } catch (err) {
      wx.showToast({ title: err.message || '操作失败', icon: 'none' });
      return false;
    }
  },

  onEditLimit() {
    wx.showModal({
      title: '设置月消费限额',
      editable: true,
      placeholderText: '请输入限额（元）',
      content: this.data.setting.monthly_limit_text || '',
      success: (res) => {
        if (res.confirm && res.content) {
          // 直接提交元字符串，由后端校验与换算；成功后重新拉取以刷新 monthly_limit_text 等展示字段
          this.updateSetting('monthly_limit', res.content, '限额已更新').then((ok) => {
            if (ok && this.data.selectedBindId) {
              this.loadChildInfo(this.data.selectedBindId);
            }
          });
        }
      }
    });
  },

  onGoBind() {
    wx.navigateTo({ url: '/pages/parent-guardian/bind/bind' });
  },

  onGoChatSummary() {
    const { selectedBindId } = this.data;
    if (!selectedBindId) return;
    wx.showToast({ title: '聊天摘要功能开发中', icon: 'none' });
  },

  onUnbind() {
    const { selectedBindId } = this.data;
    if (!selectedBindId) return;
    wx.showModal({
      title: '确认解绑',
      content: '确定要解除监护绑定吗？解绑后将无法再查看和管理孩子账号。',
      confirmColor: '#f56c6c',
      success: (res) => {
        if (res.confirm) {
          this.doUnbind(selectedBindId);
        }
      }
    });
  },

  async doUnbind(bindId) {
    try {
      await request.post('/guardian/unbind', { bind_id: bindId });
      wx.showToast({ title: '解绑成功', icon: 'success' });
      setTimeout(() => {
        this.loadBindList();
      }, 1500);
    } catch (err) {
      wx.showToast({ title: err.message || '解绑失败', icon: 'none' });
    }
  }
});
