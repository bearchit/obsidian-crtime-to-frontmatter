package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/djherbis/times"
	"gopkg.in/yaml.v3"
)

func addCreationDateToFrontmatter(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	t, err := times.Stat(filePath)
	if err != nil {
		return err
	}

	var creationTime time.Time
	if t.HasBirthTime() {
		creationTime = t.BirthTime()
	} else {
		creationTime = t.ModTime()
	}

	creationDate := creationTime.Format("2006-01-02 15:04:05")

	var newContent string
	contentStr := string(content)
	if strings.HasPrefix(contentStr, "---\n") {
		parts := strings.SplitN(contentStr[4:], "\n---\n", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid frontmatter format: missing closing '---'")
		}

		frontmatter := make(map[string]interface{})
		if len(parts[0]) > 0 {
			yamlContent := strings.TrimSpace(parts[0])
			err = yaml.Unmarshal([]byte(yamlContent), &frontmatter)
			if err != nil {
				return fmt.Errorf("invalid YAML in frontmatter: %v", err)
			}
		}

		frontmatter["created"] = creationDate

		newFrontmatter, err := yaml.Marshal(frontmatter)
		if err != nil {
			return err
		}

		newContent = fmt.Sprintf("---\n%s\n---\n%s",
			strings.TrimSpace(string(newFrontmatter)),
			parts[1])
	} else {
		frontmatter := map[string]interface{}{
			"created": creationDate,
		}

		newFrontmatter, err := yaml.Marshal(frontmatter)
		if err != nil {
			return err
		}

		newContent = fmt.Sprintf("---\n%s\n---\n\n%s",
			strings.TrimSpace(string(newFrontmatter)),
			contentStr)
	}

	return os.WriteFile(filePath, []byte(newContent), 0o644)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: program <vault-path>")
		os.Exit(1)
	}

	vaultPath := os.Args[1]
	errorFiles := make(map[string]error)
	errorTypes := make(map[string]int)

	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			if err := addCreationDateToFrontmatter(path); err != nil {
				errorFiles[path] = err
				errorTypes[err.Error()]++
				fmt.Printf("Error processing %s: %v\n", path, err)
			} else {
				fmt.Printf("Processed: %s\n", path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", vaultPath, err)
	}

	if len(errorFiles) > 0 {
		fmt.Printf("\n=== Error Summary ===\n")
		fmt.Printf("Total files processed with errors: %d\n", len(errorFiles))

		fmt.Printf("\nError types:\n")
		for errMsg, count := range errorTypes {
			fmt.Printf("- %s: %d files\n", errMsg, count)
		}

		fmt.Printf("\nFailed files:\n")
		for path, err := range errorFiles {
			fmt.Printf("- %s\n  Error: %v\n", path, err)
		}
	} else {
		fmt.Printf("\nAll files processed successfully!\n")
	}
}
