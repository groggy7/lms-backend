package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware restricts access to users with specified roles
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No role found"})
			c.Abort()
			return
		}

		userRole := role.(string)
		hasRole := slices.Contains(requiredRoles, userRole)

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
