package fd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type SelectScreen struct {
	locations map[string]string
	helpText  string
}

func NewSelectScreen(locations map[string]string) SelectScreen {
	return SelectScreen{
		locations: locations,
		helpText:  "Select a location to jump to.",
	}
}

func (this SelectScreen) Init() tea.Cmd {
	return nil
}

func (this SelectScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		default:
			key := msg.String()
			path, ok := this.locations[key]
			if !ok {
				this.helpText = fmt.Sprintf("No location bound to %s.", key)
				return this, nil
			}

			return this, SaveSelectedCmd(path)
		}
	}

	return this, nil
}

func (this SelectScreen) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n[help] %s\n\n", this.helpText))

	keys := []string{}
	for k := range this.locations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := this.locations[key]
		dir := dirName(path)
		bind := success.Render(key)
		prefix := success.Render("-->")
		builder.WriteString(fmt.Sprintf("%s [%s] %s\n", prefix, bind, dir))
	}

	builder.WriteString(fmt.Sprintf("\n[ctrl+a] Add [ctrl+d] Delete\n"))
	return builder.String()
}

type DeleteScreen struct {
	locations map[string]string
	helpText  string
}

func NewDeleteScreen(locations map[string]string) DeleteScreen {
	return DeleteScreen{
		locations: locations,
		helpText:  "Delete a location binding.",
	}
}

func (this DeleteScreen) Init() tea.Cmd {
	return nil
}

func (this DeleteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		_, ok := this.locations[key]
		if !ok {
			this.helpText = fmt.Sprintf("No location bound to %s.", key)
			return this, nil
		}

		delete(this.locations, key)
		this.helpText = fmt.Sprintf("Deleted location from %s.", key)
		return this, SaveLocationsCmd(this.locations)
	}

	return this, nil
}

func (this DeleteScreen) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n[help] %s\n\n", this.helpText))

	keys := []string{}
	for k := range this.locations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := this.locations[key]
		dir := dirName(path)
		bind := danger.Render(key)
		prefix := danger.Render("!!!")
		builder.WriteString(fmt.Sprintf("%s [%s] %s\n", prefix, bind, dir))
	}

	builder.WriteString(fmt.Sprintf("\n[ctrl+s] Select [ctrl+a] Add\n"))
	return builder.String()
}

type AddScreen struct {
	current   string
	locations map[string]string
	helpText  string
}

func NewAddScreen(current string, locations map[string]string) AddScreen {
	return AddScreen{
		locations: locations,
		current:   current,
		helpText:  "Add current location by pressing a key.",
	}
}

func (this AddScreen) Init() tea.Cmd {
	return nil
}

func (this AddScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		default:
			key := msg.String()
			_, ok := this.locations[key]
			if ok {
				this.helpText = fmt.Sprintf("Key %s is already bound to a location.", key)
				return this, nil
			}

			this.locations[key] = this.current
			this.helpText = fmt.Sprintf("Location added to %s.", key)
			return this, SaveLocationsCmd(this.locations)
		}
	}

	return this, nil
}

func (this AddScreen) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n[help] %s\n\n", this.helpText))

	keys := []string{}
	for k := range this.locations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := this.locations[key]
		dir := dirName(path)
		bind := success.Render(key)
		prefix := success.Render("+++")
		builder.WriteString(fmt.Sprintf("%s [%s] %s\n", prefix, bind, dir))
	}

	builder.WriteString(fmt.Sprintf("\n[ctrl+s] Select [ctrl+d] Delete\n"))
	return builder.String()
}

type ViewService struct {
	locationService *LocationService
	locations       map[string]string
	current         string
	selected        string
	screen          Screen
}

func NewViewService(locationService *LocationService) ViewService {
	locations, err := locationService.ReadSavedLocations()
	if err != nil {
		fmt.Println("Error initializing view service:", err)
		os.Exit(1)
	}

	currentLocation, err := locationService.CurrentLocation()
	if err != nil {
		fmt.Println("Error initializing view service:", err)
		os.Exit(1)
	}

	return ViewService{
		locationService: locationService,
		locations:       locations,
		current:         currentLocation,
		screen:          NewSelectScreen(locations),
	}
}

func (this ViewService) Init() tea.Cmd {
	return nil
}

type SaveLocationsMsg struct {
	locations map[string]string
}

func SaveLocationsCmd(locations map[string]string) tea.Cmd {
	return func() tea.Msg {
		return SaveLocationsMsg{locations}
	}
}

type SaveSelectedMsg struct {
	selected string
}

func SaveSelectedCmd(selected string) tea.Cmd {
	return func() tea.Msg {
		return SaveSelectedMsg{selected}
	}
}

type ExitAppMsg struct {
	selected string
}

func ExitAppCmd(selected string) tea.Cmd {
	return func() tea.Msg {
		return ExitAppMsg{selected}
	}
}

func (this ViewService) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SaveSelectedMsg:
		this.selected = msg.selected
		this.locationService.SaveSelectedLocation(msg.selected)
		return this, tea.Quit

	case SaveLocationsMsg:
		this.locations = msg.locations
		this.locationService.SaveLocations(msg.locations)
		return this, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return this, SaveSelectedCmd(this.current)

		case "ctrl+d":
			this.screen = NewDeleteScreen(this.locations)
			return this, nil

		case "ctrl+a":
			this.screen = NewAddScreen(this.current, this.locations)
			return this, nil

		case "ctrl+s":
			this.screen = NewSelectScreen(this.locations)
			return this, nil

		default:
			screen, cmd := this.screen.Update(msg)
			this.screen = screen
			return this, cmd
		}
	}

	return this, nil
}

func (this ViewService) View() string {
	return this.screen.View()
}

var (
	success = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	danger  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func dirName(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	return name
}
