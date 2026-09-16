"use client"

import { useEffect, useMemo, useState } from "react"
import { MessageCircleIcon, SearchIcon, UserRoundXIcon } from "lucide-react"
import { toast } from "sonner"

import { PageHeader } from "@/components/page-header"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty"
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { Spinner } from "@/components/ui/spinner"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import type { Contact } from "@/lib/types"

export default function FriendPage() {
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [query, setQuery] = useState("")
  const [target, setTarget] = useState<Contact | null>(null)
  const [message, setMessage] = useState("")
  const [sending, setSending] = useState(false)

  useEffect(() => {
    let active = true
    api<{ contacts: Contact[] }>("/api/contacts")
      .then((response) => {
        if (active) setContacts(response.contacts)
      })
      .catch((value) => {
        if (active) setError(value instanceof Error ? value.message : "读取通讯录失败")
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => { active = false }
  }, [])

  const visibleContacts = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    if (!keyword) return contacts
    return contacts.filter((contact) => `${contact.nickname} ${contact.customAccount} ${contact.wxid}`.toLowerCase().includes(keyword))
  }, [contacts, query])

  function openMessage(contact: Contact) {
    setTarget(contact)
    setMessage("")
  }

  async function sendMessage() {
    if (!target || !message.trim()) return
    setSending(true)
    try {
      const response = await api<{ message?: string }>("/api/contacts/message", {
        method: "POST",
        body: JSON.stringify({ wxId: target.wxid, message: message.trim() }),
      })
      toast.success(response.message || `消息已发送给 ${target.nickname}`)
      setTarget(null)
      setMessage("")
    } catch (value) {
      toast.error(value instanceof Error ? value.message : "私聊发送失败")
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-7xl">
      <PageHeader title="通讯录" description={`当前 ${contacts.length} 位可私聊联系人，可搜索并发送文字消息。`} />

      <div className="relative mb-4 max-w-md">
        <SearchIcon className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索昵称、微信号或 wxid" className="pl-9" />
      </div>

      {error ? <Alert variant="destructive" className="mb-4"><AlertTitle>通讯录读取失败</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}

      <Card className="overflow-hidden py-0">
        <CardContent className="p-0">
          {loading ? (
            <div className="flex flex-col gap-3 p-5">{Array.from({ length: 6 }).map((_, index) => <Skeleton key={index} className="h-14 w-full" />)}</div>
          ) : visibleContacts.length ? (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader><TableRow><TableHead>昵称</TableHead><TableHead>微信号</TableHead><TableHead>wxid</TableHead><TableHead>最近活跃</TableHead><TableHead>状态</TableHead><TableHead className="text-right">操作</TableHead></TableRow></TableHeader>
                <TableBody>
                  {visibleContacts.map((contact) => (
                    <TableRow key={contact.wxid}>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <Avatar><AvatarFallback>{(contact.nickname || "微").slice(0, 1)}</AvatarFallback></Avatar>
                          <span className="font-medium">{contact.nickname || "未命名联系人"}</span>
                        </div>
                      </TableCell>
                      <TableCell>{contact.customAccount || "—"}</TableCell>
                      <TableCell className="max-w-52 truncate text-xs text-muted-foreground">{contact.wxid}</TableCell>
                      <TableCell className="whitespace-nowrap text-xs text-muted-foreground">{contact.lastActive || "暂无活跃记录"}</TableCell>
                      <TableCell><Badge variant="secondary"><span className="size-1.5 rounded-full bg-primary" />可发送</Badge></TableCell>
                      <TableCell className="text-right"><Button size="sm" variant="outline" onClick={() => openMessage(contact)}><MessageCircleIcon data-icon="inline-start" />私聊发送</Button></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <Empty className="py-20"><EmptyHeader><EmptyMedia variant="icon"><UserRoundXIcon /></EmptyMedia><EmptyTitle>没有匹配的联系人</EmptyTitle><EmptyDescription>调整搜索词，或先同步微信数据。</EmptyDescription></EmptyHeader></Empty>
          )}
        </CardContent>
      </Card>

      <Dialog open={Boolean(target)} onOpenChange={(open) => { if (!open && !sending) setTarget(null) }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>私聊发送</DialogTitle>
            <DialogDescription>发送给 {target?.nickname || target?.wxid}</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="direct-message">消息内容</FieldLabel>
              <Textarea id="direct-message" value={message} onChange={(event) => setMessage(event.target.value.slice(0, 2000))} rows={6} placeholder="输入要发送的文字…" disabled={sending} />
              <FieldDescription className="text-right">{message.length} / 2000</FieldDescription>
            </Field>
          </FieldGroup>
          <Alert>
            <AlertTitle>发送前请确认</AlertTitle>
            <AlertDescription>确认后会通过当前已登录的微信账号真实发送。</AlertDescription>
          </Alert>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTarget(null)} disabled={sending}>取消</Button>
            <Button onClick={sendMessage} disabled={sending || !message.trim()}>{sending ? <Spinner data-icon="inline-start" /> : <MessageCircleIcon data-icon="inline-start" />}确认发送</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
