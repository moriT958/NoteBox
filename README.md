# 📓 NoteBox 📓

A terminal-based note-taking app.

![](/assets/overview.png)

## Features

- Fast and easy-to-read TUI
- Jot down notes quickly in the terminal
- Nicely rendered Markdown
- Fast fuzzy search
- Note management with boxes
- A scriptable CLI, handy for AI agents too

## Installation

**Using cURL**

```bash
curl -sSfL https://raw.githubusercontent.com/moriT958/NoteBox/main/install.sh | sh
```

## Usage

`nbx` to start

### Default keybindings

- `j/k` | `↓/↑`: move cursor
- `ctrl+l` | `ctrl+h`: Toggle focus between the preview pane and the list panel
- `n`: New note
- `e`: Edit note
- `d`: Delete note
- `/`: Find note by title
- `ctrl+b`: Open Box

### Subcommands

Every command is `nbx <target> <action>`. Output is plain text, one item per line, with tab-separated columns, so it works well in pipes and from scripts and AI agents. Commands never ask questions, except `box prune` without `-force`.

```
nbx box list                          # <title>\t<path>
nbx box add [-title <title>] <path>   # register a directory as a box (created if missing)
nbx box rename <path> <new-title>
nbx box relink <path> <new-path>      # point a box at another directory (files are not moved)
nbx box rm <path>                     # remove a box, its files are kept
nbx box prune [-force]                # delete the files of removed boxes

nbx note list [-box <path>]           # <title>\t<note-path>
nbx note new [-box <path>] <title>
nbx note rename [-box <path>] <note-path> <new-title>
nbx note view [-box <path>] <note-path>
nbx note rm [-box <path>] <note-path>

nbx version
nbx help [<command>]
```

- A box is given by its directory path. Without `-box`, note commands use the default box (`~/.notebox/notes`), which cannot be removed.
- `<note-path>` is the path relative to the box, as shown by `nbx note list`.

## Config

- config file: `~/.notebox/config.json`
