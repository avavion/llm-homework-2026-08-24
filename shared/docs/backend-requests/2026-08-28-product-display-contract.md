# Запрос к Backend: display semantics продукта для Frontend

## Контекст

Frontend не может безопасно выводить `attention`, `expired` или
`research_required` из даты в браузере: момент истечения и пригодность
рецепта зависят от подтверждённого регуляторного правила.

## Текущее расхождение

`GET /v1/products` и `GET /v1/products/{id}` возвращают `status`, равный
только lifecycle (`active`, `used`, `discarded`). В UI уже предусмотрены
display states `attention`, `expired` и `research_required`, но текущий
удалённый DTO их не содержит.

## Запрашиваемое дополнение (обратно совместимое)

Добавьте в оба product response поле `display_status` со значениями
`active | attention | expired | used | discarded | research_required`.

- Backend определяет значение через действующий реестр regulation rules,
  дату, date type и lifecycle.
- `used` и `discarded` имеют приоритет над display-статусом даты.
- Для `research_required` backend не делает медицинский или юридический вывод
  и не обещает автоматическое напоминание.
- Существующее `status` lifecycle-поле и все текущие request payload остаются
  без изменений.

## Примеры

```json
{ "id": "…", "date_type": "use_by", "status": "active", "display_status": "expired" }
{ "id": "…", "date_type": "best_before", "status": "active", "display_status": "attention" }
{ "id": "…", "date_type": "use_by", "status": "active", "display_status": "research_required" }
```

## Критерий готовности

Опубликованы JSON-примеры и contract/integration tests для list и get;
401/403/404 не раскрывают данные другого аккаунта.
