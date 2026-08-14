from app.backend import IRedParserClient
from app.backend import PasswordResponse
from app.utils import ServerConfig


class PasswordService:

    def __init__(self) -> None:
        self.client = IRedParserClient()

    def change_password(self, config: ServerConfig, mailbox: str = "", password: str = "") -> PasswordResponse:
        resp = self.client.change_password(config, mailbox, password)
        return resp
