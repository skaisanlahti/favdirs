package fd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
		helpText:        "Select location using the key in brackets.",
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
			return this, this.saveAndQuit()

		case " ":
			if this.mode == ModeSelect {
				this.mode = ModeDelete
				this.helpText = "Delete a bound directory."
				return this, nil
			}

			this.helpText = "Select directory to jump to."
			this.mode = ModeSelect
			return this, nil

		default:
			key := msg.String()
			path, ok := this.locations[key]
			if !ok {
				this.helpText = fmt.Sprintf("No directory bound to %s.", key)
				return this, nil
			}

			if this.mode == ModeDelete {
				delete(this.locations, key)
				this.helpText = fmt.Sprintf("Removed directory from %s.", key)
				return this, this.saveChanges()
			}

			this.selected = path
			return this, this.saveAndQuit()
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
		builder.WriteString(fmt.Sprintf("xxx [%s] %s\n", key, dir))
	}

	builder.WriteString(fmt.Sprintf("\n[help] %s [space] %s [esc] exit\n", this.helpText, ModeSelect))
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
		builder.WriteString(fmt.Sprintf("--> [%s] %s\n", key, dir))
	}

	builder.WriteString(fmt.Sprintf("\n[help] %s [space] %s [esc] exit\n", this.helpText, ModeDelete))
	return builder.String()
}

type saveMsg struct{}
type errMsg struct {
	err error
}

func (this ViewService) saveChanges() tea.Cmd {
	return func() tea.Msg {
		err := this.locationService.SaveLocations(this.locations)
		if err != nil {
			return errMsg{err}
		}

		err = this.locationService.SaveSelectedLocation(this.selected)
		if err != nil {
			return errMsg{err}
		}

		return saveMsg{}
	}
}

func (this ViewService) saveAndQuit() tea.Cmd {
	this.locationService.SaveLocations(this.locations)
	this.locationService.SaveSelectedLocation(this.selected)
	return tea.Quit
}
