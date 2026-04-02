package request

import (
	"encoding/json"
	"testing"
)

func FuzzCreateUserRequest(fuzzContext *testing.F) {
	fuzzContext.Add(`{"username":"john","password":"StrongPass123!","full_name":"John Doe","phone":"+998901234567"}`)
	fuzzContext.Add(`{"username":"","password":"","full_name":"","phone":""}`)
	fuzzContext.Add(`{"username":"a","password":"1","full_name":"x","phone":"123"}`)

	fuzzContext.Fuzz(func(testingContext *testing.T, input string) {
		var createUserRequest CreateUserRequest
		if unmarshalError := json.Unmarshal([]byte(input), &createUserRequest); unmarshalError != nil {
			return
		}

		testingContext.Helper()
		testingContext.Cleanup(func() {})
		createUserRequest.Normalize()
		_ = ValidateDTO(&createUserRequest)
	})
}

func FuzzLoginRequest(fuzzContext *testing.F) {
	fuzzContext.Add(`{"username":"john","password":"StrongPass123!"}`)
	fuzzContext.Add(`{"username":"","password":""}`)
	fuzzContext.Add(`{"username":"\tadmin\t","password":"\npass\n"}`)

	fuzzContext.Fuzz(func(testingContext *testing.T, input string) {
		var loginRequest LoginRequest
		if unmarshalError := json.Unmarshal([]byte(input), &loginRequest); unmarshalError != nil {
			return
		}

		loginRequest.Normalize()
		_ = ValidateDTO(&loginRequest)
	})
}
