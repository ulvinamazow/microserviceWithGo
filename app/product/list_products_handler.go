package product

import (
	"context"

	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
)

type ListProductsRequest struct {
}

type ListProductsResponse struct {
	Products []repository.Product `json:"products"`
}

type ListProductsHandler struct {
	repo repository.ProductRepository
}

func NewListProductsHandler(repo repository.ProductRepository) *ListProductsHandler {
	return &ListProductsHandler{repo: repo}
}

func (h *ListProductsHandler) Handle(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	products, err := h.repo.FindAll(ctx)

	if err != nil {
		return nil, err
	}

	return &ListProductsResponse{Products: products}, nil
}
