from dataclasses import dataclass


@dataclass
class CLIError:
    code: int
    message: str
    type: str


@dataclass
class CLIResponse:
    success: bool
    data: dict | None
    error: CLIError | None

    @classmethod
    def from_dict(cls, d: dict) -> "CLIResponse":
        error = None

        if err := d.get("error"):
            error = CLIError(
                code=err["code"],
                type=err["type"],
                message=err["message"],
            )
        return cls(
            success=d["success"],
            data=d["data"],
            error=error,
        )


@dataclass
class AuthResponse:
    authenticated: bool
    server: str
    cookie_string: str

    @classmethod
    def from_response(cls, resp: CLIResponse) -> "AuthResponse":
        return cls(
            authenticated=resp.data["authenticated"],  # type: ignore
            server=resp.data["server"],  # type: ignore
            cookie_string=resp.data["cookie_string"], # type: ignore
        )


@dataclass
class SyncResponse:
    server: str
    amount: int

    @classmethod
    def from_response(cls, resp: CLIResponse) -> "SyncResponse":
        return cls(
            server=resp.data["server"],  # type: ignore
            amount=resp.data["amount"],  # type: ignore
        )

@dataclass
class PasswordResponse:
    mailbox: str
    password: str

    @classmethod
    def from_response(cls, resp: CLIResponse) -> "PasswordResponse":
        return cls(
            mailbox=resp.data["mailbox"],   # type: ignore
            password=resp.data["password"]  # type: ignore
        )

