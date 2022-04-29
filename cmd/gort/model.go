package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type process struct {
	command string
	pid     string
	user    string
	fd      string
	ipType  string
	device  string
	sizeOff string
	node    string
	name    string
	conType string
}

type model struct {
	cursor   int             // which item the cursor is pointing at
	lines    []process       // list of running processes that are running
	selected map[int]process // which items are selected, used to figure out which processes to kill
	killed   []process       // list of processes that have been killed

	err error
}

var sliceMutex = sync.RWMutex{}
var mapMutex = sync.RWMutex{}

func createModel() model {
	return model{
		selected: make(map[int]process),
	}
}

func checkProcesses() tea.Msg {
	cmd := exec.Command("lsof", "-i", "-P")
	stdout, err := cmd.Output()
	if err != nil {
		return errMsg{err}
	}

	var lines []process
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
		lines = append(lines, process{
			command: parts[0],
			pid:     parts[1],
			user:    parts[2],
			fd:      parts[3],
			ipType:  parts[4],
			device:  parts[5],
			sizeOff: parts[6],
			node:    parts[7],
			name:    parts[8],
			conType: connectionType,
		})
	}

	return processMsg(lines)
}

func scheduleKill(processes map[int]process) tea.Cmd {
	// key is the PID of a process to killed
	toKill := make(map[string]process)
	return func() tea.Msg {
		for _, p := range processes {
			mapMutex.Lock()

			if _, ok := toKill[p.pid]; ok {
				mapMutex.Unlock()
				continue
			}
			toKill[p.pid] = p
			mapMutex.Unlock()
		}

		errs := make(chan error, 1)
		done := make(chan []process, 1)

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

func kill(processes map[string]process, done chan<- []process, errs chan<- error) {
	var wg sync.WaitGroup
	kProcs := make([]process, 0)
	for _, process := range processes {
		wg.Add(1)
		p := process
		go func() {
			defer wg.Done()

			cmd := exec.Command("kill", "-TERM", p.pid)
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

type killedMsg []process

type processMsg []process

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// ---- model functions ----

func (m model) Init() tea.Cmd {
	return checkProcesses
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case processMsg:
		m.lines = msg
		return m, nil

	case killedMsg:
		m.killed = msg
		return m, nil

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		// Quit the app if we've already killed anything and any key is pressed
		if len(m.killed) != 0 {
			return m, tea.Quit
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.lines)-1 {
				m.cursor++
			}

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
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble:\n%v\n\n", m.err)
	}
	if len(m.killed) == 0 {
		return renderSelectView(m)
	}
	return renderResultView(m)
}
