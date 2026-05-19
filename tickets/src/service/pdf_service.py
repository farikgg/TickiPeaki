import asyncio

from typing import Any, Dict
from pathlib import Path
from functools import partial

from jinja2 import Environment, FileSystemLoader
from weasyprint import HTML

from src.schemas.ticket import TicketSchema

TEMPLATES_DIR = Path(__file__).resolve().parent.parent.parent / "templates"


class PdfGeneratorService:
    def __init__(self):
        self.environment: Environment = Environment(loader=FileSystemLoader(str(TEMPLATES_DIR)))

    def _render_pdf(self, ticket_data: dict) -> bytes:
        template = self.environment.get_template("ticket.html")
        html_string = template.render(**ticket_data)
        return HTML(string=html_string).write_pdf()

    async def generate_ticket_pdf(self, ticket: TicketSchema) -> bytes:
        ticket_data: Dict[str, Any] = ticket.model_dump()
        loop = asyncio.get_running_loop()
        pdf_bytes = await loop.run_in_executor(None, partial(self._render_pdf, ticket_data))
        return pdf_bytes
