from app.database import transaction


def test_get_connection():
    with transaction(":memory:") as _:
        pass
