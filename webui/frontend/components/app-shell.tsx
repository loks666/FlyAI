"use client"

import type { ReactNode } from "react"
import Image from "next/image"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { ContactRoundIcon, HomeIcon, Settings2Icon, UsersRoundIcon } from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar"

const navigation = [
  { href: "/index.html", path: "/", label: "首页", icon: HomeIcon },
  { href: "/group.html", path: "/group", label: "群组", icon: UsersRoundIcon },
  { href: "/friend.html", path: "/friend", label: "通讯录", icon: ContactRoundIcon },
  { href: "/settings.html", path: "/settings", label: "功能配置", icon: Settings2Icon },
]

export function AppShell({ children }: { children: ReactNode }) {
  const pathname = usePathname()

  return (
    <SidebarProvider>
      <Sidebar collapsible="icon">
        <SidebarHeader className="border-b p-4">
          <div className="flex items-center gap-3 overflow-hidden">
            <div className="size-9 shrink-0 overflow-hidden rounded-xl border bg-white shadow-sm">
              <Image
                src="/flyai-logo.jpg"
                alt="FlyAI"
                width={36}
                height={36}
                unoptimized
                className="size-9 object-cover"
              />
            </div>
            <div className="min-w-0 group-data-[collapsible=icon]:hidden">
              <div className="truncate text-lg font-semibold tracking-tight">FlyAI</div>
              <div className="truncate text-xs text-muted-foreground">本地微信机器人</div>
            </div>
          </div>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {navigation.map((item) => {
                  const active = item.path === "/" ? pathname === "/" || pathname === "/index.html" : pathname.startsWith(item.path)
                  return (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton isActive={active} tooltip={item.label} render={<Link href={item.href} />}>
                        <item.icon />
                        <span>{item.label}</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter className="border-t p-3 text-xs text-muted-foreground group-data-[collapsible=icon]:hidden">
          <div className="flex items-center gap-2">
            <span className="size-2 rounded-full bg-primary" />
            本地服务运行中
          </div>
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>
      <SidebarInset className="min-w-0 bg-muted/30">
        <header className="sticky top-0 z-10 flex h-14 items-center border-b bg-background/95 px-4 backdrop-blur sm:px-6">
          <SidebarTrigger />
          <div className="ml-auto flex items-center gap-2 text-xs text-muted-foreground">
            <span className="size-2 rounded-full bg-primary" />
            FlyAI 管理服务在线
          </div>
        </header>
        <main className="min-w-0 flex-1 p-4 sm:p-6 lg:p-8">{children}</main>
      </SidebarInset>
    </SidebarProvider>
  )
}
