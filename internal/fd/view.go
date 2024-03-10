package fd

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
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
		helpText:  "Select a directory to jump to.",
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
				this.helpText = fmt.Sprintf("No directory bound to %s.", key)
				return this, nil
			}

			return this, SaveSelectedCmd(path)
		}
	}

	return this, nil
}

func (this SelectScreen) View() string {
	list := lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("[help] %s", this.helpText),
		divider(),
		locationList(this.locations, success),
		divider(),
		fmt.Sprintf("[ctrl+a] Add directory"),
		fmt.Sprintf("[ctrl+d] Delete directory"),
		fmt.Sprintf("[ctrl+c] Exit\n"),
	)

	return contentBox(list)
}

type DeleteScreen struct {
	locations map[string]string
	helpText  string
}

func NewDeleteScreen(locations map[string]string) DeleteScreen {
	return DeleteScreen{
		locations: locations,
		helpText:  "Delete a directory binding.",
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
			this.helpText = fmt.Sprintf("No directory bound to %s.", key)
			return this, nil
		}

		delete(this.locations, key)
		this.helpText = fmt.Sprintf("Deleted directory from %s.", key)
		return this, SaveLocationsCmd(this.locations)
	}

	return this, nil
}

func (this DeleteScreen) View() string {
	list := lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("[help] %s", this.helpText),
		divider(),
		locationList(this.locations, danger),
		divider(),
		fmt.Sprintf("[ctrl+a] Add directory"),
		fmt.Sprintf("[ctrl+s] Select directory"),
		fmt.Sprintf("[ctrl+c] Exit\n"),
	)

	return contentBox(list)
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
		helpText:  "Add current directory by pressing a key.",
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
				this.helpText = fmt.Sprintf("Key %s is already bound to a directory.", key)
				return this, nil
			}

			this.locations[key] = this.current
			this.helpText = fmt.Sprintf("Directory bound to %s.", key)
			return this, SaveLocationsCmd(this.locations)
		}
	}

	return this, nil
}

func (this AddScreen) View() string {
	list := lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("[help] %s", this.helpText),
		divider(),
		locationList(this.locations, warning),
		divider(),
		fmt.Sprintf("[ctrl+d] Delete directory"),
		fmt.Sprintf("[ctrl+s] Select directory"),
		fmt.Sprintf("[ctrl+c] Exit\n"),
	)

	return contentBox(list)
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
	success  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	danger   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	warning  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	faint    = lipgloss.NewStyle().Faint(true)
	frame    = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)
	maxWidth = 70
)

func contentBox(content string) string {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Error in contentBox", err)
		os.Exit(1)
	}

	return frame.Width(width).Height(height).Render(content)
}

func divider() string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Error in divider", err)
		os.Exit(1)
	}

	dividerLen := maxWidth
	if width > 0 && width < dividerLen {
		dividerLen = width
	}

	return strings.Repeat("-", dividerLen)
}

func dirName(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	return name
}

func locationList(locations map[string]string, style lipgloss.Style) string {
	var doc strings.Builder
	keys := []string{}
	for k := range locations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := truncPath(locations[key])
		dir := dirName(path)
		faintPath := faint.Render(path)
		styleDir := style.Render(dir)
		styleBind := style.Render(key)
		doc.WriteString(fmt.Sprintf("[%s] %-20s %s\n", styleBind, styleDir, faintPath))
	}

	return doc.String()
}

func truncPath(path string) string {
	pathRunes := []rune(path)
	dirRunes := []rune(dirName(path))
	maxPathLen := max(maxWidth-len(dirRunes)-5, 0)
	pathRuneCount := len(pathRunes)
	if pathRuneCount > maxPathLen {
		end := max(pathRuneCount-maxPathLen+3, 0)
		trunc := string(pathRunes[:end])
		return trunc + "..."
	}

	return path
}

func max(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}
