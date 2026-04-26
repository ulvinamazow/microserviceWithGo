package product

import (
	"context"
	"encoding/json"
	"time"

	"github.org/ulvinamazow/microservice_with_go/pkg/messaging"
	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
	"go.uber.org/zap"
)

type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type CreateProductResponse struct {
	Product *repository.Product `json:"product"`
}

type CreateProductHandler struct {
	repo     repository.ProductRepository
	producer *messaging.Producer
}

func NewCreateProductHandler(repo repository.ProductRepository, producer *messaging.Producer) *CreateProductHandler {
	return &CreateProductHandler{
		repo:     repo,
		producer: producer,
	}
}

func (h *CreateProductHandler) Handle(ctx context.Context, req *CreateProductRequest) (*CreateProductResponse, error) {
	product := &repository.Product{
		Name:  req.Name,
		Price: req.Price,
	}
	created, err := h.repo.Insert(ctx, product)
	if err != nil {
		return nil, err
	}

	event, _ := json.Marshal(map[string]interface{}{
		"event": "product_created",
		"id":    created.ID.Hex(),
		"name":  created.Name,
		"price": created.Price,
	})

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.producer.Send(sendCtx, []byte(created.ID.Hex()), event); err != nil {
			zap.L().Error("Kafka message not sent, but product created", zap.Error(err))
		}
	}()

	return &CreateProductResponse{Product: created}, nil
}
