# Менеджер задач — Go + PostgreSQL (без Docker)

Учебное fullstack-приложение для DevOps-практики: CRUD-менеджер задач.

- **Бэкенд**: Go (стандартный `net/http` + драйвер `pgx/v5`)
- **БД**: PostgreSQL
- **Фронтенд**: чистый HTML/CSS/JS (без сборщиков), статику отдаёт сам Go-сервер
- **Без Docker** — всё запускается «голым» на хосте

## Структура

```
devops-practices/
├── backend/                # Go-приложение
│   ├── main.go             # точка входа, HTTP-сервер, graceful shutdown
│   ├── go.mod
│   └── internal/
│       ├── config/         # чтение настроек из ENV
│       ├── database/       # пул соединений и запросы к PostgreSQL
│       ├── handlers/       # HTTP-обработчики (REST API) + отдача статики
│       └── models/         # модели данных
├── frontend/               # index.html, style.css, app.js
├── db/
│   ├── schema.sql          # схема таблицы tasks
│   └── seed.sql            # тестовые данные
├── .env.example
└── README.md
```

## Требования

- Go 1.22+ (используется маршрутизация `net/http` с методами; проверено на 1.26)
- PostgreSQL 13+ (проверено на 18)

## Быстрый старт

### 1. Создать роль и базу данных

```bash
sudo -u postgres psql -c "CREATE ROLE tasks_user LOGIN PASSWORD 'tasks_pass';"
sudo -u postgres psql -c "CREATE DATABASE tasks_db OWNER tasks_user;"
```

### 2. Применить схему и тестовые данные

```bash
PGPASSWORD=tasks_pass psql -h localhost -U tasks_user -d tasks_db -f db/schema.sql
PGPASSWORD=tasks_pass psql -h localhost -U tasks_user -d tasks_db -f db/seed.sql
```

### 3. Скачать зависимости и запустить

```bash
cd backend
go mod tidy
go run .
```

Откройте <http://localhost:8080>

> Порт по умолчанию — **8080**.
> Изменить можно переменной `HTTP_ADDR`, например: `HTTP_ADDR=:9000 go run .`

## REST API

| Метод  | Путь              | Описание                 |
|--------|-------------------|--------------------------|
| GET    | `/api/health`     | Проверка живости         |
| GET    | `/api/tasks`      | Список задач             |
| POST   | `/api/tasks`      | Создать задачу           |
| GET    | `/api/tasks/{id}` | Получить задачу          |
| PUT    | `/api/tasks/{id}` | Обновить задачу          |
| DELETE | `/api/tasks/{id}` | Удалить задачу           |

Пример:

```bash
curl -s localhost:8080/api/tasks | jq

curl -s -X POST localhost:8080/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Новая задача","description":"детали"}' | jq
```

## Полезные команды

```bash
cd backend
go run .                          # запустить сервер
go build -o ../bin/tasks-app .    # собрать бинарник в ./bin/tasks-app
go fmt ./...                      # форматирование
go vet ./...                      # статический анализ
```

## Конфигурация (переменные окружения)

| Переменная     | По умолчанию        | Описание                             |
|----------------|---------------------|--------------------------------------|
| `HTTP_ADDR`    | `:8080`             | адрес HTTP-сервера                   |
| `FRONTEND_DIR` | `../frontend`       | путь к статике фронтенда             |
| `DATABASE_URL` | —                   | полный DSN (имеет приоритет)         |
| `DB_HOST`      | `localhost`         | хост PostgreSQL                      |
| `DB_PORT`      | `5432`              | порт                                 |
| `DB_USER`      | `tasks_user`        | пользователь                         |
| `DB_PASSWORD`  | `tasks_pass`        | пароль                               |
| `DB_NAME`      | `tasks_db`          | имя базы                             |
| `DB_SSLMODE`   | `disable`           | режим SSL                            |
