package http

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type DocumentHandler struct {
	usecase domain.DocumentUsecase
}

func NewDocumentHandler(usecase domain.DocumentUsecase) *DocumentHandler {
	return &DocumentHandler{usecase: usecase}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No document provided"})
		return
	}
	defer file.Close()

	location, err := h.usecase.Upload(c.Request.Context(), file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Document uploaded successfully",
		"location": location,
	})
}

func (h *DocumentHandler) Download(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "anonymous"
	}

	// For this POC, we use a sample PDF
	inputPath := "./assets/sample.pdf"
	
	// Create assets dir if not exists
	if _, err := os.Stat("./assets"); os.IsNotExist(err) {
		os.Mkdir("./assets", 0755)
	}

	// Check if sample.pdf exists, if not create a dummy one
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		f, _ := os.Create(inputPath)
		f.WriteString("This is a sample PDF content.")
		f.Close()
	}

	watermarked, err := h.usecase.GetWatermarkedPDF(c.Request.Context(), inputPath, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to watermark PDF"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"watermarked_%s.pdf\"", userID))
	c.Header("Content-Type", "application/pdf")
	io.Copy(c.Writer, watermarked)
}
