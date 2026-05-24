# TickiPeaki

Микросервисная система бронирования авиабилетов.

- **aviation** — Go + Gin + GORM + PostgreSQL, REST API с JWT-аутентификацией
- **tickets** — Python FastAPI, генерация PDF-билетов
- **aviation-frontend** — React + Vite + Tailwind, UI в стиле Aviasales
- **database** — PostgreSQL 15
- **minio** — S3-совместимое хранилище для PDF
- **redis** — кэш

## Архитектура

```
aviation-frontend (:5173)
        │
        ▼
aviation (Go :8080) ──────────► database (Postgres :5432)
        │
        │  PUT /tickets/:id  (status → paid)
        ▼
tickets (Python FastAPI :8000) ──► minio (:9000)
```

При смене статуса билета на `paid` Go-сервис асинхронно вызывает
`POST {PDF_SERVICE_URL}/generate-ticket`. PDF загружается в MinIO,
ссылка сохраняется в поле `pdf_url` билета.

## Требования

- Docker + Docker Compose

Локальный запуск без Docker (опционально):
- Go 1.23+
- Node 20+
- PostgreSQL 15

## Запуск через Docker

```bash
docker-compose up --build
```

| Сервис    | Порт        | Описание                           |
|-----------|-------------|------------------------------------|
| frontend  | 5173        | React UI (nginx)                   |
| aviation  | 8080        | Go API (Gin + GORM)                |
| tickets   | 8000        | Python FastAPI, генерация PDF      |
| database  | 5432        | PostgreSQL 15                      |
| minio     | 9000 / 9001 | S3 (API / web-консоль)             |
| redis     | 6379        | Кэш                                |

Все сервисы в общей сети `aviation-net`. Данные БД — в volume `postgres_data`,
объекты MinIO — в `minio_data`. `aviation` стартует после healthcheck postgres.

Конфигурация сервисов берётся из `aviation/.env` и `tickets/.env`.

Остановка с очисткой данных:

```bash
docker-compose down -v
```

## Переменные окружения (aviation)

| Переменная        | По умолчанию                                                                          |
|-------------------|---------------------------------------------------------------------------------------|
| `DATABASE_URL`    | `host=database user=postgres password=postgres dbname=aviation port=5432 sslmode=disable` |
| `PDF_SERVICE_URL` | `http://tickets:8000`                                                                 |
| `JWT_SECRET`      | `supersecret`                                                                         |

## Локальный запуск без Docker

```bash
cd aviation
go mod tidy
go run main.go
```

```bash
cd aviation-frontend
npm install
npm run dev
```

API на `http://localhost:8080`, фронт на `http://localhost:5173`.

## Миграции

SQL-миграции лежат в `aviation/migrations/`. Последняя — `000006_add_seats_refactor`:
выносит места в отдельную таблицу `seats`, убирает `price` и `available_seats`
из таблицы `flights`. `000005_seed_data` наполняет БД тестовыми данными.

## Доменная модель

- **User** — учётка с username/password (bcrypt) и ролью (`user` / `admin`).
  Может иметь связанный профиль пассажира.
- **Passenger** — анкета пассажира (ФИО, email, телефон, паспорт).
- **Flight** — рейс. Хранит маршрут и время. **Не хранит цену и количество мест.**
- **Seat** — место на рейсе. Хранит `seat_number`, `class`, `price`, `status`
  (`available` / `booked`). Цена привязана к месту, а не к рейсу.
- **Ticket** — билет. Привязывает `Passenger` к `Seat` конкретного `Flight`.
  Статус: `reserved` / `paid` / `cancelled`. После `paid` появляется `pdf_url`.

## Аутентификация

Защищённые маршруты требуют заголовок `Authorization: Bearer <token>`.
Токен — JWT (HS256), payload `{ user_id, username, role, exp }`, срок 24 часа.

### Регистрация — POST `/register`

```json
{ "username": "ivan", "password": "secret123" }
```

### Логин — POST `/login`

```json
{ "username": "ivan", "password": "secret123" }
```

Ответ:

```json
{
  "token": "eyJhbGciOi...",
  "user": { "id": 1, "username": "ivan", "role": "user" }
}
```

### Профиль — GET `/me`

Возвращает пользователя вместе с привязанным `Passenger` (если есть).

### Заполнить пассажира — POST `/me/passenger`

```json
{
  "full_name": "Иван Иванов",
  "email": "ivan@example.com",
  "phone": "+77001234567",
  "passport_num": "N12345678"
}
```

Без заполненного профиля покупка билета вернёт `403`.

## Эндпоинты

Все маршруты ниже требуют JWT.

### Рейсы (`/flights`)

| Метод  | Путь          | Описание                       |
|--------|---------------|--------------------------------|
| GET    | /flights      | Список рейсов                  |
| GET    | /flights/:id  | Рейс + все его места + сводка  |
| POST   | /flights      | Создать рейс                   |
| PUT    | /flights/:id  | Обновить рейс                  |
| DELETE | /flights/:id  | Удалить рейс                   |

Фильтры списка: `?origin=`, `?destination=`, `?carrier=`, `?page=`, `?limit=`

`GET /flights/:id` отдаёт:

```json
{
  "flight": { "id": 1, "flight_number": "KC-101", "origin": "ALA", ... },
  "seats": [
    { "id": 1, "seat_number": "1A", "class": "first",   "price": 55000, "status": "available" },
    { "id": 5, "seat_number": "6A", "class": "economy", "price": 25000, "status": "booked" }
  ],
  "available_count": 45,
  "taken_seats": ["6A", "6B"]
}
```

### Места (`/seats`)

| Метод  | Путь                | Описание                                     |
|--------|---------------------|----------------------------------------------|
| GET    | /flights/:id/seats  | Все места рейса (`{ data: [...], total }`)   |
| GET    | /seats/:id          | Одно место                                   |
| POST   | /seats              | Создать место                                |
| PUT    | /seats/:id          | Обновить `price` / `status` (patch-style)    |
| DELETE | /seats/:id          | Удалить место (нельзя, если `booked`)        |

### Пассажиры (`/passengers`)

| Метод  | Путь            | Описание           |
|--------|-----------------|--------------------|
| GET    | /passengers     | Список пассажиров  |
| POST   | /passengers     | Создать пассажира  |
| PUT    | /passengers/:id | Обновить пассажира |
| DELETE | /passengers/:id | Удалить пассажира  |

### Билеты (`/tickets`)

| Метод  | Путь         | Описание                                      |
|--------|--------------|-----------------------------------------------|
| GET    | /tickets     | Список билетов                                |
| POST   | /tickets     | Забронировать билет (passenger из JWT)        |
| PUT    | /tickets/:id | Сменить статус билета                         |
| DELETE | /tickets/:id | Удалить билет (освобождает место)             |

Фильтры списка: `?flight_id=`, `?passenger_id=`, `?status=`, `?page=`, `?limit=`

## Примеры запросов

### Создать рейс — POST `/flights`

```json
{
  "flight_number": "KC301",
  "origin": "ALA",
  "destination": "NQZ",
  "carrier": "Air Astana",
  "departure_time": "2026-04-01T08:00:00Z",
  "arrival_time": "2026-04-01T09:30:00Z"
}
```

Места создаются отдельно (через миграции/seed или `POST /seats`) и привязываются к `flight_id`.

### Создать место — POST `/seats`

```json
{
  "flight_id": 1,
  "seat_number": "12A",
  "class": "economy",
  "price": 25000
}
```

Допустимые значения `class`: `economy`, `business`, `first`. `price` должна быть `> 0`.
Новое место получает статус `available`. Дубль `seat_number` в рамках одного рейса — `409`.

### Обновить место — PUT `/seats/:id`

```json
{ "price": 27000, "status": "available" }
```

Patch-style: передавайте только те поля, которые нужно изменить. Допустимый
`status` — `available` или `booked`.

### Забронировать билет — POST `/tickets`

```json
{
  "flight_id": 1,
  "seat_id": 3
}
```

`passenger_id` берётся из JWT автоматически. Если у пользователя нет профиля
пассажира — `403`. Если место уже забронировано — `409`.

Ответ:

```json
{
  "id": 1,
  "flight_id": 1,
  "seat_id": 3,
  "seat": {
    "id": 3,
    "seat_number": "3A",
    "class": "business",
    "price": 35000,
    "status": "booked"
  },
  "status": "reserved",
  "booked_at": "2026-04-01T07:15:00Z"
}
```

### Сменить статус — PUT `/tickets/:id`

```json
{ "status": "paid" }
```

Допустимые значения: `reserved`, `paid`, `cancelled`.

- `paid` — асинхронно запускается генерация PDF в `tickets`. После успеха
  поле `pdf_url` билета обновляется.
- `cancelled` — место освобождается (`status` → `available`).

### Удаление

```
DELETE /flights/1
DELETE /passengers/1
DELETE /tickets/1
```

`204 No Content` при успехе. При удалении билета со статусом, отличным от
`cancelled`, место на рейсе освобождается. Пассажира с активными билетами
удалить нельзя — `409`.

## Интеграция с tickets (PDF)

При `PUT /tickets/:id` со сменой статуса на `paid` запускается воркер,
который через Resty v2 вызывает `POST {PDF_SERVICE_URL}/generate-ticket`.
PDF загружается в MinIO, ссылка сохраняется в `pdf_url` билета. Hooks
`OnBeforeRequest` / `OnAfterResponse` пишут метод, URL и статус в лог.

Если `tickets` недоступен — основной запрос не падает, ошибка только в логе.
Билет к этому моменту уже сохранён в БД, можно повторить смену статуса позже.

## Структура

```
.
├── docker-compose.yaml
├── aviation/                       # Go API
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── main.go
│   ├── config/
│   │   └── database.go
│   ├── middleware/
│   │   ├── auth.go                 # JWT-мидлварь
│   │   └── cors.go
│   ├── migrations/                 # SQL-миграции (golang-migrate)
│   ├── models/
│   │   ├── flight.go
│   │   ├── seat.go
│   │   ├── passenger.go
│   │   ├── ticket.go
│   │   └── user.go
│   ├── repository/
│   │   ├── interfaces.go
│   │   └── postgres/
│   │       ├── flight_repo.go
│   │       ├── seat_repo.go
│   │       ├── passenger_repo.go
│   │       ├── ticket_repo.go
│   │       └── user_repo.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── flight_handler.go
│   │   ├── passenger_handler.go
│   │   └── ticket_handler.go
│   └── clients/
│       └── pdf_client.go           # Resty-клиент к tickets
├── tickets/                        # Python FastAPI, генерация PDF
│   ├── Dockerfile
│   ├── pyproject.toml
│   ├── src/
│   └── templates/
└── aviation-frontend/              # React + Vite + Tailwind
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    └── src/
        ├── api/
        ├── components/
        ├── context/
        └── pages/
```
