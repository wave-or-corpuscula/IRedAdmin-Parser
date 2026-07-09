from sqlite3 import Connection
from typing import List, Tuple

from app.database.models import ServerModel


class ServerRepository:
    def __init__(self, conn: Connection) -> None:
        self.conn = conn

    def save(self, name: str) -> int:
        query = """
        INSERT INTO Servers (name) 
        VALUES (?)
        ON CONFLICT(name)
        DO UPDATE SET
            name = excluded.name
        RETURNING id;
        """
        cur = self.conn.execute(query, (name,))
        return cur.fetchone()[0]

    def get(self) -> List[ServerModel]:
        query = "SELECT * FROM Servers;"
        cur = self.conn.execute(query)

        models = []
        for row in cur.fetchall():
            models.append(ServerModel.from_row(row))

        return models

    def get_tuples(self) -> List[Tuple[str, int]]:
        return [(s.name, s.id) for s in self.get()]

    def delete(self, name: str) -> None:
        query = "DELETE FROM Servers WHERE name = ?;"
        self.conn.execute(query, (name,))
