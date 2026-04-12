package http

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(usecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: usecase}
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
	// Determine if we should use secure cookies (production only)
	secure := os.Getenv("GIN_MODE") == "release"
	
	// Max age of 3 days (matching JWT expiration)
	maxAge := 3600 * 24 * 3

	// Set the cookie
	// Name, Value, MaxAge, Path, Domain, Secure, HttpOnly
	c.SetCookie(
		"lumina_auth", 
		token, 
		maxAge, 
		"/", 
		"", 
		secure, 
		true, // HttpOnly is critical for security
	)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authUsecase.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Set secure cookie
	h.setAuthCookie(c, res.Token)

	// In the response, we might still want to return user info
	// But we can clear the token from the body for extra security if we want
	res.Token = "hidden-in-cookie"

	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authUsecase.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Set secure cookie
	h.setAuthCookie(c, res.Token)

	// Optional: clear token from JSON body
	res.Token = "hidden-in-cookie"

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the cookie
	c.SetCookie("lumina_auth", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
