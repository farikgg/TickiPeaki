from pydantic import BaseModel
from enum import StrEnum


class TicketStatus(StrEnum):
    PROCESSING = "processing"
    READY = "ready"
    FAILED = "failed"


class TicketStatusResponse(BaseModel):
    status: TicketStatus
    url: str | None = None
