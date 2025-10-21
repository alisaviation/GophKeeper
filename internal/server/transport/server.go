package transport

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/alisaviation/GophKeeper/internal/server/transport/handlers"
	"github.com/alisaviation/GophKeeper/internal/server/transport/middleware"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
)

// Server gRPC сервер
type Server struct {
	server *grpc.Server
	config Config
}

// Config конфигурация gRPC сервера
type Config struct {
	Port int `yaml:"port" env:"GRPC_PORT" default:"50051"`
}

// NewServer создает новый gRPC сервер
func NewServer(
	authService *app.AuthService,
	dataService *app.DataService,
	config Config,
) *Server {
	authInterceptor := middleware.NewAuthInterceptor(authService)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.Unary(),
		),
	)

	authHandler := handlers.NewAuthHandler(authService)
	secretHandler := handlers.NewSecretHandler(dataService)

	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	pb.RegisterSecretServiceServer(grpcServer, secretHandler)

	reflection.Register(grpcServer)

	return &Server{
		server: grpcServer,
		config: config,
	}
}

// Start запускает gRPC сервер
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("gRPC server listening on :%d\n", s.config.Port)
	return s.server.Serve(lis)
}

// Stop останавливает gRPC сервер
func (s *Server) Stop() {
	s.server.GracefulStop()
}
