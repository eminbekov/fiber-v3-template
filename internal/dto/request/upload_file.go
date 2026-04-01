package request

type UploadFileRequest struct {
	Note string `form:"note" validate:"omitempty,max=500"`
}
