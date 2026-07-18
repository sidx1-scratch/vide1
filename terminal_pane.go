// Copyright (C) 2026 sidx1-scratch
// VIDE1 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type termOutputMsg struct {
	id   int
	data string
}

type termClosedMsg struct {
	id int
}

// ── TerminalPane ──────────────────────────────────────────────────────────────

const maxBufLines = 2000

type TerminalPane struct {
	id     int
	ptmx   *os.File
	cmd    *exec.Cmd
	buf    []string
	mu     sync.Mutex
	width  int
	height int
	alive  bool
}

func newTerminalPane(id, width, height int) (*TerminalPane, tea.Cmd) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		tp := &TerminalPane{id: id, width: width, height: height, alive: false}
		tp.buf = []string{"[error] failed to start shell: " + err.Error()}
		return tp, nil
	}

	// Set initial PTY size
	setTermSize(ptmx, width, height)

	tp := &TerminalPane{
		id:     id,
		ptmx:   ptmx,
		cmd:    cmd,
		width:  width,
		height: height,
		alive:  true,
	}

	return tp, tp.readLoop()
}

func setTermSize(f *os.File, w, h int) {
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	ws := &struct{ Row, Col, Xpixel, Ypixel uint16 }{uint16(h), uint16(w), 0, 0}
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

// readLoop returns a tea.Cmd that continuously reads PTY output and sends termOutputMsg
func (t *TerminalPane) readLoop() tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		for {
			n, err := t.ptmx.Read(buf)
			if n > 0 {
				return termOutputMsg{id: t.id, data: string(buf[:n])}
			}
			if err != nil {
				if err == io.EOF {
					return termClosedMsg{id: t.id}
				}
				return termClosedMsg{id: t.id}
			}
		}
	}
}

// continueRead is called after each termOutputMsg to keep the loop going
func (t *TerminalPane) continueRead() tea.Cmd {
	if !t.alive {
		return nil
	}
	return t.readLoop()
}

func (t *TerminalPane) appendLines(data string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Split on newlines but keep partial lines by appending to last entry
	parts := strings.Split(data, "\n")
	if len(t.buf) == 0 {
		t.buf = append(t.buf, "")
	}
	t.buf[len(t.buf)-1] += parts[0]
	for _, p := range parts[1:] {
		t.buf = append(t.buf, p)
	}
	// Cap buffer
	if len(t.buf) > maxBufLines {
		t.buf = t.buf[len(t.buf)-maxBufLines:]
	}
}

func (t *TerminalPane) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case termOutputMsg:
		if msg.id != t.id {
			return nil
		}
		t.appendLines(msg.data)
		return t.continueRead()

	case termClosedMsg:
		if msg.id == t.id {
			t.alive = false
		}

	case tea.KeyMsg:
		if !t.alive || t.ptmx == nil {
			return nil
		}
		var seq string
		switch msg.String() {
		case "enter":
			seq = "\r"
		case "backspace":
			seq = "\x7f"
		case "tab":
			seq = "\t"
		case "esc":
			seq = "\x1b"
		case "ctrl+c":
			seq = "\x03"
		case "ctrl+d":
			seq = "\x04"
		case "ctrl+z":
			seq = "\x1a"
		case "ctrl+l":
			seq = "\x0c"
		case "up":
			seq = "\x1b[A"
		case "down":
			seq = "\x1b[B"
		case "right":
			seq = "\x1b[C"
		case "left":
			seq = "\x1b[D"
		case "home":
			seq = "\x1b[H"
		case "end":
			seq = "\x1b[F"
		case "pgup":
			seq = "\x1b[5~"
		case "pgdown":
			seq = "\x1b[6~"
		case "delete":
			seq = "\x1b[3~"
		default:
			if msg.Type == tea.KeyRunes {
				seq = string(msg.Runes)
			} else if msg.Type == tea.KeySpace {
				seq = " "
			}
		}
		if seq != "" {
			t.ptmx.WriteString(seq)
		}
	}
	return nil
}

func (t *TerminalPane) Resize(w, h int) {
	t.width = w
	t.height = h
	if t.ptmx != nil && t.alive {
		setTermSize(t.ptmx, w, h)
	}
}

func (t *TerminalPane) View(isActive bool) string {
	w := t.width
	h := t.height
	if w <= 0 || h <= 0 {
		return ""
	}

	availW := w - 2
	availH := h - 2
	if availW <= 0 || availH <= 0 {
		return ""
	}

	// Header
	header := lipgloss.NewStyle().
		Background(colorGreen).
		Foreground(colorDarkFg).
		Bold(true).
		Width(availW).
		Render("  TERMINAL  " + fmt.Sprintf("(id:%d)", t.id))

	// Footer
	hint := " ctrl+t: new terminal  |  tab: switch pane "
	if !t.alive {
		hint = " [shell exited — tab: switch pane] "
	}
	footer := lipgloss.NewStyle().
		Background(colorBorder).
		Foreground(colorFg).
		Width(availW).
		Render(hint)

	// Content area
	contentH := availH - 2 // header + footer
	if contentH < 1 {
		contentH = 1
	}

	t.mu.Lock()
	buf := t.buf
	t.mu.Unlock()

	start := 0
	if len(buf) > contentH {
		start = len(buf) - contentH
	}
	visible := buf[start:]

	lineStyle := lipgloss.NewStyle().Width(availW)
	var lines []string
	for _, l := range visible {
		lines = append(lines, lineStyle.Render(l))
	}
	// pad to fill height
	for len(lines) < contentH {
		lines = append(lines, lineStyle.Render(""))
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(lines, "\n"),
		footer,
	)

	border := gradientBorder(isActive).Width(availW).Height(availH)
	return border.Render(content)
}
