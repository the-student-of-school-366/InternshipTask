## PR Reviewer Assignment Service

Сервис для автоматического назначения ревьюверов на Pull Request’ы внутри команд, а также управления командами и активностью пользователей. ТЗ описано в `TASK.md`, HTTP‑контракт — в `openapi.yml`.

---



## Архитектура

- `internal/domain`
  - `user` — модель пользователя и сервис (`GetByID`, `SetIsActive`).
  - `team` — модель команды и сервис (`Create`, `GetByTeamName`).
  - `pull_request` — модель PR и сервис:
    - `Create` — создаёт PR и автоматически назначает до двух активных ревьюверов из команды автора (исключая автора).
    - `Merge` — идемпотентный перевод PR в `MERGED`.
    - `Reassign` — переназначение ревьювера на активного участника его команды с проверкой доменных ограничений.
    - `GetByReviewerID` — список PR, где пользователь ревьювер.
    - `ReviewerStats` — статистика по количеству назначений ревьюверами.

- `internal/infrastructure/postgres`
  - `team/storage.go` — хранение команд, upsert пользователей при создании/обновлении команды.
  - `user/storage.go` — чтение/обновление пользователей (`GetByID`, `SetIsActive`).
  - `pull_request/storage.go` — хранение PR (`Create`, `GetByID`, `Update`, `GetByReviewerID`, `GetReviewerStats`). Поле `assigned_reviewers` хранится как `TEXT[]`.

- `internal/app`
  - `app.go` — сборка зависимостей (storages, services) и создание `gin.Engine`.
  - `dto/` — DTO для HTTP‑слоя (команды, пользователи, PR, ошибки, статистика).
  - `http/` — Gin‑хэндлеры:
    - `handlers.go` — `Handler` и `RegisterRoutes`.
    - `health_stats.go` — `/health`, `/stats`.
    - `team_handlers.go` — `/team/add`, `/team/get`.
    - `user_handlers.go` — `/users/setIsActive`, `/users/getReview`.
    - `pull_request_handlers.go` — `/pullRequest/create`, `/pullRequest/merge`, `/pullRequest/reassign`.
    - `error.go` — общий helper `writeError` для формата `ErrorResponse`.

- `cmd/main.go` — точка входа: читает `DATABASE_DSN`, создаёт приложение через `app.New` и запускает HTTP‑сервер на `:8080`.

---

## Схема БД и миграции

Миграции находятся в `db/migrations` и применяются автоматически при старте контейнера Postgres.

`001_init.sql` создаёт таблицы:

- `teams(team_name TEXT PRIMARY KEY)`;
- `users(user_id TEXT PRIMARY KEY, username TEXT, team_name TEXT REFERENCES teams(team_name), is_active BOOLEAN)`;
- `pull_requests(pull_request_id TEXT PRIMARY KEY, pull_request_name TEXT, author_id TEXT REFERENCES users(user_id), status TEXT CHECK (status IN ('OPEN', 'MERGED')), assigned_reviewers TEXT[], created_at TIMESTAMPTZ, merged_at TIMESTAMPTZ)`.

---

## Запуск через Docker Compose

Требования:

- Docker
- docker-compose

Из корня проекта:

```bash
docker-compose up --build
```

Что делает `docker-compose`:

- поднимает `postgres`:
  - БД: `pr_service`;
  - пользователь: `pr_user`;
  - пароль: `pr_password`;
  - применяет миграции из `db/migrations` через `docker-entrypoint-initdb.d`;
- после успешного healthcheck’а Postgres поднимает сервис `app`:
  - в контейнер прокидывается `DATABASE_DSN=postgres://pr_user:pr_password@postgres:5432/pr_service?sslmode=disable`;
  - HTTP‑сервер доступен на `http://localhost:8080`.

---

## Основные эндпоинты

Полное описание и примеры — в `openapi.yml`. Кратко:

- `POST /team/add` — создать команду с участниками (создаёт/обновляет пользователей).
- `GET /team/get?team_name=...` — получить команду с участниками.
- `POST /users/setIsActive` — изменить флаг активности пользователя.
- `GET /users/getReview?user_id=...` — PR, где пользователь назначен ревьювером.
- `POST /pullRequest/create` — создать PR и автоматически назначить до двух активных ревьюверов из команды автора.
- `POST /pullRequest/merge` — пометить PR как `MERGED` (идемпотентно).
- `POST /pullRequest/reassign` — переназначить ревьювера на активного участника его команды.
- `GET /health` — healthcheck.
- `GET /stats` — простая статистика по количеству назначений ревьювером.

---

## Тесты

Юнит‑тесты:

- доменные сервисы — `internal/domain/*/service_test.go`;
- HTTP‑слой — `internal/app/http/handlers_test.go` (базовые проверки).

Запуск тестов:

```bash
go test ./...
```



---

## Ручное тестирование (на основе тестовых данных из миграций)

После запуска `docker-compose up --build` в БД автоматически создаются:

- команды: `backend`, `frontend`;
- пользователи:
  - `backend`: `u1` (Alice, активен), `u2` (Bob, активен), `u3` (Charlie, неактивен);
  - `frontend`: `u4` (Dave, активен), `u5` (Eve, активен), `u6` (Frank, неактивен);
- PR:
  - `pr-1001` — `OPEN`, автор `u1`, ревьюверы `u2`, `u4`;
  - `pr-1002` — `MERGED`, автор `u1`, ревьюверы `u2`, `u5`;
  - `pr-1003` — `OPEN`, автор `u3` (неактивен), ревьюверы отсутствуют.

### Проверка health и статистики

# Healthcheck
curl http://localhost:8080/health
# => {"status":"ok"}

# Статистика по назначениям ревьюверов
curl http://localhost:8080/stats
# => {"review_assignments":{"u2":2,"u4":1,"u5":1}}   # пример### Команды

# Получить команду backend
curl "http://localhost:8080/team/get?team_name=backend"

# Создать/обновить команду с участниками
curl -X POST http://localhost:8080/team/add \
-H "Content-Type: application/json" \
-d '{
"team_name": "backend",
"members": [
{ "user_id": "u1", "username": "Alice", "is_active": true },
{ "user_id": "u2", "username": "Bob",   "is_active": true }
]
}'### Пользователи

# Сделать пользователя u2 неактивным
curl -X POST http://localhost:8080/users/setIsActive \
-H "Content-Type: application/json" \
-d '{
"user_id": "u2",
"is_active": false
}'

# Посмотреть PR, где u2 назначен ревьювером
curl "http://localhost:8080/users/getReview?user_id=u2"### Pull Request’ы

# Создать новый PR от автора u1
curl -X POST http://localhost:8080/pullRequest/create \
-H "Content-Type: application/json" \
-d '{
"pull_request_id": "pr-2001",
"pull_request_name": "New feature",
"author_id": "u1"
}'
# В ответе будут автоматически назначенные до двух активных ревьюверов из команды автора.

# Смержить существующий PR (идемпотентно)
curl -X POST http://localhost:8080/pullRequest/merge \

-H "Content-Type: application/json" \
-d '{ "pull_request_id": "pr-1001" }'

# Переназначить ревьювера u2 в PR pr-1001 на другого участника его команды
curl -X POST http://localhost:8080/pullRequest/reassign \
-H "Content-Type: application/json" \
-d '{
"pull_request_id": "pr-1001",
"old_user_id": "u2"
}'Эти запросы можно также выполнить через любой OpenAPI‑UI (например, editor.swagger.io), импортировав `openapi.yml` и указав сервер `http://localhost:8080`.
---

## Допущения

- Пользователь принадлежит только одной команде (`team_name` в таблице `users` и в модели `User`).
- Ревьюверы PR хранятся как `TEXT[]` с `user_id` в поле `assigned_reviewers` — это упрощает выборку PR по ревьюверу и расчёт статистики.
- Нагрузки из ТЗ (до 5 RPS) позволяют использовать простую реализацию без сложных оптимизаций и блокировок на уровне бизнес‑логики.


