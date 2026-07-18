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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	tea "github.com/charmbracelet/bubbletea"
)

type fileLoadedMsg struct {
	id      int
	path    string
	content string
}

func loadFileCmd(id int, path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return fileLoadedMsg{id: id, path: path, content: "Error reading file: " + err.Error()}
		}

		// Truncate to avoid massive memory usage or slow down
		if len(content) > 1024*1024 {
			content = content[:1024*1024]
		}

		var buf bytes.Buffer
		// We use Terminal 256 colors
		err = quick.Highlight(&buf, string(content), "", "terminal256", "monokai")
		if err != nil {
			// fallback
			buf.WriteString(string(content))
		}

		lines := strings.Split(buf.String(), "\n")

		// Let's add line numbers
		var out strings.Builder
		for i, line := range lines {
			out.WriteString(fmt.Sprintf("%4d │ %s\n", i+1, line))
		}

		return fileLoadedMsg{id: id, path: path, content: out.String()}
	}
}
