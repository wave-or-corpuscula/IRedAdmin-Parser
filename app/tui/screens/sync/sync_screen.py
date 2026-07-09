import asyncio
from typing import List

from textual import on, work
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.message import Message
from textual.screen import ModalScreen
from textual.widgets import Button, Label

from app.backend.exceptions import BackendError
from app.services.sync_service import SyncService
from app.storage.config_storage import Credentials
from app.utils import _create_config_service
from app.utils.config import ServerConfig


class SyncButton(Button):
    class Synced(Message):
        def __init__(self, creds: Credentials):
            super().__init__()
            self.creds = creds

        def as_server_config(self) -> ServerConfig:
            return ServerConfig(
                server=self.creds.server,
                login=self.creds.login,
                password=self.creds.password,
            )

    def __init__(self, label: str, creds: Credentials, **kwargs):
        super().__init__(label, **kwargs)
        self.creds = creds

    def _on_click(self, event):
        self.disabled = True
        self.label = "⚠️"
        self.variant = "warning"
        self.post_message(self.Synced(creds=self.creds))
        return super()._on_click(event)

    def server_config(self) -> ServerConfig:
        return ServerConfig(
            server=self.creds.server,
            login=self.creds.login,
            password=self.creds.password,
        )


class SyncScreen(ModalScreen):
    CSS_PATH = "../../styles.tcss"

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.config_service = _create_config_service()
        self.sync_service = SyncService()

    def compose(self) -> ComposeResult:
        with Vertical(id="sync-screen"):
            for creds in self.config_service.get_all():
                with Horizontal(classes="sync-row"):
                    yield Label(creds.server, classes="sync-label")
                    yield SyncButton(
                        label="♻️",
                        creds=creds,
                        variant="primary",
                    )
            with Horizontal(id="sync-nav-block"):
                yield Button("Назад", id="nav-back", variant="error")
                yield Button("Все", id="sync-all", variant="success")

    @on(SyncButton.Synced)
    def sync_signal(self, message: SyncButton.Synced) -> None:
        self._sync_one_work(message._sender)  # type: ignore

    @on(Button.Pressed, "#nav-back")
    def nav_back(self) -> None:
        self.dismiss()

    @work
    async def _sync_one_work(self, button: SyncButton) -> None:
        await self._sync_one(button)

    async def _sync_one(self, button: SyncButton) -> None:
        try:
            response = await self.sync_service.sync(button.server_config())
            self.notify(
                title="Успех!",
                message=f"Получено {response.amount} ящиков из {response.server}",
            )

            button.variant = "success"
            button.label = "✅"
            await asyncio.sleep(1.5)
            button.variant = "primary"
            button.label = "♻️"
            button.disabled = False

        except BackendError as e:
            self.notify(
                title=e.code,
                message=e.message,
                severity="error",
            )
            button.variant = "error"
            button.label = "❌"
            await asyncio.sleep(2)
            button.variant = "primary"
            button.label = "♻️"
            button.disabled = False

        except Exception as e:
            self.notify(
                title="Ошибка",
                message=f"Неизвестная ошибка: {str(e)}",
                severity="error",
            )
            button.variant = "error"
            button.label = "❌"
            await asyncio.sleep(2)
            button.variant = "primary"
            button.label = "♻️"
            button.disabled = False

    def get_all_sync_buttons(self) -> List[SyncButton]:
        return [child for child in self.query(SyncButton)]

    @on(Button.Pressed, "#sync-all")
    def sync_all(self, event: Button.Pressed) -> None:
        self.run_sync_all(event)

    @work
    async def run_sync_all(self, event: Button.Pressed) -> None:
        event.button.disabled = True

        buttons = self.get_all_sync_buttons()
        tasks = [self._sync_one(button) for button in buttons]
        await asyncio.gather(*tasks)

        event.button.disabled = False
