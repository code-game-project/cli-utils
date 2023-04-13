package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func Execute(templateText, path string, data any) error {
	err := os.MkdirAll(filepath.Join(filepath.Dir(path)), 0o755)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmpl, err := template.New(path).Parse(templateText)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	file, err := os.Create(filepath.Join(path))
	if err != nil {
		return fmt.Errorf("create template file: %w", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return nil
}
