package grpcserver

import (
	"net"

	pb "github.org/ulvinamazow/microservice_with_go/api/grpc/product"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func StartGRPCServer(addr string, service pb.ProductServiceServer) *grpc.Server {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		zap.L().Fatal("gRPC listener cannot created", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterProductServiceServer(grpcServer, service)

	zap.L().Info("gRPC server lstening", zap.String("addr", addr))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			zap.L().Fatal("gRPC server stopped", zap.Error(err))
		}
	}()
	return grpcServer
}

func ShutDownGRPCServer(server *grpc.Server) {
	if server == nil {
		return
	}

	server.GracefulStop()
}
