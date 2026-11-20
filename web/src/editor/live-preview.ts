import { ensureSyntaxTree, syntaxTree } from "@codemirror/language"
import { type EditorState, type Range, StateField } from "@codemirror/state"
import { Decoration, type DecorationSet, EditorView } from "@codemirror/view"
import type { SyntaxNode } from "@lezer/common"
import { BulletWidget, CheckboxWidget, HrWidget, ImageWidget, MermaidWidget, TableWidget } from "./widgets"

/**
 * Live preview: markdown is rendered in place, and the syntax of whatever the
 * cursor is on comes back so it can be edited. The document itself is never
 * rewritten, so the file stays exactly as typed.
 *
 * - inline syntax (#, **, `, [](), >) is revealed per line
 * - blocks (tables, mermaid) are revealed while the cursor is inside them
 */
export const livePreview = StateField.define<DecorationSet>({
  create: (state) => build(state),
  update(deco, tr) {
    if (tr.docChanged || tr.selection || syntaxTree(tr.state) !== syntaxTree(tr.startState)) {
      return build(tr.state)
    }
    return deco
  },
  provide: (f) => EditorView.decorations.from(f),
})

const hide = Decoration.replace({})
const mark = (cls: string) => Decoration.mark({ class: cls })
const line = (cls: string) => Decoration.line({ class: cls })

function build(state: EditorState): DecorationSet {
  const tree = ensureSyntaxTree(state, state.doc.length, 50) ?? syntaxTree(state)
  const doc = state.doc
  const out: Range<Decoration>[] = []

  const activeLines = new Set<number>()
  for (const r of state.selection.ranges) {
    const a = doc.lineAt(r.from).number
    const b = doc.lineAt(r.to).number
    for (let n = a; n <= b; n++) activeLines.add(n)
  }
  const lineActive = (pos: number) => activeLines.has(doc.lineAt(pos).number)
  const rangeActive = (from: number, to: number) =>
    state.selection.ranges.some((r) => r.from <= to && r.to >= from)

  // hide a syntax mark, plus one space after it when there is one
  const hideMark = (n: SyntaxNode, withSpace = false) => {
    let to = n.to
    if (withSpace && doc.sliceString(to, to + 1) === " ") to++
    if (to > n.from) out.push(hide.range(n.from, to))
  }
  const eachLine = (from: number, to: number, f: (lineFrom: number, i: number, last: boolean) => void) => {
    const first = doc.lineAt(from).number
    const lastN = doc.lineAt(to).number
    for (let n = first; n <= lastN; n++) f(doc.line(n).from, n - first, n === lastN)
  }

  tree.iterate({
    enter(ref) {
      const n = ref.node
      const name = n.name

      const heading = /^ATXHeading(\d)$/.exec(name)
      if (heading) {
        out.push(line(`cm-lp-h cm-lp-h${heading[1]}`).range(doc.lineAt(n.from).from))
        if (!lineActive(n.from)) {
          const hm = n.getChild("HeaderMark")
          if (hm) hideMark(hm, true)
          // closing #s
          const marks = n.getChildren("HeaderMark")
          if (marks.length > 1) hideMark(marks[marks.length - 1])
        }
        return
      }

      switch (name) {
        case "StrongEmphasis":
        case "Emphasis":
        case "Strikethrough": {
          const cls = name === "StrongEmphasis" ? "cm-lp-strong" : name === "Emphasis" ? "cm-lp-em" : "cm-lp-strike"
          out.push(mark(cls).range(n.from, n.to))
          if (!lineActive(n.from)) {
            for (const m of n.getChildren(name === "Strikethrough" ? "StrikethroughMark" : "EmphasisMark")) hideMark(m)
          }
          return
        }
        case "InlineCode": {
          out.push(mark("cm-lp-code").range(n.from, n.to))
          if (!lineActive(n.from)) for (const m of n.getChildren("CodeMark")) hideMark(m)
          return false
        }
        case "Link": {
          const url = n.getChild("URL")
          const marks = n.getChildren("LinkMark")
          if (marks.length >= 2) {
            const textFrom = marks[0].to
            const textTo = marks[1].from
            const href = url ? doc.sliceString(url.from, url.to) : ""
            if (textTo > textFrom) {
              out.push(Decoration.mark({ class: "cm-lp-link", attributes: { "data-href": href } }).range(textFrom, textTo))
            }
            if (!lineActive(n.from)) {
              out.push(hide.range(marks[0].from, marks[0].to))
              out.push(hide.range(marks[1].from, n.to))
            }
          }
          return
        }
        case "URL": {
          // bare autolinks
          if (n.parent?.name !== "Link" && n.parent?.name !== "Image") {
            const href = doc.sliceString(n.from, n.to).replace(/^<|>$/g, "")
            out.push(Decoration.mark({ class: "cm-lp-link", attributes: { "data-href": href } }).range(n.from, n.to))
          }
          return
        }
        case "Image": {
          if (!lineActive(n.from)) {
            const url = n.getChild("URL")
            const marks = n.getChildren("LinkMark")
            const alt = marks.length >= 2 ? doc.sliceString(marks[0].to, marks[1].from) : ""
            const src = url ? doc.sliceString(url.from, url.to) : ""
            out.push(Decoration.replace({ widget: new ImageWidget(src, alt) }).range(n.from, n.to))
          }
          return false
        }
        case "Blockquote": {
          eachLine(n.from, n.to, (from) => out.push(line("cm-lp-quote").range(from)))
          return
        }
        case "QuoteMark": {
          if (!lineActive(n.from)) hideMark(n, true)
          return
        }
        case "ListMark": {
          const item = n.parent
          if (!item || lineActive(n.from)) return
          const task = item.getChild("Task")
          const marker = task?.getChild("TaskMarker")
          if (marker) {
            const checked = /x/i.test(doc.sliceString(marker.from, marker.to))
            const end = doc.sliceString(marker.to, marker.to + 1) === " " ? marker.to + 1 : marker.to
            out.push(Decoration.replace({ widget: new CheckboxWidget(checked) }).range(n.from, end))
            if (checked && task!.to > end) out.push(mark("cm-lp-done").range(end, task!.to))
          } else if (item.parent?.name === "BulletList") {
            out.push(Decoration.replace({ widget: new BulletWidget() }).range(n.from, n.to))
          }
          return
        }
        case "HorizontalRule": {
          if (!lineActive(n.from)) out.push(Decoration.replace({ widget: new HrWidget() }).range(n.from, n.to))
          return
        }
        case "FencedCode": {
          const info = n.getChild("CodeInfo")
          const lang = info ? doc.sliceString(info.from, info.to).trim().toLowerCase() : ""
          const active = rangeActive(n.from, n.to)
          if (lang === "mermaid" && !active) {
            const text = n.getChild("CodeText")
            const src = text ? doc.sliceString(text.from, text.to) : ""
            out.push(Decoration.replace({ widget: new MermaidWidget(src), block: true }).range(n.from, n.to))
            return false
          }
          eachLine(n.from, n.to, (from, i, last) => {
            const fence = i === 0 || (last && /^\s*(```|~~~)/.test(doc.lineAt(from).text))
            out.push(
              line(
                ["cm-lp-codeblock", i === 0 && "cm-lp-codeblock-first", last && "cm-lp-codeblock-last", fence && "cm-lp-fence"]
                  .filter(Boolean)
                  .join(" "),
              ).range(from),
            )
          })
          return
        }
        case "Table": {
          if (!rangeActive(n.from, n.to)) {
            out.push(
              Decoration.replace({ widget: new TableWidget(doc.sliceString(n.from, n.to)), block: true }).range(n.from, n.to),
            )
          }
          return false
        }
      }
    },
  })

  return Decoration.set(out, true)
}

/** Ctrl/Cmd+click on a rendered link opens it. */
export const linkClick = EditorView.domEventHandlers({
  mousedown(e) {
    const el = (e.target as HTMLElement).closest<HTMLElement>(".cm-lp-link")
    const href = el?.dataset.href
    if (!href || !(e.ctrlKey || e.metaKey)) return false
    e.preventDefault()
    window.open(href, "_blank", "noopener")
    return true
  },
})
