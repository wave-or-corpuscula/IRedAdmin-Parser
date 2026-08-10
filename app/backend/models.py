from typing import Any, Dict, Optional
from pydantic import BaseModel


class CLIError(BaseModel):
    code: int
    message: str
    type: str


class CLIResponse(BaseModel):
    success: bool
    data: Optional[Dict[str, Any]]
    error: Optional[CLIError]


class AuthResponse(BaseModel):
    authenticated: bool
    server: str
    cookie_string: str

    @classmethod
    def from_cli(cls, resp: CLIResponse) -> "AuthResponse":
        return cls.model_validate(resp.data)


class SyncResponse(BaseModel):
    server: str
    amount: int

    @classmethod
    def from_cli(cls, resp: CLIResponse) -> "SyncResponse":
        return cls.model_validate(resp.data)


class PasswordResponse(BaseModel):
    mailbox: str
    password: str

    @classmethod
    def from_cli(cls, resp: CLIResponse) -> "PasswordResponse":
        return cls.model_validate(resp.data)

