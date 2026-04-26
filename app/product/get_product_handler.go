package product

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetProductRequest struct {
	ID string `params:"id"`
}

type GetProductResponse struct {
	Product *repository.Product `json:"product"`
}

type GetProductHandler struct {
	repo repository.ProductRepository
}

func NewGetProductHandler(repo repository.ProductRepository) *GetProductHandler {
	return &GetProductHandler{repo: repo}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return nil, fiber.NewError(http.StatusBadRequest, "invalid id format")
	}

	product, err := h.repo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	return &GetProductResponse{Product: product}, nil
}
