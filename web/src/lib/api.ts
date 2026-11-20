// Client of the daemon's HTTP API. Note content is plain markdown; only
// metadata is JSON.

export type Box = { id: string; title: string; path: string }
export type NoteMeta = { title: string; path: string; mtime?: string }
export type NoteContent = { text: string; etag: string }

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message)
  }
}

export class Conflict extends Error {
  constructor(readonly etag: string) {
    super("The note was changed somewhere else.")
  }
}

const notePath = (box: string, path: string) =>
  `/api/boxes/${encodeURIComponent(box)}/notes/${path.split("/").map(encodeURIComponent).join("/")}`

async function check(res: Response): Promise<Response> {
  if (res.ok || res.status === 304) return res
  let msg = res.statusText
  try {
    msg = (await res.json()).error ?? msg
  } catch {
    /* not JSON */
  }
  throw new ApiError(res.status, msg)
}

async function json<T>(res: Promise<Response>): Promise<T> {
  return (await check(await res)).json()
}

const jsonBody = (method: string, body: unknown): RequestInit => ({
  method,
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(body),
})

export const api = {
  boxes: () => json<Box[]>(fetch("/api/boxes")),
  notes: (box: string) => json<NoteMeta[]>(fetch(`/api/boxes/${encodeURIComponent(box)}/notes`)),
  createNote: (box: string, title: string) =>
    json<NoteMeta>(fetch(`/api/boxes/${encodeURIComponent(box)}/notes`, jsonBody("POST", { title }))),
  renameNote: (box: string, path: string, title: string) =>
    json<NoteMeta>(fetch(notePath(box, path), jsonBody("PATCH", { title }))),
  removeNote: async (box: string, path: string) => {
    await check(await fetch(notePath(box, path), { method: "DELETE" }))
  },

  /** Returns null when the note still has the ETag given as ifNoneMatch. */
  readNote: async (box: string, path: string, ifNoneMatch?: string): Promise<NoteContent | null> => {
    const res = await check(
      await fetch(notePath(box, path), { headers: ifNoneMatch ? { "If-None-Match": `"${ifNoneMatch}"` } : {} }),
    )
    if (res.status === 304) return null
    return { text: await res.text(), etag: unquote(res.headers.get("ETag")) }
  },

  /** Saves the note if it is still at ifMatch; throws Conflict otherwise. */
  writeNote: async (box: string, path: string, text: string, ifMatch: string): Promise<string> => {
    const res = await fetch(notePath(box, path), {
      method: "PUT",
      headers: { "Content-Type": "text/markdown; charset=utf-8", "If-Match": `"${ifMatch}"` },
      body: text,
    })
    if (res.status === 412) throw new Conflict(unquote(res.headers.get("ETag")))
    await check(res)
    return unquote(res.headers.get("ETag"))
  },
}

const unquote = (s: string | null) => (s ?? "").replace(/^W\//, "").replace(/"/g, "")

export type ServerEvent = { type: "notes"; box: string } | { type: "boxes" }

/** Subscribes to file changes. Reconnects on its own (EventSource). */
export function subscribe(onEvent: (e: ServerEvent) => void, onStatus: (connected: boolean) => void) {
  const es = new EventSource("/api/events")
  es.onopen = () => onStatus(true)
  es.onerror = () => onStatus(false)
  es.onmessage = (m) => onEvent(JSON.parse(m.data))
  return () => es.close()
}
