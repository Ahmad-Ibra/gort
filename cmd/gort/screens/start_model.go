package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	startListProcs = "List all processes running on a port"
	startOther1    = "Some other action"
	startOther2    = "And another action"
)

type startModel struct {
	choices  []string         // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
}

func StartInitialModel() startModel {
	return startModel{
		choices: []string{startListProcs, startOther1, startOther2},
	}
}

func (m startModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m startModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			switch m.cursor {
			case 0:
				return portInitialModel(), nil
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m startModel) View() string {
	// The header
	s := "What would you like to do?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
