with open("main.go", "r") as f:
    content = f.read()

target = """		case ModeEditorCommand:
			switch msg.String() {
			case "esc":"""

if "default:" not in content.split("func (m *paneModel) Update")[1].split("func (m *paneModel) View")[0]:
    # Need to add default handler
    # Find the end of switch msg := msg.(type)
    
    # Actually, let's just do a string replacement
    # We replace the end of the Update method:
    pass
