from textual import on
from textual.message import Message
from textual.containers import Horizontal, Container, Vertical, ScrollableContainer
from textual.widget import Widget
from textual.widgets import Button, Header, Footer, Label, Input, Static
from textual.app import ComposeResult

class ChangeMailboxWidget(Widget):
    CSS_PATH = "../../styles.tcss"

    class DeleteMessage(Message):
        def __init__(self) -> None:
            super().__init__()
    
    class MailboxChangedMessage(Message):
        mailbox : str
        password: str

        def __init__(self, mailbox: str, password: str) -> None:
            super().__init__()
            self.mailbox = mailbox
            self.password = password


    def __init__(self, mailbox: str = "", password: str = "", *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.mailbox_input = Input(placeholder="mailbox@domain.com", value=mailbox, classes="mailbox-input")
        self.password_input = Input(placeholder="Новый пароль", value=password, classes="password-input", password=True)


    def compose(self) -> ComposeResult:
        with Horizontal(classes="mailbox-row"):
            with Vertical(classes="mailbox-fields"):
                yield self.mailbox_input
                yield self.password_input
            with Vertical(classes="mailbox-actions"):
                yield Button("Change", variant="primary", classes="change-single-btn")
                yield Button("x", variant="error", classes="delete-btn")

    @property
    def password(self) -> str:
        return self.password_input.value

    @property
    def mailbox(self) -> str:
        return self.mailbox_input.value
    
    def set_password(self, password: str) -> None:
        self.password_input.value = password

    def set_mailbox(self, mailbox: str) -> None:
        self.mailbox_input.value = mailbox

    @on(Button.Pressed, ".delete-btn")
    def delete(self) -> None:
        self.post_message(self.DeleteMessage())
        self.remove()

    @on(Button.Pressed, ".change-single-btn")
    def change(self) -> None:
        self.post_message(self.MailboxChangedMessage(
            mailbox=self.mailbox,
            password=self.password,
        ))
        


