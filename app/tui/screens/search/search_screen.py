from dataclasses import asdict, dataclass, replace
from functools import wraps
from typing import Optional

from textual import on
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.reactive import reactive
from textual.screen import Screen
from textual.widgets import (
    Button,
    DataTable,
    Header,
    Input,
    Label,
    Select,
    Collapsible,
)

from app.database.db import transaction
from app.database.repositories import ServerRepository
from app.database.repositories.mailbox_repository import MailboxRepository
from app.tui.screens.sync.sync_screen import SyncScreen

COLUMNS = [
    "ID",
    "Сервер",
    "Домен",
    "Адрес",
    "Отображаемое имя",
    "Активен",
    "Админ",
    "Квота",
    "Использовано",
    "Занято %",
]

SORT_KEYS_BY_INDEX = [
    "id",
    "server",
    "domain",
    "address",
    "display_name",
    "disabled",
    "is_admin",
    "quota",
    "used_memory",
    "usage_percent",
]

class FilterError(Exception):
    def __init__(self, title: str = "", message: str = "") -> None:
        self.title = title
        self.message = message


@dataclass
class DataFilter:
    search: str = ""
    is_admin: Optional[bool] = None
    disabled: Optional[bool] = None
    server_id: Optional[int] = None
    quota_min: str = ""
    quota_max: str = ""

    def copy_with(self, **kwargs):
        return replace(self, **kwargs)


def update_filters(func):
    @wraps(func)
    def wrapper(self: "SearchScreen", event, *args, **kwargs):
        filters_copy = self.filters.copy_with()
        func(self, event, filters_copy, *args, **kwargs)
        self.filters = filters_copy

    return wrapper


def validate_quota(func):
    @wraps(func)
    def wrapper(self: "SearchScreen", event: Input.Changed, *args, **kwargs):
        value = event.value
        if len(value) == 0 or (value.isdigit() and int(value) >= 0):
            func(self, event, *args, **kwargs)
        else:
            self.notify(
                title="Неверная квота",
                message="Квота должна быть положительным числом",
                severity="error",
            )
            event._sender.clear()  # type: ignore
            return

    return wrapper


class SearchScreen(Screen):
    CSS_PATH = "../../styles.tcss"

    filters = reactive(DataFilter())

    sort_column = reactive("address")
    sort_desc = reactive(False)

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)

        with transaction() as conn:
            repo = ServerRepository(conn)
            self.servers = [("Все", None)] + repo.get_tuples()
        
        self.selected_rows_lb = Label(id="selected-rows-count")

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)

        with Vertical(classes="search-screen"):
            with Horizontal(classes="top-bar"):
                yield Input(placeholder="Поиск (Адрес / Имя):", id="search-input")

            # Collapsing panel (filters)
            with Collapsible(title="Фильтры"):
                if len(self.servers) > 2:
                    yield Label("Сервер:", classes="filter-label")
                    yield Select(
                        self.servers,
                        allow_blank=False,
                        id="server_select",
                        classes="filter-select",
                    )

                yield Label("Администратор:", classes="filter-label")
                yield Select(
                    [("Все", None), ("Да", 1), ("Нет", 0)],
                    allow_blank=False,
                    id="admin_select",
                    classes="filter-select",
                )

                yield Label("Отключен:", classes="filter-label")
                yield Select(
                    [("Все", None), ("Да", 1), ("Нет", 0)],
                    allow_blank=False,
                    id="disabled_select",
                    classes="filter-select",
                )

                yield Label("Квота (байт) от - до:", classes="filter-label")
                with Horizontal(classes="filter-group"):
                    yield Input(
                        placeholder="Мин", id="quota_min", classes="input-range"
                    )
                    yield Input(
                        placeholder="Макс",
                        id="quota_max",
                        classes="input-range",
                    )

            # Data table
            yield DataTable(zebra_stripes=True, cursor_type="row")

            # Footer
            with Horizontal(classes="footer-block"):
                with Horizontal(classes="left"):
                    yield self.selected_rows_lb
                with Horizontal(classes="right"):
                    yield Button(label="Синхронизовать", id="sync-mailboxes")
                    yield Button(label="Назад", id="search-nav-back")

    def on_mount(self) -> None:
        table: DataTable = self.query_one(DataTable)
        table.add_columns(*COLUMNS)
        self.update_table_data()

    @on(DataTable.RowSelected)
    def handle_select_row(self, event: DataTable.RowSelected) -> None:
        row_data = event.data_table.get_row(event.row_key)
        self.notify(title="RowSelected", message=str(row_data))

    @on(DataTable.HeaderSelected)
    def handle_highlight_column(self, event: DataTable.HeaderSelected) -> None:
        col_index = event.column_index
        clicked_column = SORT_KEYS_BY_INDEX[col_index]
        if self.sort_column == clicked_column:
            self.sort_desc = not self.sort_desc
        else:
            self.sort_desc = False
            self.sort_column = clicked_column
        self.update_table_data()

    def update_rows_count_lb(self, count: int) -> None:
        self.selected_rows_lb.update(f"Выбрано строк: {count}")

    def update_table_data(self) -> None:
        table = self.query_one(DataTable)
        table.clear()

        with transaction() as conn:
            repo = MailboxRepository(conn)
            models = repo.get_models(
                asdict(self.filters),
                self.sort_column,
                self.sort_desc,
            )

        for row in models:
            table.add_row(*row.to_row())
        self.update_rows_count_lb(len(models))

    def watch_filters(self, old, new) -> None:
        self.update_table_data()

    def watch_sort_column(self, old, new) -> None:
        self.update_table_data()

    def watch_sort_desc(self, old, new) -> None:
        self.update_table_data()

    @on(Button.Pressed, "#sync-mailboxes")
    def show_sync_screen(self) -> None:
        screen = SyncScreen()
        self.app.push_screen(screen, callback=lambda _: self.update_table_data())

    @on(Select.Changed, "#admin_select")
    @update_filters
    def admin_filter(self, event: Select.Changed, filters: DataFilter) -> None:
        filters.is_admin = event.value  # type: ignore

    @on(Select.Changed, "#disabled_select")
    @update_filters
    def disabled_filter(self, event: Select.Changed, filters: DataFilter) -> None:
        filters.disabled = event.value  # type: ignore

    @on(Select.Changed, "#server_select")
    @update_filters
    def server_filter(self, event: Select.Changed, filters: DataFilter) -> None:
        filters.server_id = event.value  # type: ignore

    @on(Input.Changed, "#quota_max")
    @validate_quota
    @update_filters
    def quota_max_filter(self, event: Input.Changed, filters: DataFilter) -> None:
        filters.quota_max = event.value

    @on(Input.Changed, "#quota_min")
    @validate_quota
    @update_filters
    def quota_min_filter(self, event: Input.Changed, filters: DataFilter) -> None:
        filters.quota_min = event.value

    @on(Input.Changed, "#search-input")
    @update_filters
    def search_changed(self, event: Input.Changed, filters: DataFilter) -> None:
        filters.search = event.value

    @on(Button.Pressed, "#search-nav-back")
    def nav_back(self):
        self.dismiss()
