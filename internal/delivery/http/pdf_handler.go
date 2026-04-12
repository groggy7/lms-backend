package http

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type PDFHandler struct {
	usecase domain.PDFUsecase
}

func NewPDFHandler(usecase domain.PDFUsecase) *PDFHandler {
	return &PDFHandler{usecase: usecase}
}

func (h *PDFHandler) Download(c *gin.Context) {
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
