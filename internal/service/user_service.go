package service

import (
	"context"
	"errors"
	"strings"

	dom "Worker/internal/domain"
	"Worker/internal/repo"
	"Worker/internal/utils"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrUsernameTaken = errors.New("username already taken")

// UserService handles user auth logic.
type UserService struct {
	repo repo.UserRepo
}

// NewUserService returns a new UserService.
func NewUserService(repo repo.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// ValidateCredentials checks username and password; returns user if valid.
func (s *UserService) ValidateCredentials(ctx context.Context, username, password string) (dom.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return dom.User{}, ErrInvalidCredentials
	}
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dom.User{}, ErrInvalidCredentials
		}
		return dom.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return dom.User{}, ErrInvalidCredentials
	}
	return u, nil
}

// Register creates a new user with hashed password.
func (s *UserService) Register(ctx context.Context, username, password string) (dom.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return dom.User{}, ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return dom.User{}, err
	}
	u, err := s.repo.Create(ctx, username, string(hash))
	if err != nil {
		if utils.IsPGUniqueViolation(err) {
			return dom.User{}, ErrUsernameTaken
		}
		return dom.User{}, err
	}
	return u, nil
}
