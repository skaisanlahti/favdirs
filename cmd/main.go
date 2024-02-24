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
	return Model{
		locations: readLocations(),
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
					return this, saveChanges(this.locations)
				}

				this.helpText = fmt.Sprintf("Failed to remove location. Nothing saved for %s.", key)
				return this, nil
			}

			path, ok := this.locations[key]
			if ok {
				this.selected = path
				return this, tea.Quit
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

func saveChanges(locations map[string]string) tea.Cmd {
	return func() tea.Msg {
		saveLocations(locations)
		return saveMsg{}
	}
}

const (
	stateDirectory = "/tmp/go-nav"
	locationFile   = stateDirectory + "/loc"
	selectedFile   = stateDirectory + "/go"
)

func currentLocation() string {
	location, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return location
}

func saveSelected(selected string) {
	err := os.WriteFile(selectedFile, []byte(selected), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readLocations() map[string]string {
	file, err := os.OpenFile(locationFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	locations := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		locations[parts[0]] = parts[1]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return locations
}

func saveLocations(locations map[string]string) {
	var builder strings.Builder
	keys := []string{}
	for k := range locations {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("%s=%s\n", key, locations[key]))
	}

	err := os.WriteFile(locationFile, []byte(builder.String()), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func addLocation(key string) {
	locations := readLocations()
	location := currentLocation()
	locations[key] = location
	saveLocations(locations)
	saveSelected(location)
}

func init() {
	err := os.MkdirAll(stateDirectory, 0755) // 0755 commonly used for directories
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) > 1 {
		addLocation(os.Args[1])
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
