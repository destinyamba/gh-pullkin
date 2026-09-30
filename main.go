package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/destinyamba/gh-pullkin/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && args[0] == "nudge" {
		return nudge()
	}
	_, err := tea.NewProgram(tui.New()).Run()
	return err
}

func nudge() error {
	fmt.Println("pullkin: nothing in progress")
	return nil
}
