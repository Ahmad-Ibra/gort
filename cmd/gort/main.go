package main

import (
	"fmt"
	"os"

	sc "github.com/Ahmad-Ibra/gort/cmd/gort/screens"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(sc.StartInitialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
