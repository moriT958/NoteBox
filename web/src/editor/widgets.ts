import { EditorView, WidgetType } from "@codemirror/view"

/** A list bullet shown instead of "-", "*" or "+". */
export class BulletWidget extends WidgetType {
  eq() {
    return true
  }
  toDOM() {
    const span = document.createElement("span")
    span.className = "cm-lp-bullet"
    span.textContent = "•"
    return span
  }
}

/** A clickable checkbox shown instead of "- [ ]" / "- [x]". */
export class CheckboxWidget extends WidgetType {
  constructor(readonly checked: boolean) {
    super()
  }
  eq(other: CheckboxWidget) {
    return other.checked === this.checked
  }
  toDOM(view: EditorView) {
    const wrap = document.createElement("span")
    wrap.className = "cm-lp-task"
    const box = document.createElement("input")
    box.type = "checkbox"
    box.checked = this.checked
    box.addEventListener("mousedown", (e) => e.preventDefault())
    box.addEventListener("click", (e) => {
      e.preventDefault()
      toggleTask(view, view.posAtDOM(wrap))
    })
    wrap.appendChild(box)
    return wrap
  }
  ignoreEvent() {
    return true
  }
}

/** Flips the "[ ]" / "[x]" of the task starting at or after pos on its line. */
function toggleTask(view: EditorView, pos: number) {
  const line = view.state.doc.lineAt(pos)
  const m = /\[( |x|X)\]/.exec(line.text.slice(pos - line.from))
  if (!m) return
  const at = pos + m.index + 1
  view.dispatch({ changes: { from: at, to: at + 1, insert: m[1] === " " ? "x" : " " } })
}

export class HrWidget extends WidgetType {
  eq() {
    return true
  }
  toDOM() {
    const el = document.createElement("span")
    el.className = "cm-lp-hr"
    return el
  }
}

export class ImageWidget extends WidgetType {
  constructor(
    readonly src: string,
    readonly alt: string,
  ) {
    super()
  }
  eq(other: ImageWidget) {
    return other.src === this.src && other.alt === this.alt
  }
  toDOM() {
    if (!/^(https?:|data:)/.test(this.src)) {
      // files inside the box are not served yet
      const span = document.createElement("span")
      span.className = "cm-lp-image-missing"
      span.textContent = `🖼 ${this.alt || this.src}`
      return span
    }
    const img = document.createElement("img")
    img.className = "cm-lp-image"
    img.src = this.src
    img.alt = this.alt
    return img
  }
}

/**
 * A block shown instead of its source (tables, mermaid). Clicking it puts the
 * cursor inside, which reveals the source.
 */
abstract class BlockWidget extends WidgetType {
  constructor(readonly source: string) {
    super()
  }
  eq(other: BlockWidget) {
    return other.source === this.source
  }
  protected wrap(view: EditorView, className: string) {
    const el = document.createElement("div")
    el.className = `cm-lp-block ${className}`
    el.addEventListener("mousedown", (e) => {
      e.preventDefault()
      const pos = view.posAtDOM(el)
      view.dispatch({ selection: { anchor: pos } })
      view.focus()
    })
    return el
  }
  ignoreEvent() {
    return true
  }
}

export class TableWidget extends BlockWidget {
  toDOM(view: EditorView) {
    const el = this.wrap(view, "cm-lp-table")
    const rows = this.source
      .split("\n")
      .filter((l) => l.trim() !== "")
      .map(splitRow)
    if (rows.length === 0) return el
    const aligns = rows.length > 1 ? rows[1].map(alignOf) : []
    const table = document.createElement("table")
    const thead = table.createTHead().insertRow()
    rows[0].forEach((c, i) => {
      const th = document.createElement("th")
      th.textContent = c
      if (aligns[i]) th.style.textAlign = aligns[i]
      thead.appendChild(th)
    })
    const tbody = table.createTBody()
    for (const r of rows.slice(2)) {
      const tr = tbody.insertRow()
      r.forEach((c, i) => {
        const td = tr.insertCell()
        td.textContent = c
        if (aligns[i]) td.style.textAlign = aligns[i]
      })
    }
    el.appendChild(table)
    return el
  }
}

function splitRow(line: string): string[] {
  let s = line.trim()
  if (s.startsWith("|")) s = s.slice(1)
  if (s.endsWith("|") && !s.endsWith("\\|")) s = s.slice(0, -1)
  return s.split(/(?<!\\)\|/).map((c) => c.trim().replace(/\\\|/g, "|"))
}

function alignOf(cell: string): string {
  const l = cell.startsWith(":")
  const r = cell.endsWith(":")
  return l && r ? "center" : r ? "right" : l ? "left" : ""
}

const mermaidCache = new Map<string, string>()
let mermaidSeq = 0

export class MermaidWidget extends BlockWidget {
  toDOM(view: EditorView) {
    const el = this.wrap(view, "cm-lp-mermaid")
    const dark = window.matchMedia("(prefers-color-scheme: dark)").matches
    const key = `${dark}:${this.source}`
    const cached = mermaidCache.get(key)
    if (cached !== undefined) {
      el.innerHTML = cached
      return el
    }
    el.textContent = "Rendering diagram…"
    import("mermaid")
      .then(async ({ default: mermaid }) => {
        mermaid.initialize({ startOnLoad: false, theme: dark ? "dark" : "default", securityLevel: "strict" })
        const { svg } = await mermaid.render(`nbx-mermaid-${++mermaidSeq}`, this.source)
        mermaidCache.set(key, svg)
        el.innerHTML = svg
      })
      .catch((err: unknown) => {
        el.classList.add("cm-lp-error")
        el.textContent = `Mermaid: ${err instanceof Error ? err.message : String(err)}`
      })
      .finally(() => view.requestMeasure())
    return el
  }
}
