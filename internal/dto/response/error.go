package response

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
