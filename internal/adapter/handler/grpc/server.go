package grpc

import (
	"fmt"
	"net"

	"DBGS_SOBERANO_BACKEND/internal/adapter/handler/grpc/interceptors"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func NewServer(port string, authInterceptor *interceptors.AuthInterceptor) *Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	}

	srv := grpc.NewServer(opts...)

	return &Server{
		grpcServer: srv,
		port:       port,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("error al iniciar listener TCP: %w", err)
	}

	fmt.Printf(" Servidor gRPC del DBGS iniciado en el puerto :%s\n", s.port)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}