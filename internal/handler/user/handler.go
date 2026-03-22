package user

import (
	"encoding/json"
	"net/http"
	"simple_microservice/internal/domain"
	"simple_microservice/internal/repository/memory"
	"simple_microservice/internal/service"
)

type Handler struct {
	userService service.UserService
}

func NewHandler(userService service.UserService) Handler {
	return Handler{
		userService: userService,
	}
}

func (h Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/{id}", h.getUser)
	mux.HandleFunc("PUT /users/{id}", h.updateUser)
	mux.HandleFunc("DELETE /users/{id}", h.deleteUser)

	mux.HandleFunc("POST /users", h.createUser)
	mux.HandleFunc("GET /users", h.getAllUsers)
}

//Обработка POST
func (h Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" {
		http.Error(w, "Email and name are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		switch err {
		case memory.ErrMailExists:
			http.Error(w, "Email already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

//Обработка GET user{id}
func (h Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		switch err {
		case memory.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

//Обработка GET allUsers
func (h Handler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users := h.userService.GetAllUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

//Обработка PUT user{id}
func (h Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" {
		http.Error(w, "Email and name are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateUser(id, req)
	if err != nil {
		switch err {
		case memory.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case memory.ErrMailExists:
			http.Error(w, "Email already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

//Обработка DELETE user{id}
func (h Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		switch err {
		case memory.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
