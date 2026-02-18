package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "session_id"

// RequireSession returns a middleware that checks for a valid session cookie.
// If missing or invalid, responds with 401 Unauthorized.
func RequireSession(sessions *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(sessionCookieName)
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		ok, err := sessions.Exists(c.Request.Context(), sessionID)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		c.Next()
	}
}
