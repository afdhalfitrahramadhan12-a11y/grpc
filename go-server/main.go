package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"example.com/grpc-polyglot-lab/go-server/pb"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedUserServiceServer
}

func (s *server) GetUserProfile(
	ctx context.Context,
	req *pb.UserRequest,
) (*pb.UserResponse, error) {
	log.Printf("Menerima request untuk User ID: %d", req.GetUserId())

	if req.GetUserId() == 1 {
		return &pb.UserResponse{
			Name:  "Afdhal Fitrah Ramadhan",
			Email: "afdhalfitrahramadhan2@gmail.com",
			Phone: "083163859104",
		}, nil
	}

	return &pb.UserResponse{
		Name:  "User Tidak Ditemukan",
		Email: "-",
		Phone: "-",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Gagal membuka port: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})

	fmt.Println("gRPC Server (Go) berjalan di port :50051...")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
