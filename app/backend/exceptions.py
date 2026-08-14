from .models import CLIError


class BackendError(Exception):
    def __init__(self, error: CLIError):
        self.message = error.message
        self.type = error.type
        self.code = error.code

        super().__init__()


class AuthError(BackendError):
    pass
