package v1

import (
	"fmt"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	requestDTO "github.com/eminbekov/fiber-v3-template/internal/dto/request"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	responseV1 "github.com/eminbekov/fiber-v3-template/internal/dto/response/v1"
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/gofiber/fiber/v3"
)

type FileHandler struct {
	fileService *service.FileService
	translator  *i18n.Translator
}

func NewFileHandler(fileService *service.FileService, translator *i18n.Translator) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		translator:  translator,
	}
}

// Upload stores a multipart file and returns signed and public URLs.
//
// @Summary      Upload file
// @Description  Accepts multipart form field "file" and returns object key and URLs.
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file    true  "File to upload"
// @Param        note  formData  string  false "Optional note"
// @Success      201   {object}  response.Response
// @Failure      400   {object}  response.ErrorResponse
// @Failure      401   {object}  response.ErrorResponse
// @Failure      403   {object}  response.ErrorResponse
// @Failure      500   {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /v1/files [post]
func (handler *FileHandler) Upload(ctx fiber.Ctx) error {
	language := extractLanguage(ctx)

	var metadata requestDTO.UploadFileRequest
	metadata.Note = ctx.FormValue("note")
	fieldErrors := requestDTO.ValidateDTO(&metadata)
	if len(fieldErrors) > 0 {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "general.validation_failed"),
				Details: fieldErrors,
			},
		}); jsonError != nil {
			return fmt.Errorf("fileHandler.Upload: %w", jsonError)
		}
		return nil
	}

	fileHeader, formError := ctx.FormFile("file")
	if formError != nil || fileHeader == nil {
		return domain.ErrValidation
	}

	uploadedFile, openError := fileHeader.Open()
	if openError != nil {
		return fmt.Errorf("fileHandler.Upload: open multipart file: %w", openError)
	}
	defer func() { _ = uploadedFile.Close() }()

	contentType := fileHeader.Header.Get("Content-Type")
	objectKey, uploadError := handler.fileService.Upload(ctx.Context(), uploadedFile, fileHeader.Filename, contentType)
	if uploadError != nil {
		return fmt.Errorf("fileHandler.Upload: %w", uploadError)
	}

	signedURL, signedError := handler.fileService.SignedDownloadURL(ctx.Context(), objectKey)
	if signedError != nil {
		return fmt.Errorf("fileHandler.Upload: %w", signedError)
	}

	publicURL := handler.fileService.PublicURL(objectKey)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if jsonError := ctx.Status(fiber.StatusCreated).JSON(response.Response{
		Data: responseV1.NewFileResponse(objectKey, signedURL, publicURL, contentType),
	}); jsonError != nil {
		return fmt.Errorf("fileHandler.Upload: %w", jsonError)
	}
	return nil
}

// Download streams a stored object after signed URL validation middleware.
//
// @Summary      Download file
// @Description  Returns file bytes when token and expires are valid.
// @Tags         Files
// @Produce      octet-stream
// @Param        filename  path      string  true  "Object key / filename"
// @Param        token     query     string  true  "HMAC token"
// @Param        expires   query     int64   true  "Unix expiry"
// @Success      200       {file}    file    "Binary content"
// @Failure      403       {object}  response.ErrorResponse
// @Failure      404       {object}  response.ErrorResponse
// @Router       /files/{filename} [get]
func (handler *FileHandler) Download(ctx fiber.Ctx) error {
	filename := ctx.Params("filename")
	reader, contentType, openError := handler.fileService.Open(ctx.Context(), filename)
	if openError != nil {
		return fmt.Errorf("fileHandler.Download: %w", openError)
	}
	defer func() { _ = reader.Close() }()

	ctx.Set("Content-Type", contentType)
	ctx.Set("Cache-Control", "private, no-store")

	if streamError := ctx.SendStream(reader); streamError != nil {
		return fmt.Errorf("fileHandler.Download: %w", streamError)
	}
	return nil
}
