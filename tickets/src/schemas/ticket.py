from datetime import datetime
from decimal import Decimal

from pydantic import BaseModel, ConfigDict, EmailStr, Field


class TicketSchema(BaseModel):
    ticket_id: int = Field(..., description="Идентификатор билета")
    flight_number: str = Field(..., description="Номер рейса")
    origin: str = Field(..., description="Место отбытия")
    destination: str = Field(..., description="Место прибытия")
    departure_time: datetime = Field(..., description="Время отправления")
    arrival_time: datetime = Field(..., description="Время прибытия")
    carrier: str = Field(..., description="Авиа компания")
    passenger_name: str = Field(..., description="Имя пассажира")
    passenger_email: EmailStr = Field(..., description="@mail пассажира")
    seat_number: str = Field(..., description="Номер посадочного места")
    flight_class: str = Field(..., description="Класс перелета", alias="class")
    price: Decimal = Field(..., description="Цена билета")

    model_config = ConfigDict(populate_by_name=True)
