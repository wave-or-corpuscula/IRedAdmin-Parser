import typing
import sqlite3
from contextlib import contextmanager


DB_PATH = "./data/ireddata.db"


def create_connection(db_path: str = DB_PATH):
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn


@contextmanager
def transaction(
    db_path: str = DB_PATH,
) -> typing.Generator[sqlite3.Connection, None, None]:
    conn = create_connection(db_path)
    try:
        init_db(conn.cursor())
        yield conn
        conn.commit()
    except:
        conn.rollback()
        raise
    finally:
        conn.close()


def transaction_factory(db_path: str = DB_PATH):
    return lambda: transaction(db_path)


def init_db(cursor: sqlite3.Cursor):
    cursor.executescript("""
    CREATE TABLE IF NOT EXISTS "Servers" (
        "id" INTEGER,
        "name" TEXT NOT NULL UNIQUE,
        PRIMARY KEY("id" AUTOINCREMENT)
	);

	CREATE TABLE IF NOT EXISTS "Domains" (
        "id" INTEGER,
        "server_id" INTEGER NOT NULL,
        "disabled" BLOB NOT NULL,
        "name" TEXT NOT NULL,
        "display_name" TEXT,
        "quota_bytes" INTEGER NOT NULL,
        "used_memory_bytes" INTEGER NOT NULL,
        PRIMARY KEY("id" AUTOINCREMENT),
        UNIQUE("server_id", "name"),
        FOREIGN KEY("server_id") REFERENCES "Servers"("id") ON DELETE CASCADE ON UPDATE CASCADE
	);

	CREATE TABLE IF NOT EXISTS "Mailboxes" (
        "id" INTEGER,
        "domain_id" INTEGER NOT NULL,
        "address" TEXT NOT NULL,
        "display_name" TEXT,
        "disabled" BLOB NOT NULL,
        "is_admin" BLOB NOT NULL,
        "quota_bytes" INTEGER NOT NULL,
        "used_memory_bytes" INTEGER NOT NULL,
        PRIMARY KEY("id" AUTOINCREMENT),
        UNIQUE("domain_id", "address"),
        FOREIGN KEY("domain_id") REFERENCES "Domains"("id") ON DELETE CASCADE ON UPDATE CASCADE
	);	
	PRAGMA foreign_keys = ON;
    """)
