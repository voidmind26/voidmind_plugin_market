import axios from 'axios'

export const api = axios.create({
  baseURL: '/api',
})

export const getHealth = () => api.get('/health')

export const apiErrorMessage = (error: any, fallback: string) =>
  error?.response?.data?.error || error?.message || fallback

export const listRoutes = () => api.get('/routes')
export const getRoute = (id: number) => api.get(`/routes/${id}`)
export const createRoute = (payload: any) => api.post('/routes', payload)
export const updateRoute = (id: number, payload: any) => api.put(`/routes/${id}`, payload)
export const deleteRoute = (id: number) => api.delete(`/routes/${id}`)

export const listKeys = () => api.get('/keys')
export const getKey = (id: number) => api.get(`/keys/${id}`)
export const createKey = (payload: any) => api.post('/keys', payload)
export const updateKey = (id: number, payload: any) => api.put(`/keys/${id}`, payload)
export const deleteKey = (id: number) => api.delete(`/keys/${id}`)

export const listRewrites = (routeId: number) => api.get(`/routes/${routeId}/rewrites`)
export const createRewrite = (routeId: number, payload: any) => api.post(`/routes/${routeId}/rewrites`, payload)
export const updateRewrite = (routeId: number, rewriteId: number, payload: any) => api.put(`/routes/${routeId}/rewrites/${rewriteId}`, payload)
export const deleteRewrite = (routeId: number, rewriteId: number) => api.delete(`/routes/${routeId}/rewrites/${rewriteId}`)

export const listReferences = () => api.get('/references')
