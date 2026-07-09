from textual import on
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.message import Message
from textual.widget import Widget
from textual.widgets import Button, Input

from app.utils.config import ServerConfig


class ServerConfigWidget(Widget):
    class ValidateRequested(Message):
        config: ServerConfig

        def __init__(self, config: ServerConfig):
            super().__init__()
            self.config = config

    class SaveRequested(Message):
        config: ServerConfig

        def __init__(self, config: ServerConfig):
            super().__init__()
            self.config = config

    class DeleteRequested(Message):
        server: str

        def __init__(self, server: str):
            super().__init__()
            self.server = server

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
        self.password_inp = Input(value=self.config.password, placeholder="Пароль")
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
        self.post_message(self.DeleteRequested(server=self.server))
        self.remove()

    @on(Button.Pressed, ".test-button")
    def validate_button(self, event: Button.Pressed) -> None:
        button = event.button
        button.disabled = True

        self.post_message(
            self.ValidateRequested(
                config=self.get_config(),
            )
        )

    @on(Button.Pressed, ".save-button")
    def save_config(self):
        self.post_message(self.SaveRequested(config=self.get_config()))

    @on(Input.Changed)
    def reset_validation(self):
        self.save_button.disabled = True
