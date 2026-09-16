import type { FeatureKey } from "@/lib/types"

export const features: Array<{
  key: FeatureKey
  shortLabel: string
  label: string
  description: string
}> = [
  { key: "weekly", shortLabel: "排行", label: "排行榜", description: "统计滚动最近 7 天活跃排行" },
  { key: "news", shortLabel: "热点", label: "知乎热榜", description: "推送知乎热榜前 10 条" },
  { key: "dormant", shortLabel: "潜水", label: "潜水统计", description: "统计低于发言阈值的成员" },
]

export const featureEnabledField: Record<FeatureKey, keyof import("@/lib/types").Contact> = {
  weekly: "enableWeeklyRank",
  yearly: "enableYearlyRank",
  news: "enableNews",
  dormant: "enableDormant",
}
