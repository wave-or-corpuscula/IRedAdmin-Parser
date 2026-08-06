from textual import on
from textual.containers import Horizontal, Container, Vertical, ScrollableContainer
from textual.widgets import Button, Header, Footer, Label, Input, Static
from textual.app import ComposeResult
from textual.screen import Screen

from app.tui.widgets import ChangeMailboxWidget


class ChangePasswordScreen(Screen):
    CSS_PATH = "../../styles.tcss"

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)

    def compose(self) -> ComposeResult:
        yield Container(
            Static("Change Passwords", classes="screen-title"),
            Horizontal(
                Button("Add", variant="primary", id="add-btn"),
                Button("From file", variant="default", id="file-btn"),
                Button("Change all", variant="success", id="change-all-btn"),
                Button("Back", variant="error", id="back-btn"),
            ),
            ScrollableContainer(
                id="mailbox-container"
            ),
            id="main-container"
        )

    def on_mount(self) -> None:
        """При монтировании добавляем виджет с подсказкой"""
        self._add_empty_state()

    def _add_empty_state(self) -> None:
        """Показывает сообщение, если нет ящиков"""
        container = self.query_one("#mailbox-container", ScrollableContainer)
        container.remove_children()
        container.mount(Static("No mailboxes added yet.\nPress 'Add' or 'From file' to start.", classes="empty-state"))

    @on(Button.Pressed, "#back-btn")
    def nav_back(self) -> None:
        self.dismiss()

    @on(Button.Pressed, "#add-btn")
    def add_mailbox(self) -> None:
        mail_container = self.query_one("#mailbox-container")
        mail_container.mount(ChangeMailboxWidget())


