package dto

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest is the JSON body for POST /auth/register.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=1,max=120"`
	Password string `json:"password" binding:"required,min=1"`
}

// UserResponse is returned when user info is needed (e.g. after login).
type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}
