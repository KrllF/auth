package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	res "github.com/KrllF/auth/cmd/db"
	"github.com/KrllF/auth/internal/config"
	desc "github.com/KrllF/auth/pkg/auth_v1"

	"github.com/jackc/pgx/v4/pgxpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

type server struct {
	desc.UnimplementedAuthV1Server
	pool *pgxpool.Pool
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	resp, err := res.GetUser(ctx, s.pool, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	log.Printf("User %v", req.GetId())
	return resp, nil

}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {

	createresp, err := res.CreateUser(ctx, s.pool, req)
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
	flag.Parse()
	ctx := context.Background()

	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	grpcConfig, err := config.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to get grpc config: %v", err)
	}

	pgConfig, err := config.NewPGConfig()
	if err != nil {
		log.Fatalf("failed to get pg config: %v", err)
	}

	pool, err := pgxpool.Connect(ctx, pgConfig.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterAuthV1Server(s, &server{pool: pool})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
