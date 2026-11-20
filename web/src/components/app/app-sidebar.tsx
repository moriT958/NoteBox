import { Check, ChevronsUpDown, FileText, MoreHorizontal, Package, Pencil, Plus, Trash2 } from "lucide-react"
import { useMemo, useState } from "react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInput,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSkeleton,
} from "@/components/ui/sidebar"
import type { Box, NoteMeta } from "@/lib/api"
import { cn } from "@/lib/utils"

type Props = {
  boxes: Box[]
  box?: Box
  notes: NoteMeta[] | undefined
  current?: string
  connected: boolean
  onSelectBox: (id: string) => void
  onSelectNote: (path: string) => void
  onNewNote: () => void
  onRenameNote: (n: NoteMeta) => void
  onRemoveNote: (n: NoteMeta) => void
}

export function AppSidebar(p: Props) {
  const [filter, setFilter] = useState("")
  const shown = useMemo(() => {
    const q = filter.trim().toLowerCase()
    const list = p.notes ?? []
    return q ? list.filter((n) => n.title.toLowerCase().includes(q) || n.path.toLowerCase().includes(q)) : list
  }, [p.notes, filter])

  return (
    <Sidebar>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton size="lg">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                    <Package className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{p.box?.title ?? "NoteBox"}</span>
                    <span className="truncate text-xs text-muted-foreground">{p.box?.path}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-(--radix-dropdown-menu-trigger-width) min-w-56" align="start">
                <DropdownMenuLabel className="text-xs text-muted-foreground">Boxes</DropdownMenuLabel>
                {p.boxes.map((b) => (
                  <DropdownMenuItem key={b.id} onSelect={() => p.onSelectBox(b.id)} className="gap-2">
                    <div className="grid flex-1">
                      <span className="truncate">{b.title}</span>
                      <span className="truncate text-xs text-muted-foreground">{b.path}</span>
                    </div>
                    {b.id === p.box?.id && <Check className="size-4" />}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
                  Add boxes with <code>nbx box add</code>
                </DropdownMenuLabel>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
        <SidebarInput placeholder="Filter notes…" value={filter} onChange={(e) => setFilter(e.target.value)} />
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {p.notes === undefined &&
                Array.from({ length: 5 }).map((_, i) => (
                  <SidebarMenuItem key={i}>
                    <SidebarMenuSkeleton showIcon />
                  </SidebarMenuItem>
                ))}
              {p.notes !== undefined && shown.length === 0 && (
                <p className="px-2 py-4 text-sm text-muted-foreground">{filter ? "No matching notes." : "No notes yet."}</p>
              )}
              {shown.map((n) => (
                <SidebarMenuItem key={n.path}>
                  <SidebarMenuButton isActive={n.path === p.current} onClick={() => p.onSelectNote(n.path)} title={n.path}>
                    <FileText />
                    <span>{n.title}</span>
                  </SidebarMenuButton>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <SidebarMenuAction showOnHover>
                        <MoreHorizontal />
                        <span className="sr-only">More</span>
                      </SidebarMenuAction>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent side="right" align="start">
                      <DropdownMenuItem onSelect={() => p.onRenameNote(n)}>
                        <Pencil /> Rename
                      </DropdownMenuItem>
                      <DropdownMenuItem variant="destructive" onSelect={() => p.onRemoveNote(n)}>
                        <Trash2 /> Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={p.onNewNote} disabled={!p.box}>
              <Plus />
              <span>New note</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <div className="flex items-center gap-2 px-2 text-xs text-muted-foreground">
          <span className={cn("size-2 rounded-full", p.connected ? "bg-emerald-500" : "bg-muted-foreground/40")} />
          {p.connected ? "Live" : "Reconnecting…"}
        </div>
      </SidebarFooter>
    </Sidebar>
  )
}
