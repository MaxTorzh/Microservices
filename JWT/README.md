# User Service with JWT Authentication

Микросервис для управления пользователями с поддержби двух протоколов: **REST API** и **gRPC**. Реализована JWT аутентификация для защиты эндпоинтов.

## 📋 Содержание

- [Архитектура](#архитектура)
- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Установка и запуск](#установка-и-запуск)
- [API Endpoints](#api-endpoints)
  - [REST API](#rest-api)
  - [gRPC API](#grpc-api)
- [Аутентификация](#аутентификация)
- [Тестирование](#тестирование)
- [Docker](#docker)

## 🏗 Архитектура

┌─────────────────────────────────────────────────────────────────────────────┐
│ User Service │
├─────────────────────────────────────────────────────────────────────────────┤
│ │
│ ┌──────────────────────────────────────────────────────────────────────┐ │
│ │ HTTP Server (port 8080) │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Публичные: │ │ │
│ │ │ POST /users/register → регистрация │ │ │
│ │ │ POST /users/login → получение JWT │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Защищенные (требуют JWT в Authorization header): │ │ │
│ │ │ GET /users → все пользователи │ │ │
│ │ │ GET /users/{id} → пользователь по ID │ │ │
│ │ │ PUT /users/{id} → обновление │ │ │
│ │ │ DELETE /users/{id} → удаление │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
│ │
│ ┌──────────────────────────────────────────────────────────────────────┐ │
│ │ gRPC Server (port 9090) │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Публичные методы: │ │ │
│ │ │ CreateUser → регистрация │ │ │
│ │ │ Login → получение JWT │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Защищенные методы (требуют JWT в metadata): │ │ │
│ │ │ GetUser, GetAllUsers, UpdateUser, DeleteUser │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ AuthInterceptor → проверяет JWT для всех защищенных методов │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
│ │
│ ┌──────────────────────────────────────────────────────────────────────┐ │
│ │ Общие компоненты │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Service Layer → бизнес-логика, хеширование паролей │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Repository → in-memory хранение пользователей │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ │ ┌────────────────────────────────────────────────────────────────┐ │ │
│ │ │ JWT Manager → генерация и валидация JWT токенов │ │ │
│ │ └────────────────────────────────────────────────────────────────┘ │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘

## 🛠 Технологии

| Технология           | Назначение                           |
| -------------------- | ------------------------------------ |
| **Go 1.26**          | Язык программирования                |
| **gRPC**             | Высокопроизводительный RPC фреймворк |
| **Protocol Buffers** | Сериализация данных                  |
| **JWT (jwt-go)**     | Аутентификация и авторизация         |
| **net/http**         | REST API                             |
| **Docker**           | Контейнеризация                      |

## 📁 Структура проекта

JWT/
├── api/
│ └── proto/
│ └── user/
│ ├── user.proto # gRPC прото файл
│ ├── user.pb.go # Сгенерированные структуры
│ └── user_grpc.pb.go # Сгенерированный gRPC код
├── cmd/
│ ├── server/
│ │ └── main.go # Точка входа сервера
│ └── client/
│ └── main.go # gRPC клиент для тестирования
├── internal/
│ ├── auth/ # Аутентификация
│ │ ├── jwt.go # JWT генерация и валидация
│ │ └── interceptor.go # gRPC интерсептор
│ ├── domain/
│ │ └── user.go # Модели и ошибки
│ ├── handler/
│ │ ├── grpc/
│ │ │ └── user.go # gRPC обработчики
│ │ └── http/
│ │ └── user.go # HTTP обработчики
│ ├── repository/
│ │ └── memory/
│ │ └── repository.go # In-memory хранилище
│ └── service/
│ └── user.go # Бизнес-логика
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md

## 🚀 Установка и запуск

### Локальный запуск

```bash
# Клонировать репозиторий
git clone <repository-url>
cd JWT

# Установить зависимости
go mod download

# Сгенерировать protobuf код
protoc --proto_path=api/proto \
       --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       api/proto/user/user.proto

# Запустить сервер
go run cmd/server/main.go
```

## Запуск через Docker

```bash
# Собрать и запустить
docker-compose up -d

# Посмотреть логи
docker-compose logs -f

# Остановить
docker-compose down
```

## API ENDPOINTS

### REST API (http://localhost:8080)

```text
Method	Endpoint	Description	Auth
POST	/users/register	Регистрация нового пользователя	❌ Нет
POST	/users/login	Аутентификация и получение JWT	❌ Нет
GET	/users	Получение всех пользователей	✅ JWT
GET	/users/{id}	Получение пользователя по ID	✅ JWT
PUT	/users/{id}	Обновление пользователя	✅ JWT
DELETE	/users/{id}	Удаление пользователя	✅ JWT
GET	/health	Health check	❌ Нет
```

### gRPC API (http://localhost:9090)

```text
Method	Description	Auth
CreateUser	Регистрация пользователя	❌ Нет
Login	Аутентификация и получение JWT	❌ Нет
GetUser	Получение пользователя по ID	✅ JWT
GetAllUsers	Получение всех пользователей	✅ JWT
UpdateUser	Обновление пользователя	✅ JWT
DeleteUser	Удаление пользователя	✅ JWT
```

## Аутентификация

1. Регистрация
   Client ──CreateUser(email, password)──► Server
   │
   ├── Хеширует пароль (SHA-256)
   └── Сохраняет пользователя

2. Логин
   Client ──Login(email, password)──────► Server
   │
   ├── Проверяет пароль
   └── Генерирует JWT токен

   Server ──{token: "eyJ..."}───────────► Client

3. Защищенный запрос
   Client ──GetUser(id) + Bearer token──► Server
   │
   ├── AuthInterceptor проверяет токен
   ├── Проверяет подпись и срок действия
   └── Выполняет запрос

## JWT структура

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "uuid",
    "email": "user@example.com",
    "exp": 174425783,
    "iat": 1744339383,
    "nbf": 1744339383
  }
}
```

## Передача токена

### REST API

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/users
```

### gRPC API

```go
md := metadata.New(map[string]string{
    "authorization": "Bearer " + token,
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

## Тестирование

### Тестирование REST API

```bash
# 1. Регистрация
curl -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123","name":"John Doe"}'

# 2. Логин (получить токен)
curl -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# 3. Получить всех пользователей (с токеном)
TOKEN="eyJhbGciOiJIUzI1NiIs..."
curl http://localhost:8080/users \
  -H "Authorization: Bearer $TOKEN"

# 4. Попробовать без токена (должна быть ошибка)
curl http://localhost:8080/users
```

### Тестирование gRPC

```bash
# Запустить gRPC клиент
go run cmd/client/main.go
```

## Полный тестовый скрипт

```bash
#!/bin/bash
# test-rest.sh

echo "=== Testing REST API ==="

# Регистрация
echo -e "\n1. Registering user..."
curl -s -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","name":"Test User"}'

# Логин
echo -e "\n\n2. Logging in..."
LOGIN_RESP=$(curl -s -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}')
echo $LOGIN_RESP

TOKEN=$(echo $LOGIN_RESP | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# Защищенный запрос
echo -e "\n3. Getting all users (with token)..."
curl -s http://localhost:8080/users -H "Authorization: Bearer $TOKEN"

# Без токена
echo -e "\n\n4. Without token (should fail)..."
curl -s http://localhost:8080/users

echo -e "\n\n=== Test completed ==="
```
