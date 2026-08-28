import type { Status } from './ui'
import { t } from './i18n'

export type Product = { id: string; name: string; location: string; group: string; dateType: 'use_by' | 'best_before'; expiryDate: string; quantity?: number; unit?: string; status: Extract<Status, 'active' | 'attention' | 'expired' | 'used' | 'discarded' | 'research_required'> }
export type ProductInput = Pick<Product, 'name' | 'location' | 'group' | 'dateType' | 'expiryDate' | 'quantity' | 'unit'> & { countryCode?: string }
export type DraftFields = Partial<ProductInput>
type Draft = { id: string; status: 'pending' | 'approved' | 'rejected'; fields: DraftFields }

let products: Product[] = [
  { id: 'milk', name: t.fixtureMilk, location: t.fixtureFridge, group: t.fixtureDairy, dateType: 'use_by', expiryDate: '2026-08-27', status: 'active' },
  { id: 'spinach', name: t.fixtureSpinach, location: t.fixtureFridge, group: t.fixtureVegetables, dateType: 'best_before', expiryDate: '2026-08-25', status: 'attention' },
  { id: 'fish', name: t.fixtureFish, location: t.fixtureFreezer, group: t.fixtureFishGroup, dateType: 'use_by', expiryDate: '2026-08-24', status: 'expired' },
]
let session: { id: string; email: string } | null = null
let profile: { country_code: string; language: 'ru' | 'en'; regulator_group: string } = { country_code: 'DE', language: navigator.language.startsWith('ru') ? 'ru' : 'en', regulator_group: 'eu_1169_2011' }
let notificationSettings = [
  { product_group: 'refrigerated_perishable', alert_threshold_minutes: 4320 },
  { product_group: 'fresh_produce', alert_threshold_minutes: 2880 },
  { product_group: 'frozen', alert_threshold_minutes: 10080 },
  { product_group: 'shelf_stable', alert_threshold_minutes: 20160 },
  { product_group: 'other', alert_threshold_minutes: 4320 },
]
const drafts = new Map<string, Draft>()
const id = () => globalThis.crypto?.randomUUID?.() ?? `fixture-${Date.now()}`

/** Development-only fixture adapter. It is selected solely when VITE_API_URL is absent. */
export const fixtureApi = {
  products: {
    list: async () => [...products],
    get: async (productId: string) => { const product = products.find((item) => item.id === productId); if (!product) throw new Error(t.productNotFound); return product },
    create: async (input: ProductInput) => { const product: Product = { ...input, id: id(), status: 'active' }; products = [product, ...products]; return product },
    complete: async (productId: string, status: 'used' | 'discarded') => { const product = products.find((item) => item.id === productId); if (!product) throw new Error(t.productNotFound); product.status = status; return product },
  },
  auth: {
    register: async (email: string, _password: string) => ({ id: id(), email }),
    login: async (email: string, _password: string) => { session = { id: id(), email }; return session },
    logout: async () => { session = null },
    session: async () => { if (!session) { throw Object.assign(new Error('Требуется вход.'), { status: 401 }) } return session },
  },
  profile: {
    get: async () => ({ ...profile }),
    save: async (input: { country_code: string; language: 'ru' | 'en' }) => { profile = { country_code: input.country_code.toUpperCase(), language: input.language, regulator_group: input.country_code.toUpperCase() === 'DE' ? 'eu_1169_2011' : '' }; return { ...profile } },
  },
  notifications: {
    list: async () => ({ settings: notificationSettings.map((setting) => ({ ...setting })) }),
    save: async (input: { product_group: string; alert_threshold_minutes: number }) => { notificationSettings = notificationSettings.map((setting) => setting.product_group === input.product_group ? { ...setting, alert_threshold_minutes: input.alert_threshold_minutes } : setting); return { settings: notificationSettings.map((setting) => ({ ...setting })) } },
  },
  drafts: {
    recognize: async (_file: File) => { const draft: Draft = { id: id(), status: 'pending', fields: { name: t.recognizedProduct, dateType: 'best_before', expiryDate: '2026-08-30', location: t.fixtureFridge, group: t.fixtureOther } }; drafts.set(draft.id, draft); return draft },
    get: async (draftId: string) => { const draft = drafts.get(draftId); if (!draft) throw new Error(t.draftNotFound); return draft },
    approve: async (draftId: string, input: ProductInput) => { const draft = drafts.get(draftId); if (!draft || draft.status !== 'pending') throw new Error(t.draftUnavailable); const product = await fixtureApi.products.create(input); draft.status = 'approved'; return product },
    reject: async (draftId: string) => { const draft = drafts.get(draftId); if (!draft || draft.status !== 'pending') throw new Error(t.draftUnavailable); draft.status = 'rejected' },
  },
  recipes: { list: async (): Promise<{ kind: 'use_up' | 'combine_group'; product_name?: string; group_name?: string; product_ids: string[] }[]> => products.filter((item) => item.status !== 'expired').slice(0, 2).map((item) => ({ kind: 'use_up', product_name: item.name, product_ids: [item.id] })) },
}
