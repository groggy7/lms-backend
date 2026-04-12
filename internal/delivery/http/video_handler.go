package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type VideoHandler struct {
	videoUsecase domain.VideoUsecase
}

func NewVideoHandler(usecase domain.VideoUsecase) *VideoHandler {
	return &VideoHandler{videoUsecase: usecase}
}

func (h *VideoHandler) UploadChunk(c *gin.Context) {
	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No chunk found"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open chunk"})
		return
	}
	defer openedFile.Close()

	chunkIndex, _ := strconv.Atoi(c.PostForm("chunkIndex"))
	totalChunks, _ := strconv.Atoi(c.PostForm("totalChunks"))
	fileName := c.PostForm("fileName")
	uploadID := c.PostForm("uploadId")

	chunk := domain.Chunk{
		Index:       chunkIndex,
		TotalChunks: totalChunks,
		FileName:    fileName,
		UploadID:    uploadID,
		Data:        openedFile,
	}

	isComplete, err := h.videoUsecase.ProcessChunk(c.Request.Context(), chunk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if isComplete {
		c.JSON(http.StatusOK, gin.H{"message": "Upload complete"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Chunk received"})
	}
}
