import json
import subprocess
from typing import Optional

from app.backend import BackendError
from app.backend.models import AuthResponse, CLIResponse, SyncResponse, PasswordResponse
from app.utils.config import ServerConfig

BINARY_PATH = "./iredparser/bin/iredparser"


class IRedParserClient:
    def __init__(self, binary: str = BINARY_PATH):
        self.binary = binary

    def auth_check(self, config: ServerConfig) -> AuthResponse:
        resp = self._run(
            "auth-check",
            "--config", config.to_json(),
        )

        return AuthResponse.from_response(resp)

    def sync_mailboxes(self, config: ServerConfig) -> SyncResponse:
        resp = self._run(
            "sync",
            "--config", config.to_json(),
        )

        return SyncResponse.from_response(resp)

    def change_password(self, config: ServerConfig, mailbox: str, password: str = ""):
        resp = self._run(
            "change-password",
            "--config", config.to_json(),
            "-m", mailbox,
            "-p", password
        )

        return PasswordResponse.from_response(resp)

    def _run(self, *args, stdin_data: Optional[str] = None) -> CLIResponse:
        process = subprocess.run(
            [self.binary, *args],
            input=stdin_data,
            capture_output=True,
            text=True,
        )

        response = CLIResponse.from_dict(
            json.loads(process.stdout),
        )

        if not response.success:
            raise BackendError(response.error)  # type: ignore

        return response
