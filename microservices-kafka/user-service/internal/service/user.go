package service

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "go.uber.org/zap"
    
    "user-service/internal/domain"
    "user-service/internal/kafka"
)

type UserRepository interface {
    Create(user domain.User) error
    GetByID(id string) (domain.User, error)
    GetByEmail(email string) (domain.User, error)
    GetAll() ([]domain.User, error)
    Update(id string, user domain.User) error
    Delete(id string) error
}

type UserService struct {
    repo     UserRepository
    producer *kafka.Producer
    logger   *zap.Logger
}

func NewUserService(repo UserRepository, producer *kafka.Producer, logger *zap.Logger) *UserService {
    return &UserService{
        repo:     repo,
        producer: producer,
        logger:   logger,
    }
}

func (s *UserService) CreateUser(ctx context.Context, req domain.CreateUserRequest) (domain.UserResponse, error) {
    if _, err := s.repo.GetByEmail(req.Email); err == nil {
        return domain.UserResponse{}, domain.ErrEmailExists
    }
    
    now := time.Now()
    user := domain.User{
        ID:        uuid.New().String(),
        Email:     req.Email,
        Name:      req.Name,
        CreatedAt: now,
        UpdatedAt: now,
    }
    
    if err := s.repo.Create(user); err != nil {
        s.logger.Error("Failed to create user", zap.Error(err))
        return domain.UserResponse{}, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Отправка события в kafka
    event := domain.UserEvent{
        EventType: "created",
        UserID:    user.ID,
        Email:     user.Email,
        Name:      user.Name,
        Timestamp: now,
    }
    
    if err := s.producer.SendUserEvent(event); err != nil {
        s.logger.Error("Failed to send Kafka event", 
            zap.Error(err),
            zap.String("user_id", user.ID),
            zap.String("event_type", "created"),
        )
    } else {
        s.logger.Info("User created event sent",
            zap.String("user_id", user.ID),
            zap.String("email", user.Email),
        )
    }
    
    return user.ToResponse(), nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (domain.UserResponse, error) {
    if id == "" {
        return domain.UserResponse{}, domain.ErrInvalidID
    }
    
    user, err := s.repo.GetByID(id)
    if err != nil {
        if err == domain.ErrUserNotFound {
            return domain.UserResponse{}, domain.ErrUserNotFound
        }
        s.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", id))
        return domain.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user.ToResponse(), nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.UserResponse, error) {
    users, err := s.repo.GetAll()
    if err != nil {
        s.logger.Error("Failed to get all users", zap.Error(err))
        return nil, fmt.Errorf("failed to get users: %w", err)
    }
    
    responses := make([]domain.UserResponse, len(users))
    for i, user := range users {
        responses[i] = user.ToResponse()
    }
    
    return responses, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (domain.UserResponse, error) {
    if id == "" {
        return domain.UserResponse{}, domain.ErrInvalidID
    }
    
    existingUser, err := s.repo.GetByID(id)
    if err != nil {
        if err == domain.ErrUserNotFound {
            return domain.UserResponse{}, domain.ErrUserNotFound
        }
        return domain.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
    }
    
    if existingUser.Email != req.Email {
        if userByEmail, err := s.repo.GetByEmail(req.Email); err == nil && userByEmail.ID != id {
            return domain.UserResponse{}, domain.ErrEmailExists
        }
    }
    
    updatedUser := domain.User{
        ID:        existingUser.ID,
        Email:     req.Email,
        Name:      req.Name,
        CreatedAt: existingUser.CreatedAt,
        UpdatedAt: time.Now(),
    }
    
    if err := s.repo.Update(id, updatedUser); err != nil {
        s.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", id))
        return domain.UserResponse{}, fmt.Errorf("failed to update user: %w", err)
    }
    
    // Отправка события об обновлении
    event := domain.UserEvent{
        EventType: "updated",
        UserID:    updatedUser.ID,
        Email:     updatedUser.Email,
        Name:      updatedUser.Name,
        Timestamp: time.Now(),
    }
    
    if err := s.producer.SendUserEvent(event); err != nil {
        s.logger.Error("Failed to send update event to Kafka",
            zap.Error(err),
            zap.String("user_id", id),
        )
    } else {
        s.logger.Info("User updated event sent", zap.String("user_id", id))
    }
    
    return updatedUser.ToResponse(), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
    if id == "" {
        return domain.ErrInvalidID
    }
    
    user, err := s.repo.GetByID(id)
    if err != nil {
        return err
    }
    
    if err := s.repo.Delete(id); err != nil {
        s.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id))
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    // Отправка события об удалении
    event := domain.UserEvent{
        EventType: "deleted",
        UserID:    user.ID,
        Email:     user.Email,
        Name:      user.Name,
        Timestamp: time.Now(),
    }
    
    if err := s.producer.SendUserEvent(event); err != nil {
        s.logger.Error("Failed to send delete event to Kafka",
            zap.Error(err),
            zap.String("user_id", id),
        )
    } else {
        s.logger.Info("User deleted event sent", zap.String("user_id", id))
    }
    
    return nil
}