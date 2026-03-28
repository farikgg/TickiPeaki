# TickiPeaki

REST API для бронирования билетов на рейсы. Построен на Go + Gin.

## Запуск

```bash
cd ticketing
go run main.go
```

Сервер стартует на `http://localhost:8080`.

## Эндпоинты

### Рейсы (`/flights`)

| Метод  | Путь         | Описание            |
|--------|--------------|---------------------|
| GET    | /flights     | Список рейсов       |
| POST   | /flights     | Создать рейс        |
| GET    | /flights/:id | Получить рейс по ID |
| PUT    | /flights/:id | Обновить рейс       |
| DELETE | /flights/:id | Удалить рейс        |

Фильтры: `?type=`, `?origin=`, `?destination=`, `?page=`, `?limit=`

### Билеты (`/tickets`)

| Метод  | Путь          | Описание             |
|--------|---------------|----------------------|
| GET    | /tickets      | Список билетов       |
| POST   | /tickets      | Забронировать билет  |
| GET    | /tickets/:id  | Получить билет по ID |
| PUT    | /tickets/:id  | Обновить билет       |
| DELETE | /tickets/:id  | Удалить билет        |

Фильтры: `?flight_id=`, `?status=`, `?class=`, `?page=`, `?limit=`

## Тестирование через Postman

Базовый URL: `http://localhost:8080`

Для всех POST/PUT запросов: Headers -> `Content-Type: application/json`

### Создать рейс — POST `/flights`

```json
{
  "origin": "Almaty",
  "destination": "Astana",
  "type": "air",
  "carrier": "Air Astana",
  "departure_time": "2026-04-01T08:00:00Z",
  "arrival_time": "2026-04-01T09:30:00Z",
  "available_seats": 120,
  "price": 25000
}
```

### Обновить рейс — PUT `/flights/1`

Тело такое же как при создании, но с изменёнными полями.

### Забронировать билет — POST `/tickets`

```json
{
  "flight_id": 1,
  "passenger_name": "Иван Иванов",
  "passenger_email": "ivan@example.com",
  "seat_number": "12A",
  "class": "economy",
  "price": 25000
}
```

Допустимые значения `class`: `economy`, `business`, `first`

### Обновить билет — PUT `/tickets/1`

```json
{
  "flight_id": 1,
  "passenger_name": "Иван Иванов",
  "passenger_email": "ivan@example.com",
  "seat_number": "12A",
  "class": "business",
  "price": 45000,
  "status": "paid"
}
```

Допустимые значения `status`: `reserved`, `paid`, `cancelled`

### GET-запросы с фильтрами

```
GET /flights?type=air
GET /flights?origin=Almaty&destination=Astana
GET /flights?page=1&limit=5

GET /tickets?flight_id=1
GET /tickets?status=reserved&class=economy
GET /tickets?page=1&limit=10
```

### Удаление

```
DELETE /flights/1
DELETE /tickets/1
```

Возвращает `204 No Content` при успехе.

## Структура

```
ticketing/
├── main.go
├── models/
│   ├── flight.go
│   └── ticket.go
├── repository/
│   ├── interfaces.go
│   └── memory/
│       └── store.go
└── handlers/
    ├── flight_handler.go
    └── ticket_handler.go
```
