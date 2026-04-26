package user

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
)

// SignupRequest istifadəçi qeydiyyatı üçün JSON gövdəsi.
type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupResponse qeydiyyatdan sonra qaytarılacaq məlumat.
type SignupResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// SignupHandler qeydiyyat məntiqi.
type SignupHandler struct {
	repo repository.UserRepository
}

// NewSignupHandler yeni handler yaradır.
func NewSignupHandler(repo repository.UserRepository) *SignupHandler {
	return &SignupHandler{repo: repo}
}

// Handle qeydiyyatı icra edir.
func (h *SignupHandler) Handle(ctx context.Context, req *SignupRequest) (*SignupResponse, error) {
	// İstifadəçi adı unikal olmalıdır
	existing, _ := h.repo.FindByUsername(ctx, req.Username)
	if existing != nil {
		return nil, fiber.NewError(http.StatusConflict, "istifadəçi adı artıq mövcuddur")
	}

	user := &repository.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	created, err := h.repo.Create(ctx, user)
	if err != nil {
		// Əgər eyni username/email varsa, MongoDB səhvi ola bilər
		if mongo.IsDuplicateKeyError(err) {
			return nil, fiber.NewError(http.StatusConflict, "istifadəçi artıq mövcuddur")
		}
		return nil, err
	}

	return &SignupResponse{
		ID:       created.ID.Hex(),
		Username: created.Username,
		Email:    created.Email,
	}, nil
}
