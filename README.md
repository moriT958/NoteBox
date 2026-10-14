# 📓 NoteBox 📓

A terminal-based note-taking app.

![](/assets/overview.png)

## Features

- Fast and easy-to-read TUI
- Jot down notes quickly in the terminal
- Nicely rendered Markdown
- Fast fuzzy search
- Note management with boxes

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

nbx keymap list                       # <action>\t<keys>
nbx theme list                        # <name>\t<builtin|file>
nbx theme export <name>               # print a built-in theme as a theme file

nbx version
nbx help [<command>]
```

- A box is given by its directory path. Without `-box`, note commands use the default box (`~/.notebox/notes`), which cannot be removed.
- `<note-path>` is the path relative to the box, as shown by `nbx note list`.

## Config

The config file is `~/.notebox/config.json`. `nbx` creates it with the defaults the first time it starts, and leaves it alone after that.

```json
{
  "editor": "vim",
  "theme": "dark",
  "keymap": {
    "list.new": ["n", "a"],
    "list.delete": ["x"]
  }
}
```

Anything left out keeps its default. When something in the config can't be used, `nbx` starts anyway with the default for that part and lists the problems in a dialog.

### Keymap

`keymap` replaces the keys of an action, which is named `<where>.<action>` such as `list.new` or `preview.close_tab`.

- `nbx keymap list` shows every action with its default keys.
- The keys you give replace the default keys of the action, so the old keys stop working.
- Keys are written as they are pressed: `n`, `?`, `enter`, `esc`, `tab`, `up`, `pgdown`, `f5`, `ctrl+c`. A shifted letter is the capital letter (`N`, not `shift+n`), and modifiers go in the order `ctrl+alt+shift+`.
- Actions that are active at the same time can't share a key. Both go back to their defaults and the conflict is reported.

### Theme

`theme` is the name of a theme: `dark`, `light`, or a file `~/.notebox/themes/<name>.json`. A file takes precedence over a built-in theme of the same name. `nbx theme list` shows the themes you can use.

To make a theme of your own, start from a built-in one:

```bash
mkdir -p ~/.notebox/themes
nbx theme export dark > ~/.notebox/themes/mine.json
```

and set `"theme": "mine"`. A theme file gives colors (`#rgb` or `#rrggbb`) by what they are for, on top of its `base` (`dark` or `light`, `dark` when left out). You only need to write the colors you want to change:

```json
{
  "base": "dark",
  "accent": "#f5c2e7",
  "danger": "#f38ba8"
}
```
