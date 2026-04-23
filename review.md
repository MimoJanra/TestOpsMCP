# Ревью MCP сервера — список правок

## 🔴 Критические проблемы

### 1. Неправильные эндпоинты Allure TestOps API

**Файл:** `internal/adapters/allure/client.go`

Используется префикс `/api/v2/`, но Allure TestOps использует Report Service — `/api/rs/`.

| Текущий (❌) | Правильный (✅) |
|---|---|
| `POST /api/v2/launch` | `POST /api/rs/launch` |
| `GET /api/v2/launch/{id}` | `GET /api/rs/launch/{id}` |
| `GET /api/v2/launch/{id}/statistics` | `GET /api/rs/launch/{id}/statistic` (без `s`) |

Спецификация: `<your-allure-domain>/swagger-ui.html`.

---

### 2. Статусы Allure не совпадают со спецификацией plan.md

Plan.md требует `"running|failed|passed"`, но Allure TestOps возвращает:
- Launch lifecycle: `CREATED`, `RUNNING`, `STOPPED`, `CLOSED`
- Результаты тестов: `passed`, `failed`, `broken`, `skipped`, `unknown`

Нужна нормализация в слое адаптера.

---

### 3. Авторизация — требуется JWT обмен

Plan говорит `Authorization: Bearer <token>`, но Allure TestOps требует:
1. `POST /api/uaa/oauth/token` с API token → получить JWT
2. Использовать JWT в `Authorization: Bearer <JWT>`
3. JWT живёт 1 час → нужен refresh mechanism

**Файл:** `internal/adapters/allure/client.go:123`

Текущий код отправляет API token как Bearer напрямую — это работает не со всеми инсталляциями.

---

## 🟡 Проблемы MCP протокола

### 4. Нет обработки `notifications/initialized`

**Файл:** `internal/mcp/server.go:68`

По спецификации MCP клиент шлёт `notifications/initialized` после `initialize`. Сервер должен молчать (notifications = нет `id`). Сейчас вернётся `-32601 Method not found`.

Нужна проверка на notification (id отсутствует или `null`).

---

### 5. Ошибки tool call должны идти в `content` с `isError: true`

**Файл:** `internal/mcp/server.go:159`

По спецификации MCP ошибки исполнения tool — не JSON-RPC error, а обычный response:

```json
{
  "content": [{"type": "text", "text": "error message"}],
  "isError": true
}
```

Сейчас возвращается JSON-RPC ошибка `-32603` — это неправильно для ошибок исполнения tool.

---

### 6. Scanner default buffer 64KB

**Файл:** `internal/mcp/server.go:34`

`bufio.NewScanner` имеет лимит 64KB на строку. Large payloads с `tools/list` могут упасть.

Нужно:
```go
buf := make([]byte, 1024*1024)
scanner.Buffer(buf, 1024*1024)
```

---

### 7. InitializeResponse не заполняется Capabilities явно

**Файлы:** `internal/mcp/protocol.go:36-45`, `internal/mcp/server.go:97`

Структура Capabilities создаётся с нулевыми значениями. Работает, но явное заполнение лучше.

---

## 🟢 Минорные замечания

### 8. Двойной таймаут

**Файл:** `internal/adapters/allure/client.go:24`

Используется и `http.Client.Timeout`, и `context.WithTimeout`. Достаточно контекста.

---

### 9. Тихое игнорирование ошибки

**Файл:** `internal/config/config.go:21`

```go
timeoutSec, _ := strconv.Atoi(timeoutStr)
```

Ошибка парсинга игнорируется, силой ставится 30. Нарушает жёсткое требование plan.md: "полная обработка ошибок".

---

### 10. `config.Load()` не валидирует `AllureBaseURL`

**Файл:** `internal/config/config.go`

Пустое значение ловится в main, но невалидный URL (например `not-a-url`) не проверяется.

---

### 11. Нет поля `Data` при отправке JSONRPCError

**Файл:** `internal/mcp/server.go:197`

`sendError` не принимает `data`, хотя структура `JSONRPCError` его поддерживает.

---

### 12. `id` в логе выводится как raw bytes

**Файл:** `internal/mcp/server.go:71`

```go
"id": string(req.ID),
```

Выведет `"1"` или `null` байтами. JSON ID может быть любой (строка, число, null). Лучше unmarshal в interface{}.

---

### 13. `resultToJSON` теряет ошибку

**Файл:** `internal/mcp/server.go:180`

Если `Marshal` упал, возвращается строка `{"error": "..."}`. Лучше вернуть error вверх по стеку.

---

### 14. `ToolCallResponse.Content` — `[]interface{}`

**Файл:** `internal/mcp/protocol.go:63`

Типобезопаснее сделать `[]ContentItem` с интерфейсом `ContentItem`.

---

## ✅ Что сделано хорошо

- Структурированное логирование в stderr (stdout резервирован под MCP)
- Mutex на запись в stdout — потокобезопасность
- Graceful shutdown через сигналы
- Context прокидывается по всему стеку
- Валидация JSON-RPC version
- Registry с sync.RWMutex
- Строгая структура директорий по plan.md
- Валидация параметров в handlers
- Корректная обработка множественных запросов в stdio loop

---

## 🎯 Приоритет исправлений

| # | Приоритет | Правка |
|---|-----------|--------|
| 1 | **Срочно** | API эндпоинты `/api/v2/` → `/api/rs/` |
| 2 | **Срочно** | Обработка `notifications/initialized` |
| 3 | **Важно** | Ошибки tool → `isError: true` в content |
| 4 | **Важно** | Буфер scanner'а до 1MB |
| 5 | **Важно** | JWT-обмен (если нужна совместимость с облаком Allure) |
| 6 | Желательно | Нормализация статусов под plan.md формат |
| 7 | Желательно | Обработка ошибки в `config.Load()` |
| 8 | Желательно | Убрать двойной таймаут в HTTP клиенте |

---

## 📚 Источники

- [API of Allure TestOps](https://docs.qameta.io/allure-testops/advanced/api/)
- [Allure TestOps API usage guide](https://help.qameta.io/support/solutions/articles/101000507439-general-approach-for-using-allure-testops-api-calls)
- [Launches | Allure TestOps Docs](https://docs.qameta.io/allure-testops/briefly/launches/)
- [allurectl GitHub](https://github.com/allure-framework/allurectl)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
