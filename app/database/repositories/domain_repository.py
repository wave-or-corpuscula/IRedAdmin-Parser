import sqlite3
from typing import List

from app.database.models import DomainModel


class DomainRepository:
    def __init__(self, conn: sqlite3.Connection) -> None:
        self.conn = conn

    def insert(self, server_id: int, domain: DomainModel) -> int:
        query = """
        INSERT INTO Domains (server_id, disabled, name, display_name, quota_bytes, used_memory_bytes)
        VALUES (?, ?, ?, ?, ?, ?) RETURNING id;"""

        cur = self.conn.execute(
            query,
            (
                server_id,
                domain.disabled,
                domain.name,
                domain.display_name,
                domain.quota_bytes,
                domain.used_memory_bytes,
            ),
        )

        return cur.fetchone()[0]

    def get(self, *server_ids: int) -> List[DomainModel]:
        """
        Gives domains from servers with provided ids
        """
        query = "SELECT * FROM Domains"
        params = []

        if server_ids:
            placeholders = ", ".join(["?"] * len(server_ids))
            query += f" WHERE server_id IN ({placeholders})"
            params = list(server_ids)
        cur = self.conn.execute(query, params)

        models = []
        for row in cur.fetchall():
            models.append(DomainModel.from_row(row))

        return models
