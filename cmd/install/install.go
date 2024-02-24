package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const (
	startComment     = "# fd-app-def"
	endComment       = "# fd-app-end"
	functionTemplate = "./cmd/install/shell-template"
	executablePath   = "/usr/local/bin/fd-app"
	stateDirectory   = "/fd-app"
)

type TemplateData struct {
	ExecutablePath string
	SelectPath     string
	StartComment   string
	EndComment     string
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding user home directory: ", err)
		return
	}

	stateDir := homeDir + stateDirectory
	err = os.MkdirAll(stateDir, 0755) // 0755 commonly used for directories
	if err != nil {
		fmt.Println("Error creating state directory: ", err)
		return
	}
	fmt.Println("Created fd-app state directory in: ", stateDir)
	selectFile := stateDir + "/select"
	locationsFile := stateDir + "/locations"

	err = os.WriteFile(selectFile, []byte{}, 0666)
	if err != nil {
		fmt.Println("Error creating select file: ", err)
		return
	}

	err = os.WriteFile(locationsFile, []byte{}, 0666)
	if err != nil {
		fmt.Println("Error creating locations file: ", err)
		return
	}

	// Prepare the template data
	data := TemplateData{
		ExecutablePath: executablePath,
		SelectPath:     homeDir + stateDirectory + "/select",
		StartComment:   startComment,
		EndComment:     endComment,
	}

	// Parse and execute the template
	tmpl, err := template.ParseFiles(functionTemplate)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Create a builder to hold the executed template
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
	functionDefinition := builder.String()

	bashrcPath := homeDir + "/.bashrc"
	tempFilePath := bashrcPath + ".tmp"
	bashrcFile, err := os.Open(bashrcPath)
	if err != nil {
		fmt.Println("Error opening .bashrc:", err)
		return
	}
	defer bashrcFile.Close()

	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}
	defer tempFile.Close()

	inFunction := false
	updated := false
	scanner := bufio.NewScanner(bashrcFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip lines until the end comment is found
		if inFunction {
			if line == endComment {
				inFunction = false
			}
			continue
		}

		// Check for the start of the function definition
		if line == startComment {
			inFunction = true
			updated = true
			// Write the new function definition
			if _, err := tempFile.WriteString(functionDefinition); err != nil {
				fmt.Println("Error writing function definition:", err)
				return
			}
			continue
		}

		// Write lines outside of the function definition
		if _, err := tempFile.WriteString(line + "\n"); err != nil {
			fmt.Println("Error writing to temporary file:", err)
			return
		}
	}

	// If the function was not found and updated, append it
	if !updated {
		if _, err := tempFile.WriteString("\n" + functionDefinition + "\n"); err != nil {
			fmt.Println("Error appending function definition:", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading .bashrc:", err)
		return
	}

	// Replace the original .bashrc with the updated version
	if err := os.Rename(tempFilePath, bashrcPath); err != nil {
		fmt.Println("Error updating .bashrc:", err)
		return
	}

	fmt.Println("Successfully updated .bashrc")
}
