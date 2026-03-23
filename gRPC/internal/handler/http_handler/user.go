package httphandler

import (
	"encoding/json"
	"gRPC/internal/domain"
	"gRPC/internal/service"
	"net/http"
)

type Handler struct {
	userService service.UserService
}

func NewHandler(userService service.UserService) Handler {
	return Handler{userService: userService}
}

func (h Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/{id}", h.getUser)
	mux.HandleFunc("PUT /users/{id}", h.updateUser)
	mux.HandleFunc("DELETE /users/{id}", h.deleteUser)
	mux.HandleFunc("POST /users", h.createUser)
	mux.HandleFunc("GET /users", h.getAllUsers)
}

func (h Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		switch err {
		case domain.ErrEmailExists:
			http.Error(w, "Email already exists", http.StatusConflict)
		case domain.ErrInvalidRequest:
			http.Error(w, "Email and name are required", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, err := h.userService.GetUser(id)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case domain.ErrEmptyID:
			http.Error(w, "User ID is required", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h Handler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users := h.userService.GetAllUsers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateUser(id, req)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case domain.ErrEmailExists:
			http.Error(w, "Email already exists", http.StatusConflict)
		case domain.ErrInvalidRequest:
			http.Error(w, "Email and name are required", http.StatusBadRequest)
		case domain.ErrEmptyID:
			http.Error(w, "User ID is required", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.userService.DeleteUser(id); err != nil {
		switch err {
		case domain.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case domain.ErrEmptyID:
			http.Error(w, "User ID is required", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}