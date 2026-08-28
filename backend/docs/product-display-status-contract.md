# Product display status contract

`GET /v1/products` and `GET /v1/products/{id}` return the existing
`status` lifecycle field and an additional `display_status` field. Existing
request bodies and response fields are unchanged.

## Values

`display_status` is one of `active`, `attention`, `expired`, `used`,
`discarded`, or `research_required`.

- `used` and `discarded` always take precedence over a date-derived value.
- `active`, `attention`, and `expired` are returned only when the regulation
  registry has a confirmed rule that permits the calculation.
- A missing country, unlisted country/date type, `research_required` rule, or
  an invalid rule produces `research_required`. This is not an assertion about
  food safety or a promise to schedule a reminder.

The current registry has no `enabled` rows, so current active products return
`research_required` until a reviewed rule is added.

## Examples

```json
{
  "id": "ca0e68e6-2166-4bc4-8119-b3971444db19",
  "name": "Milk",
  "date_type": "use_by",
  "expiry_date": "2026-03-01T00:00:00Z",
  "status": "active",
  "display_status": "research_required"
}
```

With a confirmed rule, a list or item may return:

```json
{ "date_type": "use_by", "status": "active", "display_status": "expired" }
{ "date_type": "best_before", "status": "active", "display_status": "attention" }
{ "date_type": "use_by", "status": "used", "display_status": "used" }
```

Both endpoints require a valid session. Requests without one receive `401`.
An item belonging to another account is returned as `404`; its fields are not
disclosed. There is no separate `403` path for this resource.
