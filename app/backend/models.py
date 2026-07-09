from dataclasses import dataclass


@dataclass
class CLIError:
    code: str
    message: str


@dataclass
class CLIResponse:
    success: bool
    data: dict | None
    error: CLIError | None

    @classmethod
    def from_dict(cls, d: dict) -> "CLIResponse":
        error = None

        if d.get("error"):
            error = CLIError(
                code=d["error"]["code"],
                message=d["error"]["message"],
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

    @classmethod
    def from_response(cls, resp: CLIResponse) -> "AuthResponse":
        return cls(
            authenticated=resp.data["authenticated"],  # type: ignore
            server=resp.data["server"],  # type: ignore
        )


@dataclass
class SyncResponse:
    server: str
    amount: int

    @classmethod
    def from_response(cls, resp: CLIResponse) -> "SyncResponse":
        print(resp.data)
        return cls(
            server=resp.data["server"],  # type: ignore
            amount=resp.data["amount"],  # type: ignore
        )
