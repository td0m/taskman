package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	a := newApp()
	p := tea.NewProgram(a)

	// enable full terminal mode
	p.EnterAltScreen()
	defer p.ExitAltScreen()

	// enable mouse (for scrolling)
	p.EnableMouseAllMotion()
	defer p.DisableMouseAllMotion()

	check(p.Start())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
