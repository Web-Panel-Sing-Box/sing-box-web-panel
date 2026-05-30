# Описание бэкенда Sing-Box Web Panel

Local-first / self-hosted панель управления ядром [`sing-box`](https://github.com/SagerNet/sing-box)
по аналогии с [3x-ui](https://github.com/MHSanaei/3x-ui). Бэкенд написан на Go, максимально
легковесный, с минимумом зависимостей и прямым доступом к системным ресурсам. Панель управляет
только локальным процессом `sing-box` и общается с его API исключительно через `127.0.0.1`.

> Этот документ описывает целевую архитектуру бэкенда. Часть модулей (аутентификация) уже
> реализована, остальное собирается в рамках текущей работы. Технические правила проекта —
> в [`AGENTS.md`](AGENTS.md).

---

## 1. Технологический стек

- **Go 1.26**, стандартная библиотека `net/http` (роутинг через `http.ServeMux` с методами Go 1.22+).
- **SQLite** через `modernc.org/sqlite` (чистый Go, без CGo), режим WAL, батч-запись.
- **Логирование** — `log/slog` (структурированное), `github.com/fatih/color` для dev-режима.
- **Аутентификация** — `golang-jwt/jwt/v5`, `golang.org/x/crypto` (Argon2id), `pquerna/otp` (TOTP).
- **Конфигурация** — `ilyakaznacheev/cleanenv` (YAML + переопределение через переменные окружения).
- **Swagger** — `swaggo/swag` + `swaggo/http-swagger`.
- **Опциональные зависимости** (изолированы в своих модулях):
  - `google.golang.org/grpc` + `google.golang.org/protobuf` — адаптер статистики V2Ray API.
  - `golang.org/x/crypto/acme/autocert` — Let's Encrypt для самой панели.

Системные метрики (CPU/RAM/диск) читаются напрямую из `/proc` и через `syscall` на Linux —
без сторонних библиотек.

---

## 2. Архитектура

Чистая многослойная архитектура (направление зависимостей внутрь):

```
transport (HTTP-хендлеры, middleware)
        ↓
services (бизнес-логика)
        ↓
repo (интерфейсы) → repo/sqlite (реализация)
        ↓
domain (модели предметной области)
```

Фоновые воркеры (сбор статистики, чтение логов ядра) запускаются в `cmd/main.go` на общем
`context.Context` и корректно останавливаются вместе с HTTP-сервером (graceful shutdown по
`SIGINT`/`SIGTERM`).

### Структура каталогов (целевая)

```
cmd/main.go                     # точка входа: конфиг, БД, сборка зависимостей, воркеры, сервер
internal/
  config/                       # Config + MustLoad()
  domain/                       # Admin, Inbound, Client, Setting, ConfigRevision, Metrics …
  repo/
    sqlite/                     # реализации репозиториев + миграции
  services/
    auth/                       # аутентификация (реализовано)
    inbound/                    # CRUD инбаундов, генерация ключей Reality
    client/                     # CRUD клиентов, лимиты, токены подписок
    singbox/                    # генератор конфига, проверка, процесс-менеджер, apply
    stats/                      # сбор трафика (Clash/V2Ray) + воркер лимитов
    sysstat/                    # системные метрики (Linux /proc + заглушка для разработки)
    sublink/                    # сборка ссылок и подписок
    tlsmgr/                     # TLS панели (файлы / самоподписанный / ACME)
    logbuf/                     # кольцевой буфер логов
  transport/
    handler/                    # HTTP-хендлеры по доменам
    middleware/                 # Logger, CORS, Auth, RateLimit
```

---

## 3. Модель данных (SQLite)

Миграции встроены через `//go:embed`, версионируются (`schema_migrations`), применяются на старте.

| Таблица | Назначение |
|---|---|
| `admins` | учётные записи администраторов (логин, Argon2id-хэш, секрет TOTP) — *реализовано* |
| `admin_recovery_codes` | одноразовые коды восстановления (Argon2id) — *реализовано* |
| `inbounds` | инбаунды: протокол, порт, транспорт, режим TLS, SNI, dest, `settings_json` (ключи Reality, short_id, параметры транспорта), флаг включения |
| `clients` | клиенты: привязка к инбаунду, имя (идентификатор для статистики), UUID/пароль, лимит трафика, дата истечения, статус, токен подписки, счётчики Up/Down |
| `settings` | системные настройки панели (ключ/значение JSON): публичный URL подписок, пути к сертификатам и т.п. |
| `config_revisions` | история применённых конфигов (sha256, результат проверки, ошибка) для аудита и отката |
| `traffic_rollup` | суточные агрегаты трафика (для показателей «сегодня»/«за месяц» на дашборде) |

**Протоколы v1:** VLESS (TLS/Reality; транспорты tcp/ws/grpc), Naive, Hysteria2 — ровно те,
что моделирует существующий фронтенд.

---

## 4. Ключевые модули

### 4.1. Управление процессом ядра (`services/singbox`)

`ProcessManager` — интерфейс с методами `Start / Stop / Restart / Reload / Status`
(running, pid, uptime, версия). Две реализации с авто-определением режима:

- **systemd** — `systemctl start|stop|restart|reload|is-active|show <unit>`.
- **subprocess** — прямой запуск `sing-box run -c <config>`; `SIGHUP` для reload, `SIGTERM`
  для остановки; stdout/stderr направляются в буфер логов.

Все системные команды вызываются массивом аргументов (никогда не строкой шелла из
пользовательского ввода).

> **Важная техническая поправка.** В отличие от Xray/v2ray, `sing-box` **не** выполняет
> reload без разрыва соединений: перезагрузка по `SIGHUP` / `systemctl reload` перечитывает
> конфиг, но сбрасывает активные соединения
> ([SagerNet/sing-box#3731](https://github.com/SagerNet/sing-box/issues/3731)).
> Поэтому панель не обещает «бесшовный» reload — она гарантирует валидность конфига и предлагает
> `reload` (краткий сброс) либо `restart`.

### 4.2. Генерация конфигурации (`services/singbox/generator.go`)

Конфиг `config.json` собирается из типизированных Go-структур (JSON-теги соответствуют схеме
`sing-box`) и сериализуется в JSON.

- **Статические блоки:** `log`, `dns` (Cloudflare + локальный, правила маршрутизации),
  опционально `ntp`, базовые `outbounds` (`direct`, `block`), `route` (блокировка торрентов/рекламы,
  `final: direct`). Outbound `warp`/WireGuard — на будущее.
- **`experimental.clash_api`** — всегда (`external_controller` на `127.0.0.1`, `secret`).
- **`experimental.v2ray_api`** — только если выбран источник статистики V2Ray
  (`listen` + `stats` с перечнем пользователей/инбаундов/аутбаундов).
- **`experimental.cache_file`** — для персистентности.
- **Динамические `inbounds`** — по одному на каждый включённый инбаунд из БД, со списком
  `users` из его клиентов:
  - **VLESS** — `users: [{name, uuid, flow}]`, TLS → Reality
    (`private_key`, `short_id`, `handshake.server`/`server_port` из `dest`, `server_name` из `sni`)
    либо обычный TLS; транспорт tcp/ws/grpc.
  - **Hysteria2** — `users: [{name, password}]`, TLS (сертификаты или `acme`).
  - **Naive** — `users: [{username, password}]`, обязательный TLS.

Пара ключей **Reality** (curve25519) и `short_id` генерируются на стороне сервера при создании
инбаунда и хранятся в `settings_json`. Приватный ключ никогда не логируется и не отдаётся наружу,
публичный — используется в ссылках.

### 4.3. Применение конфига (`services/singbox/apply.go`)

Перед применением нового конфига обязательна проверка:

1. Рендер конфига → запись во временный файл.
2. `sing-box check -c <tmp>` (таймаут из конфига).
3. **Успех** → атомарная замена `config_path` (write + rename), reload ядра, запись ревизии (ok).
4. **Ошибка** → конфиг не применяется (сеть не падает), ошибка проверки возвращается в UI,
   ревизия сохраняется с текстом ошибки.

Применения дебаунсятся, чтобы массовые изменения объединялись в один reload.

### 4.4. Сбор статистики и лимиты (`services/stats`)

`TrafficSource` — интерфейс источника трафика. Две реализации:

- **Clash API (REST, по умолчанию)** — опрос `/traffic` (глобальная скорость), `/connections`
  (активные соединения → «онлайн сейчас», дельты Up/Down по пользователям из `metadata.user`),
  `/version`. Работает со штатным официальным бинарником.
- **V2Ray API (gRPC, опционально)** — `StatsService.QueryStats` с именами вида
  `user>>>ИМЯ>>>traffic>>>uplink|downlink`. Требует бинарник `sing-box`, собранный с тегом
  `with_v2ray_api`. Включается автоматически, если такой источник доступен.

**Воркер** по таймеру накапливает Up/Down по клиентам и пишет в SQLite **батчами** (без
по-опросной записи на диск). Логика лимитов: если исчерпан лимит трафика
(`quota>0 && used≥quota`) или истёк срок (`now>expiry`) — клиент помечается
`expired`/`disabled`, отключается, после чего запускается перегенерация и применение конфига
(с дебаунсом) — пользователь удаляется из «живой» конфигурации ядра.

### 4.5. Сервер подписок (`services/sublink`, публичный хендлер)

- У каждого клиента есть `sub_token`. URL подписки = `subscription.public_url` + `/sub/{token}`.
- `GET /sub/{token}?format=base64|plain|json` (и `/api/subscription/{token}`) — **публичный**
  эндпоинт (без JWT). Определяет клиента по токену, проверяет статус/срок и отдаёт:
  - `plain` — список ссылок (`vless://…#name`, `hysteria2://…`, `naive+https://…`);
  - `base64` — base64 от plain (для старых клиентов);
  - `json` — готовый клиентский конфиг `sing-box` (современные клиенты парсят чистый конфиг).
- `GET /api/clients/{id}/links` (под JWT) — те же ссылки + URL подписки для UI.
  Рендер QR — на стороне фронтенда (без графических зависимостей в бэкенде).

### 4.6. Системные метрики и логи (`services/sysstat`, `services/logbuf`)

- **Метрики дашборда** — CPU/RAM/swap/диск/uptime (на Linux из `/proc` и `syscall.Statfs`;
  на macOS-разработке — заглушка), скорости Up/Down (Clash `/traffic`), «онлайн» (Clash
  `/connections`), число инбаундов/пользователей (БД), суммарный трафик (Σ счётчиков клиентов),
  «сегодня»/«за месяц» (`traffic_rollup`), статус и версия ядра.
- **Логи** — кольцевой буфер: логи ядра (stdout subprocess или `journalctl -u <unit>`) + логи
  панели (через slog-приёмник). Эндпоинт `GET /api/logs` с фильтрами; опционально SSE-стрим.
  Без websocket-зависимостей.

### 4.7. TLS / Let's Encrypt (`services/tlsmgr`)

- **Панель:** режимы `file` (готовые cert/key), `self_signed` (генерация x509 с IP в SAN —
  HTTPS работает даже на «голом» IP, без домена), `acme` (`autocert` для доменов, HTTP-01/TLS-ALPN).
- **Инбаунды:** в конфиг ядра эмитится `tls.acme {domain, email}` (выпуск делает само ядро,
  тег `with_acme` есть в официальных сборках) либо пути к сертификатам.

---

## 5. Безопасность панели

Панель имеет доступ к управлению сетевыми службами, поэтому защищена многоуровнево:

- **Аутентификация** — только по JWT (HS256, cookie/Bearer) с ограниченным сроком жизни;
  пароли — Argon2id (PHC), 2FA — TOTP, одноразовые коды восстановления.
- **Защита от брутфорса** — `middleware/ratelimit.go`: per-IP token bucket, временная блокировка
  IP после серии неудачных входов; отдельные лимиты на `/api/auth/login*` и на API в целом.
- **Изоляция** — все управляющие API ядра (Clash/V2Ray) слушают только `127.0.0.1`.
- **Секреты** — JWT-секрет, пароль администратора, секрет API ядра, UUID и приватные ключи
  никогда не логируются и не коммитятся.

---

## 6. HTTP API

JSON новых ресурсных эндпоинтов — в **camelCase**, совпадает с типами фронтенда
(`frontend/src/lib/mock/*`), чтобы переключение мок-стора на реальные запросы было бесшовным.
Эндпоинты аутентификации сохраняют исходный snake_case.

| Группа | Эндпоинты |
|---|---|
| Inbounds | `GET/POST /api/inbounds`, `GET/PUT/DELETE /api/inbounds/{id}`, `POST .../toggle`, `POST .../clone` |
| Clients | `GET/POST /api/clients` (`?inboundId=`), `GET/PUT/DELETE /api/clients/{id}`, `POST .../reset-traffic`, `POST .../status`, `GET .../links` |
| Core | `GET /api/core/status`, `POST /api/core/{start\|stop\|restart\|reload}`, `GET /api/core/version`, `GET /api/core/config` |
| Dashboard | `GET /api/dashboard/metrics`, `GET /api/dashboard/traffic` |
| Logs | `GET /api/logs` (+ опционально `/api/logs/stream`) |
| Settings | `GET /api/settings`, `PUT /api/settings` |
| Subscription | `GET /sub/{token}`, `GET /api/subscription/{token}` (**публичные**) |
| Auth (реализовано) | `/api/auth/login`, `/login/recovery`, `/me`, `/logout`, `/totp/*`, `/change-password` |

---

## 7. Конфигурация

Основной файл — `config/dev.yaml` (YAML), секреты переопределяются переменными окружения.
К существующим секциям (`runtime`, `database`, `http`, `frontend`, `auth`, `sing_box`,
`metrics`, `logging`, `subscription`) добавляются:

- `tls` — `mode (off|file|self_signed|acme)`, `cert_file`, `key_file`, `acme_email`,
  `acme_domains`, `acme_cache_dir`, `self_signed_hosts`.
- `stats` — `source (auto|clash|v2ray)`, `v2ray_api_address`.
- `sing_box` — добавлены `process_mode (auto|systemd|subprocess)` и `service_name`.

---

## 8. Соответствие исходному ТЗ

| Требование ТЗ | Реализация |
|---|---|
| Управление службой (start/stop/restart/status) | `ProcessManager` (systemd + subprocess) |
| Hot reload по сигналу | `Reload` (SIGHUP) — с поправкой: соединения сбрасываются |
| `sing-box check` перед применением | `apply.go`: проверка обязательна, иначе откат |
| Статические + динамические блоки конфига | `generator.go` |
| Маршрутизация (блокировки, обходы) | блок `route` в генераторе |
| `experimental.api` (порт 127.0.0.1:9090) | `clash_api` всегда в конфиге |
| Сбор метрик по таймеру | воркер `stats` (Clash REST / V2Ray gRPC) |
| Суммирование трафика и лимиты | воркер: батч-запись + enforcement → reload |
| Таблицы Inbounds / Clients / Settings | миграции 000003–000007 |
| Сервер подписок (URI, base64/plain/json) | `sublink` + публичный хендлер |
| JWT/сессии, защита от брутфорса | auth (реализовано) + `ratelimit` |
| HTTPS, Let's Encrypt, HTTPS на IP | `tlsmgr` (file / self_signed / acme) |
