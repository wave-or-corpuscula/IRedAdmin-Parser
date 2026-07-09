import pytest

from app.storage import ConfigStorage, Credentials

N = 10


@pytest.fixture
def test_storage() -> ConfigStorage:
    return ConfigStorage("app/tests/.iredcreds.test.json")


def server_creds_factory():
    def get_serv_cred(num: int) -> Credentials:
        return Credentials.from_dict(
            {
                "server": f"server{num}",
                "login": f"user{num}@mail.by",
                "password": f"password{num}",
            }
        )

    return get_serv_cred


def test_config_storage_save(test_storage):
    storage = test_storage

    factory = server_creds_factory()

    for i in range(N):
        creds = factory(i)
        storage.save(creds)

    for i in range(N):
        creds = factory(i)
        saved_creds = storage.get(creds.server)
        assert creds == saved_creds


def test_config_storage_get_all(test_storage):
    storage = test_storage

    creds = storage.get_all()
    assert len(creds) == N


def test_config_storage_delete(test_storage):
    storage = test_storage
    factory = server_creds_factory()

    for i in range(N):
        creds = factory(i)
        storage.delete(creds.server)

    for i in range(N):
        creds = factory(i)
        saved = storage.get(creds.server)
        assert saved is None
