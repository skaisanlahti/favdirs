package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

const (
	startComment     = "# fd-cmd-def"
	endComment       = "# fd-cmd-end"
	alias            = "fd"
	functionTemplate = "./cmd/install/function-template"
	executablePath   = "/app"
	appDir           = "/.fd-app"
)

type functionTemplateData struct {
	ExecutablePath string
	SelectPath     string
	StartComment   string
	Alias          string
	EndComment     string
}

func createFileIfNotExist(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {

		return err
	}

	defer file.Close()
	return nil
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are the same,
// then an error is returned. If dst does not exist, it is created with permissions copied
// from src.
func copyFile(src, dst string) error {
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get the source file's permissions
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create the destination file with the same permissions as the source file
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcFileInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding user home directory:", err)
		return
	}

	dir := homeDir + appDir
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("Error creating state directory:", err)
		return
	}

	executableFile := dir + executablePath
	err = copyFile("./build"+executablePath, executableFile)
	if err != nil {
		fmt.Println("Error moving executable to app directory:", err)
		return
	}

	selectFile := dir + "/select"
	err = createFileIfNotExist(selectFile)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("File at %s already exists, skipping creation.\n", selectFile)
		} else {
			fmt.Println("Error creating select file:", err)
			return
		}
	} else {
		fmt.Println("Created select file at:", selectFile)
	}

	locationsFile := dir + "/locations"
	err = createFileIfNotExist(locationsFile)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("File at %s already exists, skipping creation.\n", locationsFile)
		} else {
			fmt.Println("Error creating locations file:", err)
			return
		}
	} else {
		fmt.Println("Created locations file at:", locationsFile)
	}

	tmpl, err := template.ParseFiles(functionTemplate)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	data := functionTemplateData{
		ExecutablePath: homeDir + appDir + executablePath,
		SelectPath:     selectFile,
		StartComment:   startComment,
		Alias:          alias,
		EndComment:     endComment,
	}

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
		if inFunction {
			if line == endComment {
				inFunction = false
			}
			continue
		}

		if line == startComment {
			inFunction = true
			updated = true
			if _, err := tempFile.WriteString(functionDefinition); err != nil {
				fmt.Println("Error writing function definition:", err)
				return
			}
			continue
		}

		if _, err := tempFile.WriteString(line + "\n"); err != nil {
			fmt.Println("Error writing to temporary file:", err)
			return
		}
	}

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

	if err := os.Rename(tempFilePath, bashrcPath); err != nil {
		fmt.Println("Error updating .bashrc:", err)
		return
	}

	fmt.Println("Successfully updated .bashrc")
}
