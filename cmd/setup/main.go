package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type model struct {
	choices  []string // items on the to-do list
	cursor   int      // which to-do list item our cursor is pointing at
	funcs    map[string]func()
	selected map[int]string // which to-do items are selected
}

func initialModel() model {

	choices := []string{"*DANGER* Reset DB", "Push To DB"}

	funcs := map[string]func(){
		choices[0]: resetDb,
		choices[1]: pushToDB,
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

// ==============================================================
// ==============================================================
// ==============================================================
// ==============================================================

func pushToDB() {

	conn := loadDB()
	defer conn.Close(context.Background())

	schemaFilePath := "models/sql/schema.sql"
	contentSchema, errSchema := os.ReadFile(schemaFilePath)
	if errSchema != nil {
		log.Fatal(errSchema)
	}
	response, err := conn.Exec(context.Background(), string(contentSchema))
	if err != nil {
		log.Fatal("Error db:", err)
	}

	fmt.Println("------")
	fmt.Println(response)
	fmt.Println("------")

}

func resetDb() {

	ctx := context.Background()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error db:", err)
	}

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to parse configuration: %v", err)
	}

	// Open connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// execResponseDrop, err := pool.Exec(ctx, "DROP DATABASE IF EXISTS postgres")
	execResponseDrop, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE;")
	if err != nil {
		log.Fatalf("error dropping database: %v", err)
	}

	fmt.Println("------")
	fmt.Println(execResponseDrop)
	fmt.Println("------")

	// Create the database
	// execResponseCreate, err := pool.Exec(ctx, "CREATE DATABASE postgres")
	execResponseCreate, err := pool.Exec(ctx, "CREATE SCHEMA public;")
	if err != nil {
		log.Fatalf("error creating database: %v", err)
	}

	fmt.Println("------")
	fmt.Println(execResponseCreate)
	fmt.Println("------")

}

func loadDB() *pgx.Conn {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error db:", err)
	}

	conn, dbErr := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	// defer conn.Close(context.Background())

	// conn.Exec(context.Background(), "")

	return conn
}
