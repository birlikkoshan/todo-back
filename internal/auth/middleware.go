package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "session_id"

const contextKeyUserID = "user_id"

// UserIDFromContext returns the current user ID set by RequireSession. 0 if not set.
func UserIDFromContext(c *gin.Context) int64 {
	v, ok := c.Get(contextKeyUserID)
	if !ok {
		return 0
	}
	id, ok := v.(int64)
	if !ok {
		return 0
	}
	return id
}

// RequireSession returns a middleware that checks for a valid session cookie
// and sets the current user ID in context. If missing or invalid, responds with 401.
func RequireSession(sessions *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(sessionCookieName)
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		userID, ok := sessions.GetUserID(c.Request.Context(), sessionID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		c.Set(contextKeyUserID, userID)
		c.Next()
	}
}
