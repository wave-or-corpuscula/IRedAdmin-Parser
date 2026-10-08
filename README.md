# IRedAdmin Parser

Инструмент для **парсинга, синхронизации и управления почтовыми ящиками iRedAdmin**.

Проект состоит из двух частей:

* **Go CLI** — основной движок: авторизация в iRedAdmin, парсинг HTML, конкурентная обработка почтовых ящиков, сохранение данных в SQLite.
* **Python TUI** — терминальный интерфейс для поиска, просмотра и управления сохранёнными данными.

Обе части используют одну SQLite-базу.

---

## Возможности

* Авторизация в iRedAdmin.
* Парсинг доменов и почтовых ящиков.
* Конкурентный парсинг большого количества страниц через worker pool.
* Сохранение и обновление данных в SQLite.
* Поиск и фильтрация почтовых ящиков через TUI.
* Синхронизация данных с iRedAdmin.
* Удалённая смена паролей почтовых ящиков.
* Проверка доступности и конфигурации почтовых серверов.
* Типизированная обработка ошибок.
* Unit- и integration-тесты.

---

## Архитектура

```text
                    ┌─────────────────────┐
                    │       iRedAdmin     │
                    │    Web Interface    │
                    └──────────┬──────────┘
                               │ HTTP
                               ▼
                    ┌─────────────────────┐
                    │      Go CLI         │
                    │                     │
                    │  HTTP Client        │
                    │  Domain Parser      │
                    │  Mailbox Parser     │
                    │  Auth Service       │
                    │  Password Service   │
                    │  Sync Service       │
                    └──────────┬──────────┘
                               │
                               │ SQL
                               ▼
                    ┌─────────────────────┐
                    │       SQLite        │
                    └──────────┬──────────┘
                               │
                               │ SQL
                               ▼
                    ┌─────────────────────┐
                    │     Python TUI      │
                    │                     │
                    │  Search             │
                    │  Filtering          │
                    │  Synchronization    │
                    │  Password changes   │
                    └─────────────────────┘
```

---

# Go CLI

Основная часть проекта находится в [`iredparser/`](iredparser/).

### Структура

```text
iredparser/
├── cmd/
│   └── parser-cli/
│       └── main.go
├── common/
│   └── common.go
├── internal/
│   ├── controller/       # CLI-команды и dependency injection
│   ├── database/         # SQLite persistence
│   ├── parser/
│   │   ├── client/       # HTTP-клиент
│   │   ├── domain/       # Парсинг доменов
│   │   └── mailbox/      # Конкурентный парсинг ящиков
│   ├── services/
│   │   ├── auth_service/ # Авторизация
│   │   └── password_service/
│   │                       # Смена паролей
│   └── sync/             # Синхронизация данных
├── pkg/
│   ├── errors/           # Типизированные ошибки
│   └── utils/            # Общие утилиты
└── testing/              # Тестовые утилиты
```

## HTTP-клиент

Клиент построен поверх стандартного `net/http`.

Используются:

* `http.Client`;
* `cookiejar.Jar` для сохранения сессии;
* connection pool;
* TLS-конфигурация для серверов с self-signed сертификатами;
* автоматическое извлечение сессионных cookies;
* типизированные ошибки для HTTP-сбоев.

Все запросы выполняются в рамках `context.Context`.

---

## Конкурентный парсинг

Парсинг почтовых ящиков выполняется через фиксированный worker pool.

```text
                  URLs
                   │
                   ▼
             ┌───────────┐
             │  channel  │
             └─────┬─────┘
                   │
       ┌───────────┼───────────┐
       ▼           ▼           ▼
    Worker 1    Worker 2    Worker N
       │           │           │
       └───────────┼───────────┘
                   ▼
              parsed data
```

Количество страниц неизвестно до выполнения первого запроса, поэтому вместо создания отдельной goroutine для каждой страницы используется ограниченный worker pool.

Это позволяет контролировать количество одновременных HTTP-запросов и избежать неограниченного роста количества goroutines.

Для разбора HTML используется [`goquery`](https://github.com/PuerkitoBio/goquery).

---

## SQLite

Для работы с базой используются:

* [`sqlx`](https://github.com/jmoiron/sqlx);
* [`modernc.org/sqlite`](https://modernc.org/sqlite).

SQLite-драйвер работает без CGO.

Основные операции сохранения используют SQL `UPSERT`:

```sql
INSERT INTO ...
VALUES (...)
ON CONFLICT (...) DO UPDATE
SET ...
RETURNING id;
```

Это позволяет выполнять вставку или обновление и получать идентификатор записи одним запросом.

Массовые операции выполняются внутри транзакций:

```text
BEGIN
  ├── upsert domain
  ├── upsert mailbox
  ├── upsert mailbox
  └── ...
COMMIT
```

При ошибке транзакция откатывается целиком.

---

## CLI

CLI построен с помощью [`cobra`](https://github.com/spf13/cobra).

Доступны команды:

```text
auth-check
sync
change-password
```

### Проверка авторизации

```bash
./iredparser \
  -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}'
```

### Синхронизация

```bash
./iredparser \
  -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}' \
  sync
```

### Смена пароля

```bash
./iredparser \
  -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}' \
  change-password
```

Конфигурацию также можно передать из файла:

```bash
./iredparser -c "$(cat config.json)" sync
```

---

# Python TUI

Python-приложение находится в [`app/`](app/).

Для интерфейса используется [`Textual`](https://textual.textualize.io/).

Основные экраны:

| Экран    | Назначение                     |
| -------- | ------------------------------ |
| Main     | Навигация                      |
| Search   | Поиск и фильтрация ящиков      |
| Config   | Управление почтовыми серверами |
| Sync     | Запуск синхронизации           |
| Progress | Отображение прогресса операций |

Поиск поддерживает фильтрацию, сортировку и поиск по данным почтовых ящиков.

Для работы с SQLite используется repository pattern:

```text
ServerRepository
DomainRepository
MailboxRepository
```

Длительные операции выполняются асинхронно и не блокируют интерфейс.

---

# Обработка ошибок

В Go-части используется собственная иерархия ошибок.

Каждая ошибка содержит:

* тип ошибки;
* числовой код;
* исходную ошибку.

Основные категории:

```text
authentication
HTTP
parsing
CLI
```

Ошибки можно проверять через стандартный механизм Go:

```go
errors.Is(err, target)
```

Для агрегирования нескольких ошибок используется `IRedMultiError`, реализующий:

```go
Unwrap() []error
```

Ошибки при этом сохраняют исходную цепочку через `%w`.

---

# Тестирование

## Go

Используются:

* unit-тесты;
* integration-тесты;
* HTTP-тесты;
* in-memory SQLite;
* [`testify`](https://github.com/stretchr/testify).

Тестами покрыты:

* HTTP client;
* авторизация;
* domain parser;
* mailbox parser;
* database layer;
* domain sync;
* mailbox sync;
* utilities;
* CLI controller.

### Unit-тесты

Не требуют подключения к iRedAdmin:

```bash
cd iredparser
go test ./...
```

### Integration-тесты

Работают с реальным iRedAdmin-сервером.

Конфигурация сервера хранится в `.test.creds.json`.

```bash
cd iredparser
go test -tags=integration ./...
```

---

## Python

Python-часть тестируется через `pytest`.

Используются:

* in-memory SQLite;
* тестовые конфигурации;
* mock HTTP responses;
* Textual testing utilities.

Запуск:

```bash
pytest
```

---

# Быстрый старт

## Требования

* Go 1.26+
* Python 3.13+
* доступ к iRedAdmin

## Клонирование

```bash
git clone https://github.com/wave-or-corpuscula/IRedAdmin-Parser.git
cd IRedAdmin-Parser
```

## Go

```bash
cd iredparser

go build -o bin/iredparser ./cmd/parser-cli/main.go
```

## Python

```bash
cd ..

python3 -m venv .venv
source .venv/bin/activate

pip install -r requirements.txt
```

Создайте конфигурацию:

```bash
cp .test.creds.json.dummy config.json
```

Заполните её данными вашего iRedAdmin-сервера.

Запуск TUI:

```bash
python run.py
```

---

# Технологический стек

### Go

* Go 1.26
* Cobra
* goquery
* sqlx
* SQLite
* testify
* `net/http`
* `cookiejar`
* `crypto/tls`

### Python

* Python 3.13
* Textual
* asyncio
* sqlite3
* pytest
* ruff
* pyright

---

# Архитектурные решения

### Worker pool

Количество страниц определяется динамически, поэтому используется ограниченное количество workers вместо `goroutine-per-page`.

### SQLite без CGO

`modernc.org/sqlite` позволяет использовать SQLite без системного C toolchain.

Это упрощает сборку и развёртывание Go CLI.

### UPSERT + RETURNING

Вместо последовательности:

```text
SELECT
  ↓
INSERT / UPDATE
```

используется атомарный:

```text
INSERT ... ON CONFLICT ... DO UPDATE ... RETURNING
```

Это уменьшает количество запросов и устраняет race condition между конкурентными операциями.

### Dependency Injection

Зависимости CLI-контроллера передаются через конструктор:

```go
NewCLIController(
    client,
    storage,
    authService,
    syncService,
    passwordService,
    writer,
)
```

Глобальное состояние не используется, поэтому компоненты проще изолировать в тестах.

### Interface segregation

Зависимости описываются небольшими интерфейсами с одной ответственностью:

```go
AuthChecker
SyncService
PasswordChanger
Storage
```

---

# Лицензия

MIT

