const auth = require('./auth');

let socketTask = null;
let isConnecting = false;
let reconnectTimer = null;
let heartbeatTimer = null;
let timeoutTimer = null;
let messageQueue = [];
let listeners = {};
let offlinePuller = null;

// Go 后端 WebSocket 服务地址
const WS_URL = 'wss://your-domain.com/api/v1/ws';
const HEARTBEAT_INTERVAL = 25000;   // 心跳间隔 25 秒
const TIMEOUT = 70000;              // 超时时间 70 秒，超过未收到服务端消息则判定断线
const RECONNECT_INTERVAL = 3000;    // 重连间隔
const MAX_RECONNECT_COUNT = 10;

let reconnectCount = 0;
let lastReceivedAt = 0;

const connect = () => {
  if (isConnecting || (socketTask && socketTask.readyState === 1)) {
    return;
  }

  const token = auth.getToken();
  if (!token) {
    isConnecting = false;
    return;
  }

  isConnecting = true;

  socketTask = wx.connectSocket({
    url: `${WS_URL}?token=${token}`,
    success() {
      console.log('WebSocket 连接中...');
    },
    fail(err) {
      console.error('WebSocket 连接失败:', err);
      isConnecting = false;
      scheduleReconnect();
    }
  });

  socketTask.onOpen(() => {
    console.log('WebSocket 连接成功');
    isConnecting = false;
    reconnectCount = 0;
    lastReceivedAt = Date.now();
    const app = getApp();
    if (app && app.globalData) {
      app.globalData.wsConnected = true;
    }
    startHeartbeat();
    startTimeoutWatch();
    flushMessageQueue();
    // 断线重连后拉取离线消息
    pullOfflineMessages();
    notify('reconnected', { time: Date.now() });
  });

  socketTask.onMessage((res) => {
    lastReceivedAt = Date.now();
    try {
      const data = JSON.parse(res.data);
      handleMessage(data);
    } catch (e) {
      console.error('WebSocket 消息解析失败:', e);
    }
  });

  socketTask.onClose((res) => {
    console.log('WebSocket 连接关闭:', res);
    isConnecting = false;
    const app = getApp();
    if (app && app.globalData) {
      app.globalData.wsConnected = false;
    }
    stopHeartbeat();
    stopTimeoutWatch();
    scheduleReconnect();
  });

  socketTask.onError((err) => {
    console.error('WebSocket 错误:', err);
    isConnecting = false;
    const app = getApp();
    if (app && app.globalData) {
      app.globalData.wsConnected = false;
    }
    stopHeartbeat();
    stopTimeoutWatch();
    scheduleReconnect();
  });
};

const scheduleReconnect = () => {
  if (reconnectCount >= MAX_RECONNECT_COUNT) {
    console.log('WebSocket 重连次数已达上限');
    return;
  }

  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
  }

  reconnectCount++;
  console.log(`WebSocket 将在 ${RECONNECT_INTERVAL / 1000} 秒后重连 (第${reconnectCount}次)`);

  reconnectTimer = setTimeout(() => {
    connect();
  }, RECONNECT_INTERVAL);
};

const startHeartbeat = () => {
  stopHeartbeat();
  heartbeatTimer = setInterval(() => {
    send({ type: 'ping' });
  }, HEARTBEAT_INTERVAL);
};

const stopHeartbeat = () => {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
    heartbeatTimer = null;
  }
};

// 超时看门狗：超过 TIMEOUT 未收到任何服务端消息则主动断开并重连
const startTimeoutWatch = () => {
  stopTimeoutWatch();
  timeoutTimer = setInterval(() => {
    if (Date.now() - lastReceivedAt > TIMEOUT) {
      console.warn('WebSocket 心跳超时，主动断开重连');
      stopHeartbeat();
      stopTimeoutWatch();
      if (socketTask) {
        try {
          socketTask.close({});
        } catch (e) {}
      }
      isConnecting = false;
      scheduleReconnect();
    }
  }, HEARTBEAT_INTERVAL);
};

const stopTimeoutWatch = () => {
  if (timeoutTimer) {
    clearInterval(timeoutTimer);
    timeoutTimer = null;
  }
};

const handleMessage = (data) => {
  const { type } = data;

  if (type === 'pong') return;

  // 消息类型路由
  const typeMap = {
    'chat_message': 'chat_message',
    'group_chat': 'group_chat',
    'group_chat_message': 'group_chat',
    'after_sale': 'after_sale',
    'after_sale_message': 'after_sale',
    'platform_intervene': 'platform_intervene',
    'new_message': 'new_message',
    'message_read': 'message_read'
  };

  const mappedType = typeMap[type] || type;

  if (listeners[mappedType]) {
    listeners[mappedType].forEach(callback => callback(data));
  }

  // 同时触发原始 type 的监听器（兼容旧代码）
  if (mappedType !== type && listeners[type]) {
    listeners[type].forEach(callback => callback(data));
  }

  if (listeners['*']) {
    listeners['*'].forEach(callback => callback(data));
  }
};

const notify = (type, data) => {
  if (listeners[type]) {
    listeners[type].forEach(callback => callback(data));
  }
};

// 离线消息拉取：重连成功后调用外部注册的拉取函数
const pullOfflineMessages = () => {
  if (typeof offlinePuller === 'function') {
    try {
      offlinePuller();
    } catch (e) {
      console.error('拉取离线消息失败:', e);
    }
  }
};

const setOfflinePuller = (fn) => {
  offlinePuller = fn;
};

const flushMessageQueue = () => {
  while (messageQueue.length > 0) {
    const msg = messageQueue.shift();
    send(msg);
  }
};

const send = (data) => {
  const msg = typeof data === 'string' ? data : JSON.stringify(data);

  if (socketTask && socketTask.readyState === 1) {
    socketTask.send({
      data: msg,
      fail(err) {
        console.error('WebSocket 发送失败:', err);
        messageQueue.push(data);
      }
    });
  } else {
    messageQueue.push(data);
  }
};

const on = (type, callback) => {
  if (!listeners[type]) {
    listeners[type] = [];
  }
  listeners[type].push(callback);
};

const off = (type, callback) => {
  if (!listeners[type]) return;
  if (callback) {
    const index = listeners[type].indexOf(callback);
    if (index > -1) {
      listeners[type].splice(index, 1);
    }
  } else {
    listeners[type] = [];
  }
};

const close = () => {
  stopHeartbeat();
  stopTimeoutWatch();
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  reconnectCount = MAX_RECONNECT_COUNT;
  if (socketTask) {
    socketTask.close();
    socketTask = null;
  }
  isConnecting = false;
  const app = getApp();
  if (app && app.globalData) {
    app.globalData.wsConnected = false;
  }
};

module.exports = {
  connect,
  send,
  on,
  off,
  close,
  setOfflinePuller
};
