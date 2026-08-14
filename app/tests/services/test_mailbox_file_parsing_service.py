import tempfile

import pytest

from app.services import MailboxFileParsingService



def test_parse_valid_file_with_passwords():
    """Парсинг файла с валидными ящиками и паролями"""
    testing_data = """user1@domain.com\tpassword123
user2@domain.com\tSecurePass!@#
user3@domain.com\tqD97A6GqT&
"""
    expected = [
        ("user1@domain.com", "password123"),
        ("user2@domain.com", "SecurePass!@#"),
        ("user3@domain.com", "qD97A6GqT&"),
    ]
    
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        result = MailboxFileParsingService.parse(tmp_file.name)
        
        assert result == expected
        assert len(result) == 3

def test_parse_valid_file_without_passwords():
    """Парсинг файла с ящиками без паролей"""
    testing_data = """user1@domain.com\t
user2@domain.com\t
user3@domain.com\t
"""
    expected = [
        ("user1@domain.com", ""),
        ("user2@domain.com", ""),
        ("user3@domain.com", ""),
    ]
    
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        result = MailboxFileParsingService.parse(tmp_file.name)
        
        assert result == expected

def test_parse_valid_mixed_file():
    """Парсинг файла с ящиками: некоторые с паролями, некоторые без"""
    testing_data = """user1@domain.com\tpassword123
user2@domain.com\t
user3@domain.com\tSecurePass
user4@domain.com\t
"""
    expected = [
        ("user1@domain.com", "password123"),
        ("user2@domain.com", ""),
        ("user3@domain.com", "SecurePass"),
        ("user4@domain.com", ""),
    ]
    
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        result = MailboxFileParsingService.parse(tmp_file.name)
        
        assert result == expected
        assert len(result) == 4

def test_parse_file_with_empty_lines():
    """Парсинг файла с пустыми строками - они должны игнорироваться"""
    testing_data = """
user1@domain.com\tpassword123

user2@domain.com\tpassword456

user3@domain.com\t

"""
    expected = [
        ("user1@domain.com", "password123"),
        ("user2@domain.com", "password456"),
        ("user3@domain.com", ""),
    ]
    
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        result = MailboxFileParsingService.parse(tmp_file.name)
        
        assert result == expected
        assert len(result) == 3

def test_parse_file_with_comments():
    """Парсинг файла со строками-комментариями"""
    testing_data = """# This is a comment
user1@domain.com\tpassword123
# Another comment
user2@domain.com\tpassword456
user3@domain.com\t
"""
    expected = [
        ("user1@domain.com", "password123"),
        ("user2@domain.com", "password456"),
        ("user3@domain.com", ""),
    ]
    
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        result = MailboxFileParsingService.parse(tmp_file.name)
        
        assert result == expected

def test_parse_file_with_trailing_whitespace():
    """Парсинг файла с пробелами в начале/конце строк"""
    testing_data = """  user1@domain.com\tpassword123  
user2@domain.com\tpassword456  
user3@domain.com\t  
"""
    expected = [
        ("user1@domain.com", "password123"),
        ("user2@domain.com", "password456"),
        ("user3@domain.com", ""),
    ]
    
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        result = MailboxFileParsingService.parse(tmp_file.name)
        
        assert result == expected

def test_parse_file_with_empty_mailbox_raises_error():
    """Ошибка при пустом ящике в строке"""
    testing_data = """user1@domain.com\tpassword123
\tpassword456
user3@domain.com\t
"""
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        with pytest.raises(ValueError) as exc_info:
            MailboxFileParsingService.parse(tmp_file.name)
        
        assert "line 2" in str(exc_info.value).lower()

def test_parse_file_with_invalid_email_raises_error():
    """Ошибка при невалидном email (без @)"""
    testing_data = """user1@domain.com\tpassword123
invalid_email\tpassword456
user3@domain.com\t
"""
    with tempfile.NamedTemporaryFile(mode='w+t', delete=True, encoding='utf-8') as tmp_file:
        tmp_file.write(testing_data)
        tmp_file.flush()
        
        with pytest.raises(ValueError) as exc_info:
            MailboxFileParsingService.parse(tmp_file.name)
        
        assert "invalid mailbox" in str(exc_info.value).lower()
        assert "line 2" in str(exc_info.value).lower()
