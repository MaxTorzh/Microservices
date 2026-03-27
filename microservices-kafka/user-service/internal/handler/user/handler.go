package user

import (
    "encoding/json"
    "net/http"
    
    "go.uber.org/zap"
    
    "user-service/internal/domain"
    "user-service/internal/service"
)

type Handler struct {
    userService *service.UserService
    logger      *zap.Logger
}

func NewHandler(userService *service.UserService, logger *zap.Logger) *Handler {
    return &Handler{
        userService: userService,
        logger:      logger,
    }
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("POST /users", h.CreateUser)
    mux.HandleFunc("GET /users/{id}", h.GetUser)
    mux.HandleFunc("GET /users", h.GetAllUsers)
    mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
    mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req domain.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Invalid request body", zap.Error(err))
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.Email == "" || req.Name == "" {
        http.Error(w, "Email and name are required", http.StatusBadRequest)
        return
    }
    
    user, err := h.userService.CreateUser(r.Context(), req)
    if err != nil {
        switch err {
        case domain.ErrEmailExists:
            http.Error(w, "Email already exists", http.StatusConflict)
        default:
            h.logger.Error("Failed to create user", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    user, err := h.userService.GetUser(r.Context(), id)
    if err != nil {
        switch err {
        case domain.ErrUserNotFound:
            http.Error(w, "User not found", http.StatusNotFound)
        case domain.ErrInvalidID:
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
        default:
            h.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", id))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.userService.GetAllUsers(r.Context())
    if err != nil {
        h.logger.Error("Failed to get all users", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    var req domain.UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Invalid request body", zap.Error(err))
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.Email == "" || req.Name == "" {
        http.Error(w, "Email and name are required", http.StatusBadRequest)
        return
    }
    
    user, err := h.userService.UpdateUser(r.Context(), id, req)
    if err != nil {
        switch err {
        case domain.ErrUserNotFound:
            http.Error(w, "User not found", http.StatusNotFound)
        case domain.ErrEmailExists:
            http.Error(w, "Email already exists", http.StatusConflict)
        case domain.ErrInvalidID:
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
        default:
            h.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", id))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    if err := h.userService.DeleteUser(r.Context(), id); err != nil {
        switch err {
        case domain.ErrUserNotFound:
            http.Error(w, "User not found", http.StatusNotFound)
        case domain.ErrInvalidID:
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
        default:
            h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}