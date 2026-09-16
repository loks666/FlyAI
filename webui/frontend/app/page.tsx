"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { ActivityIcon, ContactRoundIcon, RefreshCwIcon, SendIcon, UsersRoundIcon } from "lucide-react"
import { toast } from "sonner"

import { PageHeader } from "@/components/page-header"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button, buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { cn } from "@/lib/utils"
import type { DashboardSummary } from "@/lib/types"

const summaryItems = [
  { key: "monitorCount" as const, label: "消息监控", description: "正在记录新消息的群", icon: ActivityIcon },
  { key: "groupCount" as const, label: "群组数量", description: "当前通讯录中的群聊", icon: UsersRoundIcon },
  { key: "rankCount" as const, label: "排行榜推送", description: "已开启最近 7 天排行的群", icon: SendIcon },
  { key: "contactCount" as const, label: "通讯录", description: "当前可私聊联系人", icon: ContactRoundIcon },
]

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [error, setError] = useState("")
  const [syncing, setSyncing] = useState(false)

  const load = useCallback(async () => {
    try {
      setError("")
      setSummary(await api<DashboardSummary>("/api/dashboard"))
    } catch (value) {
      setError(value instanceof Error ? value.message : "读取运行概览失败")
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  async function syncWechat() {
    setSyncing(true)
    try {
      const result = await api<{ message?: string }>("/api/sync/wechat", { method: "POST" })
      toast.success(result.message || "微信数据同步完成")
      await load()
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "同步微信数据失败")
    } finally {
      setSyncing(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-7xl">
      <PageHeader
        title="运行概览"
        description="查看当前机器人数据状态，并快速进入常用管理功能。"
        actions={
          <Button onClick={syncWechat} disabled={syncing}>
            <RefreshCwIcon data-icon="inline-start" className={cn(syncing && "animate-spin")} />
            {syncing ? "同步中" : "同步微信数据"}
          </Button>
        }
      />

      {error ? (
        <Alert variant="destructive" className="mb-6">
          <AlertTitle>运行概览读取失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4" aria-label="运行统计">
        {summaryItems.map((item) => (
          <Card key={item.key}>
            <CardHeader className="flex-row items-start justify-between gap-4">
              <div>
                <CardDescription>{item.label}</CardDescription>
                <CardTitle className="mt-2 text-3xl tabular-nums">
                  {summary ? summary[item.key] : <Skeleton className="h-9 w-16" />}
                </CardTitle>
              </div>
              <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <item.icon />
              </div>
            </CardHeader>
            <CardContent className="text-xs text-muted-foreground">{item.description}</CardContent>
          </Card>
        ))}
      </section>

      <section className="mt-6 grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>群组推送</CardTitle>
            <CardDescription>逐个群控制排行榜、热点和潜水统计，并可立即发送指定内容。</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/group.html" className={buttonVariants()}>管理群组</Link>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>全局计划</CardTitle>
            <CardDescription>设置统一推送时间、潜水阈值和主人微信号。</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/settings.html" className={buttonVariants({ variant: "outline" })}>打开功能配置</Link>
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
