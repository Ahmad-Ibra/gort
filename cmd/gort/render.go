package main

import (
	"fmt"
)

func renderResultView(m model) string {
	s := "The following processes have been killed:\n\n"
	s += fmt.Sprintf("%10s %7s\n",
		"COMMAND", "PID")

	// render rows
	for _, p := range m.killed {
		s += fmt.Sprintf("%10s %7s\n", p.command, p.pid)
	}

	s += "\nPress any key to quit.\n"
	return s
}

func renderSelectView(m model) string {
	s := fmt.Sprintf("Checking ports, this could take a moment... \n")

	if len(m.lines) > 0 {
		s = "Select process to kill with space bar. Once done hit enter to kill them:\n\n"
		s += fmt.Sprintf("%5s %10s %7s %10s %4s %5s %18s %8s %4s %14s %s\n",
			"", "COMMAND", "PID", "USER", "FD", "TYPE", "DEVICE", "SIZE/OFF", "NODE", "", "NAME")

		// render rows
		for i, p := range m.lines {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			checked := " "
			if _, ok := m.selected[i]; ok {
				checked = "x"
			}
			s += fmt.Sprintf("%s [%s] %10s %7s %10s %4s %5s %18s %8s %4s %14s %s\n",
				cursor, checked, p.command, p.pid, p.user, p.fd, p.ipType, p.device, p.sizeOff, p.node, p.conType, p.name)
		}
		s += "\nPress q to quit.\n"
	}
	return s
}
