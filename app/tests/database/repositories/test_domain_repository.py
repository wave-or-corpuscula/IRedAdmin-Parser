from collections import defaultdict
import pytest

from app.database.models import DomainModel, ServerModel
from app.database.repositories import DomainRepository, ServerRepository
from app.database import transaction


@pytest.fixture
def get_test_domain():
    def factory(num: int):
        return DomainModel(
            id=-1,
            server_id=-1,
            disabled=False,
            name=f"test_domain_{num}",
            display_name=f"Test Domain {num}",
            quota_bytes=10240,
            used_memory_bytes=2048,
        )

    return factory


def test_domains_repo(get_test_domain):
    N_servers = 3
    N_domains = 10

    with transaction(":memory:") as conn:
        server_domains = defaultdict(list)

        serv_repo = ServerRepository(conn)
        domain_repo = DomainRepository(conn)

        servers = [ServerModel(id=-1, name=f"Server {i + 1}") for i in range(N_servers)]
        for i in range(len(servers)):
            server_id = serv_repo.save(servers[i].name)
            servers[i].id = server_id

        for server in servers:
            domains = [get_test_domain(i + 1) for i in range(N_domains)]
            for i in range(len(domains)):
                domain_id = domain_repo.insert(server.id, domains[i])
                domains[i].id = domain_id
                domains[i].server_id = server.id
            server_domains[server.name] = domains

        for server in servers:
            db_domains = domain_repo.get(server.id)
            assert len(db_domains) == len(server_domains[server.name])
            for db_domain in db_domains:
                assert db_domain in server_domains[server.name]

        all_domains = domain_repo.get()
        assert len(all_domains) == sum(
            len(domains) for domains in server_domains.values()
        )
