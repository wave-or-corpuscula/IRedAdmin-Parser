from textual import on
from textual.app import App, ComposeResult
from textual.widgets import Button
from textual.containers import Vertical

from app.utils import _create_config_service
from app.database.db import transaction_factory
from app.services.config_service import ConfigService
from app.storage.config_storage import ConfigStorage
from app.tui.screens.config.config_screen import ConfigScreen
from app.tui.screens.search.search_screen import SearchScreen


class MainMenuScreen(App):
    CSS_PATH = "../../styles.tcss"

    def compose(self) -> ComposeResult:
        with Vertical(id="main-menu-container"):
            yield Button(label="Поиск", id="search-button")
            yield Button(label="Инструменты", id="tools_button")
            yield Button(label="Конфигурация", id="config-button")
            yield Button(label="Выход", id="exit-button")

    @on(Button.Pressed, "#search-button")
    def nav_search_screen(self) -> None:
        search_screen = SearchScreen()
        self.push_screen(search_screen)

    @on(Button.Pressed, "#tools_button")
    def nav_tools_screen(self) -> None:
        self.notify(
            title="Comming soon",
            message="Инструменты пока недоступны!",
            severity="information",
            timeout=5,
        )

    @on(Button.Pressed, "#config-button")
    def nav_config_screen(self) -> None:
        conf_service = _create_config_service()
        conf_screen = ConfigScreen(conf_service)
        self.push_screen(conf_screen)

    @on(Button.Pressed, "#exit-button")
    def quit_app(self) -> None:
        self.exit()


if __name__ == "__main__":
    app = MainMenuScreen()
    app.run()
