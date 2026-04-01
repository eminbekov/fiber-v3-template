package response

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message string       `json:"message" example:"validation failed"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field" example:"username"`
	Message string `json:"message" example:"username is required"`
}
