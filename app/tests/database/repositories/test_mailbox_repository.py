import pytest
from dataclasses import asdict

from app.tui.screens.search.search_screen import DataFilter
from app.database.db import transaction
from app.database.repositories import MailboxRepository, ServerRepository


@pytest.fixture
def data_filter() -> DataFilter:
    return DataFilter(
        search="",
        is_admin=None,
        disabled=None,
        server_id=None,
        quota_min="",
        quota_max="",
    )


def test_get_mailboxes():
    with transaction() as conn:
        repo = MailboxRepository(conn)
        _ = repo.get_rows({})


def test_get_filter_search(data_filter):
    filter = data_filter
    filter.search = "user@example.com"
    with transaction() as conn:
        repo = MailboxRepository(conn)
        models = repo.get_models(asdict(filter))

    for model in models:
        assert filter.search in model.address


def test_get_filter_server(data_filter):
    filter = data_filter

    with transaction() as conn:
        repo = ServerRepository(conn)
        servers = repo.get_tuples()

    for server, sid in servers:
        filter.search = "user@example.com"
        filter.server_id = sid

        with transaction() as conn:
            repo = MailboxRepository(conn)
            models = repo.get_models(asdict(filter))

        for model in models:
            assert model.server == server
