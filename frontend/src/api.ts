import { fixtureApi, type DraftFields, type Product, type ProductInput } from './mock-api'

const configuredBaseUrl = import.meta.env.VITE_API_URL?.replace(/\/$/, '')
/** Fixtures are a development aid only; a production build without VITE_API_URL uses same-origin HTTP. */
export const apiMode: 'remote' | 'fixture' = !configuredBaseUrl && import.meta.env.DEV ? 'fixture' : 'remote'

export class ApiError extends Error {
  constructor(public readonly status: number, message: string) { super(message) }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${configuredBaseUrl ?? ''}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { error?: string } | null
    throw new ApiError(response.status, payload?.error ?? 'Не удалось выполнить запрос.')
  }
  return response.status === 204 ? undefined as T : response.json() as Promise<T>
}

type ApiProduct = { id: string; name: string; date_type: Product['dateType']; expiry_date: string; product_group?: string; storage_location?: string; status: 'active' | 'used' | 'discarded' }
type Account = { id: string; email: string }
type ApiDraftFields = {
  Name?: string | null
  DateType?: Product['dateType'] | null
  ExpiryDate?: string | null
  ProductGroup?: string | null
  StorageLocation?: string | null
}
type ApiDraft = { id: string; status: 'pending' | 'approved' | 'rejected'; fields: ApiDraftFields }

const dateOnly = (value: string) => value.slice(0, 10)
const fromApiProduct = (item: ApiProduct): Product => ({ id: item.id, name: item.name, dateType: item.date_type, expiryDate: dateOnly(item.expiry_date), group: item.product_group ?? '', location: item.storage_location ?? '', status: item.status })
export const fromBackendDraft = (draft: ApiDraft): { id: string; status: ApiDraft['status']; fields: DraftFields } => ({
  id: draft.id,
  status: draft.status,
  fields: {
    ...(draft.fields.Name != null ? { name: draft.fields.Name } : {}),
    ...(draft.fields.DateType != null ? { dateType: draft.fields.DateType } : {}),
    ...(draft.fields.ExpiryDate != null ? { expiryDate: dateOnly(draft.fields.ExpiryDate) } : {}),
    ...(draft.fields.ProductGroup != null ? { group: draft.fields.ProductGroup } : {}),
    ...(draft.fields.StorageLocation != null ? { location: draft.fields.StorageLocation } : {}),
  },
})
export const toBackendProductPayload = (input: ProductInput) => ({
  name: input.name,
  date_type: input.dateType,
  expiry_date: `${input.expiryDate}T00:00:00Z`,
  product_group: input.group || undefined,
  storage_location: input.location || undefined,
})

export const api = {
  products: {
    list: async () => apiMode === 'fixture' ? fixtureApi.products.list() : (await request<ApiProduct[]>('/v1/products')).map(fromApiProduct),
    get: async (id: string) => apiMode === 'fixture' ? fixtureApi.products.get(id) : fromApiProduct(await request<ApiProduct>(`/v1/products/${id}`)),
    create: async (input: ProductInput) => apiMode === 'fixture' ? fixtureApi.products.create(input) : fromApiProduct(await request<ApiProduct>('/v1/products', { method: 'POST', body: JSON.stringify(toBackendProductPayload(input)) })),
    complete: async (id: string, action: 'used' | 'discarded') => apiMode === 'fixture' ? fixtureApi.products.complete(id, action) : fromApiProduct(await request<ApiProduct>(`/v1/products/${id}/${action === 'used' ? 'use' : 'discard'}`, { method: 'POST' })),
  },
  auth: {
    register: (email: string, password: string) => apiMode === 'fixture' ? fixtureApi.auth.register(email, password) : request<Account>('/v1/auth/register', { method: 'POST', body: JSON.stringify({ email, password }) }),
    login: (email: string, password: string) => apiMode === 'fixture' ? fixtureApi.auth.login(email, password) : request<Account>('/v1/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
    logout: () => apiMode === 'fixture' ? fixtureApi.auth.logout() : request<void>('/v1/auth/logout', { method: 'POST' }),
    session: () => apiMode === 'fixture' ? fixtureApi.auth.session() : request<Account>('/v1/auth/session'),
  },
  drafts: {
    recognize: (file: File) => apiMode === 'fixture' ? fixtureApi.drafts.recognize(file) : (() => { const form = new FormData(); form.append('image', file); form.append('locale', navigator.language); return request<ApiDraft>('/v1/product-drafts/recognize', { method: 'POST', body: form, headers: {} }).then(fromBackendDraft) })(),
    get: (id: string) => apiMode === 'fixture' ? fixtureApi.drafts.get(id) : request<ApiDraft>(`/v1/product-drafts/${id}`).then(fromBackendDraft),
    approve: (id: string, input: ProductInput) => apiMode === 'fixture' ? fixtureApi.drafts.approve(id, input) : request<ApiProduct>(`/v1/product-drafts/${id}/approve`, { method: 'POST', body: JSON.stringify(toBackendProductPayload(input)) }).then(fromApiProduct),
    reject: (id: string) => apiMode === 'fixture' ? fixtureApi.drafts.reject(id) : request<void>(`/v1/product-drafts/${id}/reject`, { method: 'POST' }),
  },
  recipes: () => apiMode === 'fixture' ? fixtureApi.recipes.list() : request<{ title: string; product_ids: string[] }[]>('/v1/recipes'),
}
