package http

import (
	"net/http"

	appUpload "kafka-order-demo/backend/internal/application/upload"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService *appUpload.Service
}

func NewUploadHandler(uploadService *appUpload.Service) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, appUpload.ErrFileRequired.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	resp, err := h.uploadService.UploadImage(c.Request.Context(), appUpload.UploadRequest{
		File:        file,
		FileName:    fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: contentType,
		Folder:      c.DefaultPostForm("folder", "products"),
	})
	if err != nil {
		status := http.StatusInternalServerError
		if err == appUpload.ErrFileRequired || err == appUpload.ErrUnsupportedFile || err == appUpload.ErrFileTooLarge {
			status = http.StatusBadRequest
		}
		errorResponse(c, status, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Upload image successfully", toUploadHTTPResponse(resp))
}
