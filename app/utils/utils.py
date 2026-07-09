from app.database.db import transaction_factory
from app.services.config_service import ConfigService
from app.storage.config_storage import ConfigStorage


def _create_config_service() -> ConfigService:
    storage = ConfigStorage(".iredcreds.json")
    return ConfigService(storage, transaction_factory())
