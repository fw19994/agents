/**
 * 与 Go 后端的 API 封装（流式对话、设置、会话等）
 */
// file:// 打开时：api-base 填服务根，如 http://127.0.0.1:9001；app-base 填 project_path，如 /translate-agent
// 或一条 api-root：http://127.0.0.1:9001/translate-agent
(function () {
  if (window.API_BASE != null) return;
  var meta = document.querySelector('meta[name="api-base"]');
  if (meta && meta.getAttribute('content')) {
    window.API_BASE = meta.getAttribute('content').replace(/\/$/, '');
    return;
  }
  if (window.location.protocol === 'file:') {
    window.API_BASE = 'http://127.0.0.1:9001';
  } else {
    window.API_BASE = '';
  }
})();

(function () {
  if (window.__API_ROOT__ != null) return;
  var rootMeta = document.querySelector('meta[name="api-root"]');
  if (rootMeta && rootMeta.getAttribute('content')) {
    window.__API_ROOT__ = rootMeta.getAttribute('content').replace(/\/$/, '');
    return;
  }
  var appBase = '';
  var appMeta = document.querySelector('meta[name="app-base"]');
  if (appMeta && appMeta.getAttribute('content')) {
    var v = appMeta.getAttribute('content').trim();
    if (v && v !== '/') appBase = v.replace(/\/$/, '');
  } else if (window.location.protocol !== 'file:') {
    var path = window.location.pathname || '';
    var m = path.match(/^(\/.+)\/(?:home|index|settings|evaluate)\.html$/i);
    if (m) appBase = m[1];
  }
  var host = window.API_BASE != null ? window.API_BASE : '';
  window.__API_ROOT__ = (host + appBase).replace(/\/$/, '') || '';
})();

const API_BASE = window.API_BASE != null ? window.API_BASE : '';

function apiUrl(path) {
  var p = path.startsWith('/') ? path : '/' + path;
  var root = window.__API_ROOT__ || '';
  return root + p;
}

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

async function streamTranslate(direction, content, sessionId, callbacks) {
  var onChunk = callbacks.onChunk;
  var onDone = callbacks.onDone;
  var onError = callbacks.onError;
  var onSessionId = callbacks.onSessionId;
  const cfg = getConfig();
  try {
    const res = await fetch(apiUrl('/api/translate/stream'), {
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

async function fetchModels() {
  const res = await fetch(apiUrl('/api/models'));
  if (!res.ok) return [];
  const data = await res.json();
  return data.models || [];
}

async function saveSettings(settings) {
  setConfig(settings);
  const res = await fetch(apiUrl('/api/settings'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings),
  });
  return res.ok;
}

async function runEvaluation(caseIds = []) {
  const res = await fetch(apiUrl('/api/evaluate/run'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ case_ids: caseIds }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

async function fetchEvaluationCases() {
  const res = await fetch(apiUrl('/api/evaluate/cases'));
  if (!res.ok) return [];
  const data = await res.json();
  return data.cases || [];
}

async function listSessions(limit) {
  const res = await fetch(apiUrl('/api/sessions') + (limit ? '?limit=' + limit : ''));
  if (!res.ok) return [];
  const data = await res.json();
  return data.sessions || [];
}

async function getSessionDetail(id) {
  const res = await fetch(apiUrl('/api/sessions/' + encodeURIComponent(id)));
  if (!res.ok) return null;
  return res.json();
}

async function createSession() {
  const res = await fetch(apiUrl('/api/sessions'), { method: 'POST' });
  if (!res.ok) return null;
  const data = await res.json();
  return data.id || null;
}

async function deleteSession(id) {
  const res = await fetch(apiUrl('/api/sessions/' + encodeURIComponent(id)), { method: 'DELETE' });
  return res.ok;
}
