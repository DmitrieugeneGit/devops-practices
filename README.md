# Менеджер задач — DevOps-практика

Учебное fullstack-приложение (CRUD-менеджер задач). Стек ориентирован на Docker, мониторинг и дальнейший переход на Docker Swarm + Traefik.

- **Бэкенд**: Go (`net/http` + `pgx/v5`)
- **БД**: PostgreSQL
- **Фронтенд**: HTML/CSS/JS (статика через nginx)
- **Мониторинг**: Prometheus + Grafana + Node Exporter
- **Оркестрация сейчас**: Docker Compose
- **Дальше (сам)**: Docker Swarm + Traefik

## Архитектура

```
   браузер → nginx:80 (статика + /api)
                │
                ▼ /api
             app:8080 → db:5432

   Мониторинг: node-exporter → prometheus → grafana
```

Бэкенд `app` наружу не публикуется — доступ только через nginx. Позже nginx можно заменить на Traefik в Swarm.

## Сервисы и порты

| Сервис          | URL                          | Роль                          |
|-----------------|------------------------------|-------------------------------|
| `nginx`         | http://localhost             | Вход: статика + прокси `/api` |
| `app`           | (внутр. `app:8080`)          | Go-бэкенд, скрыт              |
| `db`            | (внутр. `db:5432`)           | PostgreSQL                    |
| `node-exporter` | http://localhost:9100/metrics| Метрики хоста                 |
| `prometheus`    | http://localhost:9090        | Сбор и хранение метрик        |
| `grafana`       | http://localhost:3000        | Дашборды                      |

## Структура проекта

```
devops-practices/
├── backend/                # Go-приложение
├── frontend/               # статика
├── db/                     # schema.sql, seed.sql
├── nginx/                  # reverse proxy (временно, до Traefik)
├── prometheus/
├── grafana/
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── .github/workflows/      # CI + деплой на сервер
```

## Требования

- Docker 24+ и Docker Compose v2

## Быстрый старт

```bash
cp .env.example .env
docker compose up --build -d
```

- Приложение: <http://localhost>
- Grafana: <http://localhost:3000> (`admin` / `admin`)
- Prometheus: <http://localhost:9090>

Остановить:

```bash
docker compose down          # данные в томах сохраняются
docker compose down -v       # удалить и тома
```

## Мониторинг в Grafana

1. Открыть <http://localhost:3000>, войти (`admin` / `admin`).
2. **Dashboards → New → Import**.
3. ID дашборда **`1860`** (Node Exporter Full) → **Load**.
4. Datasource **Prometheus** → **Import**.

## REST API

| Метод  | Путь              | Описание          |
|--------|-------------------|-------------------|
| GET    | `/api/health`     | Проверка живости  |
| GET    | `/api/tasks`      | Список задач      |
| POST   | `/api/tasks`      | Создать задачу    |
| GET    | `/api/tasks/{id}` | Получить задачу   |
| PUT    | `/api/tasks/{id}` | Обновить задачу   |
| DELETE | `/api/tasks/{id}` | Удалить задачу    |

```bash
curl -s localhost/api/tasks | jq

curl -s -X POST localhost/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Новая задача","description":"детали"}' | jq
```

## Полезные команды

```bash
docker compose ps
docker compose logs -f
docker compose logs -f app
docker compose up --build -d
```

## Конфигурация

Значения из `.env` (Compose подхватывает автоматически).

| Переменная               | По умолчанию | Описание          |
|--------------------------|--------------|-------------------|
| `DB_PORT`                | `5432`       | порт PostgreSQL   |
| `DB_USER`                | `tasks_user` | пользователь БД   |
| `DB_PASSWORD`            | `tasks_pass` | пароль БД         |
| `DB_NAME`                | `tasks_db`   | имя базы          |
| `DB_SSLMODE`             | `disable`    | режим SSL         |
| `GRAFANA_ADMIN_USER`     | `admin`      | логин Grafana     |
| `GRAFANA_ADMIN_PASSWORD` | `admin`      | пароль Grafana    |

> В Compose `DB_HOST=db`, `FRONTEND_DIR=/app/frontend`. Для запуска Go без Docker в `.env` можно оставить `DB_HOST=localhost`.
