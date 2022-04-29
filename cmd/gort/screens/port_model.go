package screens

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Process struct {
	command   string
	processID string
	user      string
	node      string
	name      string
	conType   string
}

type Model struct {
	cursor   int             // which item the cursor is pointing at
	lines    []Process       // list of running processes that are running
	selected map[int]Process // which items are selected, used to figure out which processes to kill
	killed   []Process       // list of processes that have been killed

	err error
}

var sliceMutex = sync.RWMutex{}
var mapMutex = sync.RWMutex{}

func CreateModel() Model {
	return Model{
		selected: make(map[int]Process),
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
	for i, line := range splitStr {
		if i == 0 || line == "" {
			continue
		}
		parts := strings.Fields(line)
		connectionType := ""
		if len(parts) > 9 {
			connectionType = parts[9]
		}
		lines = append(lines, Process{
			command:   parts[0],
			processID: parts[1],
			user:      parts[2],
			node:      parts[7],
			name:      parts[8],
			conType:   connectionType,
		})
	}

	return processMsg(lines)
}

func scheduleKill(processes map[int]Process) tea.Cmd {
	// key is the PID of a process to killed
	toKill := make(map[string]Process)
	return func() tea.Msg {
		for _, p := range processes {
			mapMutex.Lock()

			if _, ok := toKill[p.processID]; ok {
				mapMutex.Unlock()
				continue
			}
			toKill[p.processID] = p
			mapMutex.Unlock()
		}

		errs := make(chan error, 1)
		done := make(chan []Process, 1)

		// call a go routine which spawns other go routines and kills the process
		go kill(toKill, done, errs)

		select {
		case err := <-errs:
			return errMsg{err}
		case pList := <-done:
			return killedMsg(pList)
		}
	}
}

func kill(processes map[string]Process, done chan<- []Process, errs chan<- error) {
	var wg sync.WaitGroup
	kProcs := make([]Process, 0)
	for _, process := range processes {
		wg.Add(1)
		p := process
		go func() {
			defer wg.Done()

			cmd := exec.Command("kill", "-TERM", p.processID)
			_, err := cmd.Output()
			if err != nil {
				errs <- err
			}

			sliceMutex.Lock()
			kProcs = append(kProcs, p)
			sliceMutex.Unlock()
		}()
	}
	wg.Wait()
	done <- kProcs
}

type killedMsg []Process

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

			// Render the title
			s += fmt.Sprintf("%5s %10s %7s %10s %4s %14s %s\n",
				"", "COMMAND", "PID", "USER", "NODE", "", "NAME")

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
				s += fmt.Sprintf("%s [%s] %10s %7s %10s %4s %14s %s\n",
					cursor, checked, choice.command, choice.processID, choice.user, choice.node, choice.conType, choice.name)
			}
			// The footer
			s += "\nPress q to quit.\n"
		}
	} else {
		// we have now killed processes

		// The header
		s = "The following processes have been killed:\n\n"

		// Render the row
		s += fmt.Sprintf("%10s %7s\n",
			"COMMAND", "PID")

		for _, p := range m.killed {
			// Render the row
			s += fmt.Sprintf("%10s %7s\n",
				p.command, p.processID)
		}
		// The footer
		s += "\nPress any key to quit.\n"
	}

	// Send the UI for rendering
	return s
}
