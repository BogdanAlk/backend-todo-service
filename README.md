# Backend ToDo Service

REST API на Go для регистрации пользователей и управления задачами.

## Tech Stack
- Go (net/http)
- PostgreSQL
- Docker Compose
- JWT (HMAC)
- bcrypt for password hashing

## Project Structure
- cmd/api — точка входа API
- internal/httpapi — HTTP handlers + middleware
- internal/db — подключение к PostgreSQL
- internal/users — репозиторий пользователей
- internal/tasks — репозиторий задач
- migrations — SQL миграции

## Run

### 1) Start PostgreSQL
```bash
docker compose up -d

