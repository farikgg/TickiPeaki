# TickiPeaki

REST API для бронирования авиабилетов. Go + Gin + GORM + PostgreSQL.

## Требования

- Go 1.22+
- PostgreSQL

## Запуск

```bash
cd aviation
go mod tidy
go run main.go
```

Сервер стартует на `http://localhost:8080`.

По умолчанию подключается к `localhost:5432`, БД `aviation`, пользователь `postgres`, пароль `postgres`.

Можно переопределить через переменную окружения:

```bash
DATABASE_URL="host=localhost user=myuser password=mypass dbname=aviation port=5432 sslmode=disable" go run main.go
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
aviation/
├── main.go
├── config/
│   └── database.go
├── models/
│   ├── flight.go
│   ├── passenger.go
│   └── ticket.go
├── repository/
│   ├── interfaces.go
│   └── postgres/
│       ├── flight_repo.go
│       ├── passenger_repo.go
│       └── ticket_repo.go
└── handlers/
    ├── flight_handler.go
    ├── passenger_handler.go
    └── ticket_handler.go
```
