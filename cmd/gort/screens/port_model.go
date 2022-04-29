package screens

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Process struct {
	command        string
	processID      string
	user           string
	fileDescriptor string
	node           string
	name           string
	conType        string
}

type Model struct {
	cursor   int             // which item the cursor is pointing at
	lines    []Process       // list of running processes that are running
	selected map[int]Process // which items are selected, used to figure out which processes to kill
	killed   map[int]Process // which items have actually been killed

	err error
}

func CreateModel() Model {
	return Model{
		selected: make(map[int]Process),
		killed:   make(map[int]Process),
	}
}

func checkProcesses() tea.Msg {
	cmd := exec.Command("lsof", "-i", "-P")
	stdout, err := cmd.Output()
	if err != nil {
		return errMsg{err}
	}

	var lines []Process
	splitStr := strings.Split(string(stdout), "\n")
	for _, line := range splitStr {
		if !strings.Contains(line, "(LISTEN)") && !strings.Contains(line, "(ESTABLISHED)") {
			continue
		}
		parts := strings.Fields(line)
		lines = append(lines, Process{
			command:        parts[0],
			processID:      parts[1],
			user:           parts[2],
			fileDescriptor: parts[3],
			node:           parts[7],
			name:           parts[8],
			conType:        parts[9],
		})
	}

	return processMsg(lines)
}

func scheduleKill(processes map[int]Process) tea.Cmd {

	killed := make(map[int]Process)
	return func() tea.Msg {

		var wg sync.WaitGroup
		for i, p := range processes {
			curProc := p
			curI := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				//fmt.Printf("\nkilling process %v with id %v\n", curProc.processID, curI)
				// TODO: actually kill it and dont just fake it

				killed[curI] = curProc
			}()
		}
		wg.Wait()
		return killedMsg(killed)
	}
}

type killedMsg map[int]Process

type processMsg []Process

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// ---- Model functions ----

func (m Model) Init() tea.Cmd {
	return checkProcesses
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case processMsg:
		m.lines = msg
		return m, nil
	case killedMsg:
		for k, v := range msg {
			m.killed[k] = v
		}
		m.killed = msg
		return m, nil
	case errMsg:
		// There was an error. Note it in the model. And tell the runtime
		// we're done and want to quit.
		m.err = msg
		return m, tea.Quit
	case tea.KeyMsg:
		// Quit the app if we've already killed anything and any key is pressed
		if len(m.killed) != 0 {
			return m, tea.Quit
		}
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
			return m, scheduleKill(m.selected)
		}
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	// If there's an error, print it out and don't do anything else.
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	s := ""
	if len(m.killed) == 0 {
		// Tell the user we're doing something.
		s = fmt.Sprintf("Checking ports, this could take a moment... \n")

		if len(m.lines) > 0 {
			// The header
			s = "Select process to kill with space bar. Once done hit enter to kill them:\n\n"

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
			s += "\nPress q to quit.\n"
		}
	} else {
		// we have now killed processes

		// The header
		s = "The following processes have been killed:\n\n"

		for _, p := range m.killed {
			// Render the row
			s += fmt.Sprintf("%s %s, %s %s %s %s %s\n",
				p.command, p.processID, p.user, p.fileDescriptor, p.node, p.name, p.conType)
		}
		// The footer
		s += "\nPress any key to quit.\n"
	}

	// Send the UI for rendering
	return s
}
