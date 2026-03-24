package grpc

import (
	pb "JWT/api/proto/user"
	"JWT/internal/auth"
	"JWT/internal/domain"
	"JWT/internal/service"
	"context"
	"time"
)

type UserGrpcHandler struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
	jwtManager  *auth.JWTManager
}

func NewUserGrpcHandler(userService service.UserService, jwtManager *auth.JWTManager) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (h *UserGrpcHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	createReq := domain.CreateUserRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Name:     req.GetName(),
	}

	user, err := h.userService.CreateUser(createReq)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserGrpcHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	loginReq := domain.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	_, err := h.userService.Login(loginReq)
	if err != nil {
		return nil, err
	}

	// Получение пользователя для генерации токена
	user, err := h.userService.GetUserByEmail(loginReq.Email)
	if err != nil {
		return nil, err
	}

	// Генерация JWT токен
	token, err := h.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (h *UserGrpcHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	userID := req.GetId()
	
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserGrpcHandler) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	users := h.userService.GetAllUsers()

	pbUsers := make([]*pb.UserResponse, len(users))
	for i, user := range users {
		pbUsers[i] = &pb.UserResponse{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &pb.GetAllUsersResponse{Users: pbUsers}, nil
}

func (h *UserGrpcHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	updateReq := domain.UpdateUserRequest{
		Email: req.GetEmail(),
		Name:  req.GetName(),
	}

	user, err := h.userService.UpdateUser(req.GetId(), updateReq)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserGrpcHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := h.userService.DeleteUser(req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{}, nil
}