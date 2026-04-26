package user

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.org/ulvinamazow/microservice_with_go/pkg/auth"
	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginHandler struct {
	repo      repository.UserRepository
	jwtSecret string
	jwtExpMin int
}

func NewLoginHandler(repo repository.UserRepository, jwtSecret string, jwtExpMin int) *LoginHandler {
	return &LoginHandler{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpMin: jwtExpMin,
	}
}

func (h *LoginHandler) Handle(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := h.repo.FindByUsername(ctx, req.Username)

	if err != nil {
		return nil, fiber.NewError(http.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fiber.NewError(http.StatusUnauthorized, "invalid credentials")
	}

	token, err := auth.GenereteToken(user.ID.Hex(), h.jwtSecret, h.jwtExpMin)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}
