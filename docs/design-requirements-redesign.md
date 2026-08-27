# Design Requirements — Premium Redesign

## Цель

Сделать учёт продуктов спокойным, точным и предсказуемым: пользователь за один взгляд видит критичные продукты, понимает различие типов срока и совершает действие без визуального шума. Направление **Calm Ledger** сочетает ясность профессиональных финтех-интерфейсов с тёплым смыслом домашнего запаса.

## Принципы

- **Сначала действие:** срок, статус и CTA находятся в первой зоне карточки.
- **Данные не декор:** даты и количество выровнены и набраны tabular figures.
- **Один акцент:** mint обозначает основное безопасное действие; warning/danger не конкурируют с ним.
- **Безопасность понятна:** `use by` и `best before` всегда различаются текстом, иконкой, цветом и последствием для рецептов.
- **Mobile-first:** ни одно действие не зависит от hover; touch target минимум 44×44px.

## Before / After

| Область | До | После | Пользовательская польза |
| --- | --- | --- | --- |
| Иерархия | Одинаковые карточки и ссылки | Три уровня: срочность → инвентарь → метаданные | Быстрее понятно, что использовать первым |
| Навигация | Набор текстовых ссылок | Desktop rail / mobile bottom bar с активным индикатором | Меньше когнитивной нагрузки |
| Статусы | Технический badge рядом с данными | Цветной сигнал + иконка + plain-language объяснение | Меньше риска ошибиться со сроком |
| Формы | Равнозначные поля | Чёткая primary-группа, optional details и inline feedback | Быстрее ручное добавление |
| Обратная связь | Мало различий между состояниями | Skeleton, inline error, toast, disabled action | Действия ощущаются надёжными |

## Дизайн-токены

### Цвет

| Токен | Значение | Роль |
| --- | --- | --- |
| `--ink-950` | `#08111F` | App canvas, desktop rail |
| `--ink-900` | `#101C2E` | Elevated dark surface |
| `--surface` | `#F7F9FC` | Основной светлый canvas |
| `--card` | `#FFFFFF` | Карточки и формы |
| `--text-primary` | `#172033` | Основной текст |
| `--text-secondary` | `#5B667A` | Метаданные |
| `--mint-500` | `#20C997` | Primary CTA, success affordance |
| `--mint-700` | `#087F5B` | Hover/active text on light surface |
| `--warning-600` | `#B54708` | Attention |
| `--danger-600` | `#D92D20` | Expired/destructive |
| `--info-600` | `#2563EB` | Draft/research_required |
| `--focus` | `#6D5DFB` | Focus ring |

Контраст текста — минимум 4.5:1; `mint-500` не используется для мелкого текста на светлом фоне. Dark rail — навигационный слой, а не переключатель темы: тёмная тема продукта остаётся отдельным scope.

### Типографика и пространство

| Роль | Значение |
| --- | --- |
| Display | `Onest`, 32/40, 700; mobile 28/36 |
| Body | `Onest`, 16/24, 400–500 |
| Data | `IBM Plex Mono`, 13/18, 500, tabular-nums |
| Scale | 4, 8, 12, 16, 24, 32, 48px |
| Radius | 10px control, 16px card, 20px dialog |
| Shadow | `0 12px 32px rgba(8,17,31,.10)` только у dialog/raised cards |

## Компоновка

```text
desktop ≥1024: [ dark rail 248 ] [ main inventory 8 cols ] [ urgency rail 4 cols ]
tablet 640–1023: [ top nav ] [ urgency ] [ inventory ]
mobile <640: [ urgency ] [ inventory ] [ fixed bottom nav ]
```

Контент ограничен 1240px; desktop gap 24px, tablet/mobile 16px. Фильтры на mobile открываются как bottom sheet; modal на mobile становится sheet с `Escape` и focus return.

## Компоненты

| Компонент | Default | Hover/active | Disabled/loading | Error/special |
| --- | --- | --- | --- | --- |
| Primary button | Mint fill, dark text | `mint-700`, 150ms | 60% opacity, spinner, no repeat click | Inline error остаётся рядом с действием |
| Secondary button | Transparent, border | ink-950 tint | muted border | Не используется для destructive action |
| Product card | White, 16px radius, 1px border | subtle raise 2px | skeleton replaces content | attention/expired имеет left status rail и текст |
| Input/select | White, 44px min height, label 14px | focus ring `--focus` | disabled calm surface | `aria-invalid`, inline error, error icon |
| Status badge | Text + icon + semantic bg | no hover required | n/a | `research_required` info tone |
| Toast | Compact elevated card, bottom-right | pause on hover | auto-dismiss 5s | Error lives 8s and has close button |
| Dialog | 20px radius, max 480px | n/a | CTA locks during mutation | Focus trap, visible title and consequence |

## Экранные требования

- **Инвентарь:** крупный title, count и primary CTA; urgency rail показывает максимум 3 продукта. В строке: название, location, date type, absolute/relative date, badge, lifecycle actions.
- **Форма:** required поля видимы сразу; optional fields в раскрываемом «Детали хранения». Под датой — explainer. Успех ведёт к обновлённому списку, ошибка не теряет данные.
- **Фото-черновик:** нейтральный upload card → progress → редактируемый draft. До approve — явный label «Черновик»; reject безопасно возвращает к ручному пути.
- **Рецепты:** карточка всегда объясняет «Использует [продукт] — [срок]»; expired `use by` не отображается как причина.
- **Настройки:** country/profile и e-mail threshold сгруппированы; `research_required` показывает information alert без медицинского вывода.

## Состояния

| Состояние | Требование |
| --- | --- |
| Loading | Skeleton повторяет layout, `aria-busy=true`; не пустой экран |
| Empty | Одна причина и одно CTA: «Добавить продукт» или «Сбросить фильтры» |
| Network/API error | Plain-language alert + retry; без технического текста |
| Mutation | CTA disabled, action-copy «Сохраняем…»; повторная отправка невозможна |
| Success | Toast и обновлённые данные/focus target |
| Forbidden/not found | Нейтральный экран без данных приватного объекта |

## Accessibility

- `<header>`, `<nav>`, `<main>`, skip link и один `<h1>` на экран.
- Native controls first; `aria-describedby` связывает error с полем; dialog использует `aria-modal` и focus trap.
- Tab follows visual order; Enter/Space activates controls; Escape closes dialog/sheet.
- `prefers-reduced-motion` отключает transitions. Статус никогда не передаётся одним цветом.

## Implementation acceptance

- [ ] Токены используются вместо локальных hex/spacing значений.
- [ ] Product list, form, status, toast и dialog имеют все состояния из таблицы.
- [ ] 320px, 768px и 1440px без горизонтального overflow.
- [ ] Keyboard smoke test и WCAG AA contrast check выполнены.
- [ ] Любой визуальный приём обслуживает CJM: добавить продукт, понять срок, завершить продукт, проверить черновик.
