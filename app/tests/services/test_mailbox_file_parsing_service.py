import pytest
import tempfile
from app.services import MailboxFileParsingService



def test_mailbox_parsing_invalid():
    testing_data = """mailbox1@domain.com password1
    mailbox_no_password@domain.com
    invalid string three_arguments

    """
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True) as name_file:
        name_file.write(testing_data)
        name_file.flush()

        with pytest.raises(ValueError):
            MailboxFileParsingService.parse(name_file.name)


def test_mailbox_parsing_valid():
    testing_data = """mailbox1@domain.com password1
mailbox1@domain.com password1
mailbox_no_password@domain.com

mailbox_no_password@domain.com


"""
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True) as name_file:
        name_file.write(testing_data)
        name_file.flush()
        boxes = MailboxFileParsingService.parse(name_file.name)

