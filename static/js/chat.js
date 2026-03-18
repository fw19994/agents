(function () {
  const messagesEl = document.getElementById('messages');
  const welcomeEl = document.getElementById('welcome');
  const userInput = document.getElementById('user-input');
  const btnSend = document.getElementById('btn-send');

  function md(text) {
    return typeof renderAssistantMarkdown === 'function'
      ? renderAssistantMarkdown(text)
      : typeof escapeHtml === 'function'
        ? '<p class="md-fallback">' + escapeHtml(text || '').replace(/\n/g, '<br>') + '</p>'
        : '';
  }

  let direction = 'product_to_dev'; // product_to_dev | dev_to_product

  document.querySelectorAll('.dir-btn').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var newDir =
        this.id === 'dir-product-to-dev'
          ? 'product_to_dev'
          : this.id === 'dir-dev-to-product'
            ? 'dev_to_product'
            : 'ops_to_product';
      if (newDir !== direction) {
        window.currentSessionId = null;
        if (typeof window.onCurrentSessionChange === 'function') {
          window.onCurrentSessionChange(null);
        }
        if (typeof window.clearMessagesAndWelcome === 'function') {
          window.clearMessagesAndWelcome();
        }
        if (typeof window.refreshSessionList === 'function') {
          window.refreshSessionList();
        }
      }
      document.querySelectorAll('.dir-btn').forEach(function (b) {
        b.classList.remove('bg-indigo-600', 'text-white');
        b.classList.add('text-slate-600');
      });
      this.classList.add('bg-indigo-600', 'text-white');
      this.classList.remove('text-slate-600');
      direction = newDir;
    });
  });

  const modelSelect = document.getElementById('model-select');
  if (modelSelect) {
    function applyModels(models) {
      if (!models || !models.length) return;
      const saved = localStorage.getItem('translate_model');
      modelSelect.innerHTML = '';
      models.forEach(function (id) {
        const opt = document.createElement('option');
        opt.value = id;
        opt.textContent = id;
        modelSelect.appendChild(opt);
      });
      if (saved && models.indexOf(saved) >= 0) modelSelect.value = saved;
      else if (models.length) modelSelect.value = models[0];
    }
    if (typeof fetchModels === 'function') {
      fetchModels().then(applyModels).catch(function () {
        applyModels(['gpt-4o-mini', 'gpt-4o', 'qwen-plus']);
      });
    } else {
      applyModels(['gpt-4o-mini', 'gpt-4o', 'qwen-plus']);
    }
    modelSelect.addEventListener('change', function () {
      if (typeof setConfig === 'function') setConfig({ model: this.value });
    });
  }

  const paramTemp = document.getElementById('param-temperature');
  const paramTempValue = document.getElementById('param-temperature-value');
  const paramMaxTokens = document.getElementById('param-max-tokens');
  if (typeof getConfig === 'function') {
    var cfg = getConfig();
    if (paramTemp) {
      paramTemp.value = Math.round((cfg.temperature || 0.7) * 100);
      if (paramTempValue) paramTempValue.textContent = (cfg.temperature || 0.7).toFixed(1);
    }
    if (paramMaxTokens) paramMaxTokens.value = cfg.maxTokens || 2048;
  }
  if (paramTemp) {
    paramTemp.addEventListener('input', function () {
      var v = (parseInt(this.value, 10) / 100).toFixed(1);
      if (paramTempValue) paramTempValue.textContent = v;
      if (typeof setConfig === 'function') setConfig({ temperature: parseFloat(v) });
    });
  }
  if (paramMaxTokens) {
    paramMaxTokens.addEventListener('change', function () {
      var v = parseInt(this.value, 10) || 2048;
      v = Math.min(8192, Math.max(256, v));
      this.value = v;
      if (typeof setConfig === 'function') setConfig({ maxTokens: v });
    });
  }

  function hideWelcome() {
    if (welcomeEl) welcomeEl.style.display = 'none';
  }

  /**
   * @returns {HTMLElement} 助手为 .msg-content（内含 .md-body）；用户为内容容器
   */
  function appendMessage(role, content, isStreaming, directionOverride) {
    hideWelcome();
    const dir = directionOverride != null ? directionOverride : direction;
    const isUser = role === 'user';
    const row = document.createElement('div');
    row.className = 'flex gap-3 ' + (isUser ? 'justify-end' : '');

    const bubble = document.createElement('div');
    bubble.className =
      'max-w-[88%] rounded-2xl px-4 py-3 ' +
      (isUser
        ? 'bg-indigo-600 text-white shadow-md ring-1 ring-indigo-500/15'
        : 'bg-white border border-slate-200/90 shadow-sm');

    const label = document.createElement('div');
    label.className =
      'text-xs mb-1.5 font-medium ' +
      (isUser ? 'text-indigo-100/95' : 'text-indigo-600');
    label.textContent = isUser
      ? dir === 'product_to_dev'
        ? '产品侧'
        : dir === 'dev_to_product'
          ? '开发侧'
          : '运营侧'
      : '翻译结果';

    const contentWrap = document.createElement('div');
    contentWrap.className = 'msg-content text-[15px] leading-relaxed break-words';

    if (isUser) {
      contentWrap.classList.add('whitespace-pre-wrap');
      contentWrap.textContent = content || '';
    } else {
      const mdEl = document.createElement('div');
      mdEl.className = 'md-body';
      mdEl.innerHTML = md(content || '');
      contentWrap.appendChild(mdEl);
      if (isStreaming) {
        const sp = document.createElement('span');
        sp.className = 'cursor-blink text-indigo-500 font-light inline-block align-baseline';
        sp.textContent = '|';
        contentWrap.appendChild(sp);
      }
    }

    bubble.appendChild(label);
    bubble.appendChild(contentWrap);
    row.appendChild(bubble);
    messagesEl.appendChild(row);
    return contentWrap;
  }

  function scrollToBottom() {
    messagesEl.scrollTop = messagesEl.scrollHeight;
  }

  function setAssistantMarkdown(contentEl, text, showCursor) {
    var mdEl = contentEl.querySelector('.md-body');
    if (!mdEl) {
      mdEl = document.createElement('div');
      mdEl.className = 'md-body';
      var cur = contentEl.querySelector('.cursor-blink');
      if (cur) contentEl.insertBefore(mdEl, cur);
      else contentEl.appendChild(mdEl);
    }
    mdEl.innerHTML = md(text || '');
    var cur = contentEl.querySelector('.cursor-blink');
    if (showCursor) {
      if (!cur) {
        cur = document.createElement('span');
        cur.className = 'cursor-blink text-indigo-500 font-light inline-block align-baseline';
        cur.textContent = '|';
        contentEl.appendChild(cur);
      }
    } else if (cur) {
      cur.remove();
    }
  }

  window.clearMessagesAndWelcome = function () {
    if (welcomeEl) welcomeEl.style.display = '';
    messagesEl.querySelectorAll(':scope > .flex.gap-3').forEach(function (el) {
      el.remove();
    });
    scrollToBottom();
  };

  window.renderHistory = function (messages) {
    if (!messages || !messages.length) return;
    hideWelcome();
    messages.forEach(function (m) {
      var d = m.direction || direction;
      appendMessage('user', m.input, false, d);
      appendMessage('assistant', m.output, false, d);
    });
    scrollToBottom();
  };

  btnSend.addEventListener('click', sendMessage);
  userInput.addEventListener('keydown', function (e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  });

  function sendMessage() {
    const text = (userInput.value || '').trim();
    if (!text) return;
    userInput.value = '';
    btnSend.disabled = true;

    appendMessage('user', text);
    const contentEl = appendMessage('assistant', '', true);
    scrollToBottom();

    var currentSessionId = window.currentSessionId || null;
    function setCurrentSessionId(id) {
      window.currentSessionId = id;
      if (typeof onCurrentSessionChange === 'function') onCurrentSessionChange(id);
    }
    let full = '';
    if (typeof streamTranslate === 'function') {
      streamTranslate(direction, text, currentSessionId, {
        onChunk: function (chunk) {
          full += chunk;
          setAssistantMarkdown(contentEl, full, true);
          scrollToBottom();
        },
        onDone: function () {
          setAssistantMarkdown(contentEl, full, false);
          btnSend.disabled = false;
          scrollToBottom();
        },
        onError: function (err) {
          var msg = (full ? full + '\n\n' : '') + '[错误] ' + (err && err.message ? err.message : '请求失败');
          var mdEl = contentEl.querySelector('.md-body');
          if (!mdEl) {
            mdEl = document.createElement('div');
            mdEl.className = 'md-body';
            contentEl.appendChild(mdEl);
          }
          mdEl.innerHTML =
            '<div class="rounded-lg bg-red-50 border border-red-100 text-red-800 px-3 py-2 text-sm whitespace-pre-wrap">' +
            (typeof escapeHtml === 'function' ? escapeHtml(msg) : msg) +
            '</div>';
          contentEl.querySelectorAll('.cursor-blink').forEach(function (n) {
            n.remove();
          });
          btnSend.disabled = false;
          scrollToBottom();
        },
        onSessionId: function (id) {
          setCurrentSessionId(id);
          if (typeof window.refreshSessionList === 'function') window.refreshSessionList();
        },
      });
    } else {
      const mock =
        direction === 'product_to_dev'
          ? '## 技术解读\n\n- **方案**：协同过滤或内容推荐\n- **数据**：需明确用户行为与物品元数据\n- **排期**：建议按埋点 → 特征 → 服务拆分估时\n\n---\n\n> 以上为示例结构化输出。'
          : '## 业务价值\n\n| 维度 | 说明 |\n|------|------|\n| 体验 | 响应更快、更流畅 |\n| 增长 | 可支撑更高并发 |\n\n**结论**：对留存与转化有正向作用。';
      let i = 0;
      const tid = setInterval(function () {
        if (i >= mock.length) {
          clearInterval(tid);
          setAssistantMarkdown(contentEl, full, false);
          btnSend.disabled = false;
          return;
        }
        full += mock[i++];
        setAssistantMarkdown(contentEl, full, true);
        scrollToBottom();
      }, 18);
    }
  }
})();
