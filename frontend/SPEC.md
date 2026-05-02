# frontend (Cadence) — техническое задание

## Назначение

Single-page Web-приложение **Cadence** — институциональный UI поверх beev.
Закрывает все user-facing сценарии: регистрация / вход, CRUD вакансий,
загрузка резюме, просмотр AI-аналитики кандидата, скачивание оригинала
резюме. Ходит исключительно через **gateway** на `:8080`; знаний о
конкретных backend-сервисах не имеет.

Дизайн-система — Coinbase-style (см. `DESIGN.md` в корне репо): белый
canvas, единственный action-цвет `#0052ff`, weight-400 заголовки,
pill-rounded CTAs, 96px section rhythm. Никаких вторичных бренд-цветов.

## Стек

- **React 19** + **TypeScript 6** (strict)
- **Vite 8** — bundler + dev server
- **Tailwind CSS v4** — `@theme` через CSS variables, без `tailwind.config.js`
- **react-router-dom v7** — flat routing
- **zod** — runtime-валидация форм
- **@fontsource-variable/jetbrains-mono** — self-hosted JetBrains Mono
- **General Sans** — через Fontshare CDN (`<link>` в index.html)

Отсутствуют намеренно: Redux, Zustand, react-query, react-hook-form,
HeroUI, shadcn — стандартный React state + reflective use-cases
покрывают весь UI scope MVP.

## Архитектура

Clean architecture, frontend-вариант (зеркальная backend-у):

```
src/
├── domain/                       чистые типы + zod-валидация
│   ├── auth/                     User, Session, credentialsSchema
│   ├── vacancy/                  Vacancy, SkillWeight, KNOWN_ROLES,
│   │                             KNOWN_ROLE_VALUES, vacancyFormSchema
│   ├── resume/                   Candidate, Resume, MAX_RESUME_BYTES,
│   │                             ACCEPTED_RESUME_TYPES
│   └── analysis/                 Analysis, AIDecision, ScoreBreakdown,
│                                 ANALYSIS_STATUS_BY_CODE, SortOrder
├── application/                  ports + use-cases
│   ├── auth/ports.ts             interface AuthGateway
│   ├── auth/useCases.ts          loginUseCase / registerUseCase / logoutUseCase
│   ├── vacancy/ports.ts          interface VacancyGateway
│   ├── resume/ports.ts           interface ResumeGateway + ResumeFile
│   └── analysis/ports.ts         interface AnalysisGateway
├── infrastructure/               адаптеры портов
│   ├── http/
│   │   ├── client.ts             fetch wrapper, single-flight refresh,
│   │   │                         FormData detection, AbortController
│   │   └── errors.ts             ApiError + grpc-status mapping
│   ├── auth/AuthHttpGateway.ts
│   ├── vacancy/VacancyHttpGateway.ts
│   ├── resume/ResumeHttpGateway.ts (chunked btoa() для base64)
│   ├── analysis/AnalysisHttpGateway.ts
│   └── storage/tokenStorage.ts   localStorage adapter (cadence:session)
├── presentation/
│   ├── ui/                       переиспользуемый kit (24 компонента)
│   ├── features/                 feature-specific compositions
│   ├── pages/                    AuthPage / VacanciesPage /
│   │                             VacancyCreatePage / VacancyDetailsPage
│   └── layouts/AppLayout.tsx     SkipLink → TopNav → main → Footer
├── app/
│   ├── App.tsx                   composition root
│   ├── router.tsx                BrowserRouter + Protected/PublicOnly
│   └── providers/                ThemeProvider / I18nProvider /
│                                 AuthProvider / GatewaysProvider
└── shared/
    ├── lib/cn.ts                 classnames helper
    ├── hooks/useDebouncedValue.ts
    └── i18n/dictionaries.ts      ru + en flat dicts + pluralKey()
```

### Правила слоёв

- `domain` ничего не импортирует
- `application` импортирует только `domain`
- `infrastructure` реализует `application` ports, может импортить
  `domain`
- `presentation` использует use-cases через DI (контексты)
- HTTP **никогда** не вызывается из UI напрямую

## UI kit

24 компонента в `presentation/ui/`. Каждый покрывает одну DESIGN-сущность:

| Компонент | DESIGN.md mapping |
|---|---|
| `Button` × 6 variants | button-primary / -cta / -secondary-light/dark / -outline-on-dark / -tertiary-text |
| `TextInput` | text-input (rounded-md, focus → 2px primary border) |
| `TextArea` | multiline вариант TextInput |
| `SearchInput` | search-input-pill |
| `BadgePill` × 4 tones | badge-pill (default / inverse / up / down) |
| `Card` × 5 variants | feature / product-light / product-dark / pricing / pricing-featured |
| `PillSwitcher` | сегментированный pill (Auth login/register, фильтры) |
| `LanguageSwitcher` | dropdown с флагами RU + US |
| `ThemeToggle` | pill 88×44 с slide-thumb |
| `Wordmark` | BrandMark + "Cadence" типографика |
| `BrandMark` | 3 ascending bars в primary square (24×24 SVG) |
| `Flag` | FlagRu (3 полосы) + FlagUs (упрощ. stars-and-stripes) |
| `TopNav` + `TopNavLink` | top-nav-light / top-nav-on-dark |
| `HeroBand` | hero-band-light / hero-band-dark |
| `CtaBandDark` | cta-band-dark |
| `AssetIconCircular` | asset-icon-circular |
| `PriceCell` × 3 tones | price-up-cell / price-down-cell / neutral |
| `Spinner` | минимальный, для loading |
| `Footer` | footer-light с 3-column grid |
| `SkipLink` | a11y skip-to-main-content |
| `Checkbox` | DESIGN-styled native checkbox |
| `WeightStepper` | `[−] 0.05 [+]` для skill weights |

Все компоненты:
- `forwardRef` где нужно
- Никаких inline-hex — только токены через Tailwind утилиты
- focus-visible 2px primary
- `cursor-pointer` на интерактиве
- aria-label / aria-pressed / aria-busy / aria-invalid

## Pages

### `/auth`
Login + register на одной странице, переключение через `PillSwitcher`.
Полноэкранный layout с `feature-card` справа, editorial copy слева,
decorative product-mockup stack под текстом. Полная инверсия в dark theme.

### `/vacancies`
- HeroBand light с CTA "Новая вакансия"
- `SearchInput` + role-filter chips (7 ролей + "Все")
- Soft-gray band с grid 1/2/3-up из `VacancyCard`
- Empty / error / loading состояния отдельные

### `/vacancies/new`
- `VacancyForm` в `feature-card`
- TextInput title (max 255), TextArea description (max 4000) с counter
- `SkillsEditor` — список рядов с
  именем + `WeightStepper` (0..1, кастомные стрелки) + Must/Nice
  чекбоксами + trash-icon delete
- На submit → `POST /api/v1/vacancies` → `navigate(/vacancies/:id)`

### `/vacancies/:id`
- HeroBand с title / role / status / skill-summary chips
- `ResumeUploader` — drag&drop + Choose file → POST + auto-trigger
  analysis в одном click flow
- 7/5 grid: список кандидатов слева, `AnalysisDetails` панель справа
- `AnalysisDetails` показывает score 44px mono, HR recommendation
  badge, RU rationale, skills breakdown (matched / missing / extra),
  candidate profile (years с правильным склонением через `pluralKey`),
  candidate feedback. Действия: **Скачать резюме** + **Удалить кандидата**
  (с inline-confirm)

## i18n

`shared/i18n/dictionaries.ts` — плоский словарь с ~200 ключами для
ru и en.

- Default: `ru`
- Persist: `localStorage["cadence:locale"]`
- HTML `<html lang>` синхронизируется при смене
- `pluralKey(baseKey, n, locale)` — RU plural-формы (1 / 2-4 / 5+):
  `vacancies.countOne` / `countFew` / `countMany`,
  `analysis.yearsOne` / `yearsFew` / `yearsMany`
- Variable interpolation через `{name}` плейсхолдеры

## Theming

- CSS variables в `:root` (light) и `[data-theme="dark"]` (override)
- `ThemeProvider` управляет атрибутом на `<html>` + persist в
  `cadence:theme`
- Tailwind v4 `@theme` экспонирует токены как утилиты
  (`bg-canvas`, `text-ink`, `rounded-pill` и т.д.)
- Pre-paint inline-script в `index.html` читает `localStorage` до
  первого кадра (no FOUC)
- Primary `#0052ff` остаётся в обеих темах per DESIGN.md

## HTTP layer

`infrastructure/http/client.ts`:

- `baseURL` через `import.meta.env.VITE_API_URL` (пустой → Vite proxy
  `/api` → `:8080`)
- Авто-Bearer из `SessionHolder.getAccessToken()`
- На `401` → single-flight `holder.refresh()` → один retry с новым
  токеном; concurrent 401-ы коалесцируются в один `/auth/refresh`
- `noRetry` флаг чтобы `/auth/refresh` сам не уходил в рекурсию
- Network errors → `ApiError{status:0, reason:'NETWORK'}`
- `FormData` определяется автоматически (для будущих multipart-загрузок)
- Все JSON-ошибки маппятся в `ApiError` с типизированным `reason`:
  `UNAUTHORIZED` / `INVALID_ARGUMENT` / `CONFLICT` / `RATE_LIMITED` /
  `NOT_FOUND` / `FORBIDDEN` / `NETWORK` / `UNKNOWN`

`ApiError` парсит `google.rpc.Status` → достаёт
`errdetails.ErrorInfo.reason/domain` (то, что бэкенд beev уже
возвращает) → UI получает чистый объект без знания о gRPC.

## Запуск и сборка

```bash
yarn install
yarn dev          # Vite dev server, http://localhost:5173
yarn build        # tsc -b && vite build → dist/
yarn lint         # eslint
```

`vite.config.ts` поднимает proxy на `/api → http://localhost:8080`.
Для прода `VITE_API_URL` указывает на gateway URL.

## Тестирование

Юнит-тесты не настроены (vitest как nice-to-have). Build-time
checking покрывает большую часть логики:
- TypeScript strict
- ESLint с react-hooks + react-refresh плагинами
- zod валидация на runtime в формах
- Manual e2e через `yarn dev`

## Bundle

- ~115 kB gzip (на момент написания SPEC) — приемлемо для
  институционального UI
- Возможные оптимизации: lazy-load `/vacancies/new` (zod chunked),
  code-split `AnalysisDetails` (не критичен на первой загрузке)
- Self-hosted JetBrains Mono (variable, ~40 kB cyrillic)
- General Sans через Fontshare CDN (один HTTPS-fetch, кэшируется)

## A11y

- `<SkipLink>` в каждом layout — keyboard skip past nav
- `focus-visible 2px primary` на всех интерактивных
- `aria-label` на icon-only buttons (theme, locale, trash, download)
- Form labels через `htmlFor`
- `role="alert"` на ошибках, `role="alertdialog"` на confirmation
- `aria-pressed` на toggle-кнопках, `aria-selected` где уместно
- Color contrast: ink `#0a0b0d` на canvas `#ffffff` = 19:1 (AAA)
- `prefers-reduced-motion` уважается в `index.css`

## Известные ограничения

- Нет восстановления пароля (нет endpoint'а на backend)
- Нет edit вакансии (PATCH endpoint есть, страница нет — TODO)
- Нет pagination на VacanciesPage (PAGE_SIZE=24 захардкожено)
- Нет real-time поллинга статуса анализа (one-shot fetch)
- Нет mobile hamburger в TopNav — одна nav-ссылка, не критично
- analyses не очищаются при удалении кандидата (orphan rows в analysis-БД)
