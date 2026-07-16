from dataclasses import dataclass
from typing import Tuple

def format_bytes(size_bytes: int) -> str:
    if size_bytes == -1:
        return "∞"
    
    if size_bytes < 0:
        return "0 Б"
        
    labels = ["Б", "КБ", "МБ", "ГБ", "ТБ"]
    if size_bytes == 0:
        return "0 Б"
        
    # Рассчитываем индекс единицы измерения
    # Используем целочисленное деление, чтобы найти нужный масштаб
    import math
    i = int(math.floor(math.log(size_bytes, 1024)))
    if i >= len(labels):
        i = len(labels) - 1
        
    p = math.pow(1024, i)
    s = round(size_bytes / p, 2)
    
    # Избавляемся от плавающей точки .0, если число целое
    if s.is_integer():
        s = int(s)
        
    return f"{s} {labels[i]}"

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
        cls, row: Tuple[int, str, str, str, str, bool, bool, int, int, int]
    ) -> "DisplayModel":
        return cls(*row)

    def to_row(self) -> Tuple[int, str, str, str, str, str, str, str, str, str]:
        return (
            self.id,
            self.server,
            self.domain,
            self.address,
            self.display_name or "",
            "✅" if not self.disabled else "❌",
            "✅" if self.is_admin else "❌",
            format_bytes(self.quota_bytes),
            format_bytes(self.used_memory_bytes),
            f"{int(self.usage_percent)}",
        )
