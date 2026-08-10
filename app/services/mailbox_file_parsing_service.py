from pathlib import Path

from typing import List, Tuple


class MailboxFileParsingService:
    DELIMITER = "\t"

    @classmethod
    def parse(cls, file_path: str) -> List[Tuple[str, str]]:

        path = Path(file_path)

        if not path.exists():
            raise FileNotFoundError(f"File not found: {file_path}")

        mailboxes = []
        line_number = 0

        with open(path, mode="r", encoding="utf-8") as file:
            for line in file:

                line_number += 1
                line = line.strip()

                if not line or line.startswith("#"):
                    continue

                parts = None
                if cls.DELIMITER in line:
                    parts = line.split(cls.DELIMITER, 1)

                if not parts:
                    mailbox = line.strip()
                    
                    if cls._validate_email(mailbox):
                        mailboxes.append(mailbox)
                    else:
                        raise ValueError(
                            f"Line {line_number}: invalid mailbox format: '{line}'"
                        )
                    continue

                mailbox = parts[0].strip()
                password = parts[1].strip() if len(parts) > 1 else ""

                if not mailbox:
                    raise ValueError(
                        f"Line {line_number}: empty mailbox: '{line}'"
                    )

                if not cls._validate_email(mailbox):
                    raise ValueError(
                        f"Line {line_number}: invalid mailbox format: '{line}'"
                    )
                mailboxes.append((mailbox, password))
            if not mailboxes:
                raise ValueError("No valid mailboxes found in file")
        return mailboxes


    @staticmethod
    def _validate_email(email: str) -> bool:
        if not email or "@" not in email:
            return False
        if email.startswith("@") or email.endswith("@"):
            return False
        return True
