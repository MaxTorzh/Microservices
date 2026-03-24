package httphandler

import (
	"JWT/internal/auth"
	"JWT/internal/domain"
	"JWT/internal/service"
	"context"
	"encoding/json"
	"net/http"
)

type Handler struct {
	userService service.UserService
	jwtManager  *auth.JWTManager
}

func NewHandler(userService service.UserService, jwtManager *auth.JWTManager) Handler {
	return Handler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (h Handler) RegisterRoutes(mux *http.ServeMux) {
	// Публичные маршруты
	mux.HandleFunc("POST /users/register", h.register)
	mux.HandleFunc("POST /users/login", h.login)
	
	// Защищенные маршруты (требуют JWT)
	mux.HandleFunc("GET /users/{id}", h.authenticate(h.getUser))
	mux.HandleFunc("GET /users", h.authenticate(h.getAllUsers))
	mux.HandleFunc("PUT /users/{id}", h.authenticate(h.updateUser))
	mux.HandleFunc("DELETE /users/{id}", h.authenticate(h.deleteUser))
}

// Проверка JWT
func (h Handler) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлечение токена из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Формат: "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := authHeader[7:]

		// Валидация токена
		claims, err := h.jwtManager.Verify(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Добавление информации о пользователе в контекст запроса
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		
		next(w, r.WithContext(ctx))
	}
}

func (h Handler) register(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Email, password and name are required", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка данных
	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Генерация токена
	token, err := h.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	user, err := h.userService.GetUserByID(id)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
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
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}