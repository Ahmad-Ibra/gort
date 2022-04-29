package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type lsof struct {
	command        string
	processID      int
	user           string
	fileDescriptor string
	node           string
	name           string
}

type portModel struct {
	cursor   int          // which item the cursor is pointing at
	lines    []lsof       // list of items
	selected map[int]lsof // which items are selected

	err error
}

func checkLsof() []lsof {
	var lines []lsof

	// TODO: get the actual lsof response and append them to lines
	testLsof := lsof{
		command:        "testCommand",
		processID:      12,
		user:           "testUser",
		fileDescriptor: "testFD",
		node:           "testNode",
		name:           "testName",
	}

	lines = append(lines, testLsof)
	lines = append(lines, testLsof)
	lines = append(lines, testLsof)
	lines = append(lines, testLsof)
	lines = append(lines, testLsof)
	lines = append(lines, testLsof)
	return lines
}

type lsofMsg []lsof

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m portModel) Init() tea.Cmd {
	//return checkLsof
	return nil
}

func (m portModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case lsofMsg:
		// the server returned lines of lsof response, Save to the model
		m.lines = msg
		return m, nil

	case errMsg:
		m.err = msg
		return m, tea.Quit

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
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = m.lines[m.cursor]
			}

		case "enter":
			// TODO figure out how to now kill all the things that are selected
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m portModel) View() string {
	// If there's an error, print it out and don't do anything else.
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	s := ""

	if len(m.lines) <= 0 {
		// Tell the user we're doing something.
		s = fmt.Sprintf("Checking ports ... \n")
	} else {
		// The header
		s = "Select port to kill with space bar. Once done hit enter to kill them:\n\n"

		// make request for ports
		// Iterate over our choices
		for i, choice := range m.lines {

			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}

			// Is this choice selected?
			checked := " " // not selected
			if _, ok := m.selected[i]; ok {
				checked = "x" // selected!
			}

			// Render the row
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.name)
		}

		// The footer
		s += "\nPress q to quit and b to go back.\n"
	}

	// Send the UI for rendering
	return s
}

func portInitialModel() portModel {
	curLines := checkLsof()
	return portModel{
		lines:    curLines,
		selected: make(map[int]lsof),
	}
}
