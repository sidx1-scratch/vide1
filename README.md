<div align="center">

<pre>
██╗   ██╗██╗██████╗ ███████╗ ██╗
██║   ██║██║██╔══██╗██╔════╝███║
██║   ██║██║██║  ██║█████╗  ╚██║
╚██╗ ██╔╝██║██║  ██║██╔══╝   ██║
 ╚████╔╝ ██║██████╔╝███████╗ ██║
  ╚═══╝  ╚═╝╚═════╝ ╚══════╝ ╚═╝
</pre>

**Vim Integrated Development Environment**

**Vim Inbuilt-editing Dynamically-tiling file Explorer**

*A modern, keyboard-driven terminal file manager and editor built in Go.*

[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Bubble Tea](https://img.shields.io/badge/Bubble%20Tea-TUI-FF69B4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue?style=flat-square)](https://www.gnu.org/licenses/gpl-3.0)

</div>

---

### VIDE1 Name Meaning

VIDE1 intentionally carries **two complementary expansions**, reflecting the dual nature of the tool:

- **Vim Integrated Development Environment**  
- **Vim Inbuilt-editing Dynamically-tiling file Explorer**

These two meanings describe the two halves of VIDE1’s identity: a modal, keyboard-driven editor inspired by Vim, and a dynamic tiling file explorer inspired by modern tiling window managers.

---

## Why VIDE1 Has Two Meanings

VIDE1 was designed from the start to be both an editor and an environment. Traditional terminal tools tend to split these roles: one program for editing, another for file navigation, another for terminal panes, another for workspace management. VIDE1 merges all of these into a single binary.

Because of that hybrid nature, the name intentionally reflects **both sides of the tool**:

- As a **Vim Integrated Development Environment**, VIDE1 behaves like a lightweight modal IDE with panes, previews, terminals, and editing modes.
- As a **Vim Inbuilt-editing Dynamically-tiling file Explorer**, VIDE1 behaves like a fast, Miller-column file navigator with dynamic tiling and inline editing.

The dual meaning isn’t a pun — it’s a declaration of scope. VIDE1 is not just an editor with a file tree, and not just a file manager with an editor bolted on. It is both, equally, by design.

This dual identity guides its architecture, keybindings, and workflow philosophy:  
**everything is a pane, everything is modal, everything is keyboard-first.**


## Overview

VIDE1 is a **fully keyboard-driven, terminal-native** development workspace that combines a Vim-style file browser, a Hyprland-style dynamic tiling layout engine, an inline code editor, an asynchronous syntax-highlighted file previewer, and a live embedded PTY shell — all inside a single TUI binary written in pure Go.

No Electron. No LSP daemon. No config files needed to get started. Just launch and go.

---

## Features

### 🗂️ Miller Column File Browser
Navigate your filesystem using the classic [Miller Columns](https://en.wikipedia.org/wiki/Miller_columns) pattern — three columns showing your parent directory, current directory, and a live preview of the selected item.

### 🪟 Hyprland-Style Dynamic Tiling
Split any pane horizontally or vertically on demand. The layout engine recursively partitions terminal dimensions using a binary tree — every pane gets exactly the space it deserves, recalculated live on resize.

### 🔍 Async Syntax-Highlighted Preview
Hover over any code file and a goroutine immediately fires to load and highlight it using [Chroma](https://github.com/alecthomas/chroma). The UI never blocks. Line numbers included.

### ✏️ Inline Vim Editor
Open any file directly inside a tiling pane. A fully modal editor — **Normal**, **Insert**, and **Command** modes — with no external process spawned. Your buffer, your pane, your workspace.

### 💻 Embedded PTY Terminal
Spawn a live interactive shell inside any pane with `ctrl+t`. Keyboard input is piped directly to the shell process via a pseudo-terminal. `stdout`/`stderr` stream back in real time.

### 🎨 TrueColor Aesthetics
- 24-bit color One Dark palette
- Nerd Font icons next to every file entry
- Per-extension neon color coding
- Rounded gradient borders (active pane: magenta, inactive: gray)
- Color-coded status bar showing current mode and path

---

## Installation

### Prerequisites
- Go 1.21 or later
- A terminal with TrueColor support (kitty, Alacritty, WezTerm, iTerm2, etc.)
- A [Nerd Font](https://www.nerdfonts.com/) installed and set as your terminal font (for icons)

### Build from source

```bash
git clone https://github.com/sidx1-scratch/vide1
cd vide1
go build -o vide1 .
./vide1
```

Or run directly without building:

```bash
go run . 
```

---

## Keybindings

### 🗂️ Explorer Mode

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `l` / `→` / `enter` | Enter directory |
| `h` / `←` | Go to parent directory |
| `e` | Open file in inline editor |
| `q` | Quit VIDE1 |
| `ctrl+c` | Force quit |

### 🪟 Tiling / Workspace

| Key | Action |
|-----|--------|
| `ctrl+w` | Split active pane (auto direction) |
| `ctrl+shift+tab` | Split active pane (alternate binding) |
| `tab` | Cycle focus to next pane |
| `ctrl+t` | Spawn a new live terminal pane |

### ✏️ Editor — Normal Mode

| Key | Action |
|-----|--------|
| `i` | Enter Insert mode |
| `:` | Open Command bar |
| `h` / `j` / `k` / `l` | Navigate cursor (left/down/up/right) |
| `esc` | Return to Normal mode |

### ✏️ Editor — Insert Mode

| Key | Action |
|-----|--------|
| `esc` | Return to Normal mode |
| Any key | Type into the buffer |

### ✏️ Editor — Command Mode

| Command | Action |
|---------|--------|
| `:w` + `enter` | Write buffer to disk |
| `:q` + `enter` | Quit editor, return to file tree |
| `:wq` + `enter` | Write and quit |
| `esc` | Cancel command |

### 💻 Terminal Pane

| Key | Action |
|-----|--------|
| Any key | Sent directly to shell |
| `ctrl+d` | EOF / exit shell |
| `ctrl+c` | Interrupt running process |
| `tab` | (focus must not be in terminal) Switch pane |

> **Tip:** Press `tab` to move focus *away* from a terminal pane before using other workspace commands.

---

## Architecture

```
vide1/
├── main.go           # Bubble Tea model, tiling WM (wmModel), pane navigator (paneModel)
├── load_file.go      # Async file loader → Chroma syntax highlighter → tea.Cmd
├── terminal_pane.go  # PTY-backed TerminalPane, goroutine read loop, key routing
├── theme.go          # TrueColor palette, Nerd Font icons, per-ext styles, status bar
├── go.mod
└── go.sum
```

### Key design decisions

- **Binary tree layout** — each pane is a leaf node. Splitting morphs a leaf into an internal node with two children. `renderNode` recurses to distribute width/height.
- **Bubble Tea command model** — all async work (file reads, PTY reads) returns `tea.Cmd` and communicates via typed messages (`fileLoadedMsg`, `termOutputMsg`). No shared state between goroutines and the main loop.
- **Single binary, zero runtime deps** — no config daemon, no server, no LSP. Just the binary.

---

## Dependencies

| Package | Purpose |
|---------|---------|
| [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) | Elm-architecture TUI framework |
| [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) | Layout, borders, TrueColor styling |
| [`charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles) | `textarea` component for the inline editor |
| [`alecthomas/chroma/v2`](https://github.com/alecthomas/chroma) | Syntax highlighting for 300+ languages |
| [`creack/pty`](https://github.com/creack/pty) | PTY creation for the embedded terminal |

---

## Roadmap

- [ ] Fuzzy file finder (`/` to search)
- [ ] Git status indicators (modified, staged, untracked)
- [ ] Bookmarks / jump list
- [ ] Config file (`~/.config/vide1/config.toml`) for custom keybindings and themes
- [ ] File operations (rename, delete, copy, move) with confirmation prompts
- [ ] Mouse support
- [ ] Remote filesystem support (SFTP/SSH)

---

## Contributing

Pull requests are welcome! Please open an issue first to discuss major changes.

```bash
# Run the dev build
go run .

# Lint
go vet ./...

# Format
gofmt -w .
```

---

## License

VIDE1 is free software, released under the **GNU General Public License v3.0**.

```
Copyright (C) 2026 sidx1-scratch

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
```

See the full license text at <https://www.gnu.org/licenses/gpl-3.0.html>.
