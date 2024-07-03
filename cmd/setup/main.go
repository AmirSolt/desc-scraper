package main

import (
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string // items on the to-do list
	cursor   int      // which to-do list item our cursor is pointing at
	funcs    map[string]func()
	selected map[int]string // which to-do items are selected
}

func initialModel() model {

	choices := []string{"*DANGER* Reset DB", "Push To DB", "Wipe reports"}

	funcs := map[string]func(){
		choices[0]: resetDb,
		choices[1]: pushToDB,
		choices[2]: wipeReports,
	}

	return model{
		// Our to-do list is a grocery list
		choices:  choices,
		funcs:    funcs,
		selected: make(map[int]string),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q", "enter":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = m.choices[m.cursor]
			}

			// return m, m.funcs[]
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Common CLI Commands:\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

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
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress Space to select, and Enter to execute.\n\n"

	// Send the UI for rendering
	return s
}

// ==============================================================
// Funcs

// ==============================================================

func main() {
	m := initialModel()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {

		os.Exit(1)
	}

	keys := make([]int, 0, len(m.selected))
	for k := range m.selected {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		selected := m.selected[k]
		m.funcs[selected]()
	}
}
