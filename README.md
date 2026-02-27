# Calendar Service

Calendar — микросервис HTTP-сервер для управления календарём событий с поддержкой JWT аутентификации, уведомлений через Email и Telegram.

## Функциональность

- **Аутентификация**: JWT токены (access + refresh)
- **Управление событиями**: создание, обновление, удаление, просмотр
- **Уведомления**: Email и Telegram напоминания
- **Периоды просмотра**: день, неделя, месяц
- **Автоочистка**: архивирование старых событий

---

## Архитектура

```
calendar/
├── cmd/Calendar/          # Точка входа
├── config/                # Конфигурация (YAML)
├── docs/                  # Swagger документация
├── internal/
│   ├── app/               # Инициализация приложения (Fx DI)
│   ├── auth/jwt/          # JWT сервис
│   ├── config/            # Загрузка конфигурации
│   ├── di/                # DI компоненты
│   ├── domain/            # Доменные модели и ошибки
│   │   ├── event/         # Модель события
│   │   └── user/          # Модель пользователя
│   ├── logger/            # Асинхронный логгер (Zap)
│   ├── pkg/
│   │   ├── cleaner/       # Сервис очистки старых событий
│   │   └── notifications/ # Email/Telegram уведомления
│   ├── repository/postgres/ # PostgreSQL репозиторий
│   ├── services/          # Бизнес-логика
│   │   ├── event/         # Сервис событий
│   │   └── user/          # Сервис пользователей
│   └── web/
│       ├── dto/           # Data Transfer Objects
│       ├── handlers/      # HTTP обработчики
│       └── routers/       # Маршрутизация и middleware
├── migrations/            # SQL миграции
└── web/                   # Статичные файлы (тестовая страница)
```

---

## Быстрый старт

### 1. Требования

- Go 1.22+
- PostgreSQL 16+
- Docker & Docker Compose (опционально)

### 2. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
DB_POSTGRES_USER=calendar
DB_POSTGRES_PASSWORD=your_secure_password
DB_POSTGRES_DB_NAME=calendar_db

TELEGRAM_BOT_TOKEN=your_telegram_bot_token

MAIL_SMTP_USER=your_email@gmail.com
MAIL_SMTP_PASSWORD=your_app_password

JWT_ACCESS_SECRET=your_access_secret_key_min_32_chars
JWT_REFRESH_SECRET=your_refresh_secret_key_min_32_chars
```

### 3. Запуск с Docker Compose

```bash
# Запуск PostgreSQL и миграций
docker-compose up -d

# Запуск сервиса
go run ./cmd/Calendar
```

### 4. Запуск без Docker

```bash
# Установка зависимостей
go mod tidy

# Запуск PostgreSQL вручную, затем миграции
# migrate -path ./migrations -database "postgres://..." up

# Запуск сервиса
go run ./cmd/Calendar
```

Сервис запустится на `http://localhost:8080`

---

## 📡 API Endpoints

### Аутентификация

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/auth/register` | Регистрация пользователя |
| POST | `/api/v1/auth/login` | Вход в систему |
| POST | `/api/v1/auth/refresh-token` | Обновление токенов |

### События (требуют авторизации)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/event/create_event` | Создание события |
| PUT | `/api/v1/event/update_event` | Обновление события |
| DELETE | `/api/v1/event/delete_event` | Удаление события |
| GET | `/api/v1/event/events_for_day?date=YYYY-MM-DD` | События за день |
| GET | `/api/v1/event/events_for_week?date=YYYY-MM-DD` | События за неделю |
| GET | `/api/v1/event/events_for_month?date=YYYY-MM-DD` | События за месяц |

### Примеры запросов

**Регистрация:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "SecurePass123",
    "email": "test@example.com",
    "telegram_chat_id": "123456789"
  }'
```

**Вход:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "SecurePass123"
  }'
```

**Создание события:**
```bash
curl -X POST http://localhost:8080/api/v1/event/create_event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "event_name": "Встреча",
    "date": "2026-03-15",
    "event": "Описание встречи",
    "reminder_time": "2026-03-14"
  }'
```

---

##  Конфигурация

Файл `config/local.yaml`:

```yaml
server:
  host: "localhost"
  port: 8080

logger:
  mode: "dev"      # "dev" или "prod"
  level: "debug"   # "debug", "info", "warn", "error"
  buffer_size: 1000

gin:
  mode: "release"

db:
  postgres:
    host: "localhost"
    port: 5433
    ssl_mode: "disable"
    max_open_conns: 5
    max_idle_conns: 10
    conn_max_lifetime: "100s"

jwt:
  exp_access_token: 15   # минуты
  exp_refresh_token: 24  # часы

cleaner:
  check_interval: 30m
  event_lifetime: 24h
```

---

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# С покрытием
go test ./... -cover

# Подробный вывод
go test ./... -v

# Генерация отчёта покрытия
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Swagger документация

Swagger документация доступна в `/swagger`.

---

## Тестовая веб-страница

Откройте `web/index.html` в браузере для интерактивного тестирования API.

Функции:
- Регистрация и авторизация
- Создание, обновление, удаление событий
- Просмотр событий по периодам
- Обновление токенов

---

## Зависимости

| Пакет | Описание |
|-------|----------|
| `github.com/gin-gonic/gin` | HTTP фреймворк |
| `go.uber.org/fx` | DI контейнер |
| `go.uber.org/zap` | Структурированный логгер |
| `github.com/jackc/pgx/v5` | PostgreSQL драйвер |
| `github.com/golang-jwt/jwt/v5` | JWT токены |
| `golang.org/x/crypto` | Bcrypt хеширование |
| `github.com/go-telegram-bot-api/telegram-bot-api/v5` | Telegram бот |

---

## 📁 Миграции

```bash
# Применить миграции
migrate -path ./migrations -database "postgres://user:pass@localhost:5433/db?sslmode=disable" up

# Откатить
migrate -path ./migrations -database "..." down
```

---

## 🛡️ Безопасность

- Пароли хешируются с bcrypt
- JWT с проверкой алгоритма подписи (защита от algorithm confusion)
- Валидация всех входных данных
- Таймауты на операции с БД


