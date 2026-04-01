package v1

// FileResponse describes an uploaded object and how to retrieve it.
type FileResponse struct {
	Key         string `json:"key" example:"019b1234-5678-7abc-8def-0123456789ab.jpg"`
	SignedURL   string `json:"signed_url" example:"/api/files/019b1234.jpg?token=...&expires=..."`
	PublicURL   string `json:"public_url"`
	ContentType string `json:"content_type" example:"image/png"`
}

func NewFileResponse(key string, signedURL string, publicURL string, contentType string) FileResponse {
	return FileResponse{
		Key:         key,
		SignedURL:   signedURL,
		PublicURL:   publicURL,
		ContentType: contentType,
	}
}
