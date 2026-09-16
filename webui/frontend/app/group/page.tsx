"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { RefreshCwIcon, SearchIcon, SendIcon, UsersIcon } from "lucide-react"
import { toast } from "sonner"

import { PageHeader } from "@/components/page-header"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Spinner } from "@/components/ui/spinner"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { api } from "@/lib/api"
import { featureEnabledField, features } from "@/lib/features"
import type { Contact, FeatureKey, GroupMember } from "@/lib/types"

type GroupResponse = { groups: Contact[] }
type GroupFilter = "all" | "monitored" | "stopped"

function featureEnabled(group: Contact, feature: FeatureKey) {
  return Boolean(group[featureEnabledField[feature]])
}

export default function GroupPage() {
  const [groups, setGroups] = useState<Contact[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [query, setQuery] = useState("")
  const [filter, setFilter] = useState<GroupFilter>("all")
  const [busy, setBusy] = useState("")
  const [membersOpen, setMembersOpen] = useState(false)
  const [membersGroup, setMembersGroup] = useState<Contact | null>(null)
  const [members, setMembers] = useState<GroupMember[]>([])
  const [membersLoading, setMembersLoading] = useState(false)

  const loadGroups = useCallback(async () => {
    setLoading(true)
    try {
      setError("")
      const response = await api<GroupResponse>("/api/groups")
      setGroups(response.groups)
    } catch (value) {
      setError(value instanceof Error ? value.message : "读取群组失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadGroups()
  }, [loadGroups])

  const visibleGroups = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    return groups.filter((group) => {
      const matchesQuery = !keyword || `${group.nickname} ${group.wxid}`.toLowerCase().includes(keyword)
      const matchesFilter = filter === "all" || (filter === "monitored" ? group.enableMonitor : !group.enableMonitor)
      return matchesQuery && matchesFilter
    })
  }, [filter, groups, query])

  async function toggleMonitor(group: Contact) {
    const key = `monitor:${group.wxid}`
    setBusy(key)
    try {
      const response = await api<{ enabled: boolean }>("/api/monitor/status", {
        method: "PUT",
        body: JSON.stringify({ wxId: group.wxid }),
      })
      setGroups((current) => current.map((item) => item.wxid === group.wxid ? { ...item, enableMonitor: response.enabled } : item))
      toast.success(response.enabled ? "消息监控已开启" : "消息监控已关闭")
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "修改消息监控失败")
    } finally {
      setBusy("")
    }
  }

  async function toggleFeature(group: Contact, feature: FeatureKey) {
    const key = `toggle:${feature}:${group.wxid}`
    setBusy(key)
    try {
      const response = await api<{ enabled: boolean }>("/api/group-feature/status", {
        method: "PUT",
        body: JSON.stringify({ wxId: group.wxid, feature }),
      })
      const field = featureEnabledField[feature]
      setGroups((current) => current.map((item) => item.wxid === group.wxid ? { ...item, [field]: response.enabled } : item))
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "修改推送状态失败")
    } finally {
      setBusy("")
    }
  }

  async function sendFeature(group: Contact, feature: FeatureKey) {
    const key = `send:${feature}:${group.wxid}`
    setBusy(key)
    try {
      const response = await api<{ message: string }>("/api/tasks/group-feature", {
        method: "POST",
        body: JSON.stringify({ wxId: group.wxid, feature }),
      })
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "发送失败")
    } finally {
      setBusy("")
    }
  }

  async function syncWechat() {
    setBusy("sync")
    try {
      const response = await api<{ message: string }>("/api/sync/wechat", { method: "POST" })
      toast.success(response.message)
      await loadGroups()
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "同步微信数据失败")
    } finally {
      setBusy("")
    }
  }

  async function sendAllWeekly() {
    setBusy("send-all")
    try {
      const response = await api<{ message: string }>("/api/tasks/weekly-rank/all", { method: "POST" })
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "批量发送失败")
    } finally {
      setBusy("")
    }
  }

  async function openMembers(group: Contact) {
    setMembersGroup(group)
    setMembersOpen(true)
    setMembersLoading(true)
    try {
      const response = await api<{ records: GroupMember[]; warning?: string }>(`/api/group/users?groupId=${encodeURIComponent(group.wxid)}`)
      setMembers(response.records)
      // 桥接器未启动时仍会返回数据库里的历史成员，提示这批数据可能不是最新的。
      if (response.warning) {
        toast.warning(response.warning)
      }
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "读取群成员失败")
      setMembers([])
    } finally {
      setMembersLoading(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1480px]">
      <PageHeader
        title="群组管理"
        description="为每个群单独控制消息监控和自动推送，也可立即发送指定推送。"
        actions={
          <>
            <Button variant="outline" onClick={syncWechat} disabled={busy === "sync"}>
              {busy === "sync" ? <Spinner data-icon="inline-start" /> : <RefreshCwIcon data-icon="inline-start" />}
              同步微信数据
            </Button>
            <Button onClick={sendAllWeekly} disabled={busy === "send-all"}>
              {busy === "send-all" ? <Spinner data-icon="inline-start" /> : <SendIcon data-icon="inline-start" />}
              发送全部排行榜
            </Button>
          </>
        }
      />

      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <div className="relative min-w-0 flex-1 sm:max-w-sm">
          <SearchIcon className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索群名称或 wxid" className="pl-9" />
        </div>
        <Select value={filter} onValueChange={(value) => setFilter(value as GroupFilter)}>
          <SelectTrigger className="w-full sm:w-40"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="monitored">监控中</SelectItem>
              <SelectItem value="stopped">已停止</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      {error ? <Alert variant="destructive" className="mb-4"><AlertTitle>群组读取失败</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}

      <Card className="overflow-hidden py-0">
        <CardContent className="p-0">
          <div className="hidden grid-cols-[minmax(190px,.9fr)_100px_minmax(390px,1.7fr)_90px] gap-3 border-b bg-muted/40 px-5 py-3 text-xs font-medium text-muted-foreground xl:grid">
            <span>群组名称</span><span>消息监控</span><span>推送功能</span><span className="text-right">成员</span>
          </div>
          {loading ? (
            <div className="flex flex-col gap-3 p-5">{Array.from({ length: 5 }).map((_, index) => <Skeleton key={index} className="h-28 w-full" />)}</div>
          ) : visibleGroups.length ? (
            <div className="divide-y">
              {visibleGroups.map((group) => (
                <article key={group.wxid} className="grid min-w-0 gap-4 px-5 py-5 xl:grid-cols-[minmax(190px,.9fr)_100px_minmax(390px,1.7fr)_90px] xl:items-center xl:gap-3">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <div className="truncate font-medium">{group.nickname || "未命名群聊"}</div>
                      <span className="size-2 shrink-0 rounded-full bg-primary" />
                    </div>
                    <div className="mt-1 truncate text-xs text-muted-foreground">{group.wxid}</div>
                    <div className="mt-2 text-xs text-muted-foreground">{group.lastActive ? `最近活跃 ${group.lastActive}` : "暂无活跃记录"}</div>
                  </div>

                  <div>
                    <Button size="sm" variant={group.enableMonitor ? "secondary" : "outline"} onClick={() => toggleMonitor(group)} disabled={busy === `monitor:${group.wxid}`}>
                      <span className={`size-2 rounded-full ${group.enableMonitor ? "bg-primary" : "bg-muted-foreground/40"}`} />
                      {group.enableMonitor ? "监控中" : "已停止"}
                    </Button>
                  </div>

                  <div className="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-2">
                    {features.map((feature) => {
                      const enabled = featureEnabled(group, feature.key)
                      const toggleKey = `toggle:${feature.key}:${group.wxid}`
                      const sendKey = `send:${feature.key}:${group.wxid}`
                      return (
                        <div key={feature.key} className="flex shrink-0 items-center gap-1.5">
                          <Button size="sm" variant={enabled ? "secondary" : "outline"} className="w-[72px] justify-between px-2" onClick={() => toggleFeature(group, feature.key)} disabled={busy === toggleKey} aria-pressed={enabled}>
                            <span>{feature.shortLabel}</span>
                            <span className={`size-2 shrink-0 rounded-full ${enabled ? "bg-primary" : "bg-muted-foreground/35"}`} />
                          </Button>
                          <Button size="xs" className="h-7 rounded-lg px-2.5" onClick={() => sendFeature(group, feature.key)} disabled={busy === sendKey} aria-label={`向${group.nickname}发送${feature.label}`}>
                            {busy === sendKey ? <Spinner /> : "发送"}
                          </Button>
                        </div>
                      )
                    })}
                  </div>

                  <div className="xl:text-right">
                    <Button variant="outline" size="sm" onClick={() => openMembers(group)}>
                      <UsersIcon data-icon="inline-start" />查看成员
                    </Button>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <Empty className="py-16"><EmptyHeader><EmptyMedia variant="icon"><UsersIcon /></EmptyMedia><EmptyTitle>没有匹配的群组</EmptyTitle><EmptyDescription>调整搜索词或状态筛选后重试。</EmptyDescription></EmptyHeader></Empty>
          )}
        </CardContent>
      </Card>

      <div className="mt-3 text-xs text-muted-foreground">共 {visibleGroups.length} 个群组</div>

      <Dialog open={membersOpen} onOpenChange={setMembersOpen}>
        <DialogContent className="max-h-[80vh] max-w-4xl overflow-hidden sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>{membersGroup?.nickname || "群成员"}</DialogTitle>
            <DialogDescription>显示当前成员及其微信号、活跃状态和排行榜设置。</DialogDescription>
          </DialogHeader>
          <div className="overflow-auto rounded-lg border">
            {membersLoading ? <div className="flex h-48 items-center justify-center"><Spinner /></div> : (
              <Table>
                <TableHeader><TableRow><TableHead>昵称</TableHead><TableHead>微信号</TableHead><TableHead>wxid</TableHead><TableHead>身份</TableHead><TableHead>最后活跃</TableHead></TableRow></TableHeader>
                <TableBody>{members.map((member) => <TableRow key={member.wxid}><TableCell>{member.nickname || "未命名"}</TableCell><TableCell>{member.account || "—"}</TableCell><TableCell className="max-w-48 truncate text-xs text-muted-foreground">{member.wxid}</TableCell><TableCell><Badge variant={member.isAdmin ? "default" : "secondary"}>{member.isAdmin ? "群主" : member.isMember ? "成员" : "已退出"}</Badge></TableCell><TableCell className="whitespace-nowrap text-xs text-muted-foreground">{member.lastActive || "—"}</TableCell></TableRow>)}</TableBody>
              </Table>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
