import asyncio
from datetime import datetime
from typing import Callable, Coroutine, List, Tuple
from textual import on
from textual.app import ComposeResult
from textual.containers import Container, Grid, Horizontal, ScrollableContainer
from textual.message import Message
from textual.widgets import Button, Footer, Header, Input, Label, Select, Static
from textual_fspicker import FileOpen

from app.backend.exceptions import BackendError
from app.services.password_service import PasswordService
from app.services import MailboxFileParsingService
from app.tui.models import BaseScreen
from app.tui.screens.progress.progress_screen import ProgressScreen
from app.tui.widgets import ChangeMailboxWidget
from app.services.config_service import _create_config_service


class ChangePasswordScreen(BaseScreen):

    TITLE = "Изменение паролей"

    progress_screen: ProgressScreen

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.empty_state = Static("Ящиков еще не добавлено.\nНажмите 'Добавить' или 'Из файла' чтобы начать.", classes="empty-state")
        self.password_service = PasswordService()

        self.config_service = _create_config_service()
        self.servers = self.config_service.get_all()
        self.current_config = self.servers[0]

    def compose(self) -> ComposeResult:
        yield Container(
            Header(show_clock=True),
            Grid(
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
    async def mailbox_change_handle(self, event: ChangeMailboxWidget.MailboxChangedMessage) -> None:
        self.run_worker(
            self._change_password_worker(event.control),
            exclusive=True,
            exit_on_error=False,
        )

    @on(Button.Pressed, "#add-btn")
    def add_mailbox(self) -> None:
        self.query(".empty-state").remove()
        self._add_mailbox()

    @on(Button.Pressed, "#from-file-btn")
    async def pick_file(self, _: Button.Pressed) -> None:
        self.run_worker(self._pick_file_worker(), exclusive=True)

    @on(Button.Pressed, "#back-btn")
    def nav_back(self) -> None:
        self.dismiss()

    @on(Select.Changed, "#server-select")
    def set_current_config(self, event: Select.Changed) -> None:
        self.current_config = self.servers[event.value] # type: ignore

    @on(Button.Pressed, "#change-all-btn")
    async def change_all_mailboxes(self, _: Button.Pressed):
        total = len(self.mailbox_container.children)
        if total == 1 and isinstance(self.mailbox_container.children[0], Static):
            self.notify_warning(message="Не добавлено ящиков для изменения")
            return

        self.cancel = False

        self.progress_screen = ProgressScreen(total=total, on_cancel=lambda: setattr(self, "cancel", True))
        await self.app.push_screen(self.progress_screen)

        self.run_worker(
            self._change_all_passwords_worker(),
            exclusive=True
        )

    @on(ProgressScreen.CancelChange)
    def handle_cancel(self) -> None:
        self.cancel = True


    def _lock_ui(self) -> None:
        self.query_one("#change-all-btn").disabled = True
        for widget in self.mailbox_container.children:
            if isinstance(widget, ChangeMailboxWidget):
                widget.set_disable()

    def _unlock_ui(self) -> None:
        self.query_one("#change-all-btn").disabled = False
        for widget in self.mailbox_container.children:
            if isinstance(widget, ChangeMailboxWidget):
                widget.set_enable()

    async def _change_all_passwords_worker(self) -> None:

        if self.progress_screen is None:
            return

        def worker_callback(mailbox, success, elapsed):
            self.progress_screen.post_message(
                ProgressScreen.ProgressUpdateMessage(
                    mailbox,
                    success,
                    elapsed,
                )
            )

        for widget in self.mailbox_container.children:
            if isinstance(widget, ChangeMailboxWidget):
                if self.cancel:
                    self.notify_warning(message="Изменение остановлено")
                    return
                await self._change_password_worker_with_callback(widget, worker_callback)

    async def _change_password(self, widget: ChangeMailboxWidget) -> str:
        loop = asyncio.get_event_loop()
        resp = await loop.run_in_executor(
            None,
            self.password_service.change_password,
            self.current_config,
            widget.mailbox,
            widget.password,
        )
        return resp.password

    async def _change_password_worker(self, widget: ChangeMailboxWidget) -> None:
        self._lock_ui()
        try:
            mailbox, password = await self._change_password(widget)

            widget.set_password(password)
            widget.enable_success()
            self.notify_success(message=f"Изменен пароль на {mailbox}")
        except BackendError as e:
            self.notify_backend_error(e)
            widget.enable_error()
        except Exception as e:
            self.notify_error(message=str(e))
            widget.enable_error()
        finally:
            self._unlock_ui()

    async def _change_password_worker_with_callback(self, widget: ChangeMailboxWidget, callback: Callable) -> None:
        start = datetime.now()
        success = False
        mailbox = widget.mailbox

        try:
            password = await self._change_password(widget)
            success = True

            widget.set_password(password)
            widget.enable_success()
        except Exception:
            widget.enable_error()
        finally:
            elapsed = (datetime.now() - start).total_seconds()
            callback(mailbox, success, elapsed)

    def _add_mailbox(self, mailbox: str = "", password: str = "") -> None:
        mail_container = self.query_one("#mailbox-container")
        mail_container.mount(ChangeMailboxWidget(mailbox, password))

    async def _pick_file_worker(self) -> None:
        mailboxes_path = await self.app.push_screen_wait(FileOpen())
        if mailboxes_path is None:
            self.notify_warning(message="Файл не был выбран")
            return

        try: 
            boxes = MailboxFileParsingService.parse(str(mailboxes_path))

            self._add_mailbox_many(boxes)
            self.notify_success(message=f"Добавлено {len(boxes)} ящиков")
        except Exception as e:
            self.notify_error(message=str(e))


    def _add_mailbox_many(self, mailboxes: List[Tuple[str, str]]) -> None:
        self.query(".empty-state").remove()
        for box, passwd in mailboxes:
            self._add_mailbox(box, passwd)


    @property
    def mailbox_container(self) -> ScrollableContainer:
        return self.query_one("#mailbox-container", ScrollableContainer)

    def _add_empty_state(self) -> None:
        container = self.mailbox_container
        if len(container.children) == 0 or len(container.children) == 1 and isinstance(container.children[0], ChangeMailboxWidget):
            container.mount(self.empty_state)

