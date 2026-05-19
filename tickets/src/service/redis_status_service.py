import json

from redis.asyncio import Redis
from typing import Union

from src.core.config import AppSettings, get_settings
from src.core.exceptions import RedisStatusServiceError
from src.core.logger import logger
from src.schemas.ticket_status import TicketStatus, TicketStatusResponse


class RedisStatusService:
    def __init__(self):
        settings: AppSettings = get_settings()
        self.redis_client: Redis = Redis.from_url(settings.redis_url)

    @staticmethod
    def _get_key(ticket_id: int) -> str:
        return f"ticket:{ticket_id}:status"

    async def get_status(self, ticket_id: int) -> Union[TicketStatusResponse, None]:
        key = self._get_key(ticket_id)
        try:
            value = await self.redis_client.get(name=key)
            if value is None:
                return None
            return json.loads(value)
        except Exception as error:
            logger.error(f"Redis get_status ошибка: {error}")
            raise RedisStatusServiceError(
                error=str(error),
                detail=f"Не удалось получить статус для ticket_id={ticket_id}",
            )

    async def set_status(
        self,
        ticket_id: int,
        status: TicketStatus,
        url: Union[str, None] = None,
    ) -> Union[dict, None]:
        key = self._get_key(ticket_id)
        value = json.dumps(
            {"status": status, "url": url}
        )

        try:
            await self.redis_client.set(name=key, value=value, ex=3600)
            logger.info(f"Статус билета {ticket_id} → {status}")
        except Exception as error:
            logger.error(f"Redis set_status ошибка: {error}")
            raise RedisStatusServiceError(
                error=str(error),
                detail=f"Не удалось записать статус для ticket_id={ticket_id}",
            )

    async def delete_status(self, ticket_id: int) -> None:
        key = self._get_key(ticket_id)
        try:
            await self.redis_client.delete(key)
            logger.info(f"Статус билета {ticket_id} удалён")
        except Exception as error:
            logger.error(f"Redis delete_status ошибка: {error}")
            raise RedisStatusServiceError(
                error=str(error),
                detail=f"Не удалось удалить статус для ticket_id={ticket_id}",
            )
