from typing import List, Tuple
from textual import on
from textual.app import ComposeResult
from textual.containers import Container, Horizontal, ScrollableContainer
from textual.widgets import Button, Footer, Header, Input, Label, Select, Static
from textual_fspicker import FileOpen

from app.backend.exceptions import BackendError
from app.services.password_service import PasswordService
from app.services import MailboxFileParsingService
from app.tui.models import BaseScreen
from app.tui.widgets import ChangeMailboxWidget
from app.services.config_service import _create_config_service


class ChangePasswordScreen(BaseScreen):

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.empty_state = Static("No mailboxes added yet.\nPress 'Add' or 'From file' to start.", classes="empty-state")
        self.password_service = PasswordService()

        self.config_service = _create_config_service()
        self.servers = self.config_service.get_all()
        self.current_config = self.servers[0]

    def compose(self) -> ComposeResult:
        yield Container(
            Static("Изменение паролей", classes="screen-title"),
            Horizontal(
                Button("Добавить", variant="primary", id="add-btn"),
                Button("Из файла", variant="default", id="from-file-btn"),
                Button("Изменить все", variant="success", id="change-all-btn"),
                Select(
                    [(s.server, i) for  i, s in enumerate(self.servers)],
                    allow_blank=False,
                    id="server-select",
                ),
                Button("Назад", variant="error", id="back-btn"),
                id="mailbox-buttons-container",
            ),
            ScrollableContainer(
                id="mailbox-container"
            ),
            id="main-container"
        )

    def on_mount(self) -> None:
        self._add_empty_state()

    @on(ChangeMailboxWidget.DeleteMessage)
    def mailbox_deleted_handle(self, _: ChangeMailboxWidget.DeleteMessage) -> None:
        self._add_empty_state()

    @on(ChangeMailboxWidget.MailboxChangedMessage)
    def mailbox_change_handle(self, event: ChangeMailboxWidget.MailboxChangedMessage) -> None:
        password = self._change_mailbox_password(event.mailbox, event.password)
        event._sender.set_password(password) # type: ignore

    def _change_mailbox_password(self, mailbox: str, password: str) -> str:
        try:
            # Выставляем в input пароль, который вернул backend
            resp = self.password_service.change_password(self.current_config, mailbox, password)
            self.notify_success(message=f"Изменен пароль на {resp.mailbox}")
            return resp.password
        except BackendError as e:
            self.notify_backend_error(e)
        except Exception as e:
            self.notify_error(message=str(e))
        return ""

    @on(Button.Pressed, "#change-all-btn")
    def change_all_mailboxes(self, _: Button.Pressed):
        mailbox_data = []

        for child in self.mailbox_container.children:
            if isinstance(child, ChangeMailboxWidget):
                mailbox_data.append((child.mailbox, child.password))

        if len(mailbox_data) == 0:
            self.notify_warning(title="Нет ящиков", message="Сначала добавьте ящики для изменения")
            return

        for box in mailbox_data:
            self._change_mailbox_password(*box)


    @on(Button.Pressed, "#from-file-btn")
    async def pick_file(self, _: Button.Pressed) -> None:
        self.run_worker(self._pick_file_worker(), exclusive=True)

    async def _pick_file_worker(self) -> None:
        mailboxes_path = await self.app.push_screen_wait(FileOpen())
        if mailboxes_path is None:
            self.notify_warning(message="Файл не был выбран")
            return

        boxes = MailboxFileParsingService.parse(str(mailboxes_path))
        self._add_mailbox_many(boxes)

        self.notify_success(message=f"Добавлено {len(boxes)} ящиков")


    def _add_mailbox_many(self, mailboxes: List[Tuple[str, str]]) -> None:
        self.query(".empty-state").remove()
        for box, passwd in mailboxes:
            self._add_mailbox(box, passwd)

    @on(Button.Pressed, "#add-btn")
    def add_mailbox(self) -> None:
        self.query(".empty-state").remove()
        self._add_mailbox()

    def _add_mailbox(self, mailbox: str = "", password: str = "") -> None:
        mail_container = self.query_one("#mailbox-container")
        mail_container.mount(ChangeMailboxWidget(mailbox, password))

    @on(Button.Pressed, "#back-btn")
    def nav_back(self) -> None:
        self.dismiss()

    @on(Select.Changed, "#server-select")
    def set_current_config(self, event: Select.Changed) -> None:
        self.current_config = self.servers[event.value] # type: ignore

    @property
    def mailbox_container(self) -> ScrollableContainer:
        return self.query_one("#mailbox-container", ScrollableContainer)

    def _add_empty_state(self) -> None:
        container = self.mailbox_container
        if len(container.children) == 0 or len(container.children) == 1 and isinstance(container.children[0], ChangeMailboxWidget):
            container.mount(self.empty_state)

