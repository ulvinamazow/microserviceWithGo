package grpcserver

import (
	"context"

	pb "github.org/ulvinamazow/microservice_with_go/api/grpc/product"
	"github.org/ulvinamazow/microservice_with_go/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductService struct {
	pb.UnimplementedProductServiceServer
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.Id)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}

	product, err := s.repo.FindByID(ctx, objID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	return &pb.GetProductResponse{
		Product: &pb.Product{
			Id:        product.ID.Hex(),
			Name:      product.Name,
			Price:     product.Price,
			CreatedAt: product.CreatedAt.String(),
		},
	}, nil
}

func (s *ProductService) ListProduct(ctx context.Context, req *pb.ListProductRequest) (*pb.ListProoductResponse, error) {
	products, err := s.repo.FindAll(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "products could not be delivered")
	}

	var pbProducts []*pb.Product

	for _, p := range products {
		pbProducts = append(pbProducts, &pb.Product{
			Id:        p.ID.Hex(),
			Name:      p.Name,
			Price:     p.Price,
			CreatedAt: p.CreatedAt.String(),
		})
	}

	return &pb.ListProoductResponse{Products: pbProducts}, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	product := &repository.Product{
		Name:  req.Name,
		Price: req.Price,
	}

	created, err := s.repo.Insert(ctx, product)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "product cannot created")
	}

	return &pb.CreateProductResponse{
		Product: &pb.Product{
			Id:        created.ID.Hex(),
			Name:      created.Name,
			Price:     created.Price,
			CreatedAt: created.CreatedAt.String(),
		},
	}, nil

}

func (s *ProductService) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.Id)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}

	if err := s.repo.Delete(ctx, objID); err != nil {
		return nil, status.Errorf(codes.Internal, "product cannot deleted")
	}

	return &pb.DeleteProductResponse{Succes: true}, nil
}
