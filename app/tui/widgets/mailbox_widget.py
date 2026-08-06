from textual import on
from textual.containers import Horizontal, Container, Vertical, ScrollableContainer
from textual.widget import Widget
from textual.widgets import Button, Header, Footer, Label, Input, Static
from textual.app import ComposeResult


class ChangeMailboxWidget(Widget):
    CSS_PATH = "../../styles.tcss"

    def __init__(self, mailbox: str = "", password: str = "", *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self._mailbox = mailbox
        self._password = password

    def compose(self) -> ComposeResult:
        with Horizontal(classes="mailbox-row"):
            with Vertical(classes="mailbox-fields"):
                yield Input(placeholder="mailbox@domain.com", value=self._mailbox, classes="mailbox-input")
                yield Input(placeholder="New password", value=self._password, classes="password-input", password=True)
            with Vertical(classes="mailbox-actions"):
                yield Button("Change", variant="primary", classes="change-single-btn")
                yield Button("x", variant="error", classes="delete-btn")

    @on(Button.Pressed, ".delete-btn")
    def delete(self) -> None:
        self.remove()


