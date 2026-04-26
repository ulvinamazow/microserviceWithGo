package product

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteProductRequest struct {
	ID string `params:"id"`
}

type DeleteProductResponse struct {
	Success bool `json:"success"`
}

type DeleteProductHandler struct {
	repo repository.ProductRepository
}

func NewDeleteProducHandler(repo repository.ProductRepository) *DeleteProductHandler {
	return &DeleteProductHandler{repo: repo}
}

func (h *DeleteProductHandler) Handle(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return nil, fiber.NewError(http.StatusBadRequest, "invalid id format")
	}

	if err := h.repo.Delete(ctx, objID); err != nil {
		return nil, err
	}
	return &DeleteProductResponse{Success: true}, nil
}
