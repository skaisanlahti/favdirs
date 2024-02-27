package fd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	ModeSelect = "select"
	ModeDelete = "delete"
)

type ViewService struct {
	locationService LocationService
	locations       map[string]string
	selected        string
	helpText        string
	mode            string
}

func NewViewService(locationService LocationService) ViewService {
	locations, err := locationService.ReadSavedLocations()
	if err != nil {
		fmt.Println("Error initializing view service:", err)
		os.Exit(1)
	}

	return ViewService{
		locationService: locationService,
		locations:       locations,
		helpText:        "Select a location to jump to.",
		mode:            ModeSelect,
	}
}

func (this ViewService) Init() tea.Cmd {
	return nil
}

func (this ViewService) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		this.helpText = msg.err.Error()
		return this, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			currentDirectory, err := this.locationService.CurrentLocation()
			if err != nil {
				return this, tea.Quit
			}

			this.selected = currentDirectory
			return this, this.saveSelectedAndQuit()

		case " ":
			if this.mode == ModeSelect {
				this.mode = ModeDelete
				this.helpText = "Delete a location binding."
				return this, nil
			}

			this.helpText = "Select a location to jump to."
			this.mode = ModeSelect
			return this, nil

		default:
			key := msg.String()
			path, ok := this.locations[key]
			if !ok {
				this.helpText = fmt.Sprintf("No location bound to %s.", key)
				return this, nil
			}

			if this.mode == ModeDelete {
				delete(this.locations, key)
				this.helpText = fmt.Sprintf("Deleted location from %s.", key)
				return this, this.saveLocations()
			}

			this.selected = path
			return this, this.saveSelectedAndQuit()
		}
	}

	return this, nil
}

func (this ViewService) View() string {
	if this.mode == ModeDelete {
		return this.deleteView()
	}

	return this.selectView()
}

func dirName(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	return name
}

var (
	success = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	danger  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func (this ViewService) deleteView() string {
	var builder strings.Builder
	builder.WriteString("Fav Dirs\n\n")

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

	space := success.Render("space")
	esc := danger.Render("esc")
	builder.WriteString(fmt.Sprintf("\n[help] %s [%s] %s [%s] exit\n", this.helpText, space, ModeSelect, esc))
	return builder.String()
}

func (this ViewService) selectView() string {
	var builder strings.Builder
	builder.WriteString("Fav Dirs\n\n")

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

	space := danger.Render("space")
	esc := danger.Render("esc")
	builder.WriteString(fmt.Sprintf("\n[help] %s [%s] %s [%s] exit\n", this.helpText, space, ModeDelete, esc))
	return builder.String()
}

type saveMsg struct{}
type errMsg struct {
	err error
}

func (this ViewService) saveLocations() tea.Cmd {
	return func() tea.Msg {
		err := this.locationService.SaveLocations(this.locations)
		if err != nil {
			return errMsg{err}
		}

		return saveMsg{}
	}
}

func (this ViewService) saveSelectedAndQuit() tea.Cmd {
	this.locationService.SaveSelectedLocation(this.selected)
	return tea.Quit
}
