import pytest
from textual.app import App
from textual.widgets import Input

from app.tui.screens.search.search_screen import SearchScreen


class AppTest(App[None]):
    def on_mount(self) -> None:
        self.push_screen(SearchScreen())


@pytest.mark.asyncio
async def test_filtering_mailboxes():

    app = AppTest()

    async with app.run_test() as pilot:
        screen = app.screen
        assert isinstance(screen, SearchScreen)

        search_input = app.screen.query_one("#search-input")
        assert isinstance(search_input, Input)

        await pilot.click("#search-input")

        await pilot.press(*"user@example.com")

        assert search_input.value == "user@example.com"

        data = screen.update_table_data()
