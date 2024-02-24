package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	locations map[string]string
	selected  string
	helpText  string
}

func NewModel() Model {
	loc, err := readLocations()
	if err != nil {
		loc = map[string]string{}
	}

	selected, err := os.UserHomeDir()
	if err != nil {
		selected = ""
	}

	return Model{
		locations: loc,
		selected:  selected,
		helpText:  "Select location using the key in brackets.",
	}
}

func (this Model) Init() tea.Cmd {
	return nil
}

func (this Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			this.selected = currentLocation()
			return this, tea.Quit
		default:
			key := msg.String()
			if strings.HasPrefix(key, "alt+") {
				key = strings.Split(key, "+")[1]
				path, ok := this.locations[key]
				if ok {
					delete(this.locations, key)
					this.selected = path
					this.helpText = fmt.Sprintf("Removed locations from %s.", key)
					return this, saveChanges(this)
				}

				this.helpText = fmt.Sprintf("Failed to remove location. Nothing saved for %s.", key)
				return this, saveChanges(this)
			}

			path, ok := this.locations[key]
			if ok {
				this.selected = path
				return this, saveAndQuit(this)
			}

			this.helpText = fmt.Sprintf("No location bound to %s, try again.", key)
		}

	}
	return this, nil
}

func (this Model) View() string {
	keys := []string{}
	for k := range this.locations {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	var builder strings.Builder

	builder.WriteString("Favourite Locations:\n\n")
	for _, key := range keys {
		path := this.locations[key]
		builder.WriteString(fmt.Sprintf("[%s] %s\n", key, path))
	}

	builder.WriteString(fmt.Sprintf("\n%s", this.helpText))
	return builder.String()
}

type saveMsg struct{}

func saveChanges(model Model) tea.Cmd {
	return func() tea.Msg {
		saveLocations(model.locations)
		saveSelected(model.selected)
		return saveMsg{}
	}
}

func saveAndQuit(model Model) tea.Cmd {
	err := saveLocations(model.locations)
	if err != nil {
		return nil
	}

	err = saveSelected(model.selected)
	if err != nil {
		return nil
	}

	return tea.Quit
}

const (
	stateDirectory = "/fd-app"
	locationFile   = stateDirectory + "/locations"
	selectedFile   = stateDirectory + "/select"
)

func currentLocation() string {
	location, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return location
}

func saveSelected(selected string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	target := homeDir + selectedFile
	err = os.WriteFile(target, []byte(selected), 0666)
	if err != nil {
		return err
	}

	return nil
}

func readLocations() (map[string]string, error) {
	locations := map[string]string{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return locations, err
	}

	target := homeDir + locationFile
	file, err := os.Open(target)
	if err != nil {
		return locations, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		locations[parts[0]] = parts[1]
	}

	if err := scanner.Err(); err != nil {
		return locations, err
	}

	return locations, err
}

func saveLocations(locations map[string]string) error {
	var builder strings.Builder
	keys := []string{}
	for k := range locations {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("%s=%s\n", key, locations[key]))
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	target := homeDir + locationFile
	err = os.WriteFile(target, []byte(builder.String()), 0666)
	if err != nil {
		return err
	}

	return nil
}

func addLocation(key string) error {
	locations, err := readLocations()
	if err != nil {
		return err
	}

	location, err := os.Getwd()
	if err != nil {
		return err
	}

	locations[key] = location
	err = saveLocations(locations)
	if err != nil {
		return err
	}

	err = saveSelected(location)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) > 1 {
		err := addLocation(os.Args[1])
		if err != nil {
			fmt.Println("Error adding location: ", err)
			return
		}

		fmt.Println("Added location successfully.")
		return
	}

	program := tea.NewProgram(NewModel())
	result, err := program.Run()
	if err != nil {
		log.Fatal(err)
	}

	model, ok := result.(Model)
	if !ok {
		log.Fatal(err)
	}

	saveSelected(model.selected)
}
