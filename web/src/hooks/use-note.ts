import { useCallback, useEffect, useRef, useState } from "react"
import { api, Conflict } from "@/lib/api"

const SAVE_DELAY = 600

export type SaveStatus = "loading" | "saved" | "dirty" | "saving" | "conflict" | "error" | "missing"

type Shown = { text: string; revision: number }

/**
 * Keeps one note in sync with the daemon:
 * - edits are saved after a short pause, with If-Match so nothing written
 *   elsewhere is overwritten silently,
 * - changes made elsewhere (refresh()) are applied when there are no unsaved
 *   edits, and reported as a conflict when there are.
 */
export function useNote(box: string | undefined, path: string | undefined) {
  const [shown, setShown] = useState<Shown>({ text: "", revision: 0 })
  const [status, setStatus] = useState<SaveStatus>("loading")
  const [error, setError] = useState<string>()
  const [loadedKey, setLoadedKey] = useState<string>()

  // latest values, for async callbacks
  const etag = useRef("") // version on disk that `base` matches
  const base = useRef("") // content on disk, as far as we know
  const draft = useRef("") // content in the editor
  const remote = useRef<{ text: string; etag: string } | null>(null) // theirs, during a conflict
  const timer = useRef<number | undefined>(undefined)
  const saving = useRef<Promise<void> | null>(null)
  const key = `${box}\u0000${path}`
  const keyRef = useRef(key)
  keyRef.current = key

  const show = (text: string) => setShown((s) => ({ text, revision: s.revision + 1 }))

  const save = useCallback(async () => {
    window.clearTimeout(timer.current)
    if (!box || !path || draft.current === base.current || remote.current) return
    if (saving.current) {
      await saving.current
      return save()
    }
    const myKey = key
    const text = draft.current
    setStatus("saving")
    saving.current = (async () => {
      try {
        const tag = await api.writeNote(box, path, text, etag.current)
        if (keyRef.current !== myKey) return
        etag.current = tag
        base.current = text
        setStatus(draft.current === text ? "saved" : "dirty")
        setError(undefined)
      } catch (e) {
        if (keyRef.current !== myKey) return
        if (e instanceof Conflict) {
          const theirs = await api.readNote(box, path)
          remote.current = theirs
          setStatus("conflict")
        } else {
          setStatus("error")
          setError(e instanceof Error ? e.message : String(e))
        }
      } finally {
        saving.current = null
      }
    })()
    await saving.current
    if (keyRef.current === myKey && draft.current !== base.current && !remote.current) {
      timer.current = window.setTimeout(save, SAVE_DELAY)
    }
  }, [box, path, key])

  // load on open; flush the previous note's edits first
  useEffect(() => {
    if (!box || !path) return
    let cancelled = false
    remote.current = null
    setStatus("loading")
    setError(undefined)
    api
      .readNote(box, path)
      .then((n) => {
        if (cancelled || !n) return
        etag.current = n.etag
        base.current = draft.current = n.text
        show(n.text)
        setLoadedKey(`${box}\u0000${path}`)
        setStatus("saved")
      })
      .catch((e) => {
        if (cancelled) return
        setStatus(e?.status === 404 ? "missing" : "error")
        setError(e instanceof Error ? e.message : String(e))
      })
    return () => {
      cancelled = true
      void save()
    }
  }, [box, path]) // eslint-disable-line react-hooks/exhaustive-deps

  const onChange = useCallback(
    (text: string) => {
      draft.current = text
      if (remote.current) return // keep the conflict until it is resolved
      setStatus("dirty")
      window.clearTimeout(timer.current)
      timer.current = window.setTimeout(save, SAVE_DELAY)
    },
    [save],
  )

  /** Something in the box changed on disk; pick it up if it is this note. */
  const refresh = useCallback(async () => {
    if (!box || !path || saving.current) return
    const myKey = key
    try {
      const n = await api.readNote(box, path, etag.current)
      if (!n || keyRef.current !== myKey || n.etag === etag.current) return
      if (draft.current === base.current) {
        etag.current = n.etag
        base.current = draft.current = n.text
        show(n.text)
        setStatus("saved")
      } else if (n.text !== draft.current) {
        remote.current = n
        setStatus("conflict")
      } else {
        // they wrote exactly what we have
        etag.current = n.etag
        base.current = n.text
        setStatus("saved")
      }
    } catch (e) {
      if (keyRef.current === myKey && (e as { status?: number }).status === 404) setStatus("missing")
    }
  }, [box, path, key])

  /** Resolve a conflict by taking the version on disk. */
  const takeTheirs = useCallback(() => {
    const r = remote.current
    if (!r) return
    remote.current = null
    etag.current = r.etag
    base.current = draft.current = r.text
    show(r.text)
    setStatus("saved")
  }, [])

  /** Resolve a conflict by writing ours over the version on disk. */
  const keepMine = useCallback(() => {
    const r = remote.current
    if (!r) return
    remote.current = null
    etag.current = r.etag
    base.current = r.text
    void save()
  }, [save])

  // warn before closing the tab with unsaved edits
  useEffect(() => {
    const on = (e: BeforeUnloadEvent) => {
      if (draft.current !== base.current) e.preventDefault()
    }
    window.addEventListener("beforeunload", on)
    return () => window.removeEventListener("beforeunload", on)
  }, [])

  return { ready: loadedKey === key, shown, status, error, onChange, refresh, takeTheirs, keepMine, flush: save }
}
