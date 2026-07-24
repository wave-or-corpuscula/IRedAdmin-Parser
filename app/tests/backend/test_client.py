import json

import pytest

from app.backend import IRedParserClient
from app.backend.exceptions import BackendError
from app.database.db import transaction
from app.database.repositories.server_repository import ServerRepository
from app.utils.config import ServerConfig

BINARY_PATH = "./iredparser/bin/iredparser"
TEST_CONFIG_PATH = ".test.creds.json"


@pytest.fixture
def get_auth_configs():
    with open(TEST_CONFIG_PATH) as f:
        data = json.load(f)

    def confs_generator():
        for raw in data:
            yield ServerConfig.from_dict(raw)

    return confs_generator()


def test_auth_check_invalid():
    invalid_config = ServerConfig(
        server="test",
        login="login",
        password="password",
    )

    client = IRedParserClient(BINARY_PATH)
    with pytest.raises(BackendError):
        client.auth_check(invalid_config)


def test_auth_check_valid(get_auth_configs):
    client = IRedParserClient(BINARY_PATH)
    for config in get_auth_configs:
        resp = client.auth_check(config)
        assert resp.authenticated
        assert resp.server == config.server
        assert len(resp.cookie_string) > 0


# TODO:: скорее всего бд не успевает закрываться и поэтому иногда выдает database is locked (5) (SQLITE_BUSY). Разобраться
def test_sync_mailboxes(get_auth_configs):
    client = IRedParserClient(BINARY_PATH)
    for config in get_auth_configs:
        with transaction() as conn:
            repo = ServerRepository(conn)
            repo.save(config.server)
        try:
            resp = client.sync_mailboxes(config)
            print(f"parsed {resp.amount} mailboxes from {resp.server}")
        except BackendError as e:
            print(e)
            raise
