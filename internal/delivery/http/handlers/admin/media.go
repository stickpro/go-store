package admin

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/medium_response"
	"github.com/stickpro/go-store/internal/service/media"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// storeFile handles file uploads
//
//	@Summary		Upload file
//	@Description	Allows users to upload files of specific types (JPEG, PDF, WEBP)
//	@Tags			Files
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			document	formData	file	true	"File to upload"
//	@Success		200			{object}	response.Result[medium_response.MediumResponse]
//	@Failure		400			{object}	apierror.Errors "Unsupported file type"
//	@Failure		500			{object}	apierror.Errors "Internal server error"
//	@Router			/v1/media/upload [post]
func (h *Handler) storeFile(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	fileData, err := file.Open()
	if err != nil {
		return err
	}
	defer fileData.Close()

	buffer := make([]byte, 512)
	_, err = fileData.Read(buffer)
	if err != nil {
		return err
	}

	fileType := http.DetectContentType(buffer)
	if !allowedTypes[fileType] {
		err := errors.New("file type not allowed")
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnsupportedMediaType)
	}

	_, err = fileData.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	content, err := io.ReadAll(fileData)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9_.-]`)
	cleanName := re.ReplaceAllString(file.Filename, "_")
	ext := filepath.Ext(cleanName)
	name := strings.TrimSuffix(cleanName, ext)
	uniqueName := fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext)

	dto := media.SaveMediumDTO{
		Name:     file.Filename,
		Path:     "public/images/" + uniqueName,
		FileType: fileType,
		Size:     file.Size,
		Data:     content,
	}
	medium, err := h.services.MediaService.Save(c.Context(), dto)
	if err != nil {
		return err
	}
	mediumResponse := medium_response.NewFromModel(medium)
	return c.JSON(response.OkByData(mediumResponse))
}

func (h *Handler) initMediaRoutes(v1 fiber.Router) {
	m := v1.Group("/media")
	m.Post("/upload", h.storeFile)
}
