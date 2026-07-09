from app.database.db import transaction_factory
from app.services.config_service import ConfigService
from app.tests.storage.test_config_storage import server_creds_factory, test_storage


def test_config_service_crud(test_storage):
    N = 10
    factory = server_creds_factory()
    service = ConfigService(test_storage, transaction_factory(":memory:"))

    server_creds = [factory(i) for i in range(N)]

    for creds in server_creds:
        service.save_config(creds)

    for creds in service.get_all():
        assert creds in server_creds

    for creds in server_creds:
        service.delete(creds.server)

    assert len(service.get_all()) == 0
