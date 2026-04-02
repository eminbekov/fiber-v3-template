package commands

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const exportUsersPageSize = 100

// ExportUsers exports users to CSV file or stdout.
func ExportUsers(ctx context.Context, dependencies *Dependencies, arguments []string) error {
	normalizedOutputPath, parseArgumentsError := parseExportUsersArguments(arguments)
	if parseArgumentsError != nil {
		return parseArgumentsError
	}

	outputWriter, outputFile, setupOutputError := setupExportUsersOutput(normalizedOutputPath)
	if setupOutputError != nil {
		return setupOutputError
	}
	if outputFile != nil {
		defer func() {
			_ = outputFile.Close()
		}()
	}

	if writeCSVError := writeUsersCSV(ctx, dependencies, outputWriter); writeCSVError != nil {
		return writeCSVError
	}
	printExportUsersCompletion(normalizedOutputPath)
	return nil
}

func parseExportUsersArguments(arguments []string) (string, error) {
	flagSet := flag.NewFlagSet("export-users", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	outputPathValue := flagSet.String("output", "", "output CSV path (stdout when empty)")
	if parseError := flagSet.Parse(arguments); parseError != nil {
		return "", fmt.Errorf("export-users parse flags: %w", parseError)
	}

	return strings.TrimSpace(*outputPathValue), nil
}

func setupExportUsersOutput(normalizedOutputPath string) (io.Writer, *os.File, error) {
	if normalizedOutputPath == "" {
		return os.Stdout, nil, nil
	}

	//nolint:gosec // console tool intentionally writes to user-provided output path.
	createdFile, createFileError := os.Create(normalizedOutputPath)
	if createFileError != nil {
		return nil, nil, fmt.Errorf("export-users create output file: %w", createFileError)
	}
	return createdFile, createdFile, nil
}

func writeUsersCSV(ctx context.Context, dependencies *Dependencies, outputWriter io.Writer) error {
	csvWriter := csv.NewWriter(outputWriter)
	defer csvWriter.Flush()

	headerValues := []string{"id", "username", "full_name", "phone", "status", "created_at"}
	if writeHeaderError := csvWriter.Write(headerValues); writeHeaderError != nil {
		return fmt.Errorf("export-users write header: %w", writeHeaderError)
	}

	pageNumber := 1
	exportedUserCount := int64(0)
	totalUserCount := int64(0)
	for {
		users, totalCount, listUsersError := dependencies.UserService.List(ctx, pageNumber, exportUsersPageSize)
		if listUsersError != nil {
			return fmt.Errorf("export-users list users: %w", listUsersError)
		}
		totalUserCount = totalCount
		if len(users) == 0 {
			break
		}

		for _, user := range users {
			recordValues := []string{
				user.ID.String(),
				user.Username,
				user.FullName,
				user.Phone,
				user.Status,
				user.CreatedAt.Format(time.RFC3339),
			}
			if writeRecordError := csvWriter.Write(recordValues); writeRecordError != nil {
				return fmt.Errorf("export-users write record: %w", writeRecordError)
			}
			exportedUserCount++
		}
		csvWriter.Flush()
		if flushError := csvWriter.Error(); flushError != nil {
			return fmt.Errorf("export-users flush writer: %w", flushError)
		}

		fmt.Fprintf(os.Stderr, "exported %d/%d users\n", exportedUserCount, totalUserCount)
		pageNumber++
	}

	return nil
}

func printExportUsersCompletion(normalizedOutputPath string) {
	if normalizedOutputPath != "" {
		fmt.Fprintf(os.Stderr, "user export completed: %s\n", normalizedOutputPath)
		return
	}
	fmt.Fprintln(os.Stderr, "user export completed: stdout")
}
