from textual import on
from textual.app import ComposeResult
from textual.widget import Widget
from textual.widgets import Input, Button
from textual.containers import Vertical, Horizontal

from app.utils import ServerConfig


class ServerConfigBlock(Widget):
    def __init__(self, config: ServerConfig | None = None, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.saved = config is not None
        self.config = config if config else ServerConfig.create_empty()

        self.server_name_inp = Input(
            value=self.config.server, placeholder="Имя сервера"
        )
        self.username_inp = Input(
            value=self.config.login, placeholder="Имя пользователя"
        )
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
    def username(self) -> str:
        return self.username_inp.value

    @property
    def password(self) -> str:
        return self.password_inp.value

    def set_config(self) -> None:
        self.config.server = self.server
        self.config.login = self.username
        self.config.password = self.password

    def get_config(self) -> ServerConfig:
        self.set_config()
        return self.config

    def compose(self) -> ComposeResult:
        with Vertical(classes="server-config-container"):
            with Vertical(classes="config-inputs"):
                yield self.server_name_inp
                yield self.username_inp
                yield self.password_inp
            with Horizontal(classes="config-buttons"):
                yield Button(
                    label="Проверить", classes="test-button", variant="warning"
                )
                yield self.save_button
                yield Button(label="Удалить", classes="remove-button", variant="error")

    # def validate_config(self) -> None:
    #     is_valid = self._validate_config(self.server, self.username, self.password)
    #     self.save_button.disabled = not is_valid
    #
    #     self.set_validation_state(is_valid)
    #
    #     if is_valid:
    #         self.notify("Успех!", severity="information")
    #     else:
    #         self.notify("Поражение!", severity="error")

    def set_validation_state(self, is_valid: bool) -> None:
        self.container.remove_class("validation-error", "validation-success")

        if is_valid:
            self.container.add_class("validation-success")
        else:
            self.container.add_class("validation-error")

    # @staticmethod
    # def _validate_config(server: str, username: str, password: str) -> bool:
    #     return server and username and password

    def reset_validation(self) -> None:
        self.container.remove_class("validation-error", "validation-success")
        self.save_button.disabled = True

    @on(Input.Changed)
    def change_input_data(self) -> None:
        self.reset_validation()

    # @on(Button.Pressed, ".save-button")
    # def save_config(self) -> None:
    #     config = self.get_config()
    #
    #     try:
    #         cfg_service.save_config(config)
    #     except InvalidConfigDataError:
    #         self.notify(
    #             message=f"Проверьте валиность файла {cfg_service.config_path}",
    #             title="Неверный формат конфигураций!",
    #             severity="error",
    #         )
    #         return
    #     self.notify(
    #         message=f"Конфигурация для {self.server} сохранена!",
    #         title="Успех",
    #         severity="information",
    #     )

    # @on(Button.Pressed, ".test-button")
    # def test_config(self) -> None:
    #     self.validate_config()

    # @on(Button.Pressed, ".remove-button")
    # def remove_block(self) -> None:
    #     config = self.get_config()
    #     cfg_service.delete_config(config)
    #     self.remove()
