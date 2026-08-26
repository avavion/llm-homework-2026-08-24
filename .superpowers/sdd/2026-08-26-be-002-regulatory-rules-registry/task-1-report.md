# BE-002 Task 1 report — regulatory rules registry

## Status

`complete` для документного объёма плана. Созданы только разрешённые продуктовые
документы и обязательные отчёты; backend code и план не изменялись.

## Deliverables

- `shared/docs/regulatory-date-rules.md`
- `shared/docs/product-group-alert-policy.md`
- `sessions/session-2026-08-26-151603.md`

## Decisions

- В реестре нет строк `enabled`: найденные групповые источники не подтверждают
  одновременно timezone и универсальный `expiry_instant_rule` для date-only
  маркировки.
- ЕС разделён на `use_by` и `best_before`, но автоматический статус, расписание
  и recipe gating выключены до подтверждения exact instant/timezone.
- Для ЕАЭС сохранены только подтверждённые формы ТР ТС 022/2011; соответствие
  типам ЕС и правило 00:00 не выведены.
- Для `AZ, MD, TJ, TM, UA, UZ` создана строка `cis_national_unverified` со
  статусом `research_required`; общая автоматизация СНГ не предполагается.
- Пять групп предупреждений — UX defaults в минутах, с пользовательским
  минимумом 60 минут и запретом обхода `research_required`.

## Official citations

Все источники проверены 2026-08-26.

1. Regulation (EU) No 1169/2011, Article 24 and Annex X:
   https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32011R1169
2. European Commission, Date marking and food waste prevention:
   https://food.ec.europa.eu/food-safety/food-waste/eu-actions-against-food-waste/date-marking-and-food-waste-prevention_en
3. Official EU country list:
   https://european-union.europa.eu/principles-countries-history/eu-countries_en
4. EAEU member states:
   https://eaeunion.org/?lang=ru
5. Eurasian Economic Commission, TR CU 022/2011 landing page:
   https://eec.eaeunion.org/comission/department/deptexreg/tr/PischevkaMarkirovka.php
6. Official TR CU 022/2011 text, section 4.7:
   https://eec.eaeunion.org/upload/medialibrary/9db/TrTsPishevkaMarkirovka.pdf
7. CIS Executive Committee report for 2025, used only for scope evidence:
   https://e-cis.info/news/564/134334/
8. ISO 3166 country-code authority:
   https://www.iso.org/iso-3166-country-codes.html

## Checks

- Plan registry `rg`: passed; all required field names, source URLs and
  `research_required` markers are present.
- Plan policy `rg`: passed; `60 минут`, `research_required` and automation gate
  are explicit.
- Semantic shell checks: passed; exactly five registry rows, all five are
  `research_required`, no enabled row, and exactly five required product groups.
- External URL check: seven sources returned HTTP `200` or `202`; ISO returned
  `403` to automated `curl` but was accessible and inspected through web search.
- Final staged `git diff --check` and session validator are run immediately
  before commit; their fresh result is reported in the handoff.
- Code tests: not run because the change is documentation-only and the plan
  defines document-specific review commands.

## Concerns and follow-up

- `research_required` is intentional, not a missing implementation: enabling
  date-derived actions without timezone/instant evidence would invent a legal
  and safety boundary.
- Before enabling EU automation, confirm exact timestamp interpretation,
  timezone source and any national/product-specific exceptions.
- Before enabling EAEU automation, confirm current consolidated amendments,
  timezone semantics and mapping of label phrases without importing EU meaning.
- Research national primary law separately for `AZ, MD, TJ, TM, UA, UZ`.
- Validate alert defaults with MVP behavior; they must remain convenience
  settings rather than food-safety advice.

## Commit

The commit SHA is generated after this report and is provided in the final
handoff to the parent agent.
