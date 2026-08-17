from textual import on
from textual.app import ComposeResult
from textual.containers import Horizontal
from textual.message import Message
from textual.widget import Widget
from textual.widgets import Button, Input


class SecretInput(Horizontal):

    DEFAULT_CSS = """
    SecretInput {
        height: 3;
        margin-bottom: 2;
    }

    SecretInput > Input {
        width: 1fr;
        margin: 0;
    }

    SecretInput > #reveal-btn {
        width: 5;
        height: 100%;
        margin: 0;
        padding: 0 1;
    }

    SecretInput > #reveal-btn {
        width: 4;
        height: 3;
        margin: 0;
        padding: 0;
        min-width: 3;
        min-height: 3;
    }
    """

    class Changed(Message):
        def __init__(self, value: str) -> None:
            super().__init__()
            self.value = value

    class Submitted(Message):
        def __init__(self, value: str) -> None:
            super().__init__()
            self.value = value

    def __init__(
        self,
        placeholder: str = "Введите пароль",
        password: bool = True,
        name: str | None = None,
        id: str | None = None,
        classes: str | None = None,
    ) -> None:
        super().__init__(name=name, id=id, classes=classes)
        self._password_visible = False
        self._placeholder = placeholder
        self._password = password

    def compose(self) -> ComposeResult:
        yield Input(
            placeholder=self._placeholder,
            password=self._password,
            id="secret-input-field"
        )
        yield Button("👁️", id="reveal-btn", variant="default")

    @on(Button.Pressed, "#reveal-btn")
    def toggle_visibility(self, event: Button.Pressed) -> None:
        input_field = self.query_one("#secret-input-field", Input)
        self._password_visible = not self._password_visible
        input_field.password = not self._password_visible
        
        if self._password_visible:
            event.button.label = "🙈"
        else:
            event.button.label = "👁️"

    @on(Input.Changed, "#secret-input-field")
    def on_input_changed(self, event: Input.Changed) -> None:
        self.post_message(self.Changed(event.value))

    @on(Input.Submitted, "#secret-input-field")
    def on_input_submitted(self, event: Input.Submitted) -> None:
        self.post_message(self.Submitted(event.value))

    @property
    def value(self) -> str:
        return self.query_one("#secret-input-field", Input).value

    @value.setter
    def value(self, new_value: str) -> None:
        self.query_one("#secret-input-field", Input).value = new_value

    def focus(self, scroll_visible: bool = True) -> Widget:
         return self.query_one("#secret-input-field", Input).focus(scroll_visible=scroll_visible)

    def clear(self) -> None:
        self.query_one("#secret-input-field", Input).value = ""

    def is_password_visible(self) -> bool:
        return self._password_visible


