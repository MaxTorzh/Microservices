package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "gRPC/api/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = flag.String("addr", "localhost:9090", "gRPC server address")

func main() {
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	ctx := context.Background()

	fmt.Println("=== gRPC Client Test ===")

	// Create user
	resp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Email: "Max@example.com",
		Name:  "Maksim Torzhkov",
	})
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}
	fmt.Printf("Created: ID=%s, Email=%s, Name=%s\n", resp.Id, resp.Email, resp.Name)

	// Get user
	getResp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: resp.Id})
	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}
	fmt.Printf("Retrieved: ID=%s, Email=%s, Name=%s\n", getResp.Id, getResp.Email, getResp.Name)

	// Get all users
	allResp, err := client.GetAllUsers(ctx, &pb.GetAllUsersRequest{})
	if err != nil {
		log.Fatalf("GetAllUsers failed: %v", err)
	}
	fmt.Printf("Total users: %d\n", len(allResp.Users))

	fmt.Println("\n=== Test completed ===")
}