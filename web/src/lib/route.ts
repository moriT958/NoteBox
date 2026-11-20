import { useEffect, useState } from "react"

// Route lives in the hash so `nbx open` can link to a note:
//   #/b/<box-id>/<note-path>
export type Route = { box?: string; note?: string }

function parse(): Route {
  const m = /^#\/b\/([^/]+)(?:\/(.+))?$/.exec(window.location.hash)
  if (!m) return {}
  return { box: decodeURIComponent(m[1]), note: m[2] ? decodeURIComponent(m[2]) : undefined }
}

export function navigate(r: Route, replace = false) {
  const hash = r.box ? `#/b/${encodeURIComponent(r.box)}${r.note ? `/${encodeURIComponent(r.note)}` : ""}` : ""
  if (replace) history.replaceState(null, "", hash || window.location.pathname)
  else window.location.hash = hash
  window.dispatchEvent(new HashChangeEvent("hashchange"))
}

export function useRoute(): Route {
  const [route, setRoute] = useState(parse)
  useEffect(() => {
    const on = () => setRoute(parse())
    window.addEventListener("hashchange", on)
    return () => window.removeEventListener("hashchange", on)
  }, [])
  return route
}
