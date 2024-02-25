package fd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type LocationService struct {
	locationFile string
	selectFile   string
}

func NewLocationService() LocationService {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}

	appDir := "/.fd-app"
	locationFile := homeDir + appDir + "/locations"
	selectFile := homeDir + appDir + "/select"
	return LocationService{locationFile, selectFile}
}

func (this LocationService) CurrentLocation() (string, error) {
	location, err := os.Getwd()
	if err != nil {
		return location, err
	}

	return location, nil
}

func (this LocationService) SaveSelectedLocation(selected string) error {
	err := os.WriteFile(this.selectFile, []byte(selected), 0666)
	if err != nil {
		return err
	}

	return nil
}

func (this *LocationService) ReadSavedLocations() (map[string]string, error) {
	locations := map[string]string{}
	file, err := os.Open(this.locationFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		locations[parts[0]] = parts[1]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}

func (this LocationService) SaveLocations(locations map[string]string) error {
	var builder strings.Builder
	keys := []string{}
	for k := range locations {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("%s=%s\n", key, locations[key]))
	}

	err := os.WriteFile(this.locationFile, []byte(builder.String()), 0666)
	if err != nil {
		return err
	}

	return nil
}

func (this LocationService) AddLocation(key string) error {
	locations, err := this.ReadSavedLocations()
	if err != nil {
		return err
	}

	location, err := this.CurrentLocation()
	if err != nil {
		return err
	}

	locations[key] = location
	err = this.SaveLocations(locations)
	if err != nil {
		return err
	}

	err = this.SaveSelectedLocation(location)
	if err != nil {
		return err
	}

	return nil
}
