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
