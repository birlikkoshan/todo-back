package handlers

import (
	"errors"
	"net/http"

	"Worker/internal/auth"
	"Worker/internal/dto"
	"Worker/internal/service"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "session_id"

// AuthHandler handles login, register and logout.
type AuthHandler struct {
	sessions *auth.Store
	userSvc  *service.UserService
}

// NewAuthHandler returns a new AuthHandler.
func NewAuthHandler(sessions *auth.Store, userSvc *service.UserService) *AuthHandler {
	return &AuthHandler{sessions: sessions, userSvc: userSvc}
}

// Login godoc
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.LoginRequest  true  "Credentials"
// @Success      200   {object}  map[string]bool
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userSvc.ValidateCredentials(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	sessionID, err := h.sessions.Create(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	c.SetCookie(sessionCookieName, sessionID, 24*60*60, "/", "", false, true) // 24h, httpOnly
	c.JSON(http.StatusOK, gin.H{"ok": true, "user": dto.UserResponse{ID: user.ID, Username: user.Username}})
}

// Register godoc
// @Summary      Register
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.RegisterRequest  true  "Credentials"
// @Success      201   {object}  map[string]bool
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userSvc.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
			return
		}
		if errors.Is(err, service.ErrUsernameTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}
	sessionID, err := h.sessions.Create(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	c.SetCookie(sessionCookieName, sessionID, 24*60*60, "/", "", false, true) // 24h, httpOnly
	c.JSON(http.StatusCreated, gin.H{"ok": true, "user": dto.UserResponse{ID: user.ID, Username: user.Username}})
}

// Logout godoc
// @Summary      Logout
// @Tags         auth
// @Success      204
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err == nil && sessionID != "" {
		_ = h.sessions.Delete(c.Request.Context(), sessionID)
	}
	c.SetCookie(sessionCookieName, "", -1, "/", "", false, true)
	c.Status(http.StatusNoContent)
}
