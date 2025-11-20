import Taro from '@tarojs/taro'
import { showToast } from './toast'

type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

interface RequestOptions<TBody = any> {
  url: string
  method?: HTTPMethod
  data?: TBody
  params?: Record<string, any>
  headers?: Record<string, string>
  timeout?: number
  requireAuth?: boolean
  rawResponse?: boolean
}

let BASE_URL = ''
let AUTH_CONFIG: { publicPaths: RegExp[] } = { publicPaths: [] }

export const setBaseURL = (url: string) => {
  BASE_URL = url || ''
}

export const setAuthConfig = (cfg: { publicPaths?: RegExp[] }) => {
  AUTH_CONFIG.publicPaths = cfg.publicPaths || []
}

const hasProtocol = (u: string) => /^https?:\/\//i.test(u)

const buildQuery = (params?: Record<string, any>) => {
  if (!params) return ''
  const s = Object.keys(params)
    .filter(k => params[k] !== undefined && params[k] !== null)
    .map(k => `${encodeURIComponent(k)}=${encodeURIComponent(String(params[k]))}`)
    .join('&')
  return s ? `?${s}` : ''
}

const resolveURL = (url: string, params?: Record<string, any>) => {
  const qs = buildQuery(params)
  if (hasProtocol(url)) return `${url}${qs}`
  return `${BASE_URL}${url}${qs}`
}

const getToken = (): string | null => {
  try {
    const v = Taro.getStorageSync('auth-token')
    return v ? String(v) : null
  } catch {
    return null
  }
}

const isPublic = (url: string): boolean => {
  return AUTH_CONFIG.publicPaths.some(re => re.test(url))
}

const getCurrentRoute = (): string => {
  try {
    const inst: any = Taro.getCurrentInstance?.()
    const path: string = inst?.router?.path || ''
    const params: any = inst?.router?.params || {}
    const qs = Object.keys(params)
      .map(k => `${encodeURIComponent(k)}=${encodeURIComponent(String(params[k]))}`)
      .join('&')
    return qs ? `${path}?${qs}` : path
  } catch {
    return ''
  }
}

export const ensureAuth = (redirect?: string): boolean => {
  const token = getToken()
  if (token) return true
  const target = encodeURIComponent(redirect || getCurrentRoute() || '/pages/index/index')
  try { Taro.navigateTo({ url: `/pages/login/index?redirect=${target}` }) } catch {}
  return false
}

export const request = async <T = any>(opts: RequestOptions): Promise<T> => {
  const url = resolveURL(opts.url, opts.params)
  const needAuth = opts.requireAuth !== false && !isPublic(opts.url)
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(opts.headers || {})
  }
  if (needAuth) {
    if (!token) {
      await Taro.showToast({ title: '请先登录', icon: 'none' })
      throw new Error('NO_TOKEN')
    }
    if (!headers['Authorization']) headers['Authorization'] = `Bearer ${token}`
  }

  const res = await Taro.request<any>({
    url,
    method: opts.method || 'GET',
    data: opts.data,
    header: headers,
    timeout: opts.timeout || 15000
  })

  const code = (res.statusCode as number) || 0
  if (code < 200 || code >= 300) {
    if (code === 401 || code === 403) {
      await Taro.showToast({ title: '请先登录', icon: 'none' })
      const redirect = encodeURIComponent(getCurrentRoute() || '/pages/index/index')
      try { Taro.navigateTo({ url: `/pages/login/index?redirect=${redirect}` }) } catch {}
      throw new Error('UNAUTHORIZED')
    }
    showToast({ title: '请求失败', icon: 'error' })
    throw new Error(`HTTP ${code}`)
  }
  const body = res.data
  if (body && typeof body === 'object' && 'code' in body) {
    const c = Number(body.code)
    if (c === 0) {
      return (opts.rawResponse ? (body as any) : (body.data ?? null)) as T
    }
    const msg = String(body.msg || '服务异常')
    showToast({ title: msg, icon: 'error' })
    if (c === 401) {
      const redirect = encodeURIComponent(getCurrentRoute() || '/pages/index/index')
      try { Taro.navigateTo({ url: `/pages/login/index?redirect=${redirect}` }) } catch {}
      throw new Error('UNAUTHORIZED')
    }
    throw new Error(msg)
  }
  return (opts.rawResponse ? (body as any) : (body as T))
}

export const get = async <T = any>(url: string, params?: Record<string, any>, headers?: Record<string, string>) => {
  const isList = (() => {
    const p = url.split('?')[0]
    return /\/list$/.test(p)
  })()
  if (isList) {
    return request<T>({ url, method: 'POST', data: params, headers })
  }
  return request<T>({ url, method: 'GET', params, headers })
}

export const post = async <T = any, B = any>(url: string, data?: B, headers?: Record<string, string>, requireAuth?: boolean) => {
  return request<T>({ url, method: 'POST', data, headers, requireAuth })
}

export const put = async <T = any, B = any>(url: string, data?: B, headers?: Record<string, string>, requireAuth?: boolean) => {
  return request<T>({ url, method: 'PUT', data, headers, requireAuth })
}

export const del = async <T = any>(url: string, params?: Record<string, any>, headers?: Record<string, string>, requireAuth?: boolean) => {
  return request<T>({ url, method: 'DELETE', params, headers, requireAuth })
}
export const requestRaw = async (opts: RequestOptions) => {
  const url = resolveURL(opts.url, opts.params)
  const needAuth = opts.requireAuth !== false && !isPublic(opts.url)
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(opts.headers || {})
  }
  if (needAuth) {
    if (!token) {
      await Taro.showToast({ title: '请先登录', icon: 'none' })
      throw new Error('NO_TOKEN')
    }
    if (!headers['Authorization']) headers['Authorization'] = `Bearer ${token}`
  }
  const res = await Taro.request({
    url,
    method: opts.method || 'GET',
    data: opts.data,
    header: headers,
    timeout: opts.timeout || 15000
  })
  return res
}

export const postRaw = async (url: string, data?: any, headers?: Record<string, string>) => {
  return requestRaw({ url, method: 'POST', data, headers, requireAuth: false })
}

// default initialization
if (!BASE_URL) {
  let envApi = ''
  try {
    envApi = process.env.API_BASE_URL as any
  } catch {}
  setBaseURL(envApi || 'http://10.30.1.53:8081')
}
if (!AUTH_CONFIG.publicPaths || AUTH_CONFIG.publicPaths.length === 0) {
  setAuthConfig({ publicPaths: [/^\/api\/auth\//] })
}
