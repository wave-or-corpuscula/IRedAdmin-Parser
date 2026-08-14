from textual import on
from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Button

from app.services.config_service import _create_config_service

from .. import BaseScreen
from .change_password_screen import ChangePasswordScreen
from .config_screen import ConfigScreen
from .search_screen import SearchScreen


class MainMenuScreen(BaseScreen):
    
    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)

    def compose(self) -> ComposeResult:
        with Vertical(id="main-menu-container"):
            yield Button(label="Поиск", id="search-button")
            yield Button(label="Инструменты", id="tools_button")
            yield Button(label="Конфигурация", id="config-button")
            yield Button(label="Выход", id="exit-button")

    @on(Button.Pressed, "#search-button")
    def nav_search_screen(self) -> None:
        search_screen = SearchScreen()
        self.app.push_screen(search_screen)

    @on(Button.Pressed, "#tools_button")
    def nav_tools_screen(self) -> None:
        passwords_change_screen = ChangePasswordScreen()
        self.app.push_screen(passwords_change_screen)

    @on(Button.Pressed, "#config-button")
    def nav_config_screen(self) -> None:
        conf_service = _create_config_service()
        conf_screen = ConfigScreen(conf_service)
        self.app.push_screen(conf_screen)

    @on(Button.Pressed, "#exit-button")
    def quit_app(self) -> None:
        self.app.exit()
