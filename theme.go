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
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ── Palette ──────────────────────────────────────────────────────────────────
var (
	colorBg     = lipgloss.Color("#282C34")
	colorBorder = lipgloss.Color("#3E4452")
	colorAccent = lipgloss.Color("#C678DD") // magenta – active border
	colorBlue   = lipgloss.Color("#61AFEF")
	colorCyan   = lipgloss.Color("#56B6C2")
	colorGreen  = lipgloss.Color("#98C379")
	colorYellow = lipgloss.Color("#E5C07B")
	colorOrange = lipgloss.Color("#D19A66")
	colorRed    = lipgloss.Color("#E06C75")
	colorGray   = lipgloss.Color("#5C6370")
	colorFg     = lipgloss.Color("#ABB2BF")
	colorDarkFg = lipgloss.Color("#1E222A")
)

// ── Nerd Font Icons ───────────────────────────────────────────────────────────

func getFileIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.ToLower(name)

	switch ext {
	case ".go":
		return "🐹"
	case ".py":
		return "🐍"
	case ".js", ".mjs", ".cjs":
		return "🟨"
	case ".ts", ".tsx":
		return "🟦"
	case ".jsx":
		return "⚛️"
	case ".html", ".htm":
		return "🌐"
	case ".css", ".scss", ".sass":
		return "🎨"
	case ".json", ".jsonc":
		return "⚙️"
	case ".yaml", ".yml":
		return "📝"
	case ".toml":
		return "🔧"
	case ".sh", ".bash", ".zsh", ".fish":
		return "🐚"
	case ".md", ".mdx":
		return "📖"
	case ".txt":
		return "📄"
	case ".rs":
		return "🦀"
	case ".c", ".h":
		return "🇨"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "➕"
	case ".java":
		return "☕"
	case ".rb":
		return "💎"
	case ".php":
		return "🐘"
	case ".lua":
		return "🌙"
	case ".vim":
		return "💚"
	case ".git":
		return "🌿"
	case ".dockerfile", ".containerfile":
		return "🐳"
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico":
		return "🖼️"
	case ".mp4", ".mkv", ".avi", ".mov":
		return "🎬"
	case ".mp3", ".flac", ".wav", ".ogg":
		return "🎵"
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar":
		return "📦"
	case ".pdf":
		return "📕"
	}

	switch base {
	case "makefile", "gnumakefile":
		return "🛠️"
	case "dockerfile":
		return "🐳"
	case "license", "licence":
		return "📜"
	case ".gitignore", ".gitmodules", ".gitattributes":
		return "🙈"
	case "readme.md", "readme.txt", "readme":
		return "📖"
	}

	return "📄"
}

// ── Per-extension Lip Gloss Styles ───────────────────────────────────────────

func getFileStyle(name string, isDir bool) lipgloss.Style {
	if isDir {
		return lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return lipgloss.NewStyle().Foreground(colorBlue)
	case ".py":
		return lipgloss.NewStyle().Foreground(colorYellow)
	case ".js", ".mjs", ".cjs", ".jsx":
		return lipgloss.NewStyle().Foreground(colorGreen)
	case ".ts", ".tsx":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC1FF"))
	case ".html", ".htm":
		return lipgloss.NewStyle().Foreground(colorRed)
	case ".css", ".scss", ".sass":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75A0"))
	case ".json", ".jsonc", ".toml", ".yaml", ".yml":
		return lipgloss.NewStyle().Foreground(colorOrange)
	case ".sh", ".bash", ".zsh", ".fish":
		return lipgloss.NewStyle().Foreground(colorGreen)
	case ".md", ".mdx", ".txt":
		return lipgloss.NewStyle().Foreground(colorFg)
	case ".rs":
		return lipgloss.NewStyle().Foreground(colorOrange)
	case ".c", ".h", ".cpp", ".cc", ".cxx", ".hpp":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6"))
	}
	return lipgloss.NewStyle().Foreground(colorGray)
}

// ── Border Styles ─────────────────────────────────────────────────────────────

func gradientBorder(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder)
}

// ── Status Bar ────────────────────────────────────────────────────────────────

func renderStatusBar(path, mode string, width int) string {
	if width <= 0 {
		return ""
	}

	modeBg := colorBlue
	switch mode {
	case "INSERT":
		modeBg = colorGreen
	case "COMMAND":
		modeBg = colorOrange
	case "TERMINAL":
		modeBg = colorCyan
	case "NORMAL":
		modeBg = colorAccent
	}

	modeLabel := lipgloss.NewStyle().
		Background(modeBg).
		Foreground(colorDarkFg).
		Bold(true).
		Padding(0, 1).
		Render(" " + mode + " ")

	modeLen := lipgloss.Width(modeLabel)

	pathAvail := width - modeLen - 2
	if pathAvail < 0 {
		pathAvail = 0
	}
	path = ansi.TruncateLeft(path, pathAvail, "…")

	pathLabel := lipgloss.NewStyle().
		Background(colorBg).
		Foreground(colorFg).
		Width(pathAvail).
		Padding(0, 1).
		Render(path)

	return lipgloss.JoinHorizontal(lipgloss.Top, modeLabel, pathLabel)
}
