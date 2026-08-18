from textual import on
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.message import Message
from textual.widget import Widget
from textual.widgets import Button, Input

from .secret_input import SecretInput

class ChangeMailboxWidget(Widget):
    CSS_PATH = "../../styles.tcss"

    class DeleteMessage(Message):
        def __init__(self) -> None:
            super().__init__()
    
    class MailboxChangedMessage(Message):
        mailbox : str
        password: str

        def __init__(self, sender: "ChangeMailboxWidget", mailbox: str, password: str) -> None:
            super().__init__()
            self.sender = sender
            self.mailbox = mailbox
            self.password = password

        @property
        def control(self) -> "ChangeMailboxWidget":
            return self.sender


    def __init__(self, mailbox: str = "", password: str = "", *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.mailbox_input = Input(placeholder="mailbox@domain.com", value=mailbox, classes="mailbox-input")
        self.password_input = SecretInput(placeholder="Новый пароль", value=password, classes="password-input")

        self.change_button = Button("Изменить", variant="primary", classes="change-single-btn", id="change-single-btn")
        self.delete_button = Button("x", variant="error", classes="delete-btn", id="delete-btn")



    def compose(self) -> ComposeResult:
        with Horizontal(classes="mailbox-row"):
            with Vertical(classes="mailbox-fields"):
                yield self.mailbox_input
                yield self.password_input
            with Vertical(classes="mailbox-actions"):
                yield self.change_button
                yield self.delete_button

    def set_disable(self) -> None:
        self._set_locked(True)

    def set_enable(self) -> None:
        self._set_locked(False)

    def _set_locked(self, locked: bool) -> None:
        self.mailbox_input.disabled = locked
        self.password_input.disabled = locked
        self.change_button.disabled = locked
        self.delete_button.disabled = locked

    def enable_success(self) -> None:
        row = self.query_one(".mailbox-row")
        row.remove_class("mailbox-error")
        row.add_class("mailbox-success")
        self.set_enable()

    def enable_error(self) -> None:
        row = self.query_one(".mailbox-row")
        row.remove_class("mailbox-success")
        row.add_class("mailbox-error")
        self.set_enable()

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
            self,
            mailbox=self.mailbox,
            password=self.password,
        ))
        


