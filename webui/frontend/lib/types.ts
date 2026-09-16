export type DashboardSummary = {
  contactCount: number
  groupCount: number
  monitorCount: number
  rankCount: number
}

export type Contact = {
  wxid: string
  customAccount: string
  nickname: string
  lastActive: string
  enableMonitor: boolean
  enableWeeklyRank: boolean
  enableYearlyRank: boolean
  enableNews: boolean
  enableDormant: boolean
}

export type FeatureKey = "weekly" | "yearly" | "news" | "dormant"

export type FeatureSetting = {
  feature: FeatureKey
  sendTime: string
  schedule: string
  messageThreshold: number
  updatedAt: string
}

export type FeatureSettings = Record<FeatureKey, FeatureSetting>

export type Owner = {
  id: number
  wechatAccount: string
  wxid: string
  nickname: string
  createdAt: string
  updatedAt: string
}

export type BridgeStatus = {
  status: "unconfigured" | "stopped" | "starting" | "running" | "degraded" | "error"
  desiredState: "stopped" | "running"
  wechatAccount: string
  resolvedWxid: string
  accountDirectory: string
  processRunning: boolean
  serviceHealthy: boolean
  databaseVerified: boolean
  windowVerified: boolean
  windowPid: number
  windowHandle: number
  sendReady: boolean
  lastError: string
  checkedAt: string
}

export type GroupMember = {
  wxid: string
  account: string
  nickname: string
  isMember: boolean
  isAdmin: boolean
  lastActive: string
  skipChatRank: boolean
}
