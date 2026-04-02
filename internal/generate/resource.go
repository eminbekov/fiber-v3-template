package generate

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type resourceOptions struct {
	WithRepository bool
	WithHandler    bool
	WithService    bool
	APIVersion     string
	Force          bool
	DryRun         bool
}

type ResourceData struct {
	ModulePath      string
	EntityName      string
	EntityNameLower string
	EntityVariable  string
	TableName       string
	FileName        string
	APIVersion      string
}

// Resource scaffolds CRUD resource files.
func Resource(arguments []string) error {
	entityInputName, options, parseOptionsError := parseResourceOptions(arguments)
	if parseOptionsError != nil {
		return parseOptionsError
	}

	modulePath, readModulePathError := readModulePath()
	if readModulePathError != nil {
		return readModulePathError
	}

	snakeCaseName := ToSnakeCase(entityInputName)
	resourceData := ResourceData{
		ModulePath:      modulePath,
		EntityName:      ToCamelCase(snakeCaseName),
		EntityNameLower: strings.ToLower(ToCamelCase(snakeCaseName)),
		EntityVariable:  ToLowerCamelCase(snakeCaseName),
		TableName:       TableName(snakeCaseName),
		FileName:        snakeCaseName,
		APIVersion:      options.APIVersion,
	}

	generatedFileSpecs := buildFileSpecs(resourceData, options)
	for _, generatedFileSpec := range generatedFileSpecs {
		if generateFileError := generateResourceFile(generatedFileSpec, resourceData, options); generateFileError != nil {
			return generateFileError
		}
	}

	return nil
}

type generatedFileSpec struct {
	OutputPath   string
	TemplatePath string
}

func parseResourceOptions(arguments []string) (string, resourceOptions, error) {
	flagSet := flag.NewFlagSet("resource", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	withRepositoryValue := flagSet.Bool("with-repo", true, "generate repository files")
	withHandlerValue := flagSet.Bool("with-handler", true, "generate handler and dto files")
	withServiceValue := flagSet.Bool("with-service", true, "generate service file")
	apiVersionValue := flagSet.String("api-version", "v1", "api version for handler/response dto")
	forceValue := flagSet.Bool("force", false, "overwrite existing files")
	dryRunValue := flagSet.Bool("dry-run", false, "print files without writing")
	if parseError := flagSet.Parse(arguments); parseError != nil {
		return "", resourceOptions{}, fmt.Errorf("resource parse flags: %w", parseError)
	}

	positionalArguments := flagSet.Args()
	if len(positionalArguments) < 1 {
		return "", resourceOptions{}, fmt.Errorf("resource requires <name>, example: generate resource order")
	}

	entityInputName := strings.TrimSpace(positionalArguments[0])
	if entityInputName == "" {
		return "", resourceOptions{}, fmt.Errorf("resource name cannot be empty")
	}

	options := resourceOptions{
		WithRepository: *withRepositoryValue,
		WithHandler:    *withHandlerValue,
		WithService:    *withServiceValue,
		APIVersion:     strings.TrimSpace(*apiVersionValue),
		Force:          *forceValue,
		DryRun:         *dryRunValue,
	}
	if options.APIVersion == "" {
		return "", resourceOptions{}, fmt.Errorf("resource: --api-version cannot be empty")
	}
	return entityInputName, options, nil
}

func readModulePath() (string, error) {
	moduleFileBytes, readModuleError := os.ReadFile("go.mod")
	if readModuleError != nil {
		return "", fmt.Errorf("read go.mod: %w", readModuleError)
	}

	for _, lineValue := range strings.Split(string(moduleFileBytes), "\n") {
		trimmedLineValue := strings.TrimSpace(lineValue)
		if strings.HasPrefix(trimmedLineValue, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmedLineValue, "module ")), nil
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}

func buildFileSpecs(resourceData ResourceData, options resourceOptions) []generatedFileSpec {
	fileSpecs := []generatedFileSpec{
		{
			OutputPath:   filepath.Join("internal", "domain", resourceData.FileName+".go"),
			TemplatePath: "templates/domain.go.tmpl",
		},
		{
			OutputPath:   filepath.Join("internal", "cache", resourceData.FileName+"_keys.go"),
			TemplatePath: "templates/cache_keys.go.tmpl",
		},
	}

	if options.WithRepository {
		fileSpecs = append(fileSpecs,
			generatedFileSpec{
				OutputPath:   filepath.Join("internal", "repository", resourceData.FileName+"_repository.go"),
				TemplatePath: "templates/repository_interface.go.tmpl",
			},
			generatedFileSpec{
				OutputPath:   filepath.Join("internal", "repository", "postgres", resourceData.FileName+".go"),
				TemplatePath: "templates/repository_postgres.go.tmpl",
			},
		)
	}

	if options.WithService {
		fileSpecs = append(fileSpecs, generatedFileSpec{
			OutputPath:   filepath.Join("internal", "service", resourceData.FileName+"_service.go"),
			TemplatePath: "templates/service.go.tmpl",
		})
	}

	if options.WithHandler {
		fileSpecs = append(fileSpecs,
			generatedFileSpec{
				OutputPath:   filepath.Join("internal", "handler", "api", options.APIVersion, resourceData.FileName+"_handler.go"),
				TemplatePath: "templates/handler.go.tmpl",
			},
			generatedFileSpec{
				OutputPath:   filepath.Join("internal", "dto", "request", resourceData.FileName+".go"),
				TemplatePath: "templates/dto_request.go.tmpl",
			},
			generatedFileSpec{
				OutputPath:   filepath.Join("internal", "dto", "response", options.APIVersion, resourceData.FileName+".go"),
				TemplatePath: "templates/dto_response.go.tmpl",
			},
		)
	}

	return fileSpecs
}

func generateResourceFile(fileSpec generatedFileSpec, resourceData ResourceData, options resourceOptions) error {
	templateBytes, readTemplateError := templatesFS.ReadFile(fileSpec.TemplatePath)
	if readTemplateError != nil {
		return fmt.Errorf("resource read template %s: %w", fileSpec.TemplatePath, readTemplateError)
	}
	outputBytes, renderTemplateError := RenderTemplate(string(templateBytes), resourceData, filepath.Base(fileSpec.TemplatePath))
	if renderTemplateError != nil {
		return renderTemplateError
	}
	if writeFileError := WriteFile(fileSpec.OutputPath, outputBytes, options.DryRun, options.Force); writeFileError != nil {
		return writeFileError
	}
	return nil
}
