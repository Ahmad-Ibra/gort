package screens

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type lsof struct {
	command        string
	processID      string
	user           string
	fileDescriptor string
	node           string
	name           string
	conType        string
}

type portModel struct {
	cursor   int          // which item the cursor is pointing at
	lines    []lsof       // list of items
	selected map[int]lsof // which items are selected

	err error
}

func checkLsof() ([]lsof, error) {

	cmd := exec.Command("lsof", "-i", "-P")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var lines []lsof
	splitStr := strings.Split(string(stdout), "\n")
	for _, line := range splitStr {
		if !strings.Contains(line, "(LISTEN)") && !strings.Contains(line, "(ESTABLISHED)") {
			continue
		}

		parts := strings.Fields(line)
		lines = append(lines, lsof{
			command:        parts[0],
			processID:      parts[1],
			user:           parts[2],
			fileDescriptor: parts[3],
			node:           parts[7],
			name:           parts[8],
			conType:        parts[9],
		})
	}

	return lines, nil
}

func (m portModel) Init() tea.Cmd {
	//return checkLsof
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

		// The space bar (a literal space) toggle
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
			s += fmt.Sprintf("%s [%s] %s %s, %s %s %s %s %s\n",
				cursor, checked, choice.command, choice.processID, choice.user, choice.fileDescriptor, choice.node, choice.name, choice.conType)
		}

		// The footer
		s += "\nPress q to quit and b to go back.\n"
	}

	// Send the UI for rendering
	return s
}

func portInitialModel() portModel {
	curLines, err := checkLsof()
	return portModel{
		lines:    curLines,
		selected: make(map[int]lsof),
		err:      err,
	}
}
