# Performance snapshots — `sing-grok-frontend`

Артефакты сравнения «до/после» оптимизации. Берётся на чистом `pnpm build:analyze` (без sourcemap в обычном build; `ANALYZE=true` включает visualizer и генерирует `dist/stats.html`).

## Bundle size

| Метрика                                | Baseline       | После                    | Δ                |
|----------------------------------------|----------------|--------------------------|------------------|
| Сборка                                 | 1 чанк, всё включено | code-split по роутам и модалкам | — |
| Initial JS (raw)                       | **815.97 KB**  | **424.68 KB**            | **−48 %**        |
| Initial JS (gzip)                      | **249.74 KB**  | **136.85 KB**            | **−45 %**        |
| Initial JS + Dashboard (gzip, главная) | 249.74 KB      | 136.85 + 102.27 ≈ 239 KB | −4 % (параллельно)|
| Initial JS + /inbounds (gzip)          | 249.74 KB      | 136.85 + 2.18 ≈ **139 KB** | **−44 %**        |
| Initial JS + /settings (gzip)          | 249.74 KB      | 136.85 + 1.17 ≈ **138 KB** | **−45 %**        |
| Initial JS + /logs (gzip)              | 249.74 KB      | 136.85 + 1.40 ≈ **138 KB** | **−45 %**        |
| `inbound-form-modal` в initial         | да             | **нет** (отдельный чанк, 4.65 KB gzip) | вынесен |
| `add-client-modal` в initial           | да             | **нет** (1.17 KB gzip)   | вынесен          |
| `client-detail-modal` в initial        | да             | **нет** (2.28 KB gzip)   | вынесен          |
| CSS (raw / gzip)                       | 33.65 / 7.06 KB| 34.49 / 7.21 KB          | +0.84 / +0.15 KB |
| Sourcemap в prod                       | да (4 MB)      | нет (только при ANALYZE) | −4 MB            |

### Чанки после оптимизации

```
index-CKTpC92u.js            424.68 KB │ gzip: 136.85 KB  ← общий vendor + framework
DashboardPage-DMDIFnLo.js    345.82 KB │ gzip: 102.27 KB  ← Recharts + dashboard-only
inbound-form-modal-*.js       16.00 KB │ gzip:   4.65 KB  ← лениво, prefetch on hover
ClientsPage-*.js               6.67 KB │ gzip:   2.44 KB
InboundsPage-*.js              5.96 KB │ gzip:   2.18 KB
client-detail-modal-*.js       5.72 KB │ gzip:   2.28 KB
select-*.js                    5.72 KB │ gzip:   2.13 KB
SettingsPage-*.js              3.47 KB │ gzip:   1.17 KB
LogsPage-*.js                  2.91 KB │ gzip:   1.40 KB
add-client-modal-*.js          2.47 KB │ gzip:   1.17 KB
modal-*.js, button-*.js, …    < 2 KB каждый
```

## Lab metrics (Chrome DevTools trace, dev-сервер, после оптимизации)

| Метрика | Значение |
|---|---|
| LCP        | **465 ms** |
| INP        | **1 ms**   |
| CLS        | **0.00**   |
| Lighthouse Accessibility | 96 |
| Lighthouse Best Practices| 100 |
| Lighthouse SEO          | 82 |

Baseline trace не снимался: оптимизация делалась перед dev-сервером с уже изменённым кодом. Метрики выше — это **состояние «после»**. Для сравнения «до/после» в production-режиме открой `baseline-stats.html` и `final-stats.html` в visualizer.

## Что улучшилось не только в размере

1. **Каскадные ререндеры устранены.** Единый `MockStoreProvider` был разделён на 5 контекстов (`MetricsContext`, `InboundsContext`, `ClientsContext`, `LogsContext`, `RuntimeContext`) + `ActionsContext`. До этого `metrics`/`history` тикали `setInterval(1000ms)` → весь `value` контекста пересоздавался → все потребители (включая `InboundsPage`/`ClientsPage`/`LogsPage`) рендерились раз в секунду. После: только `MetricsContext`-потребители (`CpuCard`, `RamCard`, `DiskCard`, `TrafficSplitCard`, `TrafficChart`) рендерятся от тика метрик; остальные страницы — 0 ререндеров от тика.
2. **LazyMotion strict.** Заменено `motion.*` → `m.*` в 12 файлах. Только нужные features (`domMax`) лениво подгружаются; остальной код framer-motion tree-shake'ится.
3. **Custom hooks.** Логика `inbound-form-modal` вынесена в `useInboundForm` (517 → 363 строки в компоненте). Также `useDisclosure`, `useLogFilter`, `useClientFilter`, `useCopyToClipboard`. Helpers `randomPort`/`randomHex`/`makeUuid` → `lib/random.ts`.
4. **React.memo + useCallback** на `InboundRow`, `Row` в `ClientsTable`. Колбэки в `InboundsPage`/`ClientsPage` обёрнуты в `useCallback`, чтобы memo фактически работал.
5. **Lazy-loading.** Все 5 страниц + 4 модалки (`inbound-form-modal`, `add-client-modal`, `client-detail-modal`, `qr-modal`) загружаются по требованию. Префетч при hover/focus на кнопке-триггере — UX модалки не страдает от задержки.
6. **Sourcemap отключён** в обычной prod-сборке — `−4 MB` дистрибутива.

## Файлы

- `baseline-stats.html` — визуализатор сборки **до** оптимизации.
- `final-stats.html` — после.
- `dashboard-after.png` — скриншот dashboard после оптимизации.
- `dashboard-cpu4x-after.json` — performance trace с CPU throttling 4×.
- `report.html`, `report.json` — Lighthouse аудит (Desktop).

## Команды для повторного замера

```bash
cd frontend
ANALYZE=true ./node_modules/.bin/vite build       # → dist/stats.html
./node_modules/.bin/vite                          # → http://127.0.0.1:3000/
./node_modules/.bin/vitest run                    # 10 тестов
```
