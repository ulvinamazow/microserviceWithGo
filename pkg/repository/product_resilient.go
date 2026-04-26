package repository

import (
	"context"

	"github.org/ulvinamazow/microservice_with_go/pkg/resilience"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResilientProductRepository struct {
	base  ProductRepository
	retry resilience.RetryConfig
	cbCfg resilience.CircuitBreakerConfig
}

func NewResilientProductRepository(
	base ProductRepository,
	retryCfg resilience.RetryConfig,
	cbCfg resilience.CircuitBreakerConfig,
) *ResilientProductRepository {
	return &ResilientProductRepository{
		base:  base,
		retry: retryCfg,
		cbCfg: cbCfg,
	}
}

func (r *ResilientProductRepository) FindAll(ctx context.Context) ([]Product, error) {
	op := resilience.NewCircoutBreaker("FindAll", r.cbCfg, func(ctx context.Context) ([]Product, error) {
		return r.base.FindAll(ctx)
	})
	return resilience.Retry(ctx, r.retry, op)
}

func (r *ResilientProductRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*Product, error) {
	op := resilience.NewCircoutBreaker("FindByID", r.cbCfg, func(ctx context.Context) (*Product, error) {
		return r.base.FindByID(ctx, id)
	})
	return resilience.Retry(ctx, r.retry, op)
}

// Insert yeni məhsul əlavə edərkən qoruyucu tədbirləri tətbiq edir.
func (r *ResilientProductRepository) Insert(ctx context.Context, p *Product) (*Product, error) {
	op := resilience.NewCircoutBreaker("Insert", r.cbCfg, func(ctx context.Context) (*Product, error) {
		return r.base.Insert(ctx, p)
	})
	return resilience.Retry(ctx, r.retry, op)
}

// Update məhsulu yeniləyərkən qoruyucu tədbirləri tətbiq edir.
func (r *ResilientProductRepository) Update(ctx context.Context, id primitive.ObjectID, p *Product) error {
	op := resilience.NewCircoutBreaker("Update", r.cbCfg, func(ctx context.Context) (any, error) {
		return nil, r.base.Update(ctx, id, p)
	})
	_, err := resilience.Retry(ctx, r.retry, op)
	return err
}

// Delete məhsulu silərkən qoruyucu tədbirləri tətbiq edir.
func (r *ResilientProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	op := resilience.NewCircoutBreaker("Delete", r.cbCfg, func(ctx context.Context) (any, error) {
		return nil, r.base.Delete(ctx, id)
	})
	_, err := resilience.Retry(ctx, r.retry, op)
	return err
}
