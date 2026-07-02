# Менеджер задач — DevOps-практика (Docker, Compose, мониторинг)

Учебное fullstack-приложение (CRUD-менеджер задач), развёрнутое как набор контейнеров через Docker Compose. Проект демонстрирует типовую production-архитектуру: reverse proxy, скрытый бэкенд, база данных и стек мониторинга.

- **Бэкенд**: Go (`net/http` + драйвер `pgx/v5`)
- **БД**: PostgreSQL
- **Фронтенд**: чистый HTML/CSS/JS, статику отдаёт nginx
- **Reverse proxy**: nginx (единый вход, проксирует `/api` на бэкенд)
- **Мониторинг**: Prometheus + Grafana + Node Exporter

## Архитектура

```
                    ┌──────────────┐
   браузер  ─────►  │    nginx     │  :8080 (единственный вход)
                    │ статика +/api│
                    └──────┬───────┘
                           │ /api → app:8080 (внутренняя сеть)
                    ┌──────▼───────┐        ┌──────────────┐
                    │  app (Go)    │ ─────► │  db (Postgres)│
                    │  скрыт       │        └──────────────┘
                    └──────────────┘

   Мониторинг:  node-exporter ─► prometheus ─► grafana
```

Бэкенд `app` наружу **не публикуется** — доступ к нему только через nginx.

## Сервисы и порты

| Сервис          | URL                          | Роль                                   |
|-----------------|------------------------------|----------------------------------------|
| `nginx`         | http://localhost:8080        | Вход: статика фронтенда + прокси `/api`|
| `app`           | (внутр. `app:8080`)          | Go-бэкенд (REST API), скрыт            |
| `db`            | (внутр. `db:5432`)           | PostgreSQL                             |
| `node-exporter` | http://localhost:9100/metrics| Метрики хоста (CPU, RAM, диск, сеть)   |
| `prometheus`    | http://localhost:9090        | Сбор и хранение метрик                 |
| `grafana`       | http://localhost:3000        | Дашборды и визуализация                |

## Структура проекта

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
├── frontend/               # index.html, style.css, app.js (статика)
├── db/
│   ├── schema.sql          # схема таблицы tasks
│   └── seed.sql            # тестовые данные
├── nginx/
│   └── nginx.conf          # конфиг reverse proxy (монтируется в контейнер)
├── prometheus/
│   └── prometheus.yml      # цели для сбора метрик
├── grafana/
│   └── provisioning/       # автонастройка источника данных (Prometheus)
├── Dockerfile              # multi-stage сборка образа приложения
├── docker-compose.yml      # оркестрация всех сервисов
├── .dockerignore
├── .env.example            # шаблон переменных окружения
└── README.md
```

## Требования

- Docker 24+ и Docker Compose v2

## Быстрый старт

### 1. Подготовить переменные окружения

```bash
cp .env.example .env
```

При необходимости отредактируйте значения (логин/пароль БД и т.д.). Файл `.env` не попадает в git.

### 2. Собрать и запустить весь стек

```bash
docker compose up --build -d
```

Флаги: `--build` — пересобрать образ приложения, `-d` — запуск в фоне.

### 3. Открыть приложение

- Приложение: <http://localhost:8080>
- Grafana: <http://localhost:3000> (логин/пароль по умолчанию `admin` / `admin`)
- Prometheus: <http://localhost:9090>

### 4. Остановить

```bash
docker compose down          # остановить (данные в томах сохраняются)
docker compose down -v       # остановить и удалить данные (тома)
```

## Мониторинг: настройка дашборда Grafana

Источник данных Prometheus подключается автоматически (provisioning). Чтобы увидеть графики:

1. Открыть <http://localhost:3000>, войти (`admin` / `admin`).
2. **Dashboards → New → Import**.
3. Ввести ID дашборда **`1860`** (Node Exporter Full) → **Load**.
4. Выбрать источник данных **Prometheus** → **Import**.

## REST API

Все запросы идут через nginx на `http://localhost:8080`.

| Метод  | Путь              | Описание          |
|--------|-------------------|-------------------|
| GET    | `/api/health`     | Проверка живости  |
| GET    | `/api/tasks`      | Список задач      |
| POST   | `/api/tasks`      | Создать задачу    |
| GET    | `/api/tasks/{id}` | Получить задачу   |
| PUT    | `/api/tasks/{id}` | Обновить задачу   |
| DELETE | `/api/tasks/{id}` | Удалить задачу    |

Пример:

```bash
curl -s localhost:8080/api/tasks | jq

curl -s -X POST localhost:8080/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Новая задача","description":"детали"}' | jq
```

## Полезные команды

```bash
docker compose ps                 # статус контейнеров
docker compose logs -f            # логи всех сервисов
docker compose logs -f app        # логи только бэкенда
docker compose restart nginx      # перезапустить nginx (после правки nginx.conf)
docker compose up --build -d      # пересобрать и перезапустить
```

## Конфигурация (переменные окружения)

Значения берутся из `.env` (Compose подхватывает автоматически).

| Переменная               | По умолчанию | Описание                          |
|--------------------------|--------------|-----------------------------------|
| `DB_PORT`                | `5432`       | порт PostgreSQL                   |
| `DB_USER`                | `tasks_user` | пользователь БД                   |
| `DB_PASSWORD`            | `tasks_pass` | пароль БД                         |
| `DB_NAME`                | `tasks_db`   | имя базы                          |
| `DB_SSLMODE`             | `disable`    | режим SSL                         |
| `GRAFANA_ADMIN_USER`     | `admin`      | логин администратора Grafana      |
| `GRAFANA_ADMIN_PASSWORD` | `admin`      | пароль администратора Grafana     |

> Внутри Compose `DB_HOST` фиксирован как `db` (имя сервиса), а `FRONTEND_DIR` — как `/app/frontend`; значения `DB_HOST=localhost` / `FRONTEND_DIR=../frontend` в `.env` предназначены для локального запуска приложения без Docker.

## Запуск без Docker (опционально)

Приложение можно запустить и напрямую на хосте:

```bash
# требуется локальный PostgreSQL с ролью tasks_user и базой tasks_db
cd backend
go mod tidy
go run .
```

Тогда Go-сервер сам отдаёт и статику, и API на <http://localhost:8080> (без nginx).
