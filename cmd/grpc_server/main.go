package main

import (
	"context"
	"fmt"
	"log"
	"net"

	res "github.com/KrllF/auth/cmd/db"
	desc "github.com/KrllF/auth/pkg/auth_v1"

	"github.com/jackc/pgx/v4/pgxpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	grpcPort = 50051
	dbDSN    = "host=localhost port=54321 dbname=auth user=auth-user password=auth-password sslmode=disable"
)

type server struct {
	desc.UnimplementedAuthV1Server
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	resp, err := res.GetUser(ctx, req.GetId())
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	log.Printf("User %v", req.GetId())
	return resp, nil

}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {

	if req.GetPassword() != req.GetPasswordConfirm() {
		log.Printf("Passwords do not match for user creation: name=%v", req.GetName())
		return nil, fmt.Errorf("passwords do not match")
	}

	createresp, err := res.CreateUser(ctx, req.GetName(), req.GetEmail(), req.GetPassword(), req.GetRole())
	if err != nil {
		log.Printf("Failed to create user in database: name=%v, email=%v, error=%v", req.GetName(), req.GetEmail(), err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created successfully: id=%v, name=%v, email=%v", createresp.Id, req.GetName(), req.GetEmail())
	return createresp, nil
}

func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("Delete user %v", req.GetId())
	return nil, nil
}

func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	log.Printf("Update user %v", req.GetId())
	return nil, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterAuthV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	res.DB = pool
	defer pool.Close()

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
