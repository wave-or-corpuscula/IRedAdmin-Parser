from textual import on
from textual.containers import Horizontal, Container, Vertical, ScrollableContainer
from textual.widget import Widget
from textual.widgets import Button, Header, Footer, Label, Input, Static
from textual.app import ComposeResult
from textual.screen import Screen

from app.tui.widgets import ChangeMailboxWidget


class ChangePasswordScreen(Screen):
    CSS_PATH = "../../styles.tcss"

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.empty_state = Static("No mailboxes added yet.\nPress 'Add' or 'From file' to start.", classes="empty-state")

    def compose(self) -> ComposeResult:
        yield Container(
            Static("Change Passwords", classes="screen-title"),
            Horizontal(
                Button("Add", variant="primary", id="add-btn"),
                Button("From file", variant="default", id="file-btn"),
                Button("Change all", variant="success", id="change-all-btn"),
                Button("Back", variant="error", id="back-btn"),
                id="mailbox-buttons-container",
            ),
            ScrollableContainer(
                id="mailbox-container"
            ),
            id="main-container"
        )

    def on_mount(self) -> None:
        self._toggle_empty_state()

    @on(ChangeMailboxWidget.Deleted)
    def mailbox_deleted_handle(self, event: ChangeMailboxWidget.Deleted) -> None:
        self._toggle_empty_state()

    # def _add_empty_state(self) -> None:
    #     container = self.query_one("#mailbox-container", ScrollableContainer)
    #     container.remove_children()
    #     container.mount(self.empty_state)
    #
    # def _clear_empty_state(self) -> None:
    #     container = self.query_one("#mailbox-container", ScrollableContainer)
    #     if container.children and isinstance(container.children[0], Static) and container.children[0].has_class("empty-state"):
    #         container.remove_children()

    def _toggle_empty_state(self) -> None:
        container = self.query_one("#mailbox-container")
        if len(container.children) == 1 and container.children[0].has_class("empty-state"):
            container.query(".empty-state").remove()
        elif len(container.children) == 1:
            container.mount(self.empty_state)

        

    @on(Button.Pressed, "#back-btn")
    def nav_back(self) -> None:
        self.dismiss()

    @on(Button.Pressed, "#add-btn")
    def add_mailbox(self) -> None:
        # self.query(".empty-state").remove()

        self._toggle_empty_state()
        mail_container = self.query_one("#mailbox-container")
        mail_container.mount(ChangeMailboxWidget())


