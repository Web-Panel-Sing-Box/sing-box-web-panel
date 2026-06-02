# Shilka · Веб-панель для Sing-box

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](./LICENSE)

[Русский](./README_RU.md) | [English](./README.md)

---

### Что такое Shilka?

Shilka — это **локальная веб-панель** для управления прокси-сервером `sing-box`. Работает как **единый бинарник** со встроенным React-фронтендом — без внешнего веб-сервера, без Node на хосте, без Docker. Создана для VPS с минимальным потреблением ресурсов (~30 МБ RAM для самой панели).

- Управление несколькими входящими соединениями **VLESS**, **Hysteria2** и **Naive**
- Создание клиентов с квотами, сроком действия и индивидуальными токенами подписки
- Мониторинг трафика, нагрузки системы и ядра sing-box в реальном времени
- TOTP 2FA с кодами восстановления для входа администратора
- Публичная конечная точка подписки для автоматического импорта клиентских конфигов

---

### Возможности

| Категория          | Детали                                                                    |
| ------------------ | ------------------------------------------------------------------------- |
| **Протоколы**      | VLESS (Reality + Flow), Hysteria2 (H3 ALPN), Naive                        |
| **Транспорты**     | TCP, WebSocket, gRPC                                                      |
| **TLS**            | None, TLS, Reality (x25519 ключи) — на каждый инбаунд                     |
| **Аутентификация** | Argon2id пароли, JWT, TOTP 2FA + коды восстановления                      |
| **Трафик**         | Счётчики байт на клиента, квоты, деактивация по истечении срока           |
| **Подписка**       | Индивидуальные токены, экспорт в plain / base64 / JSON, QR на сервере     |
| **Интерфейс**      | Тёмная минималистичная панель, дашборд в реальном времени, i18n (EN + RU) |
| **Конфигурация**   | YAML файл + переменные окружения, проверка при запуске                    |

---

### Быстрый старт (VPS)

Одна команда. Устанавливает всё: пользователя, директории, бинарник sing-box, TLS-сертификат, systemd юнит.

```bash
bash <(curl -Ls https://raw.githubusercontent.com/Web-Panel-Sing-Box/sing-box-web-panel/main/scripts/install.sh)
```

Скрипт запросит:

- Домен или IP-адрес
- Использовать ли Let's Encrypt (acme.sh) — работает для **доменов** и **голых IP**
- Порт панели (по умолчанию случайный, 10000–65535)
- Префикс пути (случайный hex для обфускации)
- Имя и пароль администратора (авто-генерация или свои)

После установки доступна CLI-команда `shilka`:

```bash
shilka           # интерактивное меню (start/stop/restart/reset-password/backup/...)
systemctl status shilka   # статус сервиса
journalctl -u shilka -f   # логи
```

---

### Сборка из исходников

Требования: **Go 1.26**, **Node 20+**, **pnpm**.

```bash
# 1. Сборка фронтенда
cd frontend
pnpm install && pnpm build

# 2. Встраивание фронтенда в Go-бинарник
cd ..
rsync -a frontend/dist/ cmd/frontend/dist/

# 3. Сборка единого бинарника
go build -ldflags="-s -w" -o shilka ./cmd/

# 4. Запуск (dev-конфигурация)
SHILKA_CONFIG_PATH=config/dev.yaml ./shilka
```

Кросс-компиляция (например, сборка для VPS с macOS):

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o shilka-linux-amd64 ./cmd/
```

Готовый бинарник полностью самодостаточен — включает SPA-фронтенд, SQLite-драйвер и миграции.

---

### Конфигурация

Панель настраивается через **YAML-конфиг** (`config/dev.yaml` или `config/prod.yaml`) с возможностью переопределения через переменные окружения. Шаблон для продакшена — `config/prod.yaml`.

Ключевые секции:

```yaml
auth:
  jwt_secret: "" # обязательно в production
  admin_user: "admin"
  admin_password: "" # обязательно в production

http:
  address: ":443" # адрес прослушивания; ":" для всех интерфейсов

tls:
  mode: "file" # off | file | self_signed | acme

sing_box:
  binary_path: "/opt/shilka/bin/sing-box"
  config_path: "/etc/shilka/config.json"

subscription:
  public_url: "https://panel.example.com"
```

Переменные окружения (примеры):

```bash
export SHILKA_AUTH_JWT_SECRET="your-secret"
export SHILKA_AUTH_ADMIN_PASSWORD="your-password"
export SHILKA_SING_BOX_API_SECRET="your-clash-secret"
```

Первый администратор создаётся автоматически из этих значений. Конфиг sing-box генерируется из базы данных, проверяется через `sing-box check` и применяется атомарно.

---

### Разработка

```bash
# Бэкенд (API + Swagger)
go run ./cmd/
# http://127.0.0.1:8080
# Swagger: http://127.0.0.1:8080/swagger/

# Фронтенд (dev-сервер, проксирует /api на :8080)
cd frontend
pnpm install
pnpm dev
# http://127.0.0.1:3000

# Тесты
go test ./tests/...          # бэкенд (11 пакетов)
cd frontend && pnpm test     # фронтенд (8 файлов / 10 тестов)

# Линтинг
cd frontend && pnpm typecheck
go vet ./...
```

---

### Утилиты командной строки

После установки через [Быстрый старт](#быстрый-старт-vps) или для бинарника с настроенными путями:

```
shilka run                   Запуск сервера (используется systemd)
shilka admin reset-password  Сброс пароля администратора
shilka setting -port PORT    Смена порта панели
shilka setting -domain DOM   Установка домена и публичного URL
shilka api-token create      Создание API-токена для нод
shilka cert set-files        Установка кастомных TLS-сертификатов
shilka cert reset            Отключение TLS панели
shilka core reload           Перезагрузка конфига sing-box
```

---

### Поддерживаемые протоколы

| Протокол      | TLS-режимы         | Аутентификация      | Транспорт     |
| ------------- | ------------------ | ------------------- | ------------- |
| **VLESS**     | Reality, TLS, None | UUID                | TCP, WS, gRPC |
| **Hysteria2** | TLS (обязателен)   | Password            | —             |
| **Naive**     | TLS (обязателен)   | Username + Password | —             |

---

### Безопасность

- Все API управления sing-box привязаны к `127.0.0.1` — никогда не выставляются наружу
- Пароли хешируются Argon2id (64 MiB память, 3 итерации)
- TOTP 2FA с 8 одноразовыми кодами восстановления (хешированы Argon2id)
- JWT с HS256, настраиваемый срок действия
- Лимит запросов на логин (5/мин) и общий API-лимит (100/с) по IP
- Тело запроса ограничено 16 КиБ
- TLS поддерживает 4 режима: off, file (готовые сертификаты), self_signed, acme (авто-сертификат + авто-продление)
