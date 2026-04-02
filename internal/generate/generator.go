package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// WriteFile writes generated output respecting dry-run and force flags.
func WriteFile(outputPath string, content []byte, dryRun bool, force bool) error {
	if dryRun {
		if _, writeLogError := fmt.Fprintf(os.Stdout, "dry-run: %s\n", outputPath); writeLogError != nil {
			return fmt.Errorf("write dry-run output: %w", writeLogError)
		}
		return nil
	}

	if !force {
		if _, statError := os.Stat(outputPath); statError == nil {
			return fmt.Errorf("output already exists: %s (use --force to overwrite)", outputPath)
		}
	}

	if mkdirError := os.MkdirAll(filepath.Dir(outputPath), 0o750); mkdirError != nil {
		return fmt.Errorf("mkdir parent for %s: %w", outputPath, mkdirError)
	}
	if writeError := os.WriteFile(outputPath, content, 0o600); writeError != nil {
		return fmt.Errorf("write file %s: %w", outputPath, writeError)
	}

	if _, writeLogError := fmt.Fprintf(os.Stdout, "created: %s\n", outputPath); writeLogError != nil {
		return fmt.Errorf("write created output: %w", writeLogError)
	}
	return nil
}

// RenderTemplate renders a text template and formats Go source files.
func RenderTemplate(templateText string, data any, fileName string) ([]byte, error) {
	compiledTemplate, parseTemplateError := template.New(fileName).Parse(templateText)
	if parseTemplateError != nil {
		return nil, fmt.Errorf("parse template %s: %w", fileName, parseTemplateError)
	}

	var buffer bytes.Buffer
	if executeTemplateError := compiledTemplate.Execute(&buffer, data); executeTemplateError != nil {
		return nil, fmt.Errorf("execute template %s: %w", fileName, executeTemplateError)
	}

	outputBytes := buffer.Bytes()
	if strings.HasSuffix(fileName, ".go.tmpl") || strings.HasSuffix(fileName, ".go") {
		formattedOutput, formatError := format.Source(outputBytes)
		if formatError != nil {
			return nil, fmt.Errorf("format generated go for %s: %w", fileName, formatError)
		}
		return formattedOutput, nil
	}

	return outputBytes, nil
}
