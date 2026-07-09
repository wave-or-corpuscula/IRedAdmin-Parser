from dataclasses import dataclass
from typing import Tuple


@dataclass
class ServerModel:
    id: int
    name: str

    @classmethod
    def from_row(cls, row: Tuple[int, str]) -> "ServerModel":
        return cls(row[0], row[1])


@dataclass
class DomainModel:
    id: int
    server_id: int
    disabled: bool
    name: str
    display_name: str
    quota_bytes: int
    used_memory_bytes: int

    @classmethod
    def from_row(cls, row: Tuple[int, int, bool, str, str, int, int]) -> "DomainModel":
        return cls(*row)


@dataclass
class MailboxModel:
    id: int
    domain_id: int
    disabled: bool
    is_admin: bool
    display_name: str
    address: str
    quota_bytes: int
    used_memory_bytes: int


@dataclass
class DisplayModel:
    id: int
    server: str
    domain: str
    address: str
    display_name: str
    disabled: bool
    is_admin: bool
    quota_bytes: int
    used_memory_bytes: int
    usage_percent: int

    @classmethod
    def from_row(
        cls, row: Tuple[int, str, str, str, str, int, int, int, int, int]
    ) -> "DisplayModel":
        return cls(*row)
