console.log("FlyAI 已加载")

async function putJson(url, data, errorMessage) {
  const response = await fetch(url, {
    method: "PUT",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    throw new Error((await response.text()) || errorMessage)
  }
  return response
}

function showPageActionStatus(message, type = "success") {
  const status = document.querySelector('[data-testid="page-action-status"]')
  status.textContent = message
  status.className = `mb-6 rounded-xl px-4 py-3 text-sm ${type === "error" ? "bg-red-50 text-red-700" : type === "loading" ? "bg-blue-50 text-blue-700" : "bg-green-50 text-green-700"}`
}

async function syncWechatCatalog() {
  const button = document.querySelector('[data-testid="sync-wechat"]')
  const originalText = button.textContent
  button.disabled = true
  button.textContent = "同步中…"
  showPageActionStatus("正在读取微信群、通讯录和当前群成员，请稍候…", "loading")
  try {
    const response = await fetch("/api/sync/wechat", {method: "POST", headers: {"Content-Type": "application/json"}, body: "{}"})
    const result = await response.json().catch(() => ({message: "服务器返回异常"}))
    if (!response.ok) throw new Error(result.message || "同步失败")
    showPageActionStatus(`${result.message}：${result.groups} 个群、${result.contacts} 位联系人、${result.members} 位当前群成员`)
  } catch (error) {
    console.error(error)
    showPageActionStatus(error.message || "同步失败", "error")
  } finally {
    button.disabled = false
    button.textContent = originalText
  }
}

// 主开关关闭后，桥接层不会读取或转发该群的新消息。
async function changeGroupMonitorStatus(wxId) {
  try {
    await putJson("/api/monitor/status", {wxId}, "修改群监控状态失败")
    window.location.reload()
  } catch (error) {
    console.error(error)
    showPageActionStatus("修改群监控状态失败", "error")
  }
}

async function changeGroupRankEnableStatus(wxId) {
  try {
    await putJson("/api/grouprank/status", {wxId}, "修改排行榜状态失败")
    window.location.reload()
  } catch (error) {
    console.error(error)
    showPageActionStatus("修改排行榜状态失败", "error")
  }
}

async function changeGroupNewsStatus(wxId) {
  try {
    await putJson("/api/news/status", {wxId}, "修改每日热点早报状态失败")
    window.location.reload()
  } catch (error) {
    console.error(error)
    showPageActionStatus("修改每日热点早报状态失败", "error")
  }
}

async function changeGroupDormantStatus(wxId) {
  try {
    await putJson("/api/dormant/status", {wxId}, "修改潜水成员统计状态失败")
    window.location.reload()
  } catch (error) {
    console.error(error)
    showPageActionStatus("修改潜水成员统计状态失败", "error")
  }
}

function sendWeeklyRank(button) {
  executeWeeklyRank(button.dataset.wxid)
}

function sendAllWeeklyRanks() {
  executeAllWeeklyRanks()
}

function sendAINews(button) {
  executeAINews(button.dataset.wxid)
}

function sendDormantReport(button) {
  executeDormantReport(button.dataset.wxid)
}

async function executeWeeklyRank(wxid) {
  const button = document.querySelector(`[data-testid="send-weekly-${CSS.escape(wxid)}"]`)
  const status = document.querySelector(`[data-testid="weekly-status-${CSS.escape(wxid)}"]`)
  const originalText = button.textContent
  button.disabled = true
  button.textContent = "发送中…"
  status.textContent = ""
  status.className = "hidden max-w-full text-xs"

  try {
    const response = await fetch("/api/tasks/weekly-rank", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({wxId: wxid}),
    })
    const result = await response.json().catch(() => ({message: "服务器返回异常"}))
    if (!response.ok) throw new Error(result.message || "发送失败")
    status.textContent = `✓ ${result.message}`
    status.className = "inline-flex max-w-full items-center rounded-full bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700"
  } catch (error) {
    console.error(error)
    status.textContent = error.message || "发送失败"
    status.className = "inline-flex max-w-full items-center rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700"
  } finally {
    button.disabled = false
    button.textContent = originalText
  }
}

async function executeAllWeeklyRanks() {
  const button = document.querySelector('[data-testid="send-all-weekly"]')
  const originalText = button.textContent
  button.disabled = true
  button.textContent = "批量发送中…"
  showPageActionStatus("正在批量生成并发送排行榜…", "loading")

  try {
    const response = await fetch("/api/tasks/weekly-rank/all", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({}),
    })
    const result = await response.json().catch(() => ({message: "服务器返回异常"}))
    if (!response.ok) throw new Error(result.message || "批量发送失败")
    showPageActionStatus(result.message, result.failed > 0 ? "error" : "success")
  } catch (error) {
    console.error(error)
    showPageActionStatus(error.message || "批量发送失败", "error")
  } finally {
    button.disabled = false
    button.textContent = originalText
  }
}

async function executeAINews(wxId) {
  const button = document.querySelector(`[data-testid="send-ai-news-${CSS.escape(wxId)}"]`)
  const status = document.querySelector(`[data-testid="ai-news-status-${CSS.escape(wxId)}"]`)
  const originalText = button.textContent
  button.disabled = true
  button.textContent = "发送中…"
  status.textContent = ""
  status.className = "hidden max-w-full text-xs"
  try {
    const response = await fetch("/api/tasks/ai-news", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({wxId}),
    })
    const result = await response.json().catch(() => ({message: "服务器返回异常"}))
    if (!response.ok) throw new Error(result.message || "每日热点早报发送失败")
    status.textContent = `✓ ${result.message}`
    status.className = "inline-flex max-w-full items-center rounded-full bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700"
  } catch (error) {
    console.error(error)
    status.textContent = error.message || "每日热点早报发送失败"
    status.className = "inline-flex max-w-full items-center rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700"
  } finally {
    button.disabled = false
    button.textContent = originalText
  }
}

async function executeDormantReport(wxId) {
  const button = document.querySelector(`[data-testid="send-dormant-${CSS.escape(wxId)}"]`)
  const status = document.querySelector(`[data-testid="dormant-status-${CSS.escape(wxId)}"]`)
  const originalText = button.textContent
  button.disabled = true
  button.textContent = "发送中…"
  status.textContent = ""
  status.className = "hidden max-w-full text-xs"
  try {
    const response = await fetch("/api/tasks/dormant", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({wxId}),
    })
    const result = await response.json().catch(() => ({message: "服务器返回异常"}))
    if (!response.ok) throw new Error(result.message || "潜水成员统计发送失败")
    status.textContent = `✓ ${result.message}`
    status.className = "inline-flex max-w-full items-center rounded-full bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700"
  } catch (error) {
    console.error(error)
    status.textContent = error.message || "潜水成员统计发送失败"
    status.className = "inline-flex max-w-full items-center rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700"
  } finally {
    button.disabled = false
    button.textContent = originalText
  }
}

async function changeUserGroupRankSkipStatus(groupId, userId) {
  try {
    await putJson("/api/grouprank/skip", {wxId: groupId, userId}, "修改成员排行状态失败")
    const modal = document.getElementById("groupUserModal")
    await loadGroupUsers(groupId, modal.dataset.groupName || "")
  } catch (error) {
    console.error(error)
    showPageActionStatus("修改成员排行状态失败", "error")
  }
}

function addTextCell(row, value) {
  const cell = row.insertCell()
  cell.textContent = value ?? ""
}

function addBooleanCell(row, value) {
  const cell = row.insertCell()
  const tag = document.createElement("span")
  tag.className = `inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${value ? "bg-green-50 text-green-700 ring-green-600/20" : "bg-red-50 text-red-700 ring-red-600/20"}`
  tag.textContent = value ? "是" : "否"
  cell.appendChild(tag)
}

function addRankSkipCell(row, groupId, userId, skipped) {
  const cell = row.insertCell()
  const button = document.createElement("button")
  button.type = "button"
  button.role = "switch"
  button.ariaChecked = String(skipped)
  button.ariaLabel = `排行榜排除 ${userId}`
  button.className = `${skipped ? "bg-gray-400" : "bg-green-600"} relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors`

  const knob = document.createElement("span")
  knob.className = `${skipped ? "translate-x-5" : "translate-x-0"} pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition`
  button.appendChild(knob)
  button.addEventListener("click", () => changeUserGroupRankSkipStatus(groupId, userId))
  cell.appendChild(button)
}

async function loadGroupUsers(groupId, groupName) {
  const tbody = document.getElementById("groupUsers")
  const title = document.getElementById("groupUserModalName")
  const count = document.getElementById("groupUserCount")
  tbody.replaceChildren()
  title.textContent = "加载中…"

  const response = await fetch(`/api/group/users?groupId=${encodeURIComponent(groupId)}`)
  if (!response.ok) {
    throw new Error((await response.text()) || "获取群成员失败")
  }

  const data = await response.json()
  count.textContent = `健在成员: ${data.isOkCount}人`
  for (const member of data.records || []) {
    const row = tbody.insertRow()
    addTextCell(row, member.wxid)
    addTextCell(row, member.nickname)
    addBooleanCell(row, member.isMember)
    addBooleanCell(row, member.isAdmin)
    addTextCell(row, member.joinTime)
    addTextCell(row, member.lastActive)
    addTextCell(row, member.leaveTime)
    addRankSkipCell(row, groupId, member.wxid, member.skipChatRank)
  }
  title.textContent = groupName
}

async function getGroupUsers(groupId, groupName) {
  const modal = document.getElementById("groupUserModal")
  modal.dataset.groupId = groupId
  modal.dataset.groupName = groupName
  modal.showModal()
  try {
    await loadGroupUsers(groupId, groupName)
  } catch (error) {
    console.error(error)
    document.getElementById("groupUserModalName").textContent = groupName
    document.getElementById("groupUserCount").textContent = "成员加载失败"
  }
}
