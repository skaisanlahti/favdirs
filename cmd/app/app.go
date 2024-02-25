package main

import (
	"os"

	"github.com/skaisanlahti/favdirs/internal/fd"
)

func main() {
	locationService := fd.NewLocationService()
	viewService := fd.NewViewService(locationService)
	app := fd.NewApp(viewService, locationService)

	if len(os.Args) > 1 {
		app.AddLocation(os.Args[1])
	} else {
		app.ViewUserInterface()
	}
}
