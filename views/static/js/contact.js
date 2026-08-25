function filterContacts(keyword) {
  const normalized = keyword.trim().toLocaleLowerCase();
  const cards = Array.from(document.querySelectorAll('.contact-card'));
  let visible = 0;
  cards.forEach((card) => {
    const matched = !normalized || (card.dataset.search || '').toLocaleLowerCase().includes(normalized);
    card.classList.toggle('hidden', !matched);
    if (matched) visible += 1;
  });
  const noResults = document.querySelector('[data-testid="contact-no-results"]');
  if (noResults) noResults.classList.toggle('hidden', visible !== 0 || cards.length === 0);
}

function openDirectMessage(button) {
  const modal = document.getElementById('directMessageModal');
  const content = document.getElementById('directMessageContent');
  modal.dataset.wxid = button.dataset.wxid;
  modal.dataset.nickname = button.dataset.nickname || button.dataset.wxid;
  modal.querySelector('[data-testid="direct-message-target"]').textContent = modal.dataset.nickname;
  modal.querySelector('[data-testid="direct-message-error"]').textContent = '';
  content.value = '';
  updateDirectMessageCount();
  modal.showModal();
  window.setTimeout(() => content.focus(), 0);
}

function closeDirectMessage() {
  document.getElementById('directMessageModal').close();
}

function updateDirectMessageCount() {
  const content = document.getElementById('directMessageContent');
  document.querySelector('[data-testid="direct-message-count"]').textContent = `${Array.from(content.value).length} / 2000`;
}

async function confirmDirectMessage() {
  const modal = document.getElementById('directMessageModal');
  const content = document.getElementById('directMessageContent');
  const error = modal.querySelector('[data-testid="direct-message-error"]');
  const confirmButton = modal.querySelector('[data-testid="confirm-direct-message"]');
  const message = content.value.trim();
  if (!message) {
    error.textContent = '请输入消息内容';
    content.focus();
    return;
  }
  const wxId = modal.dataset.wxid;
  const trigger = document.querySelector(`[data-wxid="${CSS.escape(wxId)}"]`);
  const status = document.querySelector(`[data-testid="contact-message-status-${CSS.escape(wxId)}"]`);
  error.textContent = '';
  confirmButton.disabled = true;
  if (trigger) trigger.disabled = true;
  try {
    const response = await axios.post('/api/contacts/message', { wxId, message });
    modal.close();
    if (status) {
      status.textContent = response.data.message || '发送成功';
      status.className = 'min-w-0 truncate text-xs font-medium text-emerald-600';
    }
  } catch (requestError) {
    const messageText = requestError.response?.data?.message || '发送失败，请稍后重试';
    error.textContent = messageText;
    if (status) {
      status.textContent = messageText;
      status.className = 'min-w-0 truncate text-xs font-medium text-rose-600';
    }
  } finally {
    confirmButton.disabled = false;
    if (trigger) trigger.disabled = false;
  }
}
