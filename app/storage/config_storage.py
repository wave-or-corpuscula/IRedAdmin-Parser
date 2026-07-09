import json
from dataclasses import dataclass, asdict
from typing import Dict, List


@dataclass
class Credentials:
    server: str
    login: str
    password: str

    @classmethod
    def from_dict(cls, data: dict) -> "Credentials":
        return cls(**data)


class ConfigStorage:
    def __init__(self, path: str):
        self.path = path

    def save(self, creds: Credentials):
        storage = self._load()

        storage.append(asdict(creds))

        self._write(storage)

    def delete(self, key: str):
        storage = self._load()

        for i in range(len(storage)):
            if storage[i].get("server") == key:
                storage.pop(i)
                break

        self._write(storage)

    def get(self, key: str) -> Credentials | None:
        storage = self._load()

        for i in range(len(storage)):
            if storage[i].get("server") == key:
                return Credentials.from_dict(storage[i])

    def get_all(self) -> List[Credentials]:
        storage = self._load()
        if len(storage) == 0:
            return []

        creds = []
        for data in storage:
            creds.append(Credentials.from_dict(data))

        return creds

    def _load(self) -> List[Dict[str, str]]:
        try:
            with open(self.path) as f:
                return json.load(f)
        except (FileNotFoundError, json.JSONDecodeError):
            return []

    def _write(self, data: List[Dict[str, str]]):
        with open(self.path, "w") as f:
            json.dump(data, f)
