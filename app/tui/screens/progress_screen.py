from datetime import datetime
from typing import Callable

from textual import on
from textual.app import ComposeResult
from textual.containers import Center, Container, Horizontal
from textual.message import Message
from textual.widgets import Button, Label, ProgressBar, Static

from .. import BaseModalScreen


class ProgressScreen(BaseModalScreen):
    class ProgressUpdateMessage(Message):
        def __init__(self, mailbox: str, success: bool, elapsed: float):
            super().__init__()
            self.mailbox = mailbox
            self.success = success
            self.elapsed = elapsed

    class CancelChange(Message):
        def __init__(self) -> None:
            super().__init__()

    def __init__(self, total: int, on_cancel: Callable) -> None:
        super().__init__()
        self.total = total
        self.completed = 0
        self.success = 0
        self.failed = 0
        self.start_time = datetime.now()
        self.current_mailbox = ""
        self.on_cancel = on_cancel

    def compose(self) -> ComposeResult:
        with Container():
            with Center():
                yield Label("Изменение паролей", id="progress-title")
            with Center():
                yield ProgressBar(total=self.total, show_percentage=True, show_eta=False, id="progress-bar")
            yield Static(f"0 / {self.total} обработано", id="progress-status")
            yield Static("", id="progress-details")
            yield Static("", id="progress-time")

            with Horizontal():
                yield Button("Отмена", variant="error", id="cancel-btn")
                yield Button("Выход", variant="primary", id="exit-btn", disabled=True)

    def update_progress(self, mailbox: str, success: bool, elapsed: float) -> None:
        self.completed += 1
        if success:
            self.success += 1
        else:
            self.failed += 1

        bar = self.query_one("#progress-bar", ProgressBar)
        bar.advance()

        status = self.query_one("#progress-status", Static)
        status.update(f"{self.completed} / {self.total} обработано")

        details = self.query_one("#progress-details", Static)
        details.update(f"Успешно: {self.success}, Ошибок: {self.failed}\nТекущий: {mailbox}")

        time_widget = self.query_one("#progress-time", Static)
        elapsed_total = (datetime.now() - self.start_time).total_seconds()

        if self.completed == self.total:
            time_widget.update(f"Прошло: {self._format_time(elapsed_total)}")
            self.query_one("#exit-btn", Button).disabled = False
            self.query_one("#cancel-btn", Button).disabled = True
            return

        if self.completed > 0:
            remaining = (self.total - self.completed) * elapsed
            time_widget.update(f"Прошло: {self._format_time(elapsed_total)} | Осталось: ~{self._format_time(remaining)}")
        else:
            time_widget.update(f"Прошло: {self._format_time(elapsed_total)}")

    def _format_time(self, seconds: float) -> str:
        hours = int(seconds // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)
        if hours > 0:
            return f"{hours:02d}:{minutes:02d}:{secs:02d}"
        return f"{minutes:02d}:{secs:02d}"

    @on(ProgressUpdateMessage)
    def update_handle(self, event: ProgressUpdateMessage) -> None:
        self.update_progress(event.mailbox, event.success, event.elapsed)

    @on(Button.Pressed, "#exit-btn")
    def handle_exit(self, _: Button.Pressed) -> None:
        self.dismiss()

    @on(Button.Pressed, "#cancel-btn")
    def handle_cancel(self) -> None:
        self.on_cancel()
        self.dismiss()


