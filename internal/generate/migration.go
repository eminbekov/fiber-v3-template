package generate

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var migrationFilePattern = regexp.MustCompile(`^(\d+)_.*\.(up|down)\.sql$`)

type migrationTemplateData struct {
	Name      string
	Timestamp string
	TableName string
}

// Migration creates the next sequential up/down migration files.
func Migration(arguments []string) error {
	flagSet := flag.NewFlagSet("migration", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	forceValue := flagSet.Bool("force", false, "overwrite existing files")
	dryRunValue := flagSet.Bool("dry-run", false, "print files without writing")
	if parseError := flagSet.Parse(arguments); parseError != nil {
		return fmt.Errorf("migration parse flags: %w", parseError)
	}

	positionalArguments := flagSet.Args()
	if len(positionalArguments) < 1 {
		return fmt.Errorf("migration requires <name>, example: generate migration create_orders")
	}

	migrationName := ToSnakeCase(strings.TrimSpace(positionalArguments[0]))
	if migrationName == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	nextSequenceNumber, nextSequenceError := nextMigrationSequence("migrations")
	if nextSequenceError != nil {
		return nextSequenceError
	}

	migrationPrefix := fmt.Sprintf("%06d_%s", nextSequenceNumber, migrationName)
	templateData := migrationTemplateData{
		Name:      migrationName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TableName: TableName(migrationName),
	}

	upTemplateBytes, readUpTemplateError := templatesFS.ReadFile("templates/migration_up.sql.tmpl")
	if readUpTemplateError != nil {
		return fmt.Errorf("migration read up template: %w", readUpTemplateError)
	}
	upOutputBytes, renderUpTemplateError := RenderTemplate(string(upTemplateBytes), templateData, "migration_up.sql.tmpl")
	if renderUpTemplateError != nil {
		return renderUpTemplateError
	}
	upOutputPath := filepath.Join("migrations", migrationPrefix+".up.sql")
	if writeUpFileError := WriteFile(upOutputPath, upOutputBytes, *dryRunValue, *forceValue); writeUpFileError != nil {
		return writeUpFileError
	}

	downTemplateBytes, readDownTemplateError := templatesFS.ReadFile("templates/migration_down.sql.tmpl")
	if readDownTemplateError != nil {
		return fmt.Errorf("migration read down template: %w", readDownTemplateError)
	}
	downOutputBytes, renderDownTemplateError := RenderTemplate(string(downTemplateBytes), templateData, "migration_down.sql.tmpl")
	if renderDownTemplateError != nil {
		return renderDownTemplateError
	}
	downOutputPath := filepath.Join("migrations", migrationPrefix+".down.sql")
	if writeDownFileError := WriteFile(downOutputPath, downOutputBytes, *dryRunValue, *forceValue); writeDownFileError != nil {
		return writeDownFileError
	}

	return nil
}

func nextMigrationSequence(migrationsDirectory string) (int, error) {
	directoryEntries, readDirectoryError := os.ReadDir(migrationsDirectory)
	if readDirectoryError != nil {
		return 0, fmt.Errorf("read migrations directory: %w", readDirectoryError)
	}

	highestSequenceNumber := 0
	for _, directoryEntry := range directoryEntries {
		matchValues := migrationFilePattern.FindStringSubmatch(directoryEntry.Name())
		if len(matchValues) < 2 {
			continue
		}
		sequenceNumber, parseSequenceError := strconv.Atoi(matchValues[1])
		if parseSequenceError != nil {
			return 0, fmt.Errorf("parse migration sequence from %s: %w", directoryEntry.Name(), parseSequenceError)
		}
		if sequenceNumber > highestSequenceNumber {
			highestSequenceNumber = sequenceNumber
		}
	}

	return highestSequenceNumber + 1, nil
}
