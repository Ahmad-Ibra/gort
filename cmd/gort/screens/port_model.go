package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type lsof struct {
	command string
	processID int
	user string
	fileDescriptor string
	node string
	name string
}

type portModel struct {
	cursor   int              // which to-do list item our cursor is pointing at
	lines map[int]lsof // which to-do items are selected
}

func portInitialModel() portModel {
	return portModel{
		lines:  make(map[int]lsof),
	}
}

func (m portModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."

	return nil
}

func (m portModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "b":
			return StartInitialModel(), nil

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.lines)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			// TODO: figure out how to kill the process
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m portModel) View() string {
	// The header
	s := "Here's the open ports:\n\n"

	// make request for ports
	// Iterate over our choices
	for i, line := range m.lines {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, line.name)
	}

	// The footer
	s += "\nPress q to quit and b to go back.\n"

	// Send the UI for rendering
	return s
}
