import asyncio
from typing import Callable, List

from app.backend import IRedParserClient
from app.backend.models import AuthResponse
from app.database.repositories import ServerRepository
from app.storage import ConfigStorage, Credentials
from app.utils import ServerConfig


class ConfigService:
    def __init__(self, storage: ConfigStorage, transaction: Callable):
        self.storage = storage
        self.transaction = transaction
        self.client = IRedParserClient()

    def save_config(self, creds: Credentials) -> int:
        with self.transaction() as conn:
            repo = ServerRepository(conn)
            config_id = repo.save(creds.server)
        self.storage.save(creds)
        return config_id

    def get_all(self) -> List[Credentials]:
        return self.storage.get_all()

    def delete(self, server: str):
        with self.transaction() as conn:
            repo = ServerRepository(conn)
            repo.delete(server)
        self.storage.delete(server)

    async def validate(self, config: ServerConfig) -> AuthResponse:
        loop = asyncio.get_event_loop()
        response = await loop.run_in_executor(None, self.client.auth_check, config)
        return response
