package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	databaseURLVariableName = "DATABASE_URL"
	migrationsPath          = "file://migrations"
)

func main() {
	if runError := run(); runError != nil {
		log.Fatalf("migrate: %v", runError)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <up|down|version|force>", os.Args[0])
	}

	databaseURL := os.Getenv(databaseURLVariableName)
	if databaseURL == "" {
		return fmt.Errorf("%s cannot be empty", databaseURLVariableName)
	}

	migrationInstance, newMigrationError := migrate.New(migrationsPath, databaseURL)
	if newMigrationError != nil {
		return fmt.Errorf("create migrate instance: %w", newMigrationError)
	}
	defer func() {
		sourceError, databaseError := migrationInstance.Close()
		if sourceError != nil || databaseError != nil {
			log.Printf("migrate: close source=%v database=%v", sourceError, databaseError)
		}
	}()

	command := os.Args[1]
	switch command {
	case "up":
		return runUp(migrationInstance)
	case "down":
		return runDown(migrationInstance, os.Args[2:])
	case "version":
		return runVersion(migrationInstance)
	case "force":
		return runForce(migrationInstance, os.Args[2:])
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func runUp(migrationInstance *migrate.Migrate) error {
	if upError := migrationInstance.Up(); upError != nil && !errors.Is(upError, migrate.ErrNoChange) {
		return fmt.Errorf("up: %w", upError)
	}

	return nil
}

func runDown(migrationInstance *migrate.Migrate, arguments []string) error {
	steps := 1
	if len(arguments) > 0 {
		parsedSteps, parseError := strconv.Atoi(arguments[0])
		if parseError != nil {
			return fmt.Errorf("down: invalid step count: %w", parseError)
		}
		if parsedSteps <= 0 {
			return errors.New("down: step count must be greater than zero")
		}

		steps = parsedSteps
	}

	if downError := migrationInstance.Steps(-steps); downError != nil && !errors.Is(downError, migrate.ErrNoChange) {
		return fmt.Errorf("down: %w", downError)
	}

	return nil
}

func runVersion(migrationInstance *migrate.Migrate) error {
	version, isDirty, versionError := migrationInstance.Version()
	if versionError != nil {
		if errors.Is(versionError, migrate.ErrNilVersion) {
			fmt.Println("version: none")
			return nil
		}

		return fmt.Errorf("version: %w", versionError)
	}

	fmt.Printf("version: %d dirty: %t\n", version, isDirty)
	return nil
}

func runForce(migrationInstance *migrate.Migrate, arguments []string) error {
	if len(arguments) != 1 {
		return errors.New("force: usage force <version>")
	}

	version, parseError := strconv.Atoi(arguments[0])
	if parseError != nil {
		return fmt.Errorf("force: invalid version: %w", parseError)
	}

	if forceError := migrationInstance.Force(version); forceError != nil {
		return fmt.Errorf("force: %w", forceError)
	}

	return nil
}
