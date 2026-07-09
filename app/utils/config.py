from dataclasses import dataclass, asdict
import json


class Config:
    pass


@dataclass
class ServerConfig:
    server: str = ""
    login: str = ""
    password: str = ""

    def to_dict(self) -> dict:
        return asdict(self)

    def to_json(self) -> str:
        return json.dumps(self.to_dict())

    @classmethod
    def from_dict(cls, data: dict) -> "ServerConfig":
        return cls(**data)

    @classmethod
    def create_empty(cls) -> "ServerConfig":
        return cls(server="", login="", password="")
