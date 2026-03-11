# ⏳ Load Tester

Нагрузочный тестировщик **HTTP-сервисов**, написанный на **Go**.

---

# 🏗 Архитектура


```
cmd/         — точки входа приложений
config/      — конфигурация и валидация
engine/      — оркестратор теста
worker/      — воркер, выполняющий запросы
httpclient/  — HTTP клиент (интерфейс + реализация)
metrics/     — сбор и агрегация метрик
reporter/    — вывод результатов
mockserver/  — мок-сервер для тестирования
```

---

# ▶️ Запуск

## Mock Server

```bash
go run cmd/mockserver/main.go

Доступные endpoints:

Endpoint	    Описание
GET /fast	    быстрый ответ
GET /slow	    задержка 100-150ms
GET /flaky	    30% ошибок
GET /health	    health check

Примеры запуска тестов

Тест на 100 запросов, 10 воркеров
go run cmd/loadtester/main.go \
  -url http://localhost:8080/fast \
  -c 10 \
  -n 100


Тест длительностью 30 секунд, 20 воркеров

go run cmd/loadtester/main.go \
  -url http://localhost:8080/slow \
  -c 20 \
  -d 30s

Unit Тесты
go test ./...

📊 Пример вывода
──────────────────────────────────────────────────
           ## LOAD TEST RESULTS
──────────────────────────────────────────────────
  Total Requests:    1000
  Success:           970
  Failed:            30
  Duration:          5.2s
  Req/sec:           192.30
──────────────────────────────────────────────────
  ## LATENCY
  Min:               1.2ms
  Avg:               48.3ms
  Max:               312ms
  P50:               45ms
  P90:               98ms
  P95:               145ms
  P99:               287ms
──────────────────────────────────────────────────
  ## STATUS CODES
  [200]:             970
  [500]:             30
──────────────────────────────────────────────────