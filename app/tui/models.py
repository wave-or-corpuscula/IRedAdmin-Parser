from typing import Protocol
from textual.screen import ModalScreen, Screen

from app.backend.exceptions import BackendError


TCSS_PATH = "../../styles.tcss"

class NotifyMixin:

    def notify_success(self, title: str = "Успех!", message: str = "") -> None:
        self.notify( # type: ignore
            title=title,
            message=message,
            severity="information"
        )

    def notify_warning(self, title: str = "Предупреждение!", message: str = "") -> None:
        self.notify( # type: ignore
            title=title,
            message=message,
            severity="warning"
        )

    def notify_error(self, title: str = "Ошибка!", message: str = "") -> None:
        self.notify( # type: ignore
            title=title,
            message=message,
            severity="error"
        )

    def notify_backend_error(self, error: BackendError) -> None:
        self.notify_error(
            title=error.type,
            message=f"[{error.code}] {error.message}",
        )


class BaseScreen(NotifyMixin, Screen):
    CSS_PATH = TCSS_PATH


class BaseModalScreen(NotifyMixin, ModalScreen):
    CSS_PATH = TCSS_PATH

