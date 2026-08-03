/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    serviceList: [],
    loading: true,
    showEditModal: false,
    editingService: {
      id: '',
      gameId: '',
      gameName: '',
      rankName: '',
      price: '',
      priceText: '',
      description: '',
      tags: [],
      status: 1
    },
    gameList: [],
    gameIndex: -1,
    rankList: [],
    rankIndex: -1,
    tagOptions: ['高效', '耐心', '专业', '新手友好', '包教包会', '全程直播', '语音沟通', '快速完成'],
    util: {
      isTagSelected: (tags, tag) => (tags || []).indexOf(tag) > -1
    }
  },

  onLoad() {
    this.loadGameList();
    this.loadServiceList();
  },

  /* ========== 数据加载 ========== */
  async loadGameList() {
    try {
      const res = await request.get('/common/games');
      const games = res.list || [];
      const rankList = [];
      this.setData({ gameList: games, rankList });
    } catch (err) {
      // 忽略
    }
  },

  async loadServiceList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/player/services');
      const list = (res.list || []).map(item => this.formatServiceItem(item));
      this.setData({ serviceList: list, loading: false });
    } catch (err) {
      this.setData({ loading: false });
    }
  },

  formatServiceItem(item) {
    return {
      ...item,
      gameIcon: item.gameIcon || '/assets/images/default-game.png',
      priceText: item.price_text || item.priceText || '',
      tags: item.tags || []
    };
  },

  /* ========== 上架/下架 ========== */
  async toggleServiceStatus(e) {
    const { id, status } = e.currentTarget.dataset;
    const newStatus = status === 1 ? 0 : 1;
    const statusText = newStatus === 1 ? '上架' : '下架';

    wx.showLoading({ title: `${statusText}中...` });
    try {
      await request.put(`/player/services/${id}/status`, { status: newStatus });
      wx.hideLoading();
      wx.showToast({ title: `${statusText}成功`, icon: 'success' });
      const list = this.data.serviceList.map(item => {
        if (item.id === id) {
          return { ...item, status: newStatus };
        }
        return item;
      });
      this.setData({ serviceList: list });
    } catch (err) {
      wx.hideLoading();
    }
  },

  /* ========== 新增/编辑 ========== */
  addService() {
    this.setData({
      showEditModal: true,
      editingService: {
        id: '',
        gameId: '',
        gameName: '',
        rankName: '',
        price: '',
        priceText: '',
        description: '',
        tags: [],
        status: 1
      },
      gameIndex: -1,
      rankIndex: -1
    });
  },

  editService(e) {
    const { id } = e.currentTarget.dataset;
    const item = this.data.serviceList.find(s => s.id === id);
    if (!item) return;

    const gameIndex = this.data.gameList.findIndex(g => g.id === item.gameId);
    const rankIndex = this.data.rankList.findIndex(r => r === item.rankName);

    this.setData({
      showEditModal: true,
      editingService: {
        id: item.id,
        gameId: item.gameId || '',
        gameName: item.gameName || '',
        rankName: item.rankName || '',
        price: item.priceText || '',
        priceText: item.priceText || '',
        description: item.description || '',
        tags: item.tags || [],
        status: item.status
      },
      gameIndex: gameIndex >= 0 ? gameIndex : -1,
      rankIndex: rankIndex >= 0 ? rankIndex : -1
    });
  },

  onGameChange(e) {
    const index = parseInt(e.detail.value);
    const game = this.data.gameList[index];
    const ranks = game.ranks || [];
    this.setData({
      gameIndex: index,
      rankList: ranks,
      rankIndex: -1,
      'editingService.gameId': game.id,
      'editingService.gameName': game.name,
      'editingService.rankName': ''
    });
  },

  onRankChange(e) {
    const index = parseInt(e.detail.value);
    const rankName = this.data.rankList[index];
    this.setData({
      rankIndex: index,
      'editingService.rankName': rankName
    });
  },

  onPriceInput(e) {
    this.setData({ 'editingService.price': e.detail.value });
  },

  onDescInput(e) {
    this.setData({ 'editingService.description': e.detail.value });
  },

  toggleTag(e) {
    const { tag } = e.currentTarget.dataset;
    const tags = [...this.data.editingService.tags];
    const index = tags.indexOf(tag);
    if (index > -1) {
      tags.splice(index, 1);
    } else {
      tags.push(tag);
    }
    this.setData({ 'editingService.tags': tags });
  },

  closeEditModal() {
    this.setData({ showEditModal: false });
  },

  async saveService() {
    const { editingService } = this.data;
    if (!editingService.gameId) {
      wx.showToast({ title: '请选择游戏', icon: 'none' });
      return;
    }
    if (!editingService.rankName) {
      wx.showToast({ title: '请选择段位', icon: 'none' });
      return;
    }
    if (!editingService.price || parseFloat(editingService.price) <= 0) {
      wx.showToast({ title: '请输入有效价格', icon: 'none' });
      return;
    }

    wx.showLoading({ title: '保存中...' });
    try {
      const data = {
        gameId: editingService.gameId,
        rankName: editingService.rankName,
        price: editingService.price,
        description: editingService.description || '',
        tags: editingService.tags || []
      };

      if (editingService.id) {
        await request.put(`/player/services/${editingService.id}`, data);
      } else {
        await request.post('/player/services', data);
      }

      wx.hideLoading();
      wx.showToast({ title: '保存成功', icon: 'success' });
      this.setData({ showEditModal: false });
      this.loadServiceList();
    } catch (err) {
      wx.hideLoading();
    }
  },

  /* ========== 删除服务 ========== */
  deleteService(e) {
    const { id } = e.currentTarget.dataset;
    wx.showModal({
      title: '确认删除',
      content: '删除后不可恢复，确定要删除该服务吗？',
      confirmColor: '#e94560',
      success: async (res) => {
        if (res.confirm) {
          wx.showLoading({ title: '删除中...' });
          try {
            await request.del(`/player/services/${id}`);
            wx.hideLoading();
            wx.showToast({ title: '已删除', icon: 'success' });
            this.loadServiceList();
          } catch (err) {
            wx.hideLoading();
          }
        }
      }
    });
  },

  noop() {}
});
