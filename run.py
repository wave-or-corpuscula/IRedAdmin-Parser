from textual.app import App
from app import MainMenuScreen


class MyApp(App):
    CSS_PATH = "./app/tui/styles.tcss"

    def on_mount(self) -> None:
        self.push_screen(MainMenuScreen())

if __name__ == "__main__":
    MyApp().run()
