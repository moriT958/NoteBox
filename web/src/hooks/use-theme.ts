import { useSyncExternalStore } from "react"

const query = "(prefers-color-scheme: dark)"

function subscribe(cb: () => void) {
  const mql = window.matchMedia(query)
  mql.addEventListener("change", cb)
  return () => mql.removeEventListener("change", cb)
}

/** Follows the OS color scheme. A theme switcher can come later. */
export function useTheme(): "light" | "dark" {
  return useSyncExternalStore(
    subscribe,
    () => (window.matchMedia(query).matches ? "dark" : "light"),
    () => "light",
  )
}
