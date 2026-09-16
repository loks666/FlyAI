"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { BotIcon, CableIcon, Clock3Icon, DatabaseIcon, MonitorIcon, PencilIcon, PlayIcon, PlusIcon, RefreshCwIcon, SaveIcon, ShieldCheckIcon, SquareIcon, Trash2Icon } from "lucide-react"
import { toast } from "sonner"

import { PageHeader } from "@/components/page-header"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { Spinner } from "@/components/ui/spinner"
import { api } from "@/lib/api"
import { features } from "@/lib/features"
import type { BridgeStatus, FeatureKey, FeatureSettings, Owner } from "@/lib/types"

const scheduleOptions = [
  ["daily", "每天"], ["monday", "每周一"], ["tuesday", "每周二"],
  ["wednesday", "每周三"], ["thursday", "每周四"], ["friday", "每周五"],
  ["saturday", "每周六"], ["sunday", "每周日"],
] as const

const commandGroups = [
  { title: "管理员命令（兼容 / 前缀）", values: ["排行榜", "周排", "年排", "潜水", "热点"] },
  { title: "群聊指令", values: ["给我骂他@昵称", "给我骂他.@昵称"] },
]

const bridgeStatusLabels: Record<BridgeStatus["status"], string> = {
  unconfigured: "未配置",
  stopped: "已停止",
  starting: "启动中",
  running: "运行正常",
  degraded: "运行异常",
  error: "启动失败",
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<FeatureSettings | null>(null)
  const [owners, setOwners] = useState<Owner[]>([])
  const [bridgeStatus, setBridgeStatus] = useState<BridgeStatus | null>(null)
  const [bridgeAccount, setBridgeAccount] = useState("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [busy, setBusy] = useState("")
  const [ownerDialog, setOwnerDialog] = useState<Owner | "new" | null>(null)
  const [ownerAccount, setOwnerAccount] = useState("")
  const [deleteOwner, setDeleteOwner] = useState<Owner | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setError("")
      const [nextSettings, ownerResponse, nextBridgeStatus] = await Promise.all([
        api<FeatureSettings>("/api/settings"),
        api<{ owners: Owner[] }>("/api/owners"),
        api<BridgeStatus>("/api/bridge/status"),
      ])
      setSettings(nextSettings)
      setOwners(ownerResponse.owners)
      setBridgeStatus(nextBridgeStatus)
      setBridgeAccount(nextBridgeStatus.wechatAccount || "")
    } catch (value) {
      setError(value instanceof Error ? value.message : "读取配置失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const ownerDialogTitle = useMemo(() => ownerDialog === "new" ? "添加主人" : "修改主人微信号", [ownerDialog])

  async function refreshBridgeStatus() {
    setBusy("bridge-refresh")
    try {
      const status = await api<BridgeStatus>("/api/bridge/status")
      setBridgeStatus(status)
      setBridgeAccount((current) => current || status.wechatAccount || "")
      toast.success("桥接器状态已刷新")
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "刷新桥接器状态失败")
    } finally {
      setBusy("")
    }
  }

  async function saveBridgeAccount() {
    if (!bridgeAccount.trim()) return
    setBusy("bridge-save")
    try {
      const response = await api<{ message: string }>("/api/bridge/settings", {
        method: "PUT",
        body: JSON.stringify({ wechatAccount: bridgeAccount.trim() }),
      })
      const status = await api<BridgeStatus>("/api/bridge/status")
      setBridgeStatus(status)
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "保存机器人微信号失败")
    } finally {
      setBusy("")
    }
  }

  async function changeBridgeState(action: "start" | "stop") {
    setBusy(`bridge-${action}`)
    try {
      const response = await api<{ message: string; status: BridgeStatus }>(`/api/bridge/${action}`, { method: "POST" })
      setBridgeStatus(response.status)
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : `${action === "start" ? "启动" : "停止"}桥接器失败`)
      try {
        setBridgeStatus(await api<BridgeStatus>("/api/bridge/status"))
      } catch {}
    } finally {
      setBusy("")
    }
  }

  function updateSetting(feature: FeatureKey, field: "schedule" | "sendTime" | "messageThreshold", value: string | number) {
    setSettings((current) => current ? { ...current, [feature]: { ...current[feature], [field]: value } } : current)
  }

  async function saveSetting(feature: FeatureKey) {
    if (!settings) return
    const setting = settings[feature]
    setBusy(`setting:${feature}`)
    try {
      const response = await api<{ message: string; setting: FeatureSettings[FeatureKey] }>("/api/settings/time", {
        method: "PUT",
        body: JSON.stringify({
          feature,
          schedule: setting.schedule,
          sendTime: setting.sendTime,
          ...(feature === "dormant" ? { messageThreshold: setting.messageThreshold } : {}),
        }),
      })
      updateSetting(feature, "sendTime", response.setting.sendTime)
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "保存推送计划失败")
    } finally {
      setBusy("")
    }
  }

  function openOwner(owner: Owner | "new") {
    setOwnerDialog(owner)
    setOwnerAccount(owner === "new" ? "" : owner.wechatAccount)
  }

  async function saveOwner() {
    if (!ownerDialog || !ownerAccount.trim()) return
    setBusy("owner")
    try {
      const existing = ownerDialog === "new" ? null : ownerDialog
      const response = await api<{ message: string; owner: Owner }>(existing ? `/api/owners/${existing.id}` : "/api/owners", {
        method: existing ? "PUT" : "POST",
        body: JSON.stringify({ wechatAccount: ownerAccount.trim() }),
      })
      setOwners((current) => existing ? current.map((item) => item.id === existing.id ? response.owner : item) : [...current, response.owner])
      setOwnerDialog(null)
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "保存主人失败")
    } finally {
      setBusy("")
    }
  }

  async function confirmDeleteOwner() {
    if (!deleteOwner) return
    setBusy("delete-owner")
    try {
      const response = await api<{ message: string }>(`/api/owners/${deleteOwner.id}`, { method: "DELETE" })
      setOwners((current) => current.filter((item) => item.id !== deleteOwner.id))
      setDeleteOwner(null)
      toast.success(response.message)
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "删除主人失败")
    } finally {
      setBusy("")
    }
  }

  return (
    <div className="mx-auto w-full max-w-7xl">
      <PageHeader title="功能配置" description="这里只配置全局推送时间与主人身份；每个群的开关请到群组管理设置。" />
      {error ? <Alert variant="destructive" className="mb-5"><AlertTitle>配置读取失败</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}

      <div className="grid gap-5 xl:grid-cols-[minmax(0,1.7fr)_minmax(320px,1fr)]">
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><Clock3Icon />推送计划</CardTitle><CardDescription>保存后立即生效，群聊开关不会在这里重复出现。</CardDescription></CardHeader>
          <CardContent>
            {loading || !settings ? <div className="space-y-3">{Array.from({ length: 3 }).map((_, index) => <Skeleton key={index} className="h-20 w-full" />)}</div> : (
              <div className="divide-y rounded-xl border">
                {features.map((feature) => {
                  const setting = settings[feature.key]
                  return (
                    <div key={feature.key} className="grid gap-4 p-4 md:grid-cols-[minmax(150px,1fr)_155px_120px_auto] md:items-end">
                      <div><div className="font-medium">{feature.label}</div><div className="mt-1 text-xs text-muted-foreground">{feature.description}</div></div>
                      <Field><FieldLabel htmlFor={`schedule-${feature.key}`}>周期</FieldLabel><Select value={setting.schedule} onValueChange={(value) => updateSetting(feature.key, "schedule", String(value))}><SelectTrigger id={`schedule-${feature.key}`} className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectGroup>{scheduleOptions.map(([value, label]) => <SelectItem key={value} value={value}>{label}</SelectItem>)}</SelectGroup></SelectContent></Select></Field>
                      <Field><FieldLabel htmlFor={`time-${feature.key}`}>发送时间</FieldLabel><Input id={`time-${feature.key}`} type="time" value={setting.sendTime} onChange={(event) => updateSetting(feature.key, "sendTime", event.target.value)} /></Field>
                      <Button size="sm" onClick={() => saveSetting(feature.key)} disabled={busy === `setting:${feature.key}`}>
                        {busy === `setting:${feature.key}` ? <Spinner data-icon="inline-start" /> : <SaveIcon data-icon="inline-start" />}保存
                      </Button>
                      {feature.key === "dormant" ? <Field className="md:col-start-2 md:col-span-2"><FieldLabel htmlFor="dormant-threshold">潜水消息阈值</FieldLabel><Input id="dormant-threshold" type="number" min={1} max={100000} value={setting.messageThreshold} onChange={(event) => updateSetting(feature.key, "messageThreshold", Number(event.target.value))} /><FieldDescription>低于该消息数的成员会被计入潜水统计。</FieldDescription></Field> : null}
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>

        <div className="space-y-5">
          <Card>
            <CardHeader className="flex-row items-start justify-between gap-3">
              <div>
                <CardTitle className="flex items-center gap-2"><CableIcon />微信桥接器</CardTitle>
                <CardDescription className="mt-1.5">多开微信时，只控制与机器人账号匹配的窗口。</CardDescription>
              </div>
              <Button
                size="icon-sm"
                variant="outline"
                aria-label="刷新桥接器状态"
                title="刷新桥接器状态"
                onClick={refreshBridgeStatus}
                disabled={Boolean(busy)}
              >
                <RefreshCwIcon className={busy === "bridge-refresh" ? "animate-spin" : ""} />
              </Button>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-xl border bg-muted/30 px-3 py-2.5">
                <div className="flex items-center gap-2 text-sm font-medium">
                  <span className={`size-2.5 rounded-full ${bridgeStatus?.status === "running" ? "bg-emerald-500" : bridgeStatus?.status === "degraded" || bridgeStatus?.status === "error" ? "bg-amber-500" : "bg-muted-foreground/40"}`} />
                  {bridgeStatus ? bridgeStatusLabels[bridgeStatus.status] : "读取中"}
                </div>
                <span className="text-xs text-muted-foreground">{bridgeStatus?.checkedAt ? `更新于 ${bridgeStatus.checkedAt}` : "尚未同步"}</span>
              </div>

              <Field>
                <FieldLabel htmlFor="bridge-account">机器人微信号</FieldLabel>
                <div className="flex gap-2">
                  <Input
                    id="bridge-account"
                    value={bridgeAccount}
                    onChange={(event) => setBridgeAccount(event.target.value)}
                    placeholder="例如 niubaidely"
                    disabled={Boolean(busy)}
                  />
                  <Button variant="outline" onClick={saveBridgeAccount} disabled={Boolean(busy) || !bridgeAccount.trim()}>
                    {busy === "bridge-save" ? <Spinner data-icon="inline-start" /> : <SaveIcon data-icon="inline-start" />}保存
                  </Button>
                </div>
                <FieldDescription>填写微信资料中的唯一微信号；保存后启动桥接器会自动解析 wxid。</FieldDescription>
              </Field>

              <div className="grid grid-cols-2 gap-2 text-xs">
                <div className="rounded-lg border p-2.5">
                  <div className="flex items-center gap-1.5 text-muted-foreground"><DatabaseIcon className="size-3.5" />账号数据库</div>
                  <div className="mt-1 truncate font-medium" title={bridgeStatus?.resolvedWxid || ""}>{bridgeStatus?.databaseVerified ? bridgeStatus.resolvedWxid : "未验证"}</div>
                </div>
                <div className="rounded-lg border p-2.5">
                  <div className="flex items-center gap-1.5 text-muted-foreground"><MonitorIcon className="size-3.5" />目标窗口</div>
                  <div className="mt-1 truncate font-medium">{bridgeStatus?.windowVerified ? `PID ${bridgeStatus.windowPid}` : "未绑定"}</div>
                </div>
              </div>

              {bridgeStatus?.lastError ? <Alert variant="destructive"><AlertTitle>桥接器不可用</AlertTitle><AlertDescription>{bridgeStatus.lastError}</AlertDescription></Alert> : null}

              <div className="flex gap-2">
                <Button
                  className="flex-1"
                  onClick={() => changeBridgeState("start")}
                  disabled={Boolean(busy) || !bridgeAccount.trim() || bridgeStatus?.status === "running"}
                >
                  {busy === "bridge-start" ? <Spinner data-icon="inline-start" /> : <PlayIcon data-icon="inline-start" />}启动桥接器
                </Button>
                <Button
                  className="flex-1"
                  variant="outline"
                  onClick={() => changeBridgeState("stop")}
                  disabled={Boolean(busy) || !bridgeStatus || ["unconfigured", "stopped"].includes(bridgeStatus.status)}
                >
                  {busy === "bridge-stop" ? <Spinner data-icon="inline-start" /> : <SquareIcon data-icon="inline-start" />}停止
                </Button>
              </div>
              <p className="text-xs leading-relaxed text-muted-foreground">状态不会定时轮询。页面首次打开、启动/停止后会同步一次，其他时间请点击右上角刷新。</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex-row items-start justify-between gap-3"><div><CardTitle className="flex items-center gap-2"><ShieldCheckIcon />主人账号</CardTitle><CardDescription className="mt-1.5">仅通过唯一微信号绑定管理员权限。</CardDescription></div><Button size="sm" onClick={() => openOwner("new")}><PlusIcon data-icon="inline-start" />添加</Button></CardHeader>
            <CardContent className="space-y-3">
              {loading ? <Skeleton className="h-24 w-full" /> : owners.length ? owners.map((owner) => (
                <div key={owner.id} className="rounded-xl border p-3"><div className="flex items-start justify-between gap-3"><div className="min-w-0"><div className="truncate font-medium">{owner.nickname || "未命名主人"}</div><div className="mt-1 truncate text-xs text-muted-foreground">微信号：{owner.wechatAccount}</div><div className="mt-0.5 truncate text-xs text-muted-foreground">{owner.wxid}</div></div><Badge>主人</Badge></div><div className="mt-3 flex gap-2"><Button size="xs" variant="outline" onClick={() => openOwner(owner)}><PencilIcon />编辑</Button><Button size="xs" variant="outline" onClick={() => setDeleteOwner(owner)}><Trash2Icon />删除</Button></div></div>
              )) : <div className="rounded-xl border border-dashed p-5 text-center text-sm text-muted-foreground">尚未绑定主人微信号</div>}
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2"><BotIcon />机器人指令</CardTitle><CardDescription>主人在群聊中可直接调用这些指令。</CardDescription></CardHeader>
            <CardContent className="space-y-4">{commandGroups.map((group, index) => <div key={group.title}>{index ? <Separator className="mb-4" /> : null}<div className="mb-2 text-xs font-medium text-muted-foreground">{group.title}</div><div className="flex flex-wrap gap-2">{group.values.map((command) => <Badge key={command} variant="secondary">{command}</Badge>)}</div></div>)}</CardContent>
          </Card>
        </div>
      </div>

      <Dialog open={Boolean(ownerDialog)} onOpenChange={(open) => { if (!open && busy !== "owner") setOwnerDialog(null) }}>
        <DialogContent className="sm:max-w-md"><DialogHeader><DialogTitle>{ownerDialogTitle}</DialogTitle><DialogDescription>填写通讯录中的唯一微信号，系统会自动校验并绑定 wxid。</DialogDescription></DialogHeader><FieldGroup><Field><FieldLabel htmlFor="owner-account">微信号</FieldLabel><Input id="owner-account" value={ownerAccount} onChange={(event) => setOwnerAccount(event.target.value)} placeholder="例如 wxid 对应的微信号" disabled={busy === "owner"} /></Field></FieldGroup><DialogFooter><Button variant="outline" onClick={() => setOwnerDialog(null)} disabled={busy === "owner"}>取消</Button><Button onClick={saveOwner} disabled={busy === "owner" || !ownerAccount.trim()}>{busy === "owner" ? <Spinner data-icon="inline-start" /> : <SaveIcon data-icon="inline-start" />}保存</Button></DialogFooter></DialogContent>
      </Dialog>

      <Dialog open={Boolean(deleteOwner)} onOpenChange={(open) => { if (!open && busy !== "delete-owner") setDeleteOwner(null) }}>
        <DialogContent className="sm:max-w-md"><DialogHeader><DialogTitle>删除主人账号？</DialogTitle><DialogDescription>删除后，{deleteOwner?.nickname || deleteOwner?.wechatAccount} 将无法再使用管理员指令。此操作需要明确确认。</DialogDescription></DialogHeader><DialogFooter><Button variant="outline" onClick={() => setDeleteOwner(null)} disabled={busy === "delete-owner"}>取消</Button><Button variant="destructive" onClick={confirmDeleteOwner} disabled={busy === "delete-owner"}>{busy === "delete-owner" ? <Spinner data-icon="inline-start" /> : <Trash2Icon data-icon="inline-start" />}确认删除</Button></DialogFooter></DialogContent>
      </Dialog>
    </div>
  )
}
