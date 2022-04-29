package main

import (
	"fmt"
	"os"

	sc "github.com/Ahmad-Ibra/gort/cmd/gort/screens"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := tea.NewProgram(sc.CreateModel()).Start(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
