package grpc

import (
	"context"
	"time"

	pb "gRPC/api/proto/user"
	"gRPC/internal/domain"
	"gRPC/internal/service"
)

type UserGrpcHandler struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserGrpcHandler(userService service.UserService) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
	}
}

func (h *UserGrpcHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	createReq := domain.CreateUserRequest{
		Email: req.GetEmail(),
		Name:  req.GetName(),
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

func (h *UserGrpcHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := h.userService.GetUser(req.GetId())
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