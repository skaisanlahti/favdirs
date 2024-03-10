package fd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	viewService     ViewService
	locationService *LocationService
}

func NewApp(viewService ViewService, locationService *LocationService) *App {
	return &App{viewService, locationService}
}

func (this *App) AddLocation(key string) {
	locations, err := this.locationService.ReadSavedLocations()
	if err != nil {
		this.handleError("Error reading saved locations:", err)
	}

	location, err := this.locationService.CurrentLocation()
	if err != nil {
		this.handleError("Error reading current directory:", err)
	}

	locations[key] = location
	err = this.locationService.SaveLocations(locations)
	if err != nil {
		this.handleError("Error saving locations:", err)
	}

	err = this.locationService.SaveSelectedLocation(location)
	if err != nil {
		this.handleError("Error saving selected location:", err)
	}

	fmt.Println("Added location successfully.")
}

func (this *App) ViewUserInterface() {
	program := tea.NewProgram(this.viewService, tea.WithAltScreen())
	model, err := program.Run()
	if err != nil {
		this.handleError("Error creating user interface:", err)
	}

	lastState := model.(ViewService)
	if lastState.current != lastState.selected {
		fmt.Printf("\nMoving to %s...\n", lastState.selected)
	}
}

func (this *App) handleError(message string, err error) {
	fmt.Println(message, err)
	os.Exit(1)
}
