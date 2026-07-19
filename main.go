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
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	// columnStyle: thin right-border separator between miller columns
	columnStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(colorBorder)

	// titleStyle: directory path titles inside columns
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			MarginBottom(1)

	// focusedStyle: high-contrast inverted block selection for current file/folder
	focusedStyle = lipgloss.NewStyle().
			Foreground(colorDarkFg).
			Background(colorAccent).
			Bold(true)

	// normalStyle: non-highlighted rows
	normalStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	// paneActiveBorder / paneInactiveBorder used by TerminalPane.View
	paneActiveBorder   = gradientBorder(true)
	paneInactiveBorder = gradientBorder(false)
)

type PaneMode int

const (
	ModeExplorer PaneMode = iota
	ModeEditorNormal
	ModeEditorInsert
	ModeEditorCommand
)

type paneModel struct {
	id           int
	currentPath  string
	parentItems  []os.DirEntry
	currentItems []os.DirEntry
	childItems   []os.DirEntry

	previewContent string
	previewPath    string
	previewLoading bool

	cursor int
	scroll int
	width  int
	height int

	mode     PaneMode
	editor   textarea.Model
	editPath string
	cmdInput string
}

func initialPaneModel(id int) *paneModel {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}
	cwd = filepath.Clean(cwd)

	ti := textarea.New()
	ti.ShowLineNumbers = true

	m := &paneModel{
		id:          id,
		currentPath: cwd,
		mode:        ModeExplorer,
		editor:      ti,
	}
	return m
}

func (m *paneModel) updateDirs() tea.Cmd {
	m.currentItems = readDir(m.currentPath)

	parentDir := filepath.Dir(m.currentPath)
	if parentDir == m.currentPath { // At root
		m.parentItems = nil
	} else {
		m.parentItems = readDir(parentDir)
	}

	return m.updateChildDir()
}

func (m *paneModel) updateChildDir() tea.Cmd {
	if len(m.currentItems) == 0 {
		m.childItems = nil
		m.cursor = 0
		m.scroll = 0
		return nil
	}
	if m.cursor >= len(m.currentItems) {
		m.cursor = len(m.currentItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureCursorVisible()

	selected := m.currentItems[m.cursor]
	if selected.IsDir() {
		childPath := filepath.Join(m.currentPath, selected.Name())
		m.childItems = readDir(childPath)
		return nil
	} else {
		m.childItems = nil
		m.previewPath = filepath.Join(m.currentPath, selected.Name())
		m.previewLoading = true
		return loadFileCmd(m.id, m.previewPath)
	}
}

func (m *paneModel) visibleExplorerRows() int {
	contentH := m.height - 3 // border plus one status line
	rows := contentH - 2     // title, spacer, and breathing room
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m *paneModel) ensureCursorVisible() {
	if len(m.currentItems) == 0 {
		m.scroll = 0
		return
	}

	rows := m.visibleExplorerRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+rows-1 {
		m.scroll = m.cursor - rows + 1
	}

	maxScroll := len(m.currentItems) - rows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}

func readDir(path string) []os.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	return entries
}

func (m *paneModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case fileLoadedMsg:
		if msg.id == m.id && m.previewPath == msg.path {
			m.previewContent = msg.content
			m.previewLoading = false
		}
	case tea.KeyMsg:
		switch m.mode {
		case ModeExplorer:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					cmd = m.updateChildDir()
				}

			case "down", "j":
				if m.cursor < len(m.currentItems)-1 {
					m.cursor++
					cmd = m.updateChildDir()
				}

			case "left", "h":
				parentDir := filepath.Dir(m.currentPath)
				if parentDir != m.currentPath {
					oldBase := filepath.Base(m.currentPath)
					m.currentPath = parentDir
					m.updateDirs()

					m.cursor = 0
					for i, item := range m.currentItems {
						if item.Name() == oldBase {
							m.cursor = i
							break
						}
					}
					cmd = m.updateChildDir()
				}

			case "right", "l", "enter", "e":
				if len(m.currentItems) > 0 {
					selected := m.currentItems[m.cursor]
					if selected.IsDir() {
						if msg.String() != "e" {
							m.currentPath = filepath.Join(m.currentPath, selected.Name())
							m.cursor = 0
							cmd = m.updateDirs()
						}
					} else {
						if msg.String() == "enter" || msg.String() == "e" || msg.String() == "l" || msg.String() == "right" {
							if msg.String() == "e" || msg.String() == "enter" {
								m.editPath = filepath.Join(m.currentPath, selected.Name())
								content, err := os.ReadFile(m.editPath)
								if err == nil {
									m.editor.SetValue(string(content))
								} else {
									m.editor.SetValue("")
								}
								m.mode = ModeEditorNormal
								m.editor.Blur()
							}
						}
					}
				}
			}

		case ModeEditorNormal:
			switch msg.String() {
			case "i":
				m.mode = ModeEditorInsert
				m.editor.Focus()
				m.editor.ShowLineNumbers = true
			case "a": // Append right
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyRight})
				m.mode = ModeEditorInsert
				m.editor.Focus()
			case "A": // Append at end of line
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyEnd})
				m.mode = ModeEditorInsert
				m.editor.Focus()
			case "I": // Insert at beginning of line
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyHome})
				m.mode = ModeEditorInsert
				m.editor.Focus()
			case "o": // Open line below
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyEnd})
				m.editor.InsertString("\n")
				m.mode = ModeEditorInsert
				m.editor.Focus()
			case "O": // Open line above
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyHome})
				m.editor.InsertString("\n")
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyUp})
				m.mode = ModeEditorInsert
				m.editor.Focus()
			case "x": // Delete char under cursor
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyDelete})
			case "0": // Start of line
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyHome})
			case "$": // End of line
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyEnd})
			case ":":
				m.mode = ModeEditorCommand
				m.cmdInput = ":"
			case "h":
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyLeft})
			case "j":
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyDown})
			case "k":
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyUp})
			case "l":
				m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: tea.KeyRight})
			}

		case ModeEditorInsert:
			switch msg.String() {
			case "esc":
				m.mode = ModeEditorNormal
				m.editor.Blur()
			default:
				m.editor, cmd = m.editor.Update(msg)
			}

		case ModeEditorCommand:
			switch msg.String() {
			case "esc":
				m.mode = ModeEditorNormal
				m.cmdInput = ""
			case "enter":
				if m.cmdInput == ":w" {
					_ = os.WriteFile(m.editPath, []byte(m.editor.Value()), 0644)
				} else if m.cmdInput == ":q" {
					m.mode = ModeExplorer
					// Reset the preview content so it reflects any changes
					cmd = m.updateChildDir()
				} else if m.cmdInput == ":wq" {
					_ = os.WriteFile(m.editPath, []byte(m.editor.Value()), 0644)
					m.mode = ModeExplorer
					cmd = m.updateChildDir()
				}
				if m.mode == ModeEditorCommand {
					m.mode = ModeEditorNormal
				}
				m.cmdInput = ""
			case "backspace", "delete":
				if len(m.cmdInput) > 0 {
					m.cmdInput = m.cmdInput[:len(m.cmdInput)-1]
				}
				if len(m.cmdInput) == 0 {
					m.mode = ModeEditorNormal
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.cmdInput += string(msg.Runes)
				}
			}
		}
	}
	return cmd
}

func (m *paneModel) modeString() string {
	switch m.mode {
	case ModeEditorNormal:
		return "NORMAL"
	case ModeEditorInsert:
		return "INSERT"
	case ModeEditorCommand:
		return "COMMAND"
	default:
		return "EXPLORER"
	}
}

func (m *paneModel) View(isActive bool) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	availWidth := m.width - 2
	availHeight := m.height - 2

	if availWidth <= 0 || availHeight <= 0 {
		return ""
	}

	statusBar := renderStatusBar(m.currentPath, m.modeString(), availWidth)
	contentH := availHeight - 1 // reserve one line for status bar
	if contentH < 1 {
		contentH = 1
	}

	if m.mode != ModeExplorer {
		editorHeight := contentH - 1 // extra line for file path header
		if editorHeight < 1 {
			editorHeight = 1
		}
		m.editor.SetWidth(availWidth)
		m.editor.SetHeight(editorHeight)

		header := lipgloss.NewStyle().
			Background(colorBorder).
			Foreground(colorFg).
			Width(availWidth).
			Render(ansi.Truncate(" "+getFileIcon(m.editPath, false)+" "+m.editPath, availWidth, "…"))

		content := lipgloss.JoinVertical(lipgloss.Left, header, m.editor.View(), statusBar)
		if isActive {
			return gradientBorder(true).Width(availWidth).Height(availHeight).Render(content)
		}
		return gradientBorder(false).Width(availWidth).Height(availHeight).Render(content)
	}

	// Adjusting column width to completely account for layout mathematics padding/borders.
	// 3 columns * 3 cells structural tax = 9 cells total.
	colWidth := (availWidth - 9) / 3
	if colWidth <= 0 {
		colWidth = 1
	}

	m.ensureCursorVisible()
	parentCol := renderColumn(filepath.Dir(m.currentPath), m.parentItems, -1, 0, colWidth, contentH)
	currentCol := renderColumn(m.currentPath, m.currentItems, m.cursor, m.scroll, colWidth, contentH)

	var childTitle string
	var childContent string

	if len(m.currentItems) > 0 {
		selected := m.currentItems[m.cursor]
		if selected.IsDir() {
			childTitle = filepath.Join(m.currentPath, selected.Name())
			childContent = renderColumn(childTitle, m.childItems, -1, 0, colWidth, contentH)
		} else {
			childTitle = "Preview: " + selected.Name()
			if m.previewLoading && m.previewPath == filepath.Join(m.currentPath, selected.Name()) {
				childContent = titleStyle.Render(ansi.Truncate(childTitle, colWidth, "…")) + "\n\nLoading..."
			} else if m.previewPath == filepath.Join(m.currentPath, selected.Name()) {
				lines := strings.Split(m.previewContent, "\n")
				displayHeight := contentH - 2
				if displayHeight < 1 {
					displayHeight = 1
				}
				var dispLines []string
				for i := 0; i < len(lines) && i < displayHeight; i++ {
					dispLines = append(dispLines, ansi.Truncate(lines[i], colWidth, ""))
				}
				childContent = titleStyle.Render(ansi.Truncate(childTitle, colWidth, "…")) + "\n\n" + strings.Join(dispLines, "\n")
			} else {
				childContent = titleStyle.Render(ansi.Truncate(childTitle, colWidth, "…")) + "\n\n"
			}
		}
	} else {
		childTitle = "Empty"
		childContent = renderColumn(childTitle, m.childItems, -1, 0, colWidth, contentH)
	}

	cols := lipgloss.JoinHorizontal(
		lipgloss.Top,
		columnStyle.Width(colWidth).Render(parentCol),
		columnStyle.Width(colWidth).Render(currentCol),
		columnStyle.Width(colWidth).Render(childContent),
	)
	content := lipgloss.JoinVertical(lipgloss.Left, cols, statusBar)

	if isActive {
		return gradientBorder(true).Width(availWidth).Height(availHeight).Render(content)
	}
	return gradientBorder(false).Width(availWidth).Height(availHeight).Render(content)
}

func renderColumn(title string, items []os.DirEntry, cursor int, scroll int, width int, height int) string {
	var s strings.Builder

	titleWidth := width
	if titleWidth < 1 {
		titleWidth = 1
	}
	shortTitle := ansi.TruncateLeft(title, titleWidth, "…")
	s.WriteString(titleStyle.Render(shortTitle) + "\n\n")

	displayHeight := height - 2
	if displayHeight < 1 {
		displayHeight = 1
	}

	start := 0
	end := len(items)

	if cursor >= 0 {
		maxScroll := len(items) - displayHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if scroll < 0 {
			scroll = 0
		}
		if scroll > maxScroll {
			scroll = maxScroll
		}
		start = scroll
		end = start + displayHeight
		if end > len(items) {
			end = len(items)
		}
	} else {
		if end > displayHeight {
			end = displayHeight
		}
	}

	for i := start; i < end; i++ {
		item := items[i]
		name := item.Name()
		isDir := item.IsDir()

		icon := getFileIcon(name, isDir)
		fileStyle := getFileStyle(name, isDir)

		if isDir {
			name += "/"
		}
		maxNameLen := width - 6
		if maxNameLen < 1 {
			maxNameLen = 1
		}
		name = ansi.Truncate(name, maxNameLen, "…")

		if i == cursor {
			entry := fmt.Sprintf(" %s %s ", icon, name)
			s.WriteString(focusedStyle.Render(ansi.Truncate(entry, width, "")) + "\n")
		} else {
			entry := fmt.Sprintf("  %s %s", icon, name)
			s.WriteString(fileStyle.Render(ansi.Truncate(entry, width, "")) + "\n")
		}
	}

	return s.String()
}

type SplitDirection int

const (
	DirHorizontal SplitDirection = iota // side-by-side
	DirVertical                         // top-and-bottom
)

type Node struct {
	IsLeaf     bool
	IsTerminal bool
	Pane       *paneModel
	TermPane   *TerminalPane
	Dir        SplitDirection
	Child1     *Node
	Child2     *Node
	Parent     *Node
}

// Leaves returns all leaf nodes (file-browser panes and terminal panes)
func (n *Node) Leaves() []*Node {
	if n.IsLeaf {
		return []*Node{n}
	}
	var leaves []*Node
	if n.Child1 != nil {
		leaves = append(leaves, n.Child1.Leaves()...)
	}
	if n.Child2 != nil {
		leaves = append(leaves, n.Child2.Leaves()...)
	}
	return leaves
}

type wmModel struct {
	root       *Node
	activePane *Node
	width      int
	height     int
	paneIDSeq  int
}

func initialWMModel() wmModel {
	firstPane := initialPaneModel(1)
	root := &Node{
		IsLeaf: true,
		Pane:   firstPane,
	}
	return wmModel{
		root:       root,
		activePane: root,
		paneIDSeq:  1,
	}
}

func (m wmModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, leaf := range m.root.Leaves() {
		cmds = append(cmds, leaf.Pane.updateDirs())
	}
	cmds = append(cmds, textarea.Blink)
	return tea.Batch(cmds...)
}

func (m wmModel) activePaneIsEditor() bool {
	return m.activePane != nil &&
		!m.activePane.IsTerminal &&
		m.activePane.Pane != nil &&
		m.activePane.Pane.mode != ModeExplorer
}

func (m wmModel) activePaneIsTerminal() bool {
	return m.activePane != nil && m.activePane.IsTerminal && m.activePane.TermPane != nil
}

// splitActive splits the currently active leaf node, placing oldPane in Child1
// and newNode in Child2, using a direction based on pane geometry.
func (m *wmModel) splitActive(newNode *Node) SplitDirection {
	dir := DirHorizontal
	var aw, ah int
	if m.activePane.IsTerminal && m.activePane.TermPane != nil {
		aw = m.activePane.TermPane.width
		ah = m.activePane.TermPane.height
	} else if m.activePane.Pane != nil {
		aw = m.activePane.Pane.width
		ah = m.activePane.Pane.height
	}
	if aw <= ah*2 {
		dir = DirVertical
	}

	// Preserve existing content in child1
	child1 := &Node{
		IsLeaf:     true,
		IsTerminal: m.activePane.IsTerminal,
		Pane:       m.activePane.Pane,
		TermPane:   m.activePane.TermPane,
		Parent:     m.activePane,
	}
	newNode.Parent = m.activePane

	m.activePane.IsLeaf = false
m.activePane.IsTerminal = false
m.activePane.Dir = dir
m.activePane.Child1 = child1
m.activePane.Child2 = newNode
m.activePane.Pane = nil
m.activePane.TermPane = nil
	return dir
}

func (m wmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case fileLoadedMsg:
		for _, leaf := range m.root.Leaves() {
			if !leaf.IsTerminal && leaf.Pane != nil && leaf.Pane.id == msg.id {
				cmds = append(cmds, leaf.Pane.Update(msg))
			}
		}

	case termOutputMsg:
		for _, leaf := range m.root.Leaves() {
			if leaf.IsTerminal && leaf.TermPane != nil && leaf.TermPane.id == msg.id {
				cmds = append(cmds, leaf.TermPane.Update(msg))
			}
		}

	case termClosedMsg:
		for _, leaf := range m.root.Leaves() {
			if leaf.IsTerminal && leaf.TermPane != nil && leaf.TermPane.id == msg.id {
				leaf.TermPane.Update(msg)
			}
		}

	case tea.KeyMsg:
		isEditorActive := m.activePaneIsEditor()
		isTermActive := m.activePaneIsTerminal()

		// Always-global: ctrl+c quits (except when terminal is active, terminal gets it)
		if msg.String() == "ctrl+c" && !isTermActive {
			return m, tea.Quit
		}

		// q quits only in explorer mode
		if msg.String() == "q" && !isEditorActive && !isTermActive {
			return m, tea.Quit
		}

		// ctrl+t: spawn a new terminal pane (works in any mode)
		if msg.String() == "ctrl+t" {
			m.paneIDSeq++
			tp, readCmd := newTerminalPane(m.paneIDSeq, m.width/2, m.height/2)
			newNode := &Node{
				IsLeaf:     true,
				IsTerminal: true,
				TermPane:   tp,
			}
			m.splitActive(newNode)
			if readCmd != nil {
				cmds = append(cmds, readCmd)
			}
			return m, tea.Batch(cmds...)
		}

		// tab: cycle focus outside editor modes
		if msg.String() == "tab" && !isEditorActive {
			leaves := m.root.Leaves()
			for i, leaf := range leaves {
				if leaf == m.activePane {
					m.activePane = leaves[(i+1)%len(leaves)]
					break
				}
			}
			return m, nil
		}

		// ctrl+shift+tab / ctrl+w: split into new explorer pane
		if (msg.String() == "ctrl+shift+tab" || msg.String() == "ctrl+w") && !isEditorActive && !isTermActive {
			m.paneIDSeq++
			newPane := initialPaneModel(m.paneIDSeq)
			if m.activePane.Pane != nil {
				newPane.currentPath = m.activePane.Pane.currentPath
			}
			newNode := &Node{IsLeaf: true, Pane: newPane}
			m.splitActive(newNode)
			cmds = append(cmds, newPane.updateDirs())
			return m, tea.Batch(cmds...)
		}

		// Route to active terminal pane
		if isTermActive {
			cmds = append(cmds, m.activePane.TermPane.Update(msg))
			return m, tea.Batch(cmds...)
		}

		// Route to active file-browser/editor pane
		if m.activePane != nil && !m.activePane.IsTerminal && m.activePane.Pane != nil {
			if paneCmd := m.activePane.Pane.Update(msg); paneCmd != nil { cmds = append(cmds, paneCmd) }
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	default:
		// Blink and other broadcast messages go to all editor panes
		for _, leaf := range m.root.Leaves() {
			if !leaf.IsTerminal && leaf.Pane != nil {
				cmds = append(cmds, leaf.Pane.Update(msg))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m wmModel) View() string {
	if m.width == 0 {
		return "Initializing WM..."
	}
	return renderNode(m.root, m.width, m.height, m.activePane)
}

func renderNode(n *Node, w, h int, active *Node) string {
	if n.IsLeaf {
		if n.IsTerminal && n.TermPane != nil {
			n.TermPane.Resize(w, h)
			return n.TermPane.View(n == active)
		}
		if n.Pane != nil {
			n.Pane.width = w
			n.Pane.height = h
			return n.Pane.View(n == active)
		}
		return ""
	}

	if n.Dir == DirHorizontal {
		w1 := w / 2
		w2 := w - w1
		left := renderNode(n.Child1, w1, h, active)
		right := renderNode(n.Child2, w2, h, active)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	h1 := h / 2
	h2 := h - h1
	top := renderNode(n.Child1, w, h1, active)
	bottom := renderNode(n.Child2, w, h2, active)
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func main() {
	p := tea.NewProgram(initialWMModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
