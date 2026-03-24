package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "JWT/api/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	addr = flag.String("addr", "localhost:9090", "gRPC server address")
)

func main() {
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	ctx := context.Background()

	fmt.Println("=== gRPC Client Test with JWT Authentication ===")
	fmt.Println()

	// 1. Регистрация пользователя
	fmt.Println("1. Registering user...")
	registerResp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    "max@example.com",
		Password: "secret123",
		Name:     "Max Torj",
	})
	if err != nil {
		log.Fatalf("Register failed: %v", err)
	}
	fmt.Printf("Registered: ID=%s, Email=%s, Name=%s\n", registerResp.Id, registerResp.Email, registerResp.Name)

	// 2. Логин
	fmt.Println("\n2. Logging in...")
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Email:    "max@example.com",
		Password: "secret123",
	})
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Printf("Login successful! Token: %s...\n", loginResp.Token[:50])
	token := loginResp.Token

	// 3. Создание контекста с токеном для защищенных запросов
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	authCtx := metadata.NewOutgoingContext(ctx, md)

	// 4. Получение пользователя (защищенный метод)
	fmt.Println("\n3. Getting user (protected)...")
	getResp, err := client.GetUser(authCtx, &pb.GetUserRequest{Id: registerResp.Id})
	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}
	fmt.Printf("Got user: %s (%s)\n", getResp.Name, getResp.Email)

	// 5. Получение всех пользователей (защищенный метод)
	fmt.Println("\n4. Getting all users (protected)...")
	allResp, err := client.GetAllUsers(authCtx, &pb.GetAllUsersRequest{})
	if err != nil {
		log.Fatalf("GetAllUsers failed: %v", err)
	}
	fmt.Printf("Total users: %d\n", len(allResp.Users))
	for _, user := range allResp.Users {
		fmt.Printf("  - %s: %s (%s)\n", user.Id, user.Name, user.Email)
	}

	// 6. Тест без токена (должен вернуть ошибку)
	fmt.Println("\n5. Testing without token (should fail)...")
	_, err = client.GetUser(ctx, &pb.GetUserRequest{Id: registerResp.Id})
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// 7. Тест с неверным паролем
	fmt.Println("\n6. Testing with wrong password...")
	_, err = client.Login(ctx, &pb.LoginRequest{
		Email:    "max@example.com",
		Password: "wrongpassword",
	})
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	fmt.Println("\n=== Test completed ===")
}