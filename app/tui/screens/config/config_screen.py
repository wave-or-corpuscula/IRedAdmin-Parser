from textual import on, work
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical, VerticalScroll
from textual.screen import Screen
from textual.widgets import Button

from app.backend.exceptions import BackendError
from app.services import ConfigService
from app.storage import Credentials
from app.tui.widgets import ServerConfigWidget
from app.utils.config import ServerConfig


class ConfigScreen(Screen):
    CSS_PATH = "../../styles.tcss"

    def __init__(self, service: ConfigService, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.config_service = service

    def add_server_config(self, config: ServerConfig | None = None) -> None:
        configs_container = self.query_one("#servers-list-container")
        configs_container.mount(ServerConfigWidget(config))

    def compose(self) -> ComposeResult:
        with Vertical(id="config-container"):
            with VerticalScroll(id="servers-list-container"):
                pass
            with Horizontal(id="config-footer-container"):
                yield Button(
                    label="Добавить", id="add_config_button", variant="success"
                )
                yield Button(label="Назад", id="back_button")

    def on_mount(self) -> None:
        for creds in self.config_service.get_all():
            config = ServerConfig(
                server=creds.server, login=creds.login, password=creds.password
            )
            self.add_server_config(config)

    @on(ServerConfigWidget.ValidateRequested)
    @work(exclusive=True)
    async def validate_config(
        self, message: ServerConfigWidget.ValidateRequested
    ) -> None:

        if message._sender is None:
            return

        widget: ServerConfigWidget = message._sender  # type: ignore
        config: ServerConfig = message.config

        try:
            response = await self.config_service.validate(config)
            self.notify(
                severity="information",
                title="Успех!",
                message=f"Успешная авторизация на '{response.server}'",
            )
            widget.set_validation_result(True)
        except BackendError as e:
            self.notify(
                severity="error",
                title=e.code,
                message=e.message,
            )
            widget.set_validation_result(False)
            return
        finally:
            try:
                test_button = widget.query_one(".test-button", Button)
                test_button.disabled = False
            except Exception:
                pass

    @on(ServerConfigWidget.SaveRequested)
    def save_config(self, message: ServerConfigWidget.SaveRequested):
        config = message.config
        self.config_service.save_config(
            Credentials(
                server=config.server, login=config.login, password=config.password
            ),
        )
        self.notify(
            title="Успех!", message=f"Конфигурация для {config.server} сохранена"
        )

    @on(ServerConfigWidget.DeleteRequested)
    def delete_config(self, message: ServerConfigWidget.DeleteRequested):
        if message._sender is None:
            return

        server = message.server
        self.config_service.delete(server)

    @on(Button.Pressed, "#add_config_button")
    def add_server(self) -> None:
        self.add_server_config()

    @on(Button.Pressed, "#back_button")
    def nav_back(self) -> None:
        self.dismiss()
