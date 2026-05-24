# TickiPeaki

> Микросервисная система бронирования авиабилетов с JWT-аутентификацией,
> 5-минутным холдом места под оплату и асинхронной генерацией PDF-билетов.

## Содержание

- [Стек](#стек)
- [Архитектура](#архитектура)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация](#конфигурация)
- [Миграции](#миграции)
- [Доменная модель](#доменная-модель)
- [Аутентификация](#аутентификация)
- [Эндпоинты](#эндпоинты)
- [Платёжный flow](#платёжный-flow)
- [Интеграция с tickets (PDF)](#интеграция-с-tickets-pdf)
- [Структура репозитория](#структура-репозитория)

---

## Стек

| Сервис              | Технологии                                        | Назначение                       |
|---------------------|---------------------------------------------------|----------------------------------|
| `aviation`          | Go 1.23 · Gin · GORM · JWT (HS256) · Resty v2     | REST API                         |
| `tickets`           | Python · FastAPI                                  | Генерация PDF-билетов            |
| `aviation-frontend` | React · Vite · Tailwind · React Router · axios    | UI в стиле Aviasales             |
| `database`          | PostgreSQL 15                                     | Основное хранилище               |
| `minio`             | MinIO (S3-совместимый)                            | Хранение PDF-файлов              |
| `redis`             | Redis 7                                           | Кэш                              |

## Архитектура

```
                  ┌──────────────────────────┐
                  │  aviation-frontend :5173 │
                  └────────────┬─────────────┘
                               │ REST + JWT
                               ▼
   ┌───────────────────────────────────────────────────────┐
   │                 aviation (Go) :8080                   │
   │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │
   │  │  auth    │  │ flights  │  │  seats   │  │tickets │ │
   │  └──────────┘  └──────────┘  └──────────┘  └────────┘ │
   │       reservation timer (5 min, in-memory)            │
   └────┬───────────────────────┬──────────────────────────┘
        │ GORM                  │ POST /generate-ticket
        ▼                       ▼
  ┌──────────────┐      ┌─────────────────┐      ┌────────┐
  │ database     │      │ tickets :8000   │ ───► │ minio  │
  │ Postgres 15  │      │ FastAPI         │      │  S3    │
  └──────────────┘      └─────────────────┘      └────────┘
```

При смене статуса билета на `paid` Go-сервис асинхронно вызывает
`POST {PDF_SERVICE_URL}/generate-ticket`. PDF загружается в MinIO, ссылка
сохраняется в поле `pdf_url` билета.

## Быстрый старт

### Через Docker

```bash
docker-compose up --build
```

Поднимутся все сервисы из таблицы ниже. Конфигурация берётся из
`aviation/.env` и `tickets/.env`. `aviation` стартует после healthcheck
postgres.

| Сервис    | Порт        | Описание                       |
|-----------|-------------|--------------------------------|
| frontend  | 5173        | React UI (nginx)               |
| aviation  | 8080        | Go API                         |
| tickets   | 8000        | Python FastAPI                 |
| database  | 5432        | PostgreSQL 15                  |
| minio     | 9000 / 9001 | S3 API / web-консоль           |
| redis     | 6379        | Кэш                            |

Все сервисы — в общей сети `aviation-net`. Данные БД — в volume
`postgres_data`, объекты MinIO — в `minio_data`.

Остановка с очисткой данных:

```bash
docker-compose down -v
```

### Без Docker

```bash
# backend
cd aviation && go mod tidy && go run main.go

# frontend
cd aviation-frontend && npm install && npm run dev
```

API доступен на `http://localhost:8080`, фронт — на `http://localhost:5173`.

## Конфигурация

Переменные окружения сервиса `aviation`:

| Переменная        | По умолчанию                                                                              |
|-------------------|-------------------------------------------------------------------------------------------|
| `DATABASE_URL`    | `host=database user=postgres password=postgres dbname=aviation port=5432 sslmode=disable` |
| `PDF_SERVICE_URL` | `http://tickets:8000`                                                                     |
| `JWT_SECRET`      | `supersecret`                                                                             |

## Миграции

SQL-миграции лежат в `aviation/migrations/` и применяются через
[`golang-migrate`](https://github.com/golang-migrate/migrate):

| №       | Файл                                | Что делает                                                                          |
|---------|-------------------------------------|-------------------------------------------------------------------------------------|
| 000005  | `seed_flights_and_seats`            | Сидит 5 рейсов и 640 мест (128 на рейс: 8 first + 30 business + 90 economy)         |

## Доменная модель

| Модель      | Описание                                                                                  |
|-------------|-------------------------------------------------------------------------------------------|
| **User**    | Учётка (username + bcrypt-password), роль `user` / `admin`, опциональная связь с `Passenger` |
| **Passenger** | Анкета пассажира: ФИО, email, телефон, паспорт                                          |
| **Flight**  | Рейс: маршрут и время. **Не хранит цену и количество мест**                               |
| **Seat**    | Место на рейсе: `seat_number`, `class`, `price`, `status` (`available` / `booked`). Цена живёт на месте, а не на рейсе |
| **Ticket**  | Привязывает `Passenger` к `Seat` конкретного `Flight`. Статус: `reserved` / `paid` / `cancelled`. После `paid` появляется `pdf_url` |

## Аутентификация

Защищённые маршруты требуют заголовок:

```
Authorization: Bearer <token>
```

Токен — JWT (HS256), payload `{ user_id, username, role, exp }`, срок действия
24 часа.

### Регистрация — `POST /register`

```json
{ "username": "ivan", "password": "secret123" }
```

### Логин — `POST /login`

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

### Профиль — `GET /me`

Возвращает пользователя вместе с привязанным `Passenger` (если есть).

### Заполнить пассажира — `POST /me/passenger`

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

### Рейсы — `/flights`

| Метод  | Путь          | Описание                       |
|--------|---------------|--------------------------------|
| GET    | /flights      | Список рейсов                  |
| GET    | /flights/:id  | Рейс + все его места + сводка  |
| POST   | /flights      | Создать рейс                   |
| PUT    | /flights/:id  | Обновить рейс                  |
| DELETE | /flights/:id  | Удалить рейс                   |

Фильтры списка: `?origin=`, `?destination=`, `?carrier=`, `?page=`, `?limit=`.

`GET /flights/:id` отдаёт:

```json
{
  "flight": { "id": 1, "flight_number": "KC-101", "origin": "ALA", "...": "..." },
  "seats": [
    { "id": 1, "seat_number": "1A", "class": "first",   "price": 65000, "status": "available" },
    { "id": 5, "seat_number": "6A", "class": "economy", "price": 25000, "status": "booked" }
  ],
  "available_count": 45,
  "taken_seats": ["6A", "6B"]
}
```

### Места — `/seats`

| Метод  | Путь                | Описание                                     |
|--------|---------------------|----------------------------------------------|
| GET    | /flights/:id/seats  | Все места рейса (`{ data: [...], total }`)   |
| GET    | /seats/:id          | Одно место                                   |
| POST   | /seats              | Создать место                                |
| PUT    | /seats/:id          | Обновить `price` / `status` (patch-style)    |
| DELETE | /seats/:id          | Удалить место (нельзя, если `booked`)        |

### Пассажиры — `/passengers`

| Метод  | Путь            | Описание           |
|--------|-----------------|--------------------|
| GET    | /passengers     | Список пассажиров  |
| POST   | /passengers     | Создать пассажира  |
| PUT    | /passengers/:id | Обновить пассажира |
| DELETE | /passengers/:id | Удалить пассажира  |

### Билеты — `/tickets`

| Метод  | Путь                | Описание                                          |
|--------|---------------------|---------------------------------------------------|
| GET    | /tickets            | Список билетов                                    |
| POST   | /tickets            | Забронировать билет (passenger из JWT)            |
| POST   | /tickets/:id/pay    | Оплатить свой билет (`reserved` → `paid`)         |
| PUT    | /tickets/:id        | Сменить статус (`paid` / `cancelled`)             |
| DELETE | /tickets/:id        | Удалить билет (освобождает место)                 |

Фильтры списка: `?flight_id=`, `?passenger_id=`, `?status=`, `?page=`, `?limit=`.

## Платёжный flow

```
POST /tickets                  →  status: reserved
                                  место → booked
                                  старт 5-минутного таймера
                                          │
                  ┌───────────────────────┴────────────────────────┐
                  │                                                │
   POST /tickets/:id/pay                                   таймер истёк
   (или PUT { status: "paid" })                                    │
                  │                                                │
                  ▼                                                ▼
       status: paid                                      status: cancelled
       PDF-воркер стартует                              место → available
       pdf_url появится позже
```

Таймер в `aviation` хранится в памяти (`map[uint]*time.Timer` под `sync.Mutex`):
после `paid` или `cancelled` он останавливается. Это даёт «холд места под
оплату» в духе настоящих авиасайтов.

### Примеры

**Забронировать билет — `POST /tickets`**

```json
{ "flight_id": 1, "seat_id": 3 }
```

`passenger_id` берётся из JWT автоматически. Ошибки: `403` (нет профиля
пассажира), `409` (место уже занято).

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

**Оплатить билет — `POST /tickets/:id/pay`**

Тело не требуется — пользователь определяется по JWT. Ошибки:
- `403` — билет принадлежит другому пользователю
- `409` — билет уже оплачен или отменён

**Создать рейс — `POST /flights`**

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

Места создаются отдельно — через миграции/seed или `POST /seats`.

**Создать место — `POST /seats`**

```json
{
  "flight_id": 1,
  "seat_number": "12A",
  "class": "economy",
  "price": 25000
}
```

Допустимые `class`: `economy`, `business`, `first`. `price` должна быть `> 0`.
Дубль `seat_number` в рамках одного рейса вернёт `409`.

**Обновить место — `PUT /seats/:id`**

```json
{ "price": 27000, "status": "available" }
```

Patch-style: передавайте только те поля, которые нужно изменить.

**Сменить статус билета — `PUT /tickets/:id`**

```json
{ "status": "paid" }
```

- `paid` — асинхронно запускается генерация PDF в `tickets`. Таймер брони
  останавливается.
- `cancelled` — место освобождается (`status` → `available`). Таймер брони
  останавливается.

**Удаление**

```
DELETE /flights/:id
DELETE /passengers/:id
DELETE /tickets/:id
```

`204 No Content` при успехе. При удалении билета со статусом, отличным от
`cancelled`, место на рейсе освобождается. Пассажира с активными билетами
удалить нельзя — `409`.

## Интеграция с tickets (PDF)

При оплате билета (`POST /tickets/:id/pay` или `PUT /tickets/:id { status: "paid" }`)
запускается воркер, который через Resty v2 вызывает
`POST {PDF_SERVICE_URL}/generate-ticket`. PDF загружается в MinIO, ссылка
сохраняется в `pdf_url` билета. Hooks `OnBeforeRequest` / `OnAfterResponse`
пишут метод, URL и статус в лог.

Если `tickets` недоступен — основной запрос не падает, ошибка только в логе.
Билет к этому моменту уже сохранён в БД, можно повторить смену статуса позже.

## Структура репозитория

```
.
├── docker-compose.yaml
├── aviation/                       # Go API
│   ├── main.go
│   ├── Dockerfile
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
│   │   ├── seat_handler.go
│   │   ├── passenger_handler.go
│   │   └── ticket_handler.go       # + reservation timer registry
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
