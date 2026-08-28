import { expect, test } from 'vitest'
import * as apiModule from './api'

type ProductPayloadAdapter = {
  toBackendProductPayload: (input: { name: string; dateType: 'use_by' | 'best_before'; expiryDate: string; group: string; location: string }) => unknown
}

type DraftAdapter = {
  fromBackendDraft: (draft: { id: string; status: 'pending'; fields: { Name: string; DateType: 'best_before'; ExpiryDate: string; ProductGroup: string; StorageLocation: string } }) => unknown
}

type ProductDisplayAdapter = {
  fromApiProduct: (item: {
    id: string
    name: string
    date_type: 'use_by'
    expiry_date: string
    status: 'active'
    display_status: 'research_required'
  }) => unknown
}

test('serializes the date-only form value as the RFC 3339 timestamp required by the backend', () => {
  const adapter = apiModule as unknown as ProductPayloadAdapter

  expect(adapter.toBackendProductPayload({
    name: 'Milk',
    dateType: 'use_by',
    expiryDate: '2026-09-01',
    group: 'Dairy',
    location: 'Fridge',
  })).toEqual({
    name: 'Milk',
    date_type: 'use_by',
    expiry_date: '2026-09-01T00:00:00Z',
    product_group: 'Dairy',
    storage_location: 'Fridge',
  })
})

test('maps the backend PascalCase draft fields to the form fields', () => {
  const adapter = apiModule as unknown as DraftAdapter

  expect(adapter.fromBackendDraft({
    id: 'draft-1',
    status: 'pending',
    fields: {
      Name: 'Yogurt',
      DateType: 'best_before',
      ExpiryDate: '2026-09-01T00:00:00Z',
      ProductGroup: 'Dairy',
      StorageLocation: 'Fridge',
    },
  })).toEqual({
    id: 'draft-1',
    status: 'pending',
    fields: {
      name: 'Yogurt',
      dateType: 'best_before',
      expiryDate: '2026-09-01',
      group: 'Dairy',
      location: 'Fridge',
    },
  })
})

test('uses the backend display status instead of the lifecycle status', () => {
  const adapter = apiModule as unknown as ProductDisplayAdapter

  expect(adapter.fromApiProduct({
    id: 'product-1',
    name: 'Milk',
    date_type: 'use_by',
    expiry_date: '2026-09-01T00:00:00Z',
    status: 'active',
    display_status: 'research_required',
  })).toMatchObject({ status: 'research_required' })
})
