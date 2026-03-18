/**
 * 与 Go 后端的 API 封装（流式对话、设置、评测）
 */
// 与后端同源时留空；file:// 打开时自动指向本地服务
// 覆盖方式：页面内 <meta name="api-base" content="http://127.0.0.1:8081"> 或脚本前设置 window.API_BASE
(function () {
  if (window.API_BASE != null) return;
  var meta = document.querySelector('meta[name="api-base"]');
  if (meta && meta.getAttribute('content')) {
    window.API_BASE = meta.getAttribute('content').replace(/\/$/, '');
    return;
  }
  if (window.location.protocol === 'file:') {
    window.API_BASE = 'http://127.0.0.1:8080';
  } else {
    window.API_BASE = '';
  }
})();
const API_BASE = window.API_BASE != null ? window.API_BASE : '';

function getConfig() {
  return {
    model: localStorage.getItem('translate_model') || 'gpt-4o-mini',
    temperature: parseFloat(localStorage.getItem('translate_temperature') || '0.7'),
    maxTokens: parseInt(localStorage.getItem('translate_max_tokens') || '2048', 10),
  };
}

function setConfig(config) {
  if (config.model != null) localStorage.setItem('translate_model', config.model);
  if (config.temperature != null) localStorage.setItem('translate_temperature', String(config.temperature));
  if (config.maxTokens != null) localStorage.setItem('translate_max_tokens', String(config.maxTokens));
}

/**
 * 流式翻译
 * @param {string} direction - 'product_to_dev' | 'dev_to_product'
 * @param {string} content - 用户输入
 * @param {string} [sessionId] - 可选，当前会话 ID；空则服务端创建新会话并在流结束时通过 onSessionId 回传
 * @param {object} callbacks - onChunk, onDone, onError, onSessionId(sessionId)
 */
async function streamTranslate(direction, content, sessionId, callbacks) {
  var onChunk = callbacks.onChunk;
  var onDone = callbacks.onDone;
  var onError = callbacks.onError;
  var onSessionId = callbacks.onSessionId;
  const cfg = getConfig();
  try {
    const res = await fetch(API_BASE + '/api/translate/stream', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_id: sessionId || undefined,
        direction,
        content,
        model: cfg.model,
        temperature: cfg.temperature,
        max_tokens: cfg.maxTokens,
      }),
    });
    if (!res.ok) {
      const t = await res.text();
      onError(new Error(t || res.statusText));
      return;
    }
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6);
          if (data === '[DONE]') continue;
          try {
            const j = JSON.parse(data);
            if (j.error) {
              var errMsg = j.error;
              if (j.detail) errMsg += '\n' + j.detail;
              if (j.source) errMsg += ' [' + j.source + ']';
              onError(new Error(errMsg));
              return;
            }
            if (j.session_id && onSessionId) onSessionId(j.session_id);
            const text = j.choices?.[0]?.delta?.content ?? j.text ?? '';
            if (text) onChunk(text);
          } catch (_) {}
        }
      }
    }
    if (onDone) onDone();
  } catch (e) {
    onError(e);
  }
}

/**
 * 获取模型列表（可选，用于设置页）
 */
async function fetchModels() {
  const res = await fetch(API_BASE + '/api/models');
  if (!res.ok) return [];
  const data = await res.json();
  return data.models || [];
}

/**
 * 保存设置
 */
async function saveSettings(settings) {
  setConfig(settings);
  const res = await fetch(API_BASE + '/api/settings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings),
  });
  return res.ok;
}

/**
 * 运行 Agent 评测
 * @param {string[]} caseIds - 用例 id 列表，空则全部
 */
async function runEvaluation(caseIds = []) {
  const res = await fetch(API_BASE + '/api/evaluate/run', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ case_ids: caseIds }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

/**
 * 获取评测用例列表
 */
async function fetchEvaluationCases() {
  const res = await fetch(API_BASE + '/api/evaluate/cases');
  if (!res.ok) return [];
  const data = await res.json();
  return data.cases || [];
}

/** 会话列表 */
async function listSessions(limit) {
  const res = await fetch(API_BASE + '/api/sessions' + (limit ? '?limit=' + limit : ''));
  if (!res.ok) return [];
  const data = await res.json();
  return data.sessions || [];
}

/** 会话详情（含消息历史） */
async function getSessionDetail(id) {
  const res = await fetch(API_BASE + '/api/sessions/' + encodeURIComponent(id));
  if (!res.ok) return null;
  return res.json();
}

/** 新建会话 */
async function createSession() {
  const res = await fetch(API_BASE + '/api/sessions', { method: 'POST' });
  if (!res.ok) return null;
  const data = await res.json();
  return data.id || null;
}

/** 删除会话 */
async function deleteSession(id) {
  const res = await fetch(API_BASE + '/api/sessions/' + encodeURIComponent(id), { method: 'DELETE' });
  return res.ok;
}
