function showSettingsStatus(message, type = "success") {
  const status = document.querySelector('[data-testid="settings-status"]')
  status.textContent = message
  status.className = `mt-5 rounded-lg px-4 py-3 text-sm ${type === "error" ? "bg-red-50 text-red-700" : type === "loading" ? "bg-blue-50 text-blue-700" : "bg-green-50 text-green-700"}`
}

async function jsonRequest(url, options) {
  const response = await fetch(url, options)
  const result = await response.json().catch(() => ({message: "服务器返回异常"}))
  if (!response.ok) throw new Error(result.message || "操作失败")
  return result
}

async function saveFeatureSetting(feature, scheduleId, timeId, button, thresholdId = "") {
  const schedule = document.getElementById(scheduleId)
  const time = document.getElementById(timeId)
  const threshold = thresholdId ? document.getElementById(thresholdId) : null
  const payload = {feature, schedule: schedule.value, sendTime: time.value}
  if (threshold) {
    const value = Number(threshold.value)
    if (!Number.isInteger(value) || value < 1 || value > 100000) {
      showSettingsStatus("潜水消息阈值必须为 1～100000；配置 1 表示只统计 0 条", "error")
      threshold.focus()
      return
    }
    payload.messageThreshold = value
  }
  const originalText = button.textContent
  button.disabled = true
  button.textContent = "保存中…"
  showSettingsStatus("正在保存推送计划…", "loading")
  try {
    const result = await jsonRequest("/api/settings/time", {
      method: "PUT",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(payload),
    })
    showSettingsStatus(result.message)
  } catch (error) {
    console.error(error)
    showSettingsStatus(error.message || "保存失败", "error")
  } finally {
    button.disabled = false
    button.textContent = originalText
  }
}

function showOwnerStatus(message, type = "success") {
  const status = document.querySelector('[data-testid="owner-status"]')
  status.textContent = message
  status.className = `mt-4 rounded-lg px-4 py-3 text-sm ${type === "error" ? "bg-red-50 text-red-700" : type === "loading" ? "bg-blue-50 text-blue-700" : "bg-green-50 text-green-700"}`
}

async function runOwnerAction(button, loadingText, action) {
  const originalText = button.textContent
  button.disabled = true
  button.textContent = loadingText
  try {
    const result = await action()
    showOwnerStatus(result.message)
    window.setTimeout(() => window.location.reload(), 350)
  } catch (error) {
    showOwnerStatus(error.message || "操作失败", "error")
    button.disabled = false
    button.textContent = originalText
  }
}

async function addOwner(event) {
  event.preventDefault()
  const button = event.currentTarget.querySelector('[data-testid="add-owner"]')
  const input = document.getElementById("owner-account")
  showOwnerStatus("正在从通讯录查询微信号…", "loading")
  await runOwnerAction(button, "查询中…", () => jsonRequest("/api/owners", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({wechatAccount: input.value}),
  }))
}

async function updateOwner(id, button) {
  const input = document.getElementById(`owner-account-${id}`)
  showOwnerStatus("正在更新主人微信号…", "loading")
  await runOwnerAction(button, "保存中…", () => jsonRequest(`/api/owners/${id}`, {
    method: "PUT",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({wechatAccount: input.value}),
  }))
}

async function deleteOwner(id, button) {
  showOwnerStatus("正在删除主人微信号…", "loading")
  await runOwnerAction(button, "删除中…", () => jsonRequest(`/api/owners/${id}`, {method: "DELETE"}))
}
