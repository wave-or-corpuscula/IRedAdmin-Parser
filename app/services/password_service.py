from app.backend.client import IRedParserClient
from app.backend.models import PasswordResponse
from app.utils.config import ServerConfig


class PasswordService:

    def __init__(self) -> None:
        self.client = IRedParserClient()

    def change_password(self, config: ServerConfig, mailbox: str = "", password: str = "") -> PasswordResponse:
        resp = self.client.change_password(config, mailbox, password)
        return resp
