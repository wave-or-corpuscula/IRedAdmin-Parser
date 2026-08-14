from pathlib import Path

import pytest

from app.storage import ConfigStorage
from app.utils import ServerConfig

N = 10 # Number of test credentials created
test_storage_path = "app/tests/.iredcreds.test.json"


@pytest.fixture(scope="module", autouse=True)
def cleanup_test_file():
    path = Path(test_storage_path)
    path.unlink(missing_ok=True)

@pytest.fixture
def test_storage() -> ConfigStorage:
    return ConfigStorage(test_storage_path)


def server_creds_factory():
    def get_serv_cred(num: int) -> ServerConfig:
        return ServerConfig.from_dict(
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


def test_duplicate_credentials(test_storage):
    storage = test_storage
    factory = server_creds_factory()

    cred = factory(1)
    storage.save(cred)
    storage.save(cred)
    
    creds = storage.get_all()
    assert len(creds) == 1

