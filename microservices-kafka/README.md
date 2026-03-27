# Микросервисная архитектура с Kafka и PostgreSQL

## Описание проекта

Проект представляет собой микросервисную архитектуру, состоящую из двух сервисов, взаимодействующих через Apache Kafka:

1. **User Service** - сервис управления пользователями (CRUD операции)
2. **Notification Service** - сервис отправки уведомлений

При создании, обновлении или удалении пользователя, User Service отправляет событие в Kafka, которое обрабатывается Notification Service для отправки соответствующего уведомления.

## Технологии

- **Go 1.26** - язык программирования
- **PostgreSQL 15** - реляционная база данных
- **Apache Kafka** - брокер сообщений для асинхронного взаимодействия
- **Docker & Docker Compose** - контейнеризация и оркестрация
- **Zap** - структурированное логирование
- **Sarama** - Kafka клиент для Go

## Архитектура

```text
┌─────────────────┐ ┌──────────────┐ ┌─────────────────────┐
│ │ │ │ │ │
│ User Service │────▶│ Kafka │────▶│ Notification Service│
│ (Port 8080) │ │ (Port 9092) │ │ (Port 8081) │
│ │ │ │ │ │
└────────┬────────┘ └──────────────┘ └─────────────────────┘
│ │
▼ │
┌─────────────────┐ │
│ │ │
│ PostgreSQL │◀────────────────────────────────────┘
│ (Port 5432) │
│ │
└─────────────────┘
```

### События

При операциях с пользователями генерируются следующие события:

- `created` - при создании пользователя → отправляется приветственное уведомление
- `updated` - при обновлении пользователя → отправляется уведомление об изменении профиля
- `deleted` - при удалении пользователя → отправляется прощальное уведомление

## Структура проекта

```text
microservices-kafka/
├── user-service/ # Сервис управления пользователями
│ ├── cmd/server/
│ │ └── main.go # Точка входа
│ ├── internal/
│ │ ├── domain/ # Модели данных
│ │ │ ├── errors.go
│ │ │ └── user.go
│ │ ├── handler/user/ # HTTP обработчики
│ │ │ └── handler.go
│ │ ├── repository/
│ │ │ └── postgres/ # PostgreSQL репозиторий
│ │ │ ├── repository.go
│ │ │ └── migrations/
│ │ │ ├── 001_create_users_table.up.sql
│ │ │ └── 001_create_users_table.down.sql
│ │ ├── service/ # Бизнес-логика
│ │ │ └── user.go
│ │ └── kafka/ # Kafka producer
│ │ └── producer.go
│ ├── pkg/logger/ # Логирование
│ │ └── logger.go
│ ├── Dockerfile
│ ├── .env
│ └── go.mod
│
├── notification-service/ # Сервис уведомлений
│ ├── cmd/server/
│ │ └── main.go
│ ├── internal/
│ │ ├── domain/ # Модели данных
│ │ │ └── notification.go
│ │ ├── handler/notification/ # HTTP обработчики
│ │ │ └── handler.go
│ │ ├── service/ # Бизнес-логика
│ │ │ └── notifier.go
│ │ └── kafka/ # Kafka consumer
│ │ └── consumer.go
│ ├── pkg/logger/
│ │ └── logger.go
│ ├── Dockerfile
│ ├── .env
│ └── go.mod
│
├── docker-compose.yml # Оркестрация контейнеров
├── .gitignore
└── README.md
```

## Установка и запуск

### Требования

- Docker (версия 20.10+)
- Docker Compose (версия 2.0+)
- Git

### Клонирование репозитория

```bash
git clone <repository-url>
cd microservices-kafka
Запуск приложения
bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Просмотр логов конкретного сервиса
docker-compose logs -f user-service
docker-compose logs -f notification-service
Остановка
bash
# Остановка сервисов
docker-compose down

# Остановка с удалением томов (очистка БД)
docker-compose down -v
API Эндпоинты
User Service (http://localhost:8080)
Метод	Эндпоинт	Описание
POST	/users	Создание пользователя
GET	/users	Получение всех пользователей
GET	/users/{id}	Получение пользователя по ID
PUT	/users/{id}	Обновление пользователя
DELETE	/users/{id}	Удаление пользователя
GET	/health	Проверка здоровья сервиса
GET	/ready	Проверка готовности сервиса
Notification Service (http://localhost:8081)
Метод	Эндпоинт	Описание
GET	/health	Проверка здоровья сервиса
GET	/ready	Проверка готовности сервиса
GET	/stats	Статистика обработки уведомлений
GET	/metrics	Метрики сервиса
GET	/last-notification	Последнее отправленное уведомление
```

## Примеры запросов

### Создание пользователя

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "ivan@example.com",
    "name": "Ivan Petrov"
  }'
```

### Ответ:

```json
{
  "id": "89d3c3ee-9903-4107-98a3-ef90f738270a",
  "email": "ivan@example.com",
  "name": "Ivan Petrov",
  "created_at": "2026-03-26T16:36:44.608195Z",
  "updated_at": "2026-03-26T16:36:44.608195Z"
}
```

### Получение всех пользователей

```bash
curl http://localhost:8080/users
```

### Обновление пользователя

```bash
curl -X PUT http://localhost:8080/users/89d3c3ee-9903-4107-98a3-ef90f738270a \
  -H "Content-Type: application/json" \
  -d '{
    "email": "ivan.updated@example.com",
    "name": "Ivan Petrovich"
  }'
```

### Удаление пользователя

```bash
curl -X DELETE http://localhost:8080/users/89d3c3ee-9903-4107-98a3-ef90f738270a
```

### Проверка статистики уведомлений

```bash
curl http://localhost:8081/stats
```

### Ответ:

```json
{
  "status": "running",
  "service": "notification-service",
  "processed_count": 3,
  "error_count": 0,
  "notifier_sent_count": 3,
  "notifier_error_count": 0,
  "last_event_type": "deleted",
  "last_event_time": "2026-03-26T16:36:44Z",
  "uptime_seconds": 230.05
}
```

# Логирование

### Оба сервиса используют структурированное логирование в формате JSON. Пример логов:

## User Service:

```json
{
  "level": "info",
  "timestamp": "2026-03-26T16:36:44.608Z",
  "message": "User created event sent",
  "user_id": "89d3c3ee-9903-4107-98a3-ef90f738270a",
  "email": "ivan@example.com"
}
```

## Notification Service:

```json
{
  "level": "info",
  "timestamp": "2026-03-26T16:36:44.627Z",
  "message": "Notification sent",
  "notification_id": "b88cc7fd-28e8-4d8c-8bbb-56eedbf2b575",
  "type": "welcome",
  "user_id": "89d3c3ee-9903-4107-98a3-ef90f738270a",
  "duration": 0.000000233
}
```

# Конфигурация

### Переменные окружения

## User Service (.env)

```env
# Server
HTTP_PORT=8080
LOG_LEVEL=info

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=user_service

# Kafka
KAFKA_BROKER=kafka:9092
KAFKA_TOPIC=user-events
Notification Service (.env)
```

```env
# Server
HTTP_PORT=8081
LOG_LEVEL=info

# Kafka
KAFKA_BROKER=kafka:9092
KAFKA_TOPIC=user-events
KAFKA_GROUP_ID=notification-group
```

# Тестирование

### Полный тестовый сценарий

```bash
# 1. Создание пользователя
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User"}'

# 2. Получение всех пользователей
curl http://localhost:8080/users

# 3. Обновление пользователя
curl -X PUT http://localhost:8080/users/{id} \
  -H "Content-Type: application/json" \
  -d '{"email":"updated@example.com","name":"Updated User"}'

# 4. Проверка статистики уведомлений
curl http://localhost:8081/stats

# 5. Удаление пользователя
curl -X DELETE http://localhost:8080/users/{id}
```

# Проверка Kafka сообщений

```bash
# Просмотр сообщений в топике
docker exec kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic user-events \
  --from-beginning
```

# Проверка PostgreSQL

```bash
# Подключение к БД
docker exec -it postgres psql -U postgres -d user_service

# Просмотр таблиц
\dt

# Просмотр данных
SELECT * FROM users;

# Выход
\q
```

# Мониторинг

### Health Checks

```bash
# Проверка User Service
curl http://localhost:8080/health
# {"status":"healthy","service":"user-service"}

# Проверка Notification Service
curl http://localhost:8081/health
# {"status":"healthy","service":"notification-service"}
```

# Метрики

```bash
# Получение метрик
curl http://localhost:8081/metrics
```

### Ответ:

{
"notification_processed_total": 3,
"notification_errors_total": 0,
"notifier_sent_total": 3,
"notifier_errors_total": 0,
"uptime_seconds": 360.5
}
