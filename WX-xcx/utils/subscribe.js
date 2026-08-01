/**
 * 订阅消息工具
 */
const request = require('./request');

/**
 * 存储订阅消息模板ID
 */
const TEMPLATE_IDS = {
  appeal_notify: 'TEMPLATE_ID_PLACEHOLDER_01',
  order_notify: 'TEMPLATE_ID_PLACEHOLDER_02',
  chat_notify: 'TEMPLATE_ID_PLACEHOLDER_03',
  platform_intervene: 'TEMPLATE_ID_PLACEHOLDER_04',
  after_sale_remind: 'TEMPLATE_ID_PLACEHOLDER_05'
};

/**
 * 获取指定场景的模板ID
 */
function getTemplateIds(scenes) {
  const ids = [];
  if (typeof scenes === 'string') {
    scenes = scenes.split(',').map(s => s.trim());
  }
  scenes.forEach(scene => {
    if (TEMPLATE_IDS[scene]) {
      ids.push(TEMPLATE_IDS[scene]);
    }
  });
  return ids;
}

/**
 * 请求订阅消息授权
 * @param {Array|String} scenes - 场景列表，如 ['order_notify', 'chat_notify']
 * @returns {Promise}
 */
function requestSubscribe(scenes) {
  return new Promise((resolve, reject) => {
    const ids = getTemplateIds(scenes);
    if (ids.length === 0) {
      reject(new Error('无有效模板ID'));
      return;
    }

    wx.requestSubscribeMessage({
      tmplIds: ids,
      success(res) {
        // 上报授权结果
        const accepted = [];
        const rejected = [];
        ids.forEach(id => {
          if (res[id] === 'accept') {
            accepted.push(id);
          } else {
            rejected.push(id);
          }
        });

        request.post('/subscribe/report', {
          accepted,
          rejected,
          scenes: Array.isArray(scenes) ? scenes : [scenes]
        }).catch(() => {});

        resolve({ accepted, rejected, result: res });
      },
      fail(err) {
        reject(err);
      }
    });
  });
}

/**
 * 在合适的时机统一请求订阅授权
 * 建议在用户完成关键操作后调用
 */
function requestAllSubscribe() {
  const keyScenes = ['order_notify', 'chat_notify', 'platform_intervene'];
  return requestSubscribe(keyScenes);
}

module.exports = {
  TEMPLATE_IDS,
  getTemplateIds,
  requestSubscribe,
  requestAllSubscribe
};