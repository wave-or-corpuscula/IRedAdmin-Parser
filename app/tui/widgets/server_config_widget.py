from app.utils import ServerConfig
from textual import on
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.events import Event
from textual.message import Message
from textual.widget import Widget
from textual.widgets import Button, Input

from .secret_input_widget import SecretInput


class ServerConfigWidget(Widget):
    class ValidateRequested(Event):
        config: ServerConfig

        def __init__(self, sender: "ServerConfigWidget", config: ServerConfig):
            super().__init__()
            self.sender = sender
            self.config = config

        @property
        def control(self) -> "ServerConfigWidget":
            return self.sender

    class SaveRequested(Message):
        config: ServerConfig

        def __init__(self, sender: "ServerConfigWidget",config: ServerConfig):
            super().__init__()
            self.sender = sender
            self.config = config

        @property
        def control(self) -> "ServerConfigWidget":
            return self.sender

    class DeleteRequested(Message):
        server: str

        def __init__(self, sender: "ServerConfigWidget",server: str):
            super().__init__()
            self.sender = sender
            self.server = server

        @property
        def control(self) -> "ServerConfigWidget":
            return self.sender

    def compose(self) -> ComposeResult:
        with Vertical(classes="server-config-container"):
            with Vertical(classes="config-inputs"):
                yield self.server_name_inp
                yield self.login_inp
                yield self.password_inp
            with Horizontal(classes="config-buttons"):
                yield Button(
                    label="Проверить", classes="test-button", variant="warning"
                )
                yield self.save_button
                yield Button(label="Удалить", classes="remove-button", variant="error")

    def __init__(self, config: ServerConfig | None = None, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.config = config if config else ServerConfig.create_empty()

        self.server_name_inp = Input(
            value=self.config.server, placeholder="Имя сервера"
        )
        self.login_inp = Input(value=self.config.login, placeholder="Имя пользователя")
        self.password_inp = SecretInput(value=self.config.password, placeholder="Пароль")
        self.save_button = Button(
            label="Сохранить",
            classes="save-button",
            variant="success",
            disabled=True,
        )

    @property
    def container(self) -> Widget:
        return self.query_one(".server-config-container")

    @property
    def server(self) -> str:
        return self.server_name_inp.value

    @property
    def login(self) -> str:
        return self.login_inp.value

    @property
    def password(self) -> str:
        return self.password_inp.value

    def set_config(self) -> None:
        self.config.server = self.server
        self.config.login = self.login
        self.config.password = self.password

    def get_config(self) -> ServerConfig:
        self.set_config()
        return self.config

    def set_validation_result(self, valid: bool):
        self.save_button.disabled = not valid

    @on(Button.Pressed, ".remove-button")
    def remove_widget(self) -> None:
        self.post_message(self.DeleteRequested(self, server=self.server))
        self.remove()

    @on(Button.Pressed, ".test-button")
    def validate_button(self, event: Button.Pressed) -> None:
        button = event.button
        button.disabled = True

        self.post_message(
            self.ValidateRequested(
                self, config=self.get_config(),
            )
        )

    @on(Button.Pressed, ".save-button")
    def save_config(self):
        self.post_message(self.SaveRequested(self, config=self.get_config()))

    @on(Input.Changed)
    def reset_validation(self):
        self.save_button.disabled = True
