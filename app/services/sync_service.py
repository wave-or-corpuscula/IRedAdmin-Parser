import asyncio

from app.backend.client import IRedParserClient
from app.backend.models import SyncResponse
from app.utils.config import ServerConfig


class SyncService:
    def __init__(self) -> None:
        self.client = IRedParserClient()

    async def sync(self, config: ServerConfig) -> SyncResponse:
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(
            None,
            self.client.sync_mailboxes,
            config,
        )
