import io

from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse

from src.schemas.ticket import TicketSchema
from src.service.pdf_service import generate_ticket_pdf
from src.core.logger import logger

router = APIRouter(tags=["Generate PDF ticket"])

@router.post("/generate_ticket")
async def generate_ticket(ticket: TicketSchema) -> StreamingResponse:
    try:
        pdf_bytes = await generate_ticket_pdf(ticket)
        filename = f"{ticket.ticket_id}_{ticket.passenger_name.replace(' ', '_')}"
        logger.info(f"Созданный PDF файл: {filename}")

        headers = {"Content-Disposition": f"attachment; filename={filename}.pdf"}
        return StreamingResponse(
            content=io.BytesIO(pdf_bytes),
            media_type="application/pdf",
            headers=headers,
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
