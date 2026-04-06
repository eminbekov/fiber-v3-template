package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

func executeErrorHandlerRequest(testingContext *testing.T, inputError error) (int, string) {
	testingContext.Helper()

	application := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	application.Get("/test", func(ctx fiber.Ctx) error { return inputError })

	request := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	httpResponse, requestError := application.Test(request)
	if requestError != nil {
		testingContext.Fatalf("request failed: %v", requestError)
	}

	bodyBytes, readError := io.ReadAll(httpResponse.Body)
	if closeError := httpResponse.Body.Close(); closeError != nil {
		testingContext.Fatalf("failed to close body: %v", closeError)
	}
	if readError != nil {
		testingContext.Fatalf("failed to read body: %v", readError)
	}

	return httpResponse.StatusCode, string(bodyBytes)
}

func TestErrorHandler_DomainErrors(testingContext *testing.T) {
	testingContext.Parallel()

	tests := []struct {
		name               string
		inputError         error
		expectedStatusCode int
	}{
		{"not found returns 404", fmt.Errorf("wrapped: %w", domain.ErrNotFound), fiber.StatusNotFound},
		{"unauthorized returns 401", fmt.Errorf("wrapped: %w", domain.ErrUnauthorized), fiber.StatusUnauthorized},
		{"forbidden returns 403", fmt.Errorf("wrapped: %w", domain.ErrForbidden), fiber.StatusForbidden},
		{"conflict returns 409", fmt.Errorf("wrapped: %w", domain.ErrConflict), fiber.StatusConflict},
		{"validation returns 400", fmt.Errorf("wrapped: %w", domain.ErrValidation), fiber.StatusBadRequest},
		{"unknown error returns 500", errors.New("unexpected"), fiber.StatusInternalServerError},
	}

	for _, testCase := range tests {
		testCase := testCase
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			testingContext.Parallel()

			statusCode, _ := executeErrorHandlerRequest(testingContext, testCase.inputError)
			if statusCode != testCase.expectedStatusCode {
				testingContext.Fatalf("expected status %d, got %d", testCase.expectedStatusCode, statusCode)
			}
		})
	}
}

func TestErrorHandler_InternalErrorNotLeaked(testingContext *testing.T) {
	testingContext.Parallel()

	_, body := executeErrorHandlerRequest(testingContext, errors.New("secret database details"))
	if strings.Contains(body, "secret database details") {
		testingContext.Fatalf("internal error message leaked to client")
	}
}
