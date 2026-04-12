package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allowing all origins for POC
	},
}

type ProgressHandler struct {
	usecase domain.ProgressUsecase
}

func NewProgressHandler(usecase domain.ProgressUsecase) *ProgressHandler {
	return &ProgressHandler{usecase: usecase}
}

func (h *ProgressHandler) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade connection: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		var msg domain.UserProgress
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("Error reading JSON: %v\n", err)
			break
		}

		// Update progress in the background
		err = h.usecase.UpdateProgress(c.Request.Context(), msg)
		if err != nil {
			fmt.Printf("Error updating progress: %v\n", err)
		}
	}
}
