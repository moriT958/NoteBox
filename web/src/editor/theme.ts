import { HighlightStyle, syntaxHighlighting } from "@codemirror/language"
import { EditorView } from "@codemirror/view"
import { tags as t } from "@lezer/highlight"

// Colors come from the shadcn CSS variables, so the editor follows the app
// theme (and later, the design system) without its own palette.
export const editorTheme = EditorView.theme({
  "&": {
    height: "100%",
    fontSize: "15px",
    backgroundColor: "transparent",
    color: "var(--foreground)",
  },
  "&.cm-focused": { outline: "none" },
  ".cm-scroller": {
    fontFamily: "ui-sans-serif, system-ui, sans-serif",
    lineHeight: "1.7",
  },
  ".cm-content": {
    maxWidth: "48rem",
    margin: "0 auto",
    padding: "2rem 1.5rem 40vh",
    caretColor: "var(--foreground)",
  },
  ".cm-line": { padding: "0" },
  ".cm-cursor": { borderLeftColor: "var(--foreground)" },
  "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, ::selection": {
    backgroundColor: "color-mix(in oklch, var(--ring) 35%, transparent) !important",
  },
  ".cm-activeLine": { backgroundColor: "transparent" },
  ".cm-gutters": { display: "none" },

  ".cm-lp-h": { fontWeight: "700", lineHeight: "1.3" },
  ".cm-lp-h1": { fontSize: "2em", paddingTop: "0.6em" },
  ".cm-lp-h2": { fontSize: "1.5em", paddingTop: "0.5em" },
  ".cm-lp-h3": { fontSize: "1.25em", paddingTop: "0.4em" },
  ".cm-lp-h4, .cm-lp-h5, .cm-lp-h6": { fontSize: "1.05em" },
  ".cm-lp-strong": { fontWeight: "700" },
  ".cm-lp-em": { fontStyle: "italic" },
  ".cm-lp-strike, .cm-lp-done": { textDecoration: "line-through", color: "var(--muted-foreground)" },
  ".cm-lp-code": {
    fontFamily: "ui-monospace, SFMono-Regular, monospace",
    fontSize: "0.9em",
    backgroundColor: "var(--muted)",
    borderRadius: "4px",
    padding: "0.1em 0.3em",
  },
  ".cm-lp-link": { color: "var(--primary)", textDecoration: "underline", textUnderlineOffset: "3px", cursor: "text" },
  ".cm-lp-quote": {
    borderLeft: "3px solid var(--border)",
    paddingLeft: "1em !important",
    color: "var(--muted-foreground)",
  },
  ".cm-lp-bullet": { display: "inline-block", width: "1ch", color: "var(--muted-foreground)" },
  ".cm-lp-task input": { margin: "0 0.4em 0 0", verticalAlign: "middle", cursor: "pointer" },
  ".cm-lp-hr": { display: "inline-block", width: "100%", borderTop: "1px solid var(--border)", verticalAlign: "middle" },
  ".cm-lp-image": { maxWidth: "100%", borderRadius: "var(--radius)" },
  ".cm-lp-image-missing": { color: "var(--muted-foreground)" },
  ".cm-lp-codeblock": {
    fontFamily: "ui-monospace, SFMono-Regular, monospace",
    fontSize: "0.88em",
    backgroundColor: "var(--muted)",
    padding: "0 1em !important",
  },
  ".cm-lp-codeblock-first": { borderTopLeftRadius: "var(--radius)", borderTopRightRadius: "var(--radius)", paddingTop: "0.4em !important" },
  ".cm-lp-codeblock-last": { borderBottomLeftRadius: "var(--radius)", borderBottomRightRadius: "var(--radius)", paddingBottom: "0.4em !important" },
  ".cm-lp-fence": { color: "var(--muted-foreground)" },
  ".cm-lp-block": { cursor: "text", margin: "0.5em 0" },
  ".cm-lp-table table": { borderCollapse: "collapse", width: "100%", fontSize: "0.95em" },
  ".cm-lp-table th, .cm-lp-table td": { border: "1px solid var(--border)", padding: "0.35em 0.7em" },
  ".cm-lp-table th": { backgroundColor: "var(--muted)", fontWeight: "600", textAlign: "left" },
  ".cm-lp-mermaid": { display: "flex", justifyContent: "center", color: "var(--muted-foreground)" },
  ".cm-lp-mermaid svg": { maxWidth: "100%", height: "auto" },
  ".cm-lp-error": { color: "var(--destructive)", fontFamily: "ui-monospace, monospace", fontSize: "0.85em" },
})

const highlight = HighlightStyle.define([
  { tag: t.processingInstruction, color: "var(--muted-foreground)" }, // marks shown while editing
  { tag: [t.keyword, t.modifier, t.operatorKeyword], color: "oklch(0.55 0.18 300)" },
  { tag: [t.string, t.special(t.string)], color: "oklch(0.55 0.13 150)" },
  { tag: [t.number, t.bool, t.null, t.atom], color: "oklch(0.6 0.15 50)" },
  { tag: [t.comment], color: "var(--muted-foreground)", fontStyle: "italic" },
  { tag: [t.function(t.variableName), t.function(t.propertyName)], color: "oklch(0.55 0.15 250)" },
  { tag: [t.typeName, t.className], color: "oklch(0.6 0.13 200)" },
  { tag: t.url, color: "var(--muted-foreground)" },
])

export const editorHighlight = syntaxHighlighting(highlight)
