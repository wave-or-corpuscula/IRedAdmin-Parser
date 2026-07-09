import sqlite3
from typing import Any, Dict, List

from pytest import param

from app.database.models import DisplayModel


class MailboxRepository:
    def __init__(self, conn: sqlite3.Connection) -> None:
        self.conn = conn

    def get_rows(
        self,
        filters: Dict[str, Any],
        sort_key: str = "address",
        sort_desc: bool = False,
    ) -> List[tuple]:
        ALLOWED_SORTS = {
            "id": "m.id",
            "server": "s.name",
            "domain": "d.name",
            "address": "m.address",
            "display_name": "m.display_name",
            "disabled": "m.disabled",
            "is_admin": "m.is_admin",
            "quota": "m.quota_bytes",
            "used_memory": "m.used_memory_bytes",
        }
        query = """
        SELECT 
            m.id, 
            s.name AS server_name, 
            d.name AS domain_name, 
            m.address, 
            m.display_name, 
            m.disabled, 
            m.is_admin, 
            m.quota_bytes, 
            m.used_memory_bytes,
            CASE 
                WHEN m.quota_bytes > 0 THEN ROUND((CAST(m.used_memory_bytes AS REAL) / m.quota_bytes) * 100, 1)
                ELSE 0.0
            END AS usage_percent
        FROM Mailboxes m
        JOIN Domains d ON m.domain_id = d.id
        JOIN Servers s ON d.server_id = s.id
        WHERE 1=1
        """

        params = {}

        if filters.get("server_id") is not None:
            query += " AND s.id = :server_id"
            params["server_id"] = int(filters["server_id"])

        if filters.get("search"):
            query += (
                " AND (m.address LIKE :search_str OR m.display_name LIKE :search_str)"
            )
            params["search_str"] = f"%{filters['search']}%"

        if filters.get("disabled") is not None:
            query += " AND m.disabled = :disabled"
            params["disabled"] = int(filters["disabled"])

        if filters.get("is_admin") is not None:
            query += " AND m.is_admin = :is_admin"
            params["is_admin"] = int(filters["is_admin"])

        if filters.get("quota_min") is not None and filters["quota_min"] != "":
            query += " AND m.quota_bytes >= :quota_min"
            params["quota_min"] = int(filters["quota_min"])

        if filters.get("quota_max") is not None and filters["quota_max"] != "":
            query += " AND m.quota_bytes >= :quota_max"
            params["quota_max"] = int(filters["quota_max"])

        db_sort_column = ALLOWED_SORTS.get(sort_key, "m.address")
        direction = "DESC" if sort_desc else "ASC"
        query += f" ORDER BY {db_sort_column} {direction}"

        cur = self.conn.execute(query, params)

        return cur.fetchall()

    def get_models(
        self,
        filters: Dict[str, Any],
        sort_key: str = "address",
        sort_desc: bool = False,
    ) -> List[DisplayModel]:
        rows = self.get_rows(filters, sort_key, sort_desc)

        models = []
        for row in rows:
            models.append(DisplayModel.from_row(row))

        return models
