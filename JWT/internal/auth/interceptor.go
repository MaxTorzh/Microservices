package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Список методов, которые не требуют аутентификации
var publicMethods = map[string]bool{
	"/user.UserService/CreateUser": true,
	"/user.UserService/Login":      true,
}

type AuthInterceptor struct {
	jwtManager *JWTManager
}

func NewAuthInterceptor(jwtManager *JWTManager) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
	}
}

// Unary интерсептор для проверки аутентификации
func (i *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Проверка аутентификации
		if i.isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Извлечение токена из метаданных
		token, err := i.extractToken(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing token: %v", err)
		}

		// Валидация токена
		claims, err := i.jwtManager.Verify(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Добавление информации о пользователе в контекст
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)

		return handler(ctx, req)
	}
}

// Проверка на публичный метод
func (i *AuthInterceptor) isPublicMethod(method string) bool {
	return publicMethods[method]
}

// Извлечение токена из gRPC метаданных
func (i *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := authHeaders[0]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	return parts[1], nil
}