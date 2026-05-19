import logging

from pathlib import Path
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import Literal
from functools import lru_cache

from src.core.constants import FILE_ENCODER

BASE_DIR = Path(__file__).resolve().parent.parent.parent
ENV_FILE = BASE_DIR / ".env"

if not ENV_FILE.exists():
    logging.error("Отсутствует файл .env в корне проекта.")


class _ProjectBaseSettings(BaseSettings):
    """Базовые настройки."""
    model_config = SettingsConfigDict(
        env_file=ENV_FILE,
        env_file_encoding=FILE_ENCODER,
        populate_by_name=True,
        extra="ignore",
    )


class AppSettings(_ProjectBaseSettings):
    """Настройки сервиса по созданию PDF билетов"""
    environment: Literal["local", "dev", "prod"] = Field(default="local", validation_alias="APP_ENV")
    app_name: str = Field(default="Generate ticket PDF service")
    debug: bool = Field(default=False, validation_alias="APP_DEBUG")

    # redis
    redis_url: str = Field(..., alias="REDIS_URL")

    # s3
    endpoint_url: str = Field(..., alias="S3_ENDPOINT_URL")
    access_key: str = Field(..., alias="S3_ACCESS_KEY")
    secret_key: str = Field(..., alias="S3_SECRET_KEY")
    bucket: str = Field(..., alias="S3_BUCKET")

@lru_cache
def get_settings() -> AppSettings:
    return AppSettings()
