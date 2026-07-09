# IRedHelp — Parser & Browser for IRedAdmin Mail Servers

Утилита для парсинга данных из веб-панели **IRedAdmin** с последующим просмотром, фильтрацией и поиском почтовых ящиков, доменов и серверов.

Проект состоит из двух компонентов:

- **CLI-парсер на Go** — авторизуется в IRedAdmin, парсит HTML-страницы доменов и почтовых ящиков, сохраняет данные в SQLite
- **TUI-приложение на Python** — Textual-based интерфейс для просмотра, фильтрации, сортировки и синхронизации данных

---

## Архитектура

```
┌─────────────────────┐     ┌──────────────────────────────┐
│   Python TUI App    │     │      Go CLI Parser           │
│                     │     │                              │
│  ┌───────────────┐  │     │  ┌──────────┐ ┌──────────┐   │
│  │  SearchScreen │  │     │  │  Parser  │ │  Sync    │   │
│  │  ConfigScreen │──┼─────┼──┤  Client  │ │  Service │   │
│  │  SyncScreen   │  │     │  └────┬─────┘ └────┬─────┘   │
│  └───────┬───────┘  │     │       │            │         │
│          │          │     │  ┌────▼────────────▼──────┐  │
│  ┌───────▼───────┐  │     │  │    Database (SQLite)   │  │
│  │  SQLite (RO)  │  │     │  └────────────────────────┘  │
│  └───────────────┘  │     │                              │
└─────────────────────┘     └──────────────────────────────┘
```

Оба компонента используют общую SQLite-базу: Go-парсер пишет, Python TUI читает.

---

## Go CLI (`iredparser/`)

### Структура пакетов

```
iredparser/
├── cmd/parser-cli/main.go          # Точка входа
├── common/                          # Общие типы (ServerConfig)
├── internal/
│   ├── controller/                  # CLI контроллер (cobra commands)
│   ├── database/                    # SQLite (sqlx + modernc.org/sqlite)
│   ├── parser/
│   │   ├── client/                  # HTTP-клиент с cookie jar + TLS
│   │   ├── domain/                  # Парсинг списка доменов
│   │   └── mailbox/                 # Конкурентный парсинг ящиков
│   ├── services/auth_service/       # Сервис аутентификации
│   └── sync/                        # Синхронизация доменов и ящиков
├── pkg/
│   ├── errors/                      # Кастомные sentinel-ошибки
│   └── utils/                       # Конвертация размеров памяти
└── testing/                         # Общие утилиты для тестов
```

### Ключевые особенности

| Особенность | Реализация |
|---|---|
| **HTTP-клиент** | Cookie Jar для сессий, `InsecureSkipVerify` для self-signed сертов, User-Agent маскировка |
| **HTML-парсинг** | `goquery` — jQuery-style селекторы для извлечения данных из таблиц |
| **Конкурентный парсинг** | Worker pool (30 горутин), каналы `jobs`/`results`, `sync.WaitGroup` для асинхронного парсинга страниц почтовых ящиков |
| **База данных** | `sqlx` + `modernc.org/sqlite` (pure Go, без CGO), UPSERT через `ON CONFLICT ... DO UPDATE ... RETURNING`, транзакции для массовых операций |
| **CLI Framework** | `cobra` — команды `auth-check` и `sync`, middleware для аутентификации через `PersistentPreRun` |
| **Пользовательские ошибки** | Sentinel errors (`ErrInvalidCredentials`, `ErrPostRequestCreation`) для явного разбора иерархии ошибок |
| **Тестирование** | Интеграционные тесты с реальным сервером через `.test.creds.json`, in-memory SQLite для unit-тестов репозитория |

### Пример использования

```bash
# Проверка авторизации
./iredparser/bin/iredparser auth-check \
  --config '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}'

# Синхронизация доменов и ящиков
./iredparser/bin/iredparser sync \
  --config '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}'
```

---

## Python TUI

Терминальный интерфейс на [Textual](https://textual.textualize.io/) для просмотра результатов парсинга.

- **Главное меню** — навигация по разделам
- **Экран поиска** — таблица с фильтрацией (сервер, админ, бан, квота), сортировкой по колонкам, живым поиском
- **Экран конфигурации** — добавление/удаление/проверка серверов для синхронизации
- **Экран синхронизации** — запуск синхронизации для выбранного сервера или всех сразу

### Использованные практики

- **Repository pattern** для слоя БД (`ServerRepository`, `DomainRepository`, `MailboxRepository`)
- **Dataclasses + `reactive`** для реактивного обновления UI при изменении фильтров
- **Декораторы** для валидации квоты и обновления фильтров
- **`@work` (worker decorator)** для асинхронного запуска синхронизации без блокировки UI

---

## Использованные технологии

### Go
- **Go 1.26** — generics, `t.Context()`, `wg.Go()`
- **`cobra`** — CLI framework
- **`goquery`** — HTML парсинг (jQuery-like selectors)
- **`sqlx`** — расширение database/sql с named queries и StructScan
- **`modernc.org/sqlite`** — pure-Go SQLite (без CGO)
- **`stretchr/testify`** — assertions

### Python
- **Textual** — TUI framework
- **SQLite3** (stdlib) — чтение общей БД
- **`asyncio`** — конкурентная синхронизация

---

## Go-практики, демонстрируемые в проекте

1. **Чистая архитектура пакетов** — `cmd/` для точки входа, `internal/` для закрытой логики, `pkg/` для публичных утилит
2. **Dependency Injection** — контроллер получает готовые сервисы через конструктор (`NewCLIController`)
3. **Interface Segregation** — маленькие интерфейсы (`AuthChecker`, `SyncService`, `Storage`) вместо одного монолитного
4. **Sentinel Errors** — пакет `pkg/errors` с `errors.Is()` для точной типизации ошибок
5. **Асинхронная обработка** — worker pool через каналы + `sync.WaitGroup`, конкурентный парсинг пагинированных страниц
6. **SQL Upsert с RETURNING** — атомарная вставка/обновление с получением ID, без лишних SELECT
7. **Транзакции** — массовые операции (`UpsertDomainMany`, `UpsertMailboxMany`) в одной транзакции для согласованности
8. **Встраивание (embedding)** — модель `ServerModel` встраивает `parser.Server` через композицию структур
9. **Модульное тестирование** — in-memory SQLite для быстрых тестов; интеграционные тесты через конфигурационный файл-заглушку

---

## Установка и запуск

```bash
# Go-парсер
cd iredparser
go build -o ./bin/iredparser ./cmd/parser-cli/main.go

# Python TUI
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python run.py
```

Скопируйте `.test.creds.json.dummy` в `.test.creds.json` и заполните своими данными перед запуском интеграционных тестов.

---

## Лицензия

MIT
