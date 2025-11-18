## PR Reviewer Assignment Service

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


---

## Допущения и вопросы

- CRUD операции не реализовывал, тк в тз не просили :).
- Странно что в openapi.yml есть описание сущностей, которые не описаны в README (вроде PullRequestShort).
- Было бы славно написать о short сущностях что-то в тз.
- Сервис рассчитан на небольшую нагрузку, поэтому многопоточку не реализовывал.
