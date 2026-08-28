import type { Status } from './ui'

export type Product = { id: string; name: string; location: string; group: string; dateType: 'use_by' | 'best_before'; expiryDate: string; status: Extract<Status, 'active' | 'attention' | 'expired' | 'used' | 'discarded' | 'research_required'> }
export type ProductInput = Pick<Product, 'name' | 'location' | 'group' | 'dateType' | 'expiryDate'>
export type DraftFields = Partial<ProductInput>
type Draft = { id: string; status: 'pending' | 'approved' | 'rejected'; fields: DraftFields }

let products: Product[] = [
  { id: 'milk', name: 'Молоко', location: 'Холодильник', group: 'Молочные', dateType: 'use_by', expiryDate: '2026-08-27', status: 'active' },
  { id: 'spinach', name: 'Шпинат', location: 'Холодильник', group: 'Овощи', dateType: 'best_before', expiryDate: '2026-08-25', status: 'attention' },
  { id: 'fish', name: 'Рыба', location: 'Морозилка', group: 'Рыба', dateType: 'use_by', expiryDate: '2026-08-24', status: 'expired' },
]
let session: { id: string; email: string } | null = null
const drafts = new Map<string, Draft>()
const id = () => globalThis.crypto?.randomUUID?.() ?? `fixture-${Date.now()}`

/** Development-only fixture adapter. It is selected solely when VITE_API_URL is absent. */
export const fixtureApi = {
  products: {
    list: async () => [...products],
    get: async (productId: string) => { const product = products.find((item) => item.id === productId); if (!product) throw new Error('Продукт не найден.'); return product },
    create: async (input: ProductInput) => { const product: Product = { ...input, id: id(), status: 'active' }; products = [product, ...products]; return product },
    complete: async (productId: string, status: 'used' | 'discarded') => { const product = products.find((item) => item.id === productId); if (!product) throw new Error('Продукт не найден.'); product.status = status; return product },
  },
  auth: {
    register: async (email: string, _password: string) => ({ id: id(), email }),
    login: async (email: string, _password: string) => { session = { id: id(), email }; return session },
    logout: async () => { session = null },
    session: async () => { if (!session) { throw Object.assign(new Error('Требуется вход.'), { status: 401 }) } return session },
  },
  drafts: {
    recognize: async (_file: File) => { const draft: Draft = { id: id(), status: 'pending', fields: { name: 'Распознанный продукт', dateType: 'best_before', expiryDate: '2026-08-30', location: 'Холодильник', group: 'Другое' } }; drafts.set(draft.id, draft); return draft },
    get: async (draftId: string) => { const draft = drafts.get(draftId); if (!draft) throw new Error('Черновик не найден.'); return draft },
    approve: async (draftId: string, input: ProductInput) => { const draft = drafts.get(draftId); if (!draft || draft.status !== 'pending') throw new Error('Черновик недоступен.'); const product = await fixtureApi.products.create(input); draft.status = 'approved'; return product },
    reject: async (draftId: string) => { const draft = drafts.get(draftId); if (!draft || draft.status !== 'pending') throw new Error('Черновик недоступен.'); draft.status = 'rejected' },
  },
  recipes: { list: async () => products.filter((item) => item.status !== 'expired').slice(0, 2).map((item) => ({ title: `Используйте ${item.name}`, product_ids: [item.id] })) },
}
