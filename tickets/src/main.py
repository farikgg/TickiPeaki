from contextlib import asynccontextmanager
from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware

from src.core.config import get_settings
from src.core.logger import logger
from src.api.v1.ticket import router as generate_ticket_router

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Запуск сервиса PDF")
    yield
    logger.info("Остановка сервиса...")

async def validation_exception_handler(
        request: Request,
        exc: RequestValidationError
):
    error_details = exc.errors()
    logger.error(f"Ошибка валидации (422) при запросе к {request.url.path}: {error_details}")
    return JSONResponse(
        status_code=422,
        content={"detail": error_details},
    )

def create_application() -> FastAPI:
    settings = get_settings()
    application = FastAPI(
        title=settings.app_name,
        version="1.0.0",
        lifespan=lifespan,
        debug=settings.debug
    )

    application.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    application.include_router(generate_ticket_router, prefix="/api/v1")
    application.add_exception_handler(RequestValidationError, validation_exception_handler)

    return application

app = create_application()

@app.get("/")
def health_check():
    return {"status": "ok", "message": "Сервис PDF работает!"}
