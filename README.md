# 📓 NoteBox 📓

A terminal-based note-taking app.

![](/assets/overview.png)

## Features

- 高速で見やすい TUI
- ターミナル上でサクッとメモが取れる
- Markdown の見やすい表示
- FuzzyFinder で高速検索
- Box 機能によるノート管理
  - (Obsidian Vault みたいな感じの機能)

## Installation

**Using cURL**

```bash
curl -sSfL https://raw.githubusercontent.com/moriT958/NoteBox/main/install.sh | sh
```

## Usage

`notebox` to start

### defalut keybindings

- `j/k` | `↓/↑`: move cursor
- `ctrl+l` | `ctrl+h`: Toggle focus between the preview pane and the list panel
- `n`: New note
- `e`: Edit note
- `d`: Delete note
- `/`: Find note by title
- `ctrl+b`: Open Box

### Subcommands

- `note version`: show version

## Config

- config file: `~/.notebox/config.json`
