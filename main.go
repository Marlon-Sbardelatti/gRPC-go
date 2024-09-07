package main

import (
	"context"
	"grpc/pb"
	"grpc/server/db"
	"grpc/server/models"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	db *gorm.DB
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	result := s.db.Create(&user)
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
	}

	response := &pb.UserResponse{
		Status: "User created!",
	}

	return response, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	id := req.Id

	var user models.User

	result := s.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Retorna um erro gRPC "NotFound" com uma mensagem descritiva
			return nil, status.Errorf(codes.NotFound, "User with ID %d not found", id)
		} else {
			// Retorna um erro gRPC "Internal" para erros de consulta internos
			return nil, status.Errorf(codes.Internal, "Error querying user: %v", result.Error)
		}
	}

	return &pb.User{
		Id:       int32(user.ID),
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func main() {
	db := db.InitDB()
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, &Server{db: db})

	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}

	log.Printf("Server is listening on port 5000...")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
