# IRedAdmin Parser

Парсинг, синхронизация и просмотр данных iRedAdmin — набор инструментов для управления iRedMail-инфраструктурой.

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](./iredparser/)
[![Python Version](https://img.shields.io/badge/Python-3.13+-3776AB?logo=python)](.)
[![SQLite](https://img.shields.io/badge/SQLite-modernc_/_stdlib-003B57?logo=sqlite)](./iredparser/internal/database/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Обзор

iRedAdmin Parser подключается к iRedAdmin — веб-панели управления iRedMail-серверами — собирает метаданные доменов и почтовых ящиков, сохраняет в локальную SQLite. Два интерфейса, общая база:

- **[`iredparser/`](iredparser/) — Go CLI**: Конкурентный движок парсинга. Авторизуется в iRedAdmin, парсит HTML в структурированные данные, upsert'ит в SQLite.
- **[`app/`](app/) — Python TUI**: Терминальный браузер на Textual. Поиск, фильтрация, сортировка данных; запуск синхронизации и смена паролей.

---

## Go CLI (`iredparser/`)

Основной движок парсинга на Go.

### Архитектура

```
iredparser/
  cmd/parser-cli/   # Точка входа
  common/           # Общие типы (ServerConfig)
  internal/
    controller/     # Cobra-команды, корень DI
    database/       # SQLite (sqlx + modernc.org/sqlite)
    parser/
      client/       # HTTP-клиент с cookie jar
      domain/       # Парсинг списка доменов
      mailbox/      # Конкурентный парсинг ящиков (worker pool)
    services/
      auth_service/       # Аутентификация в iRedAdmin
      password_service/   # Смена пароля удалённо
    sync/           # Оркестрация: парсинг → сохранение
  pkg/
    errors/         # Типизированная иерархия sentinel-ошибок
    utils/          # Парсинг размера памяти, извлечение CSRF
  testing/          # Вспомогательные утилиты для тестов
```

### Технологии и подходы

**HTTP-клиент** (`internal/parser/client/`)

- `http.Client` + `*cookiejar.Jar` — куки сессии сохраняются между запросами
- `TLSClientConfig{InsecureSkipVerify: true}` — self-signed сертификаты типичны для mail-серверов
- Пул соединений: 100 idle, 10 на хост, 90s таймаут
- Маскировка User-Agent (Chrome 149) — обход наивной антибот-защиты
- Каждый HTTP-сбой маппится на типизированную sentinel-ошибку (`ErrPostRequestCreation`, `ErrUnexpectedStatusCode`, `ErrInternalServerError`)
- Извлечение сессионной куки из cookie jar после успешной аутентификации

**Конкурентный парсинг HTML** (`internal/parser/mailbox/`)

- Определение числа страниц по `.pages`-навигации, конкурентный обход всех страниц
- Worker pool: 30 горутин читают из `chan string` (URL-задачи), пишут в `chan []*parser.Mailbox`
- `sync.WaitGroup` через `wg.Go()` (Go 1.26) — координированное завершение
- `goquery` — jQuery-стиль обхода DOM для извлечения данных из `<tbody> <tr>`
- Устойчивость к ошибкам: сбой парсинга одной страницы не убивает весь процесс

**База данных** (`internal/database/`)

- `jmoiron/sqlx` + `modernc.org/sqlite` — pure Go, без CGO
- `ON CONFLICT ... DO UPDATE ... RETURNING` — атомарный upsert с возвратом ID за один запрос, без race condition
- Транзакционные batch-операции (`Beginx`/`Commit`/`Rollback`) для `UpsertDomainMany` / `UpsertMailboxMany` — откат целиком при любой ошибке
- Встраивание через композицию структур: `DomainModel` содержит `parser.Domain`, `MailboxModel` содержит `parser.Mailbox` — без ручного маппинга
- Автоинициализация схемы при подключении (`CREATE TABLE IF NOT EXISTS`) + foreign keys

**CLI-фреймворк** (`internal/controller/`)

- `spf13/cobra` с подкомандами: `auth-check`, `sync`, `change-password`
- Middleware через `PersistentPreRunE`: парсинг JSON-конфига → аутентификация → внедрение конфига в состояние контроллера — выполняется перед каждой командой
- Dependency Injection в `NewCLIController(client, storage, authService, syncService, passwordService, writer)` — без глобального состояния, тестируемо изолированно
- Интерфейс сегрегирован: `AuthChecker`, `SyncService`, `PasswordChanger`, `Storage` — каждый интерфейс = один метод, одна ответственность

**Обработка ошибок** (`pkg/errors/`)

- Кастомный тип `IRedError` с `ErrType` (authentication, HTTP, parsing, CLI) и `ErrCode` (числовой, сгруппирован по домену)
- Sentinel-ошибки — `errors.Is()` матчится по паре (Type, Code), а не по строке сообщения
- `IsType()` — вспомогательная функция для проверки типа ошибки
- `IRedMultiError` реализует `Unwrap() []error` для агрегации нескольких ошибок
- Консистентное обёртывание через `%w` по всей цепочке вызовов

### Использование CLI

```bash
# Сборка
cd iredparser && go build -o bin/iredparser ./cmd/parser-cli/main.go

# Проверка авторизации
./iredparser -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}'

# Синхронизация ящиков
./iredparser -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}' sync

# Смена пароля
./iredparser -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}' change-password

# С конфиг-файлом
./iredparser -c "$(cat config.json)" sync
```

---

## Python TUI (`app/`)

Терминальный интерфейс на [Textual](https://textual.textualize.io/) для просмотра и управления данными почтовых серверов.

### Экраны

| Экран | Назначение |
|---|---|
| **Главное меню** | Навигация по разделам |
| **Поиск** | Таблица с фильтрацией (сервер, админ, бан, квота), сортировка по колонкам, полнотекстовый поиск |
| **Конфигурация** | Добавление / удаление / проверка серверов |
| **Синхронизация** | Запуск синхронизации для выбранного сервера или всех сразу |
| **Прогресс** | Статус смены паролей в реальном времени по каждому ящику |

### Подходы

- `@dataclass` + Textual `reactive` — автоматическое обновление UI при изменении фильтров/состояния
- Repository pattern на слое БД (`ServerRepository`, `DomainRepository`, `MailboxRepository`)
- `@work`-декоратор — асинхронный запуск синхронизации без блокировки UI
- Async/await — Textual'ный async event loop управляет всем I/O

---

## Стек технологий

### Go

| Зависимость | Назначение |
|---|---|
| **Go 1.26.3** | Generics, `t.Context()`, `wg.Go()` |
| **[cobra](https://github.com/spf13/cobra)** v1.10 | CLI-фреймворк, подкоманды, middleware |
| **[goquery](https://github.com/PuerkitoBio/goquery)** v1.12 | jQuery-стиль парсинга HTML |
| **[sqlx](https://github.com/jmoiron/sqlx)** v1.4 | Расширение database/sql: named queries, StructScan |
| **[modernc.org/sqlite](https://modernc.org/sqlite)** v1.52 | Pure Go SQLite-драйвер (без CGO) |
| **[testify](https://github.com/stretchr/testify)** v1.11 | Ассерты и test suites |
| stdlib `net/http/cookiejar` | Управление сессионными куками |
| stdlib `crypto/tls` | TLS с insecure-skip для self-signed сертов |

### Python

| Зависимость | Назначение |
|---|---|
| **[Textual](https://textual.textualize.io/)** | TUI-фреймворк |
| **stdlib sqlite3** | Доступ к SQLite |
| **stdlib asyncio** | Асинхронная конкурентность |

### Инструменты

| Инструмент | Назначение |
|---|---|
| **SQLite** | Встраиваемая БД — общая для Go CLI и Python TUI |
| **ruff** | Линтер + форматтер Python |
| **pyright** | Статический типизатор Python |

---

## Тестирование

### Go (`iredparser/`)

- **Unit-тесты**: in-memory SQLite (`:memory:`) — быстрые тесты БД без файловой системы
- **Интеграционные тесты**: реальный iRedAdmin-сервер через `.test.creds.json`
- **HTTP-тесты**: проверка маппинга HTTP-статусов на sentinel-ошибки
- Покрытие: client, domain parser, mailbox parser, domain sync, mailbox sync, utils, CLI controller

### Python (`app/`)

- **pytest** + Textual testing utilities
- **Фикстуры**: in-memory БД, конфиг-сторадж, мок HTTP-ответов
- **Тесты экранов**: состояние виджетов для search, config, sync

```bash
# Go unit-тесты (без сервера)
cd iredparser && go test ./...

# Go интеграционные тесты (требуется .test.creds.json)
cd iredparser && go test -tags=integration ./...

# Python тесты
pytest
```

---

## Быстрый старт

### Требования

- Go 1.26+
- Python 3.13+
- Доступ к веб-интерфейсу iRedAdmin

### Запуск

```bash
# Клонирование
git clone <repo-url> && cd IRedAdmin-Parser

# Сборка Go CLI
cd iredparser && go build -o bin/iredparser ./cmd/parser-cli/main.go

# Установка Python TUI
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt

# Создание файла с кредами
cp .test.creds.json.dummy config.json
# Отредактируйте config.json с данными вашего сервера

# Запуск
source .venv/bin/activate && python run.py
```

---

## Структура проекта

```
IRedAdmin-Parser/
  iredparser/               # Go CLI — движок парсинга
    cmd/parser-cli/main.go  # Точка входа
    common/                 # Общие типы конфига
    internal/
      controller/           # Cobra-команды + DI
      database/             # SQLite persistence
      parser/
        client/             # HTTP-клиент с cookie jar
        domain/             # Парсер доменов
        mailbox/            # Конкурентный парсер ящиков
      services/
        auth_service/       # Аутентификация
        password_service/   # Смена пароля
      sync/                 # Оркестрация парсинга
    pkg/
      errors/               # Типизированные ошибки
      utils/                # Утилиты
  app/                      # Python TUI — браузер данных
    tui/screens/            # Экраны Textual
    tui/widgets/            # Переиспользуемые компоненты
    backend/                # HTTP-клиент
    database/               # Репозитории SQLite
    services/               # Бизнес-логика
    storage/                # Хранение конфигов
  data/                     # Файлы SQLite (в .gitignore)
```

---

## Архитектурные решения

- **SQLite без CGO** — `modernc.org/sqlite` транслирует SQLite в Go-ассемблер. Нулевая зависимость от C toolchain, тривиальная кросскомпиляция (проблема на mail-серверах с устаревшими Linux).
- **Worker pool вместо goroutine-per-page** — число страниц неизвестно до первого запроса. Фиксированный пул (30 воркеров) предотвращает неограниченный рост горутин.
- **UPSERT с RETURNING** — устраняет race condition SELECT-потом-INSERT при конкурентных синхронизациях.
- **Встраивание структур** — `DomainModel` содержит `parser.Domain`: добавление поля требует ровно двух изменений (поле в модели + колонка в схеме). Без ручного маппинга.

---

## Лицензия

MIT
