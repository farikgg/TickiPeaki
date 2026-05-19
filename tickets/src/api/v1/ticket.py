from fastapi import APIRouter, BackgroundTasks, Depends, HTTPException
from fastapi.responses import JSONResponse

from src.schemas.ticket import TicketSchema
from src.schemas.ticket_status import TicketStatus, TicketStatusResponse
from src.service.pdf_service import PdfGeneratorService
from src.service.s3_service import S3StorageService
from src.service.redis_status_service import RedisStatusService
from src.core.exceptions import BaseExceptionError
from src.core.logger import logger

router = APIRouter(tags=["Generate PDF ticket"])


async def process_ticket(
    ticket: TicketSchema,
    pdf_service: PdfGeneratorService,
    s3_service: S3StorageService,
    status_service: RedisStatusService,
) -> None:
    try:
        pdf_bytes: bytes = await pdf_service.generate_ticket_pdf(ticket)

        filename = f"{ticket.ticket_id}_{ticket.passenger_name.replace(' ', '_')}.pdf"
        url = await s3_service.upload_file(file=pdf_bytes, filename=filename)

        await status_service.set_status(
            ticket_id=ticket.ticket_id,
            status=TicketStatus.READY,
            url=url,
        )
        logger.info(f"Билет {ticket.ticket_id} успешно сгенерирован и загружен")
    except BaseExceptionError as error:
        logger.error(f"Ошибка обработки билета {ticket.ticket_id}: {error}")
        await status_service.set_status(
            ticket_id=ticket.ticket_id,
            status=TicketStatus.FAILED,
        )

@router.post("/ticket_generate", status_code=202)
async def generate_ticket(
        ticket: TicketSchema,
        background_tasks: BackgroundTasks,
        pdf_service: PdfGeneratorService = Depends(PdfGeneratorService),
        s3_service: S3StorageService = Depends(S3StorageService),
        status_service: RedisStatusService = Depends(RedisStatusService),
):
    await status_service.set_status(
        ticket_id=ticket.ticket_id,
        status=TicketStatus.PROCESSING,
    )

    background_tasks.add_task(
        process_ticket,
        ticket,
        pdf_service,
        s3_service,
        status_service,
    )

    return JSONResponse(
        status_code=202,
        content={
            "ticket_id": ticket.ticket_id,
            "status": "processing",
        },
    )

@router.get("/ticket_status/{ticket_id}")
async def get_ticket_status(
        ticket_id: int,
        status_service: RedisStatusService = Depends(RedisStatusService),
) -> TicketStatusResponse:
    ticket_status: TicketStatusResponse = await status_service.get_status(ticket_id)

    if ticket_status is None:
        raise HTTPException(
            status_code=404,
            detail=f"Статус для билета {ticket_id} не найден",
        )

    return ticket_status
