# Ticketing API

REST API для бронирования билетов на рейсы. Assignment 1.

## Запуск

```bash
go run main.go
```

Сервер стартует на `http://localhost:8080`.

## Эндпоинты

### Рейсы

| Метод  | Путь          | Описание             |
|--------|---------------|----------------------|
| GET    | /flights      | Список рейсов        |
| POST   | /flights      | Создать рейс         |
| GET    | /flights/{id} | Получить рейс по ID  |
| PUT    | /flights/{id} | Обновить рейс        |
| DELETE | /flights/{id} | Удалить рейс         |

Фильтры для `GET /flights`: `?type=`, `?origin=`, `?destination=`, `?page=`, `?limit=`

### Билеты

| Метод  | Путь           | Описание              |
|--------|----------------|-----------------------|
| GET    | /tickets       | Список билетов        |
| POST   | /tickets       | Забронировать билет   |
| GET    | /tickets/{id}  | Получить билет по ID  |
| PUT    | /tickets/{id}  | Обновить билет        |
| DELETE | /tickets/{id}  | Удалить билет         |

Фильтры для `GET /tickets`: `?flight_id=`, `?status=`, `?class=`, `?page=`, `?limit=`

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

## Примеры

```bash
# все рейсы
curl http://localhost:8080/flights

# только авиа
curl "http://localhost:8080/flights?type=air"

# создать билет
curl -X POST http://localhost:8080/tickets \
  -H "Content-Type: application/json" \
  -d '{
    "flight_id": 1,
    "passenger_name": "Криштиану Роналду",
    "passenger_email": "cr7@example.com",
    "seat_number": "7A",
    "class": "economy",
    "price": 25000
  }'
```
