from typing import Callable

from textual import on
from textual.app import ComposeResult
from textual.containers import Center, Container, Horizontal
from textual.widgets import Button, Input, Label, Static

from app.backend.exceptions import BackendError

from ..models import BaseModalScreen


class ChangeUserPasswordScreen(BaseModalScreen):

    def __init__(self, server: str, mailbox: str, change: Callable[[str], str]) -> None:
        super().__init__()
        self.server = server
        self.mailbox = mailbox
        self.change = change

    def compose(self) -> ComposeResult:
        with Container():
            with Center():
                yield Label("Смена пароля", id="change-password-title")
            yield Static(f"Сервер: {self.server}", id="server-info")
            yield Static(f"Почтовый ящик: {self.mailbox}", id="mailbox-info")
            yield Input(placeholder="Новый пароль (пусто: сгенерированный)", password=True, id="new-password-input")
            with Horizontal(id="change-password-buttons"):
                yield Button("Изменить", variant="success", id="change-password-confirm-btn")
                yield Button("Выйти", variant="default", id="change-password-exit-btn")

    @on(Button.Pressed, "#change-password-exit-btn")
    def nav_back(self, _: Button.Pressed) -> None:
        self.dismiss()

    @on(Button.Pressed, "#change-password-confirm-btn")
    def change_password(self, _: Button.Pressed) -> None:
        pass_input = self.query_one("#new-password-input", Input)

        password = pass_input.value
        try:
            new_password = self.change(password)
            pass_input.value = new_password
            self.notify_success(message="Пароль успешно изменен")
        except BackendError as e:
            self.notify_backend_error(e)
            return
        except Exception as e:
            self.notify_error(message=str(e))
            return

