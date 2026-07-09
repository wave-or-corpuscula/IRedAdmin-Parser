from app.database import transaction
from app.database.repositories import ServerRepository


def test_get_servers():
    with transaction(":memory:") as conn:
        repo = ServerRepository(conn)
        _ = repo.get()


def test_upsert_servers():
    serv_name = "test_server"
    with transaction(":memory:") as conn:
        repo = ServerRepository(conn)
        first_id = repo.save(serv_name)
        next_id = repo.save(serv_name)

        assert first_id == next_id

        db_server = repo.get()

        assert db_server[-1].name == serv_name


def test_delete_server():
    serv_name = "test_server"
    with transaction(":memory:") as conn:
        repo = ServerRepository(conn)
        repo.delete(serv_name)
