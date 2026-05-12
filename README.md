# TickiPeaki

REST API для бронирования авиабилетов. Go + Gin + GORM + PostgreSQL.
PDF-билеты генерируются отдельным Python-сервисом, который вызывается из Go через Resty.

## Архитектура

```
aviation/ (Go :8080)               db/ (PostgreSQL :5432)
    └── PUT /tickets/:id (status→paid)
            └── clients.PDFClient.GenerateTicket()
                    └── POST http://pdf-service:8000/generate-ticket
                            └── tickets/ (Python FastAPI :8000)
```

## Требования

- Docker + Docker Compose

Локальный запуск без Docker (опционально):
- Go 1.23+
- PostgreSQL

## Запуск через Docker

```bash
docker-compose up --build
```

Поднимутся три сервиса:

| Сервис      | Порт | Описание                       |
|-------------|------|--------------------------------|
| aviation    | 8080 | Go API (Gin + GORM)            |
| pdf-service | 8000 | Python FastAPI, генерация PDF  |
| db          | 5432 | PostgreSQL 15                  |

Все сервисы в общей сети `aviation-net`. БД хранится в volume `postgres_data`.
`aviation` стартует только после healthcheck postgres.

Остановка с очисткой данных:

```bash
docker-compose down -v
```

## Переменные окружения (aviation)

| Переменная        | По умолчанию                                                                  |
|-------------------|-------------------------------------------------------------------------------|
| `DATABASE_URL`    | `host=db user=postgres password=postgres dbname=aviation port=5432 sslmode=disable` |
| `PDF_SERVICE_URL` | `http://pdf-service:8000`                                                     |
| `JWT_SECRET`      | `supersecret`                                                                 |

## Локальный запуск без Docker

```bash
cd aviation
go mod tidy
go run main.go
```

Сервер стартует на `http://localhost:8080`. По умолчанию подключается к `localhost:5432`,
БД `aviation`, пользователь `postgres`, пароль `postgres`.

```bash
DATABASE_URL="host=localhost user=myuser password=mypass dbname=aviation port=5432 sslmode=disable" \
PDF_SERVICE_URL="http://localhost:8000" \
go run main.go
```

## Эндпоинты

### Рейсы (`/flights`)

| Метод  | Путь         | Описание       |
|--------|--------------|----------------|
| GET    | /flights     | Список рейсов  |
| POST   | /flights     | Создать рейс   |
| PUT    | /flights/:id | Обновить рейс  |
| DELETE | /flights/:id | Удалить рейс   |

Фильтры: `?origin=`, `?destination=`, `?carrier=`, `?page=`, `?limit=`

### Пассажиры (`/passengers`)

| Метод  | Путь            | Описание           |
|--------|-----------------|--------------------|
| GET    | /passengers     | Список пассажиров  |
| POST   | /passengers     | Создать пассажира  |
| PUT    | /passengers/:id | Обновить пассажира |
| DELETE | /passengers/:id | Удалить пассажира  |

Фильтры: `?page=`, `?limit=`

### Билеты (`/tickets`)

| Метод  | Путь         | Описание            |
|--------|--------------|---------------------|
| GET    | /tickets     | Список билетов      |
| POST   | /tickets     | Забронировать билет |
| PUT    | /tickets/:id | Обновить билет      |
| DELETE | /tickets/:id | Удалить билет       |

Фильтры: `?flight_id=`, `?passenger_id=`, `?status=`, `?class=`, `?page=`, `?limit=`

## Интеграция с pdf-service

При обновлении билета через `PUT /tickets/:id` со статусом `paid` (если предыдущий
статус был не `paid`) Go-сервис вызывает `POST {PDF_SERVICE_URL}/generate-ticket`
через Resty v2. Hooks `OnBeforeRequest` / `OnAfterResponse` логируют метод, URL и
статус ответа.

Если pdf-service недоступен — основной запрос не падает, ошибка только пишется в лог.
Билет к этому моменту уже сохранён в БД.

## Тестирование через Postman

Базовый URL: `http://localhost:8080`

Для всех POST/PUT запросов: Headers -> `Content-Type: application/json`

### Создать рейс — POST `/flights`

```json
{
  "flight_number": "KC301",
  "origin": "ALA",
  "destination": "NQZ",
  "carrier": "Air Astana",
  "departure_time": "2026-04-01T08:00:00Z",
  "arrival_time": "2026-04-01T09:30:00Z",
  "available_seats": 120,
  "price": 25000
}
```

### Создать пассажира — POST `/passengers`

```json
{
  "full_name": "Иван Иванов",
  "email": "ivan@example.com",
  "phone": "+77001234567",
  "passport_num": "N12345678"
}
```

### Забронировать билет — POST `/tickets`

```json
{
  "flight_id": 1,
  "passenger_id": 1,
  "seat_number": "12A",
  "class": "economy",
  "price": 25000
}
```

Допустимые значения `class`: `economy`, `business`, `first`

При бронировании `available_seats` рейса уменьшается на 1. Если мест нет — `409`.

### Обновить билет — PUT `/tickets/1`

```json
{
  "seat_number": "12A",
  "class": "business",
  "price": 45000,
  "status": "paid"
}
```

Допустимые значения `status`: `reserved`, `paid`, `cancelled`

При смене статуса на `paid` запускается генерация PDF в pdf-service.
При смене статуса на `cancelled` место на рейсе освобождается.

### GET-запросы с фильтрами

```
GET /flights?origin=ALA&destination=NQZ
GET /flights?carrier=Air%20Astana&page=1&limit=5

GET /passengers?page=1&limit=10

GET /tickets?flight_id=1
GET /tickets?status=reserved&class=economy
GET /tickets?passenger_id=1&page=1&limit=10
```

### Удаление

```
DELETE /flights/1
DELETE /passengers/1
DELETE /tickets/1
```

Возвращает `204 No Content` при успехе.

Удаление пассажира с активными билетами вернёт `409`.

При удалении билета со статусом, отличным от `cancelled`, место на рейсе освобождается.

## Структура

```
.
├── docker-compose.yaml
├── aviation/                 # Go API
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── main.go
│   ├── config/
│   │   └── database.go
│   ├── models/
│   │   ├── flight.go
│   │   ├── passenger.go
│   │   └── ticket.go
│   ├── repository/
│   │   ├── interfaces.go
│   │   └── postgres/
│   │       ├── flight_repo.go
│   │       ├── passenger_repo.go
│   │       └── ticket_repo.go
│   ├── handlers/
│   │   ├── flight_handler.go
│   │   ├── passenger_handler.go
│   │   └── ticket_handler.go
│   └── clients/
│       └── pdf_client.go     # Resty-клиент к pdf-service
└── tickets/                  # Python FastAPI, генерация PDF
    ├── Dockerfile
    ├── pyproject.toml
    ├── src/
    └── templates/
```
