import { AlertTriangle, Check, CircleAlert, FileQuestion, Loader2, NotebookPen } from "lucide-react"
import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "sonner"
import { AppSidebar } from "@/components/app/app-sidebar"
import { TitleDialog } from "@/components/app/title-dialog"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { Toaster } from "@/components/ui/sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import { Editor } from "@/editor/Editor"
import { type SaveStatus, useNote } from "@/hooks/use-note"
import { api, type Box, type NoteMeta, subscribe } from "@/lib/api"
import { navigate, useRoute } from "@/lib/route"

const LAST_BOX = "nbx.lastBox"

function readLastBox() {
  try {
    return localStorage.getItem(LAST_BOX) ?? undefined
  } catch {
    return undefined
  }
}

export function App() {
  const route = useRoute()
  const [boxes, setBoxes] = useState<Box[]>([])
  const [notes, setNotes] = useState<NoteMeta[]>()
  const [connected, setConnected] = useState(false)
  const [dialog, setDialog] = useState<
    { kind: "new" } | { kind: "rename"; note: NoteMeta } | { kind: "remove"; note: NoteMeta } | null
  >(null)

  const justCreated = useRef<string | undefined>(undefined)
  const box = boxes.find((b) => b.id === route.box)
  const note = useNote(box?.id, box ? route.note : undefined)
  const current = notes?.find((n) => n.path === route.note)

  const loadBoxes = useCallback(async () => {
    try {
      setBoxes(await api.boxes())
    } catch (e) {
      toast.error(`Failed to load boxes: ${(e as Error).message}`)
    }
  }, [])

  const loadNotes = useCallback(async (id: string) => {
    try {
      const list = await api.notes(id)
      setNotes((prev) => (routeBox.current === id ? list : prev))
    } catch (e) {
      toast.error(`Failed to load notes: ${(e as Error).message}`)
    }
  }, [])

  const routeBox = useRef(route.box)
  routeBox.current = route.box

  useEffect(() => {
    void loadBoxes()
  }, [loadBoxes])

  // pick a box when the URL has none (or an unknown one)
  useEffect(() => {
    if (boxes.length === 0 || box) return
    const last = readLastBox()
    const fallback = boxes.find((b) => b.id === last) ?? boxes[0]
    navigate({ box: fallback.id }, true)
  }, [boxes, box])

  useEffect(() => {
    if (!box) return
    try {
      localStorage.setItem(LAST_BOX, box.id)
    } catch {
      /* storage unavailable */
    }
    setNotes(undefined)
    void loadNotes(box.id)
  }, [box?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  // live updates from the daemon
  const refreshNote = useRef(note.refresh)
  refreshNote.current = note.refresh
  useEffect(
    () =>
      subscribe(
        (e) => {
          if (e.type === "boxes") void loadBoxes()
          if (e.type === "notes" && e.box === routeBox.current) {
            void loadNotes(e.box)
            void refreshNote.current()
          }
        },
        (ok) => {
          setConnected(ok)
          // catch up on whatever was missed while disconnected
          if (ok && routeBox.current) {
            void loadNotes(routeBox.current)
            void refreshNote.current()
          }
        },
      ),
    [loadBoxes, loadNotes],
  )

  const createNote = async (title: string) => {
    if (!box) return
    const n = await api.createNote(box.id, title)
    await loadNotes(box.id)
    justCreated.current = n.path
    navigate({ box: box.id, note: n.path })
  }

  const renameNote = async (n: NoteMeta, title: string) => {
    if (!box) return
    if (n.path === route.note) await note.flush()
    const renamed = await api.renameNote(box.id, n.path, title)
    await loadNotes(box.id)
    if (n.path === route.note) navigate({ box: box.id, note: renamed.path }, true)
  }

  const removeNote = async (n: NoteMeta) => {
    if (!box) return
    try {
      await api.removeNote(box.id, n.path)
      if (n.path === route.note) navigate({ box: box.id }, true)
      await loadNotes(box.id)
      toast.success(`Deleted “${n.title}”`)
    } catch (e) {
      toast.error((e as Error).message)
    }
  }

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar
          boxes={boxes}
          box={box}
          notes={notes}
          current={route.note}
          connected={connected}
          onSelectBox={(id) => navigate({ box: id })}
          onSelectNote={(path) => box && navigate({ box: box.id, note: path })}
          onNewNote={() => setDialog({ kind: "new" })}
          onRenameNote={(n) => setDialog({ kind: "rename", note: n })}
          onRemoveNote={(n) => setDialog({ kind: "remove", note: n })}
        />
        <SidebarInset className="h-svh overflow-hidden">
          <header className="flex h-12 shrink-0 items-center gap-2 border-b px-3">
            <SidebarTrigger />
            <Separator orientation="vertical" className="mr-1 data-[orientation=vertical]:h-4" />
            <div className="min-w-0 flex-1 truncate text-sm">
              {route.note ? (
                <>
                  <span className="font-medium">{current?.title ?? route.note}</span>
                  <span className="ml-2 text-muted-foreground">{route.note}</span>
                </>
              ) : (
                <span className="text-muted-foreground">{box?.title}</span>
              )}
            </div>
            {route.note && <StatusBadge status={note.status} error={note.error} />}
          </header>

          {note.status === "conflict" && (
            <div className="flex flex-wrap items-center gap-3 border-b bg-amber-500/10 px-4 py-2 text-sm">
              <AlertTriangle className="size-4 text-amber-600" />
              <span className="flex-1">This note was changed somewhere else while you were editing.</span>
              <Button size="sm" variant="outline" onClick={note.takeTheirs}>
                Use theirs
              </Button>
              <Button size="sm" onClick={note.keepMine}>
                Keep mine
              </Button>
            </div>
          )}

          <main className="min-h-0 flex-1">
            {!route.note ? (
              <Empty icon={<NotebookPen />} text="Select a note, or create one." />
            ) : note.status === "missing" ? (
              <Empty icon={<FileQuestion />} text="This note does not exist anymore." />
            ) : note.ready ? (
              <Editor
                docId={`${box?.id}/${route.note}`}
                value={note.shown.text}
                revision={note.shown.revision}
                onChange={note.onChange}
                cursorAtEnd={justCreated.current === route.note}
              />
            ) : note.status === "error" ? (
              <Empty icon={<CircleAlert />} text={note.error ?? "Failed to load the note."} />
            ) : null}
          </main>
        </SidebarInset>
      </SidebarProvider>

      <TitleDialog
        open={dialog?.kind === "new"}
        onOpenChange={(o) => !o && setDialog(null)}
        title="New note"
        submitLabel="Create"
        onSubmit={createNote}
      />
      <TitleDialog
        open={dialog?.kind === "rename"}
        onOpenChange={(o) => !o && setDialog(null)}
        title="Rename note"
        submitLabel="Rename"
        initial={dialog?.kind === "rename" ? dialog.note.title : ""}
        onSubmit={(t) => (dialog?.kind === "rename" ? renameNote(dialog.note, t) : Promise.resolve())}
      />
      <AlertDialog open={dialog?.kind === "remove"} onOpenChange={(o) => !o && setDialog(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete “{dialog?.kind === "remove" ? dialog.note.title : ""}”?</AlertDialogTitle>
            <AlertDialogDescription>The file is deleted. This cannot be undone.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => dialog?.kind === "remove" && void removeNote(dialog.note)}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <Toaster position="bottom-right" />
    </TooltipProvider>
  )
}

function StatusBadge({ status, error }: { status: SaveStatus; error?: string }) {
  const map: Record<SaveStatus, [React.ReactNode, string]> = {
    loading: [<Loader2 className="size-3.5 animate-spin" />, "Loading"],
    saved: [<Check className="size-3.5" />, "Saved"],
    dirty: [<span className="size-2 rounded-full bg-muted-foreground" />, "Edited"],
    saving: [<Loader2 className="size-3.5 animate-spin" />, "Saving"],
    conflict: [<AlertTriangle className="size-3.5 text-amber-600" />, "Conflict"],
    error: [<CircleAlert className="size-3.5 text-destructive" />, "Not saved"],
    missing: [<FileQuestion className="size-3.5" />, "Missing"],
  }
  const [icon, label] = map[status]
  return (
    <span className="flex items-center gap-1.5 text-xs text-muted-foreground" title={error}>
      {icon}
      {label}
    </span>
  )
}

function Empty({ icon, text }: { icon: React.ReactNode; text: string }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 text-sm text-muted-foreground [&>svg]:size-8">
      {icon}
      <p>{text}</p>
    </div>
  )
}
