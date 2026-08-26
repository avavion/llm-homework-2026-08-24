# Реестр регуляторных правил дат

## Назначение и безопасное значение по умолчанию

Этот документ — проверяемый источник данных для будущей конфигурации статусов,
уведомлений и допуска продуктов в рецепты. Он не является юридической или
медицинской консультацией и не определяет фактическую пригодность конкретного
продукта.

Автоматизация разрешается только для строки со статусом `enabled`, когда
первичный официальный источник однозначно подтверждает все поля строки:
`regulator_group`, ISO-коды стран, `date_type`, `expiry_timezone_source`,
`expiry_instant_rule`, `post_expiry_status` и `recipe_eligibility`. Если хотя бы
одно поле не подтверждено, строка получает `research_required`: автоматический
статус, исключение из рецептов и расписание уведомлений для неё отключены.

В текущей версии строк `enabled` нет. В частности, дата без времени не
преобразуется в 00:00 или 23:59:59 без прямого регуляторного основания. Минимум
пользовательского порога в 60 минут из
[политики групп продуктов](product-group-alert-policy.md) не обходит этот
evidence gate.

## Реестр

ISO-коды ниже — двухбуквенные `ISO 3166-1 alpha-2`. Значение `none` означает
отсутствие исполняемого правила, а не подразумеваемое поведение.

| regulator_group | ISO country codes | date_type | expiry_timezone_source | expiry_instant_rule | post_expiry_status | recipe_eligibility | research_status | source URL | accessed_on |
|---|---|---|---|---|---|---|---|---|---|
| `eu_1169_2011` | `AT, BE, BG, HR, CY, CZ, DK, EE, FI, FR, DE, GR, HU, IE, IT, LV, LT, LU, MT, NL, PL, PT, RO, SK, SI, ES, SE` | `use_by` — дата безопасности; семантика подтверждена статьёй 24 | `none` — регламент не задаёт единую часовую зону для напечатанной даты | `none` — приложение X задаёт календарное представление, но не универсальный instant; 00:00/23:59:59 не выводятся | `none` — статья 24 считает продукт небезопасным после даты, но приложение не вычисляет момент перехода автоматически | `none` — автоматическое исключение из рецептов выключено до подтверждения instant/timezone | `research_required` | [Regulation (EU) No 1169/2011, Article 24 and Annex X](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32011R1169); [European Commission date-marking guidance](https://food.ec.europa.eu/food-safety/food-waste/eu-actions-against-food-waste/date-marking-and-food-waste-prevention_en); [EU countries](https://european-union.europa.eu/principles-countries-history/eu-countries_en); [ISO 3166](https://www.iso.org/iso-3166-country-codes.html) | `2026-08-26` |
| `eu_1169_2011` | `AT, BE, BG, HR, CY, CZ, DK, EE, FI, FR, DE, GR, HU, IE, IT, LV, LT, LU, MT, NL, PL, PT, RO, SK, SI, ES, SE` | `best_before` — минимальная долговечность и качество, а не автоматический вывод о безопасности | `none` — регламент не задаёт единую часовую зону | `none` — приложение X допускает day/month, month/year или year; единого instant нет | `none` — можно показать исходную маркировку, но автоматический статус после даты выключен | `none` — автоматическое включение или исключение из рецептов выключено; решение не выводится только из даты | `research_required` | [Regulation (EU) No 1169/2011, Article 24 and Annex X](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32011R1169); [European Commission date-marking guidance](https://food.ec.europa.eu/food-safety/food-waste/eu-actions-against-food-waste/date-marking-and-food-waste-prevention_en); [EU countries](https://european-union.europa.eu/principles-countries-history/eu-countries_en); [ISO 3166](https://www.iso.org/iso-3166-country-codes.html) | `2026-08-26` |
| `eaeu_tr_ts_022_2011` | `AM, BY, KZ, KG, RU` | `shelf_life_lt_72h` — «годен до» с часом, числом и месяцем; не приравнивается автоматически к `use_by` или `best_before` | `none` — на маркировке есть час, но групповой источник не задаёт часовую зону | `none` — без timezone точный instant не вычисляется | `none` — автоматический статус выключен | `none` — автоматический допуск или исключение из рецептов выключены | `research_required` | [EAEU member states](https://eaeunion.org/?lang=ru); [EEC page for TR CU 022/2011](https://eec.eaeunion.org/comission/department/deptexreg/tr/PischevkaMarkirovka.php); [official TR CU 022/2011 text, section 4.7](https://eec.eaeunion.org/upload/medialibrary/9db/TrTsPishevkaMarkirovka.pdf); [ISO 3166](https://www.iso.org/iso-3166-country-codes.html) | `2026-08-26` |
| `eaeu_tr_ts_022_2011` | `AM, BY, KZ, KG, RU` | `shelf_life_ge_72h` — «годен до», «годен до конца» или длительность; не приравнивается автоматически к `use_by` или `best_before` | `none` — единая часовая зона не установлена найденным групповым источником | `none` — для date-only/month-only/year-only маркировки нельзя выдумывать правило 00:00 | `none` — автоматический статус выключен | `none` — автоматический допуск или исключение из рецептов выключены | `research_required` | [EAEU member states](https://eaeunion.org/?lang=ru); [EEC page for TR CU 022/2011](https://eec.eaeunion.org/comission/department/deptexreg/tr/PischevkaMarkirovka.php); [official TR CU 022/2011 text, section 4.7](https://eec.eaeunion.org/upload/medialibrary/9db/TrTsPishevkaMarkirovka.pdf); [ISO 3166](https://www.iso.org/iso-3166-country-codes.html) | `2026-08-26` |
| `cis_national_unverified` | `AZ, MD, TJ, TM, UA, UZ` | `none` — единое межгосударственное правило типов дат не подтверждено | `none` | `none` | `none` — автоматический статус выключен | `none` — автоматический допуск или исключение из рецептов выключены | `research_required` | [CIS Executive Committee report for 2025](https://e-cis.info/news/564/134334/); [ISO 3166](https://www.iso.org/iso-3166-country-codes.html) | `2026-08-26` |

## Что именно подтверждают источники

- ЕС: статья 24 Regulation (EU) No 1169/2011 связывает `use by` с
  безопасностью и указывает, что после этой даты пищевой продукт считается
  небезопасным; приложение X задаёт формы `best before`, `best before end` и
  `use by`. Для модели приложения надписи `best before …` и
  `best before end …` — варианты представления одного канонического
  `date_type = best_before`, а не два значения enum. Разъяснение Еврокомиссии
  также различает безопасность (`use by`) и качество (`best before`). Ни один
  из этих источников не задаёт единую часовую зону или универсальный instant
  для календарной даты.
- ЕАЭС: раздел 4.7 ТР ТС 022/2011 требует час для срока менее 72 часов и
  допускает календарные либо длительные формы для больших сроков. Источник не
  подтверждает эквивалентность этих форм типам ЕС, единую часовую зону или
  универсальное правило 00:00.
- Остальная область СНГ: отчёт Исполнительного комитета документирует участие
  перечисленных стран в органах СНГ, но не является общим регуляторным актом о
  маркировке пищевых дат. Поэтому страны ЕАЭС исключены из этой строки, а для
  `AZ, MD, TJ, TM, UA, UZ` требуется отдельное исследование национальных
  первичных актов. Грузия не включена, потому что использованный источник её в
  этой области охвата не перечисляет.

## Условия перевода строки в `enabled`

Для каждой строки исследователь должен приложить прямой актуальный первичный
источник и подтвердить одновременно:

1. применимость правила ко всем перечисленным ISO-кодам и конкретному
   `date_type`;
2. источник часовой зоны и точный `expiry_instant_rule`, включая неполные даты;
3. регуляторно допустимый `post_expiry_status` и соответствующий
   `recipe_eligibility` без медицинского вывода приложения;
4. отсутствие противоречащего национального или продуктового исключения;
5. URL и новую дату доступа, после чего изменение проходит юридическое и QA
   ревью.

До выполнения всех пяти условий система может хранить и показывать введённую
пользователем маркировку, но не запускает основанную на ней автоматизацию.
