import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands"
import { markdown, markdownLanguage } from "@codemirror/lang-markdown"
import { languages } from "@codemirror/language-data"
import { searchKeymap } from "@codemirror/search"
import { Annotation, EditorState, Transaction } from "@codemirror/state"
import { drawSelection, EditorView, keymap } from "@codemirror/view"
import { useEffect, useRef } from "react"
import { linkClick, livePreview } from "./live-preview"
import { editorHighlight, editorTheme } from "./theme"

/** Marks changes that came from the server, so they are not echoed back. */
const external = Annotation.define<boolean>()

type Props = {
  /** Identity of the document; a new id starts a fresh editor state. */
  docId: string
  /** Content to show. It is applied when docId or revision changes, so typing
   * doesn't need to feed it back. */
  value: string
  /** Bump to replace the text in place (e.g. the note changed on disk). */
  revision: number
  onChange: (value: string) => void
  /** Start with the cursor at the end (a note that was just created). */
  cursorAtEnd?: boolean
}

export function Editor({ docId, value, revision, onChange, cursorAtEnd }: Props) {
  const host = useRef<HTMLDivElement>(null)
  const view = useRef<EditorView | null>(null)
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  const newState = (doc: string) =>
    EditorState.create({
      doc,
      selection: cursorAtEnd ? { anchor: doc.length } : undefined,
      extensions: [
        history(),
        drawSelection(),
        EditorView.lineWrapping,
        keymap.of([indentWithTab, ...defaultKeymap, ...historyKeymap, ...searchKeymap]),
        markdown({ base: markdownLanguage, codeLanguages: languages }),
        livePreview,
        linkClick,
        editorTheme,
        editorHighlight,
        EditorView.updateListener.of((u) => {
          if (u.docChanged && !u.transactions.some((tr) => tr.annotation(external))) {
            onChangeRef.current(u.state.doc.toString())
          }
        }),
      ],
    })

  useEffect(() => {
    const v = new EditorView({ parent: host.current!, state: newState(value) })
    view.current = v
    v.focus()
    return () => {
      v.destroy()
      view.current = null
    }
    // the editor is created once per document
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [docId])

  // content changed outside the editor (another client, an agent, vim)
  useEffect(() => {
    const v = view.current
    if (!v) return
    const cur = v.state.doc.toString()
    if (cur === value) return
    const sel = v.state.selection.main
    v.dispatch({
      changes: { from: 0, to: cur.length, insert: value },
      selection: { anchor: Math.min(sel.anchor, value.length), head: Math.min(sel.head, value.length) },
      annotations: [external.of(true), Transaction.addToHistory.of(false)],
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [revision])

  return <div ref={host} className="h-full overflow-hidden [&_.cm-editor]:h-full" />
}
